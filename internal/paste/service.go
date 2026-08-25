package paste

import (
	"context"
	"errors"
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yueli-official/foundation/go/identifier"
	"golang.org/x/crypto/bcrypt"
)

type Options struct {
	Now func() time.Time
}

type Service struct {
	store Store
	now   func() time.Time
}

func New(store Store, options Options) (*Service, error) {
	if store == nil {
		return nil, errors.New("paste: Store is required")
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}, nil
}

func (service *Service) Create(ctx context.Context, input CreateInput) (Paste, error) {
	now := service.now().UTC()
	normalized, err := normalizeCreate(input, now)
	if err != nil {
		return Paste{}, err
	}

	id, err := identifier.New()
	if err != nil {
		return Paste{}, err
	}
	var passwordHash []byte
	if normalized.Password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(normalized.Password), bcrypt.DefaultCost)
		if err != nil {
			return Paste{}, err
		}
	}

	base := Paste{
		ID:            id.String(),
		OwnerUserKey:  normalized.OwnerUserKey,
		Title:         normalized.Title,
		Description:   normalized.Description,
		Tags:          normalized.Tags,
		Files:         normalized.Files,
		Visibility:    normalized.Visibility,
		PasswordHash:  passwordHash,
		PasswordGuard: len(passwordHash) > 0,
		State:         StateActive,
		Revision:      1,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     normalized.ExpiresAt,
	}

	var created Paste
	code, err := identifier.Allocate(ctx, identifier.CompactURLV1, func(ctx context.Context, candidate identifier.Key) (identifier.ClaimResult, error) {
		value := clonePaste(base)
		value.Code = candidate.String()
		stored, insertErr := service.store.Insert(ctx, value)
		if errors.Is(insertErr, ErrCodeCollision) {
			return identifier.Collision, nil
		}
		if insertErr != nil {
			return 0, insertErr
		}
		created = stored
		return identifier.Claimed, nil
	})
	if err != nil {
		return Paste{}, err
	}
	created.Code = code.String()
	return publicPaste(created), nil
}

func (service *Service) Open(ctx context.Context, code string, access Access) (Paste, error) {
	parsed, err := identifier.CompactURLV1.Parse(code)
	if err != nil {
		return Paste{}, ErrNotFound
	}
	value, err := service.store.GetByCode(ctx, parsed.String())
	if err != nil {
		return Paste{}, err
	}
	if value.State == StateDeleted {
		return Paste{}, ErrDeleted
	}
	if value.ExpiresAt != nil && !service.now().UTC().Before(*value.ExpiresAt) {
		return Paste{}, ErrExpired
	}
	if value.Visibility == VisibilityPrivate && (access.UserKey == "" || access.UserKey != value.OwnerUserKey) {
		return Paste{}, ErrForbidden
	}
	if len(value.PasswordHash) > 0 {
		if access.Password == "" {
			return Paste{}, ErrPasswordNeeded
		}
		if bcrypt.CompareHashAndPassword(value.PasswordHash, []byte(access.Password)) != nil {
			return Paste{}, ErrPasswordInvalid
		}
	}
	return publicPaste(value), nil
}

func (service *Service) ListMine(ctx context.Context, userKey string) ([]Paste, error) {
	userKey = strings.TrimSpace(userKey)
	if userKey == "" {
		return nil, ErrForbidden
	}
	values, err := service.store.ListByOwner(ctx, userKey)
	if err != nil {
		return nil, err
	}
	result := make([]Paste, 0, len(values))
	for _, value := range values {
		if value.State == StateDeleted {
			continue
		}
		result = append(result, publicPaste(value))
	}
	sort.Slice(result, func(left, right int) bool { return result[left].CreatedAt.After(result[right].CreatedAt) })
	return result, nil
}

