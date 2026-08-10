package pasteerr

import (
	"errors"
	"net/http"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/paste/internal/paste"
)

const typeRoot = "https://errors.yuelili.com/problems/"

var (
	RateLimited     = descriptor("common.rate_limited", http.StatusTooManyRequests)
	Validation      = descriptor("validation.failed", http.StatusBadRequest)
	Internal        = descriptor("common.internal", http.StatusInternalServerError)
	Unauthorized    = descriptor("paste.not_authenticated", http.StatusUnauthorized)
	Forbidden       = descriptor("paste.forbidden", http.StatusForbidden)
	NotFound        = descriptor("paste.not_found", http.StatusNotFound)
	Gone            = descriptor("paste.gone", http.StatusGone)
	PasswordNeeded  = descriptor("paste.password_required", http.StatusLocked)
	PasswordInvalid = descriptor("paste.password_invalid", http.StatusForbidden)
	Conflict        = descriptor("paste.conflict", http.StatusConflict)
)

func descriptor(code string, status int) problem.Descriptor {
	return problem.MustDescriptor(problem.MustKind(code, status), typeRoot+code)
}

func Map(err error) error {
	var selected problem.Descriptor
	switch {
	case errors.Is(err, paste.ErrNotFound):
		selected = NotFound
	case errors.Is(err, paste.ErrExpired), errors.Is(err, paste.ErrDeleted):
		selected = Gone
	case errors.Is(err, paste.ErrPasswordNeeded):
		selected = PasswordNeeded
	case errors.Is(err, paste.ErrPasswordInvalid):
		selected = PasswordInvalid
	case errors.Is(err, paste.ErrForbidden):
		selected = Forbidden
	case errors.Is(err, paste.ErrConflict), errors.Is(err, paste.ErrCodeCollision):
		selected = Conflict
	case errors.Is(err, paste.ErrInvalid):
		selected = Validation
	default:
		return err
	}
	mapped, mapErr := problem.WrapError(selected, err, nil)
	if mapErr != nil {
		return err
	}
	return mapped
}
