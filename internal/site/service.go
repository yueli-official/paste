package site

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yueli-official/foundation/go/siteprofile"
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
	value, err := service.Get(ctx)
	if err != nil {
		return Settings{}, err
	}
	if value.Revision != input.ExpectedRevision {
		return Settings{}, ErrConflict
	}
	if _, err := applyBranding(&value, input.BrandImagePatch); err != nil {
		return Settings{}, err
	}
	value.Name, value.Description = name, description
	return service.storeSettings(ctx, value, updatedBy)
}

var BrandingDefaults = siteprofile.BrandImageURLs{LogoURL: "/favicon.svg", FaviconURL: "/favicon.svg"}

func (service *Service) PatchBranding(ctx context.Context, patch siteprofile.BrandImagePatch, updatedBy string) (Settings, error) {
	value, err := service.Get(ctx)
	if err != nil {
		return Settings{}, err
	}
	changed, err := applyBranding(&value, patch)
	if err != nil || !changed {
		return value, err
	}
	return service.storeSettings(ctx, value, updatedBy)
}

func (service *Service) storeSettings(ctx context.Context, value Settings, updatedBy string) (Settings, error) {
	expected := value.Revision
	value.Revision++
	value.UpdatedAt, value.UpdatedBy = service.now().UTC(), updatedBy
	return service.store.UpdateSettings(ctx, value, expected)
}

func applyBranding(value *Settings, patch siteprofile.BrandImagePatch) (bool, error) {
	branding := siteprofile.Branding{BrandImageURLs: value.BrandImageURLs}
	changed := patch.Apply(&branding, BrandingDefaults)
	for _, entry := range []struct{ field, url string }{
		{"logoUrl", branding.LogoURL}, {"faviconUrl", branding.FaviconURL},
	} {
		if entry.url != "" && !siteprofile.ValidPublicBrandImageURL(entry.url) {
			return false, ValidationError{Field: entry.field, Message: "must be a public HTTPS image URL"}
		}
	}
	value.BrandImageURLs = branding.BrandImageURLs
	return changed, nil
}
