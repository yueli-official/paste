package paste

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/yueli-official/paste/internal/governance"
)

func (store *MemoryStore) ReadGovernanceSettings(ctx context.Context) (governance.Settings, error) {
	if err := ctx.Err(); err != nil {
		return governance.Settings{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.governanceSettings, nil
}

func (store *MemoryStore) UpdateGovernanceSettings(ctx context.Context, value governance.Settings, expectedRevision int64) (governance.Settings, error) {
	if err := ctx.Err(); err != nil {
		return governance.Settings{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.governanceSettings.Revision != expectedRevision {
		return governance.Settings{}, governance.ErrConflict
	}
	store.governanceSettings = value
	return store.governanceSettings, nil
}

func (store *MemoryStore) ListGovernanceUsers(ctx context.Context, query governance.UserQuery, now time.Time) (governance.UserPage, error) {
	if err := ctx.Err(); err != nil {
		return governance.UserPage{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	keys := make(map[string]struct{})
	for _, value := range store.values {
		if value.OwnerUserKey != "" {
			keys[value.OwnerUserKey] = struct{}{}
		}
	}
	for key := range store.governancePolicies {
		keys[key] = struct{}{}
	}
	needle := strings.ToLower(query.Query)
	users := make([]governance.User, 0, len(keys))
	for key := range keys {
		policy, exists := store.governancePolicies[key]
		if !exists {
			policy = governance.UserPolicy{UserKey: key, State: governance.UserStateActive}
		}
		if query.State != "" && policy.State != query.State {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(key), needle) {
			continue
		}
		user := governance.User{
			UserKey: key, State: policy.State, DailyLimitOverride: cloneGovernanceInt(policy.DailyLimitOverride),
			EffectiveDailyLimit: governance.EffectiveDailyLimit(store.governanceSettings, policy),
			UsedToday:           store.dailyUsage[now.UTC().Format("2006-01-02")+"|user|"+key],
			Reason:              policy.Reason, Revision: policy.Revision, UpdatedBy: policy.UpdatedBy,
		}
		if !policy.UpdatedAt.IsZero() {
			updatedAt := policy.UpdatedAt
			user.UpdatedAt = &updatedAt
		}
		for _, value := range store.values {
			if value.OwnerUserKey != key {
				continue
			}
			user.TotalPastes++
			if value.State == StateActive {
				user.ActivePastes++
			}
			if user.LastCreatedAt == nil || value.CreatedAt.After(*user.LastCreatedAt) {
				createdAt := value.CreatedAt
				user.LastCreatedAt = &createdAt
			}
		}
		users = append(users, user)
	}
	sort.Slice(users, func(left, right int) bool {
		if users[left].UsedToday != users[right].UsedToday {
			return users[left].UsedToday > users[right].UsedToday
		}
		leftTime, rightTime := users[left].LastCreatedAt, users[right].LastCreatedAt
		if leftTime != nil && rightTime != nil && !leftTime.Equal(*rightTime) {
			return leftTime.After(*rightTime)
		}
		if leftTime != nil && rightTime == nil {
			return true
		}
		if leftTime == nil && rightTime != nil {
			return false
		}
		return users[left].UserKey < users[right].UserKey
	})
	total := len(users)
	start := min(query.Offset, total)
	end := min(start+query.Limit, total)
	return governance.UserPage{Users: append([]governance.User(nil), users[start:end]...), Total: total, Limit: query.Limit, Offset: query.Offset}, nil
}

func (store *MemoryStore) UpdateGovernanceUser(ctx context.Context, value governance.UserPolicy, expectedRevision int64) (governance.UserPolicy, error) {
	if err := ctx.Err(); err != nil {
		return governance.UserPolicy{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.governancePolicies[value.UserKey]
	if expectedRevision == 0 {
		if exists {
			return governance.UserPolicy{}, governance.ErrConflict
		}
	} else if !exists || current.Revision != expectedRevision {
		return governance.UserPolicy{}, governance.ErrConflict
	}
	value.DailyLimitOverride = cloneGovernanceInt(value.DailyLimitOverride)
	store.governancePolicies[value.UserKey] = value
	return value, nil
}

func cloneGovernanceInt(value *int) *int {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
