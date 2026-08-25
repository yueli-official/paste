package governance

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrConflict             = errors.New("paste governance: revision conflict")
	ErrInvalid              = errors.New("paste governance: invalid input")
	ErrCreationSuspended    = errors.New("paste governance: creation suspended")
	ErrDailyLimitReached    = errors.New("paste governance: daily creation limit reached")
	ErrAnonymousCreationOff = errors.New("paste governance: anonymous creation disabled")
)

type ValidationError struct {
	Field   string
	Message string
}

func (err ValidationError) Error() string {
	return ErrInvalid.Error() + ": " + err.Field + " " + err.Message
}

func (err ValidationError) Unwrap() error { return ErrInvalid }

type Store interface {
	ReadGovernanceSettings(context.Context) (Settings, error)
	UpdateGovernanceSettings(context.Context, Settings, int64) (Settings, error)
	ListGovernanceUsers(context.Context, UserQuery, time.Time) (UserPage, error)
	UpdateGovernanceUser(context.Context, UserPolicy, int64) (UserPolicy, error)
}

type Options struct {
	Now func() time.Time
}

type Service struct {
	store Store
	now   func() time.Time
}

func New(store Store, options Options) (*Service, error) {
	if store == nil {
		return nil, errors.New("paste governance: Store is required")
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}, nil
}

func (service *Service) GetSettings(ctx context.Context) (Settings, error) {
	return service.store.ReadGovernanceSettings(ctx)
}

func (service *Service) UpdateSettings(ctx context.Context, input UpdateSettingsInput) (Settings, error) {
	if input.UserDailyLimit < 1 || input.UserDailyLimit > MaxUserDailyLimit {
		return Settings{}, ValidationError{Field: "userDailyLimit", Message: "is out of range"}
	}
	if input.AnonymousDailyLimit < 0 || input.AnonymousDailyLimit > MaxAnonymousDailyLimit {
		return Settings{}, ValidationError{Field: "anonymousDailyLimit", Message: "is out of range"}
	}
	if input.ExpectedRevision < 1 {
		return Settings{}, ErrConflict
	}
	value := Settings{
		UserDailyLimit: input.UserDailyLimit, AnonymousDailyLimit: input.AnonymousDailyLimit,
		Revision: input.ExpectedRevision + 1, UpdatedAt: service.now().UTC(), UpdatedBy: strings.TrimSpace(input.UpdatedBy),
	}
	return service.store.UpdateGovernanceSettings(ctx, value, input.ExpectedRevision)
}

func (service *Service) ListUsers(ctx context.Context, query UserQuery) (UserPage, error) {
	query.Query = strings.TrimSpace(query.Query)
	if query.State != "" && query.State != UserStateActive && query.State != UserStateSuspended {
		return UserPage{}, ValidationError{Field: "state", Message: "is invalid"}
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
	page, err := service.store.ListGovernanceUsers(ctx, query, service.now().UTC())
	if err != nil {
		return UserPage{}, err
	}
	if page.Users == nil {
		page.Users = []User{}
	}
	return page, nil
}

func (service *Service) UpdateUser(ctx context.Context, input UpdateUserInput) (UserPolicy, error) {
	input.UserKey = strings.TrimSpace(input.UserKey)
	input.Reason = strings.TrimSpace(input.Reason)
	input.UpdatedBy = strings.TrimSpace(input.UpdatedBy)
	if input.UserKey == "" || len(input.UserKey) > 64 {
		return UserPolicy{}, ValidationError{Field: "userKey", Message: "is invalid"}
	}
	if input.State != UserStateActive && input.State != UserStateSuspended {
		return UserPolicy{}, ValidationError{Field: "state", Message: "is invalid"}
	}
	if input.ExpectedRevision < 0 {
		return UserPolicy{}, ErrConflict
	}
	if utf8.RuneCountInString(input.Reason) > MaxReasonRunes {
		return UserPolicy{}, ValidationError{Field: "reason", Message: "is too long"}
	}
	limit := input.DailyLimitOverride
	if input.ClearDailyLimit {
		limit = nil
	}
	if limit != nil && (*limit < 1 || *limit > MaxUserDailyLimit) {
		return UserPolicy{}, ValidationError{Field: "dailyLimitOverride", Message: "is out of range"}
	}
	value := UserPolicy{
		UserKey: input.UserKey, State: input.State, DailyLimitOverride: cloneInt(limit), Reason: input.Reason,
		Revision: input.ExpectedRevision + 1, UpdatedAt: service.now().UTC(), UpdatedBy: input.UpdatedBy,
	}
	return service.store.UpdateGovernanceUser(ctx, value, input.ExpectedRevision)
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