func (service *Service) ListForAdministration(ctx context.Context, query AdministrationQuery) (AdministrationPage, error) {
	query.Query = strings.TrimSpace(query.Query)
	query.Ownership = strings.TrimSpace(strings.ToLower(query.Ownership))
	if query.Visibility != "" && query.Visibility != VisibilityUnlisted && query.Visibility != VisibilityPrivate {
		return AdministrationPage{}, ValidationError{Field: "visibility", Message: "is invalid"}
	}
	if query.State != "" && query.State != StateActive && query.State != StateDeleted {
		return AdministrationPage{}, ValidationError{Field: "state", Message: "is invalid"}
	}
	if query.Ownership != "" && query.Ownership != "anonymous" && query.Ownership != "owned" {
		return AdministrationPage{}, ValidationError{Field: "ownership", Message: "is invalid"}
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	page, err := service.store.ListForAdministration(ctx, query)
	if err != nil {
		return AdministrationPage{}, err
	}
	if page.Items == nil {
		page.Items = []AdministrationItem{}
	}
	return page, nil
}

func (service *Service) Govern(ctx context.Context, code string, input GovernanceInput) (Paste, error) {
	value, err := service.managedValue(ctx, code)
	if err != nil {
		return Paste{}, err
	}
	if input.ExpectedRevision < 1 || input.ExpectedRevision != value.Revision {
		return Paste{}, ErrConflict
	}
	if input.Visibility != nil {
		if *input.Visibility != VisibilityUnlisted && *input.Visibility != VisibilityPrivate {
			return Paste{}, ValidationError{Field: "visibility", Message: "is invalid"}
		}
		if *input.Visibility == VisibilityPrivate && value.OwnerUserKey == "" {
			return Paste{}, ValidationError{Field: "visibility", Message: "anonymous Paste cannot be private"}
		}
		value.Visibility = *input.Visibility
	}
	now := service.now().UTC()
	if input.ClearExpiry {
		value.ExpiresAt = nil
	} else if input.ExpiresAt != nil {
		expires := input.ExpiresAt.UTC()
		if !expires.After(now) {
			return Paste{}, ValidationError{Field: "expiresAt", Message: "must be in the future"}
		}
		value.ExpiresAt = &expires
	}
	value.Revision++
	value.UpdatedAt = now
	updated, err := service.store.Update(ctx, value, input.ExpectedRevision)
	if err != nil {
		return Paste{}, err
	}
	return publicPaste(updated), nil
}

// GetMine returns an owned Paste for management without requiring its public
// access password. Ownership is the authorization boundary for this view.
func (service *Service) GetMine(ctx context.Context, code, userKey string) (Paste, error) {
	value, err := service.owned(ctx, code, userKey)
	if err != nil {
		return Paste{}, err
	}
	return publicPaste(value), nil
}

func (service *Service) Update(ctx context.Context, code string, input UpdateInput) (Paste, error) {
	value, err := service.owned(ctx, code, input.OwnerUserKey)
	if err != nil {
		return Paste{}, err
	}
	if input.ExpectedRevision < 1 || input.ExpectedRevision != value.Revision {
		return Paste{}, ErrConflict
	}

	now := service.now().UTC()
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			title = "未命名片段"
		}
		if utf8.RuneCountInString(title) > MaxTitleRunes {
			return Paste{}, ValidationError{Field: "title", Message: "is too long"}
		}
		value.Title = title
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		if utf8.RuneCountInString(description) > MaxDescription {
			return Paste{}, ValidationError{Field: "description", Message: "is too long"}
		}
		value.Description = description
	}
	if input.Tags != nil {
		value.Tags, err = normalizeTags(*input.Tags)
		if err != nil {
			return Paste{}, err
		}
	}
	if input.Files != nil {
		value.Files, err = normalizeFiles(*input.Files)
		if err != nil {
			return Paste{}, err
		}
	}
	if input.Visibility != nil {
		if *input.Visibility != VisibilityUnlisted && *input.Visibility != VisibilityPrivate {
			return Paste{}, ValidationError{Field: "visibility", Message: "is invalid"}
		}
		value.Visibility = *input.Visibility
	}
	if input.Password != nil {
		password := *input.Password
		if password != "" && (utf8.RuneCountInString(password) < 8 || utf8.RuneCountInString(password) > 128) {
			return Paste{}, ValidationError{Field: "password", Message: "must contain 8 to 128 characters"}
		}
		if password == "" {
			value.PasswordHash = nil
		} else {
			value.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return Paste{}, err
			}
		}
	}
	if input.ClearExpiry {
		value.ExpiresAt = nil
	} else if input.ExpiresAt != nil {
		expires := input.ExpiresAt.UTC()
		if !expires.After(now) {
			return Paste{}, ValidationError{Field: "expiresAt", Message: "must be in the future"}
		}
		value.ExpiresAt = &expires
	}

	value.Revision++
	value.UpdatedAt = now
	value.PasswordGuard = len(value.PasswordHash) > 0
	updated, err := service.store.Update(ctx, value, input.ExpectedRevision)
	if err != nil {
		return Paste{}, err
	}
	return publicPaste(updated), nil
}

