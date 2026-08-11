package pasteerr

import (
	"errors"
	"net/http"
	"strings"

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
	var violations []problem.Violation
	var validation paste.ValidationError
	switch {
	case errors.As(err, &validation):
		selected = Validation
		violations = []problem.Violation{validationViolation(validation)}
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
	default:
		return err
	}
	mapped, mapErr := problem.WrapError(selected, err, nil, violations...)
	if mapErr != nil {
		return err
	}
	return mapped
}

func validationViolation(err paste.ValidationError) problem.Violation {
	pointer := "/" + strings.ReplaceAll(strings.ReplaceAll(err.Field, "~", "~0"), "/", "~1")
	pointer = strings.ReplaceAll(pointer, ".", "/")
	violation := problem.Violation{
		Pointer: pointer,
		Code:    "validation.invalid",
		Params:  problem.Parameters{"detail": err.Message},
	}
	switch {
	case err.Field == "files.content" && err.Message == "exceeds the per-file limit":
		violation.Code = "validation.max_bytes"
		violation.Params = problem.Parameters{"maxBytes": paste.MaxFileBytes}
	case err.Field == "files.content" && err.Message == "exceeds the Paste limit":
		violation.Code = "validation.max_total_bytes"
		violation.Params = problem.Parameters{"maxBytes": paste.MaxContentBytes}
	case err.Field == "title" && err.Message == "is too long":
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": paste.MaxTitleRunes}
	case err.Field == "description" && err.Message == "is too long":
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": paste.MaxDescription}
	case err.Field == "tags" && err.Message == "contains too many values":
		violation.Code = "validation.max_items"
		violation.Params = problem.Parameters{"maxItems": paste.MaxTags}
	case err.Field == "tags" && err.Message == "contains a value that is too long":
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": paste.MaxTagRunes}
	case err.Field == "files.path" && err.Message == "is too long":
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": paste.MaxPathRunes}
	}
	return violation
}
