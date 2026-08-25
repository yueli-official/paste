package site

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrConflict = errors.New("paste site: revision conflict")
	ErrInvalid  = errors.New("paste site: invalid settings")
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
	ReadSettings(context.Context) (Settings, error)
	UpdateSettings(context.Context, Settings, int64) (Settings, error)
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
		return nil, errors.New("paste site: Store is required")
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}, nil
}

func (service *Service) Get(ctx context.Context) (Settings, error) {
	return service.store.ReadSettings(ctx)
}

func (service *Service) Update(ctx context.Context, input UpdateInput) (Settings, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	updatedBy := strings.TrimSpace(input.UpdatedBy)
	if name == "" {
		return Settings{}, ValidationError{Field: "name", Message: "is required"}
	}
	if utf8.RuneCountInString(name) > MaxNameRunes {
		return Settings{}, ValidationError{Field: "name", Message: "is too long"}
	}
	if utf8.RuneCountInString(description) > MaxDescriptionRunes {
		return Settings{}, ValidationError{Field: "description", Message: "is too long"}
	}
	if input.ExpectedRevision < 1 {
		return Settings{}, ErrConflict
	}
	value := Settings{
		Name: name, Description: description, Revision: input.ExpectedRevision + 1,
		UpdatedAt: service.now().UTC(), UpdatedBy: updatedBy,
	}
	return service.store.UpdateSettings(ctx, value, input.ExpectedRevision)
}