func (service *Service) Delete(ctx context.Context, code, ownerUserKey string, expectedRevision int64) error {
	value, err := service.owned(ctx, code, ownerUserKey)
	if err != nil {
		return err
	}
	return service.deleteValue(ctx, value, expectedRevision)
}

func (service *Service) DeleteAsAdministrator(ctx context.Context, code string, expectedRevision int64) error {
	value, err := service.managedValue(ctx, code)
	if err != nil {
		return err
	}
	return service.deleteValue(ctx, value, expectedRevision)
}

func (service *Service) deleteValue(ctx context.Context, value Paste, expectedRevision int64) error {
	if expectedRevision < 1 || expectedRevision != value.Revision {
		return ErrConflict
	}
	now := service.now().UTC()
	value.State = StateDeleted
	value.DeletedAt = &now
	value.UpdatedAt = now
	value.Revision++
	value.Files = nil
	value.Description = ""
	value.Tags = []string{}
	value.PasswordHash = nil
	value.PasswordGuard = false
	_, err := service.store.Update(ctx, value, expectedRevision)
	return err
}

func (service *Service) managedValue(ctx context.Context, code string) (Paste, error) {
	parsed, err := identifier.CompactURLV1.Parse(code)
	if err != nil {
		return Paste{}, ErrNotFound
	}
	value, err := service.store.GetByCode(ctx, parsed.String())
	if err != nil {
		return Paste{}, err
	}
	if value.State == StateDeleted {
		return Paste{}, ErrDeleted
	}
	return value, nil
}

func (service *Service) owned(ctx context.Context, code, ownerUserKey string) (Paste, error) {
	ownerUserKey = strings.TrimSpace(ownerUserKey)
	if ownerUserKey == "" {
		return Paste{}, ErrForbidden
	}
	parsed, err := identifier.CompactURLV1.Parse(code)
	if err != nil {
		return Paste{}, ErrNotFound
	}
	value, err := service.store.GetByCode(ctx, parsed.String())
	if err != nil {
		return Paste{}, err
	}
	if value.State == StateDeleted {
		return Paste{}, ErrDeleted
	}
	if value.OwnerUserKey == "" || value.OwnerUserKey != ownerUserKey {
		return Paste{}, ErrForbidden
	}
	return value, nil
}

