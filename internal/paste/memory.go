package paste

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yueli-official/paste/internal/governance"
)

type MemoryStore struct {
	mu                 sync.RWMutex
	values             map[string]Paste
	governanceSettings governance.Settings
	governancePolicies map[string]governance.UserPolicy
	dailyUsage         map[string]int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		values: make(map[string]Paste),
		governanceSettings: governance.Settings{
			UserDailyLimit: governance.DefaultUserDailyLimit, AnonymousDailyLimit: governance.DefaultAnonymousDailyLimit,
			Revision: 1, UpdatedAt: time.Unix(0, 0).UTC(),
		},
		governancePolicies: make(map[string]governance.UserPolicy),
		dailyUsage:         make(map[string]int),
	}
}

func (store *MemoryStore) Insert(ctx context.Context, value Paste) (Paste, error) {
	if err := ctx.Err(); err != nil {
		return Paste{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.values[value.Code]; exists {
		return Paste{}, ErrCodeCollision
	}
	if err := store.claimCreationLocked(value); err != nil {
		return Paste{}, err
	}
	store.values[value.Code] = clonePaste(value)
	return clonePaste(value), nil
}

func (store *MemoryStore) claimCreationLocked(value Paste) error {
	actorKind := "anonymous"
	actorKey := "global"
	limit := store.governanceSettings.AnonymousDailyLimit
	if value.OwnerUserKey != "" {
		actorKind = "user"
		actorKey = value.OwnerUserKey
		policy, exists := store.governancePolicies[actorKey]
		if !exists {
			policy = governance.UserPolicy{UserKey: actorKey, State: governance.UserStateActive}
		}
		if policy.State == governance.UserStateSuspended {
			return governance.ErrCreationSuspended
		}
		limit = governance.EffectiveDailyLimit(store.governanceSettings, policy)
	} else if limit == 0 {
		return governance.ErrAnonymousCreationOff
	}
	key := value.CreatedAt.UTC().Format("2006-01-02") + "|" + actorKind + "|" + actorKey
	if store.dailyUsage[key] >= limit {
		return governance.ErrDailyLimitReached
	}
	store.dailyUsage[key]++
	return nil
}

func (store *MemoryStore) GetByCode(ctx context.Context, code string) (Paste, error) {
	if err := ctx.Err(); err != nil {
		return Paste{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, exists := store.values[code]
	if !exists {
		return Paste{}, ErrNotFound
	}
	return clonePaste(value), nil
}

func (store *MemoryStore) ListByOwner(ctx context.Context, userKey string) ([]Paste, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]Paste, 0)
	for _, value := range store.values {
		if value.OwnerUserKey == userKey {
			result = append(result, clonePaste(value))
		}
	}
	sort.Slice(result, func(left, right int) bool { return result[left].CreatedAt.Before(result[right].CreatedAt) })
	return result, nil
}

func (store *MemoryStore) ListForAdministration(ctx context.Context, query AdministrationQuery) (AdministrationPage, error) {
	if err := ctx.Err(); err != nil {
		return AdministrationPage{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	needle := strings.ToLower(query.Query)
	values := make([]Paste, 0, len(store.values))
	for _, value := range store.values {
		if query.OwnerUserKey != "" && value.OwnerUserKey != query.OwnerUserKey {
			continue
		}
		if query.Visibility != "" && value.Visibility != query.Visibility {
			continue
		}
		if query.State != "" && value.State != query.State {
			continue
		}
		if query.Ownership == "anonymous" && value.OwnerUserKey != "" {
			continue
		}
		if query.Ownership == "owned" && value.OwnerUserKey == "" {
			continue
		}
		if needle != "" && !administrationMatches(value, needle) {
			continue
		}
		values = append(values, clonePaste(value))
	}
	sort.Slice(values, func(left, right int) bool { return values[left].CreatedAt.After(values[right].CreatedAt) })
	total := len(values)
	start := min(query.Offset, total)
	end := min(start+query.Limit, total)
	items := make([]AdministrationItem, 0, end-start)
	for _, value := range values[start:end] {
		items = append(items, administrationItem(value))
	}
	return AdministrationPage{Items: items, Total: total, Limit: query.Limit, Offset: query.Offset}, nil
}

func (store *MemoryStore) Update(ctx context.Context, value Paste, expectedRevision int64) (Paste, error) {
	if err := ctx.Err(); err != nil {
		return Paste{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.values[value.Code]
	if !exists {
		return Paste{}, ErrNotFound
	}
	if current.Revision != expectedRevision {
		return Paste{}, ErrConflict
	}
	store.values[value.Code] = clonePaste(value)
	return clonePaste(value), nil
}

func administrationMatches(value Paste, needle string) bool {
	values := []string{value.Code, value.OwnerUserKey, value.Title}
	values = append(values, value.Tags...)
	for _, file := range value.Files {
		values = append(values, file.Language)
	}
	for _, candidate := range values {
		if strings.Contains(strings.ToLower(candidate), needle) {
			return true
		}
	}
	return false
}

func administrationItem(value Paste) AdministrationItem {
	language := "text"
	if len(value.Files) > 0 {
		language = value.Files[0].Language
	}
	return AdministrationItem{
		Code: value.Code, OwnerUserKey: value.OwnerUserKey, Title: value.Title,
		Tags: append([]string{}, value.Tags...), FileCount: len(value.Files), PrimaryLanguage: language,
		Visibility: value.Visibility, PasswordProtected: len(value.PasswordHash) > 0, State: value.State,
		Revision: value.Revision, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ExpiresAt: cloneTime(value.ExpiresAt),
	}
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