func normalizeCreate(input CreateInput, now time.Time) (CreateInput, error) {
	input.OwnerUserKey = strings.TrimSpace(input.OwnerUserKey)
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		input.Title = "未命名片段"
	}
	if utf8.RuneCountInString(input.Title) > MaxTitleRunes {
		return CreateInput{}, ValidationError{Field: "title", Message: "is too long"}
	}
	input.Description = strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(input.Description) > MaxDescription {
		return CreateInput{}, ValidationError{Field: "description", Message: "is too long"}
	}
	if input.Visibility == "" {
		input.Visibility = VisibilityUnlisted
	}
	if input.Visibility != VisibilityUnlisted && input.Visibility != VisibilityPrivate {
		return CreateInput{}, ValidationError{Field: "visibility", Message: "is invalid"}
	}
	if input.Visibility == VisibilityPrivate && input.OwnerUserKey == "" {
		return CreateInput{}, ValidationError{Field: "visibility", Message: "private Paste requires an authenticated owner"}
	}
	if input.Password != "" && (utf8.RuneCountInString(input.Password) < 8 || utf8.RuneCountInString(input.Password) > 128) {
		return CreateInput{}, ValidationError{Field: "password", Message: "must contain 8 to 128 characters"}
	}
	if input.ExpiresAt != nil {
		value := input.ExpiresAt.UTC()
		if !value.After(now) {
			return CreateInput{}, ValidationError{Field: "expiresAt", Message: "must be in the future"}
		}
		input.ExpiresAt = &value
	}

	tags, err := normalizeTags(input.Tags)
	if err != nil {
		return CreateInput{}, err
	}
	files, err := normalizeFiles(input.Files)
	if err != nil {
		return CreateInput{}, err
	}
	input.Tags = tags
	input.Files = files
	return input, nil
}

func normalizeTags(values []string) ([]string, error) {
	if len(values) > MaxTags {
		return nil, ValidationError{Field: "tags", Message: "contains too many values"}
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if utf8.RuneCountInString(value) > MaxTagRunes {
			return nil, ValidationError{Field: "tags", Message: "contains a value that is too long"}
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func normalizeFiles(values []File) ([]File, error) {
	if len(values) == 0 || len(values) > MaxFiles {
		return nil, ValidationError{Field: "files", Message: "must contain 1 to 20 files"}
	}
	result := make([]File, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	total := 0
	meaningful := false
	for index, raw := range values {
		filePath := strings.ReplaceAll(strings.TrimSpace(raw.Path), "\\", "/")
		filePath = path.Clean(filePath)
		if filePath == "." || filePath == ".." || strings.HasPrefix(filePath, "../") || strings.HasPrefix(filePath, "/") {
			return nil, ValidationError{Field: "files.path", Message: "must be a relative display path"}
		}
		if utf8.RuneCountInString(filePath) > MaxPathRunes {
			return nil, ValidationError{Field: "files.path", Message: "is too long"}
		}
		key := strings.ToLower(filePath)
		if _, exists := seen[key]; exists {
			return nil, ValidationError{Field: "files.path", Message: "must be unique"}
		}
		seen[key] = struct{}{}
		if len(raw.Content) > MaxFileBytes {
			return nil, ValidationError{Field: "files.content", Message: "exceeds the per-file limit"}
		}
		total += len(raw.Content)
		if total > MaxContentBytes {
			return nil, ValidationError{Field: "files.content", Message: "exceeds the Paste limit"}
		}
		if strings.TrimSpace(raw.Content) != "" {
			meaningful = true
		}
		language := strings.ToLower(strings.TrimSpace(raw.Language))
		if language == "" {
			language = "text"
		}
		result = append(result, File{Path: filePath, Language: language, Content: raw.Content, Order: index})
	}
	if !meaningful {
		return nil, ValidationError{Field: "files.content", Message: "must contain text"}
	}
	return result, nil
}

func publicPaste(value Paste) Paste {
	value = clonePaste(value)
	value.PasswordGuard = len(value.PasswordHash) > 0
	value.PasswordHash = nil
	return value
}

func clonePaste(value Paste) Paste {
	value.Tags = append([]string{}, value.Tags...)
	value.Files = append([]File(nil), value.Files...)
	value.PasswordHash = append([]byte(nil), value.PasswordHash...)
	if value.ExpiresAt != nil {
		expires := *value.ExpiresAt
		value.ExpiresAt = &expires
	}
	if value.DeletedAt != nil {
		deleted := *value.DeletedAt
		value.DeletedAt = &deleted
	}
	return value
}
