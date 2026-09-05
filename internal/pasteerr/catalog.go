package pasteerr

import (
	"errors"
	"net/http"
	"strings"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/paste"
	"github.com/yueli-official/paste/internal/site"
)

var (
	RateLimited          = descriptor("common.rate_limited", http.StatusTooManyRequests)
	Validation           = descriptor("common.validation_failed", http.StatusBadRequest)
	DailyLimitReached    = descriptors[CodeDailyLimitReached]
	Internal             = descriptor("common.internal", http.StatusInternalServerError)
	Unauthorized         = descriptors[CodeUnauthorized]
	Forbidden            = descriptors[CodeForbidden]
	CreationSuspended    = descriptors[CodeCreationSuspended]
	AnonymousCreationOff = descriptors[CodeAnonymousCreationOff]
	NotFound             = descriptors[CodeNotFound]
	Gone                 = descriptors[CodeGone]
	PasswordNeeded       = descriptors[CodePasswordNeeded]
	PasswordInvalid      = descriptors[CodePasswordInvalid]
	Conflict             = descriptors[CodeConflict]
)

func Map(err error) error {
	if err == nil {
		return nil
	}
	var existing *problem.Error
	if errors.As(err, &existing) {
		return err
	}
	var selected problem.Descriptor
	var violations []problem.Violation
	var validation paste.ValidationError
	var siteValidation site.ValidationError
	var governanceValidation governance.ValidationError
	switch {
	case errors.As(err, &validation):
		selected = Validation
		violations = []problem.Violation{validationViolation(validation)}
	case errors.As(err, &siteValidation):
		selected = Validation
		violations = []problem.Violation{siteValidationViolation(siteValidation)}
	case errors.As(err, &governanceValidation):
		selected = Validation
		violations = []problem.Violation{governanceValidationViolation(governanceValidation)}
	case errors.Is(err, governance.ErrDailyLimitReached):
		selected = DailyLimitReached
	case errors.Is(err, governance.ErrCreationSuspended):
		selected = CreationSuspended
	case errors.Is(err, governance.ErrAnonymousCreationOff):
		selected = AnonymousCreationOff
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
	case errors.Is(err, paste.ErrConflict), errors.Is(err, paste.ErrCodeCollision), errors.Is(err, site.ErrConflict), errors.Is(err, governance.ErrConflict):
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

func governanceValidationViolation(err governance.ValidationError) problem.Violation {
	violation := problem.Violation{
		Pointer: "/" + err.Field,
		Code:    "validation.invalid",
	}
	if err.Message == "is out of range" {
		switch err.Field {
		case "userDailyLimit", "dailyLimitOverride":
			violation.Code = "validation.range"
			violation.Params = problem.Parameters{"min": 1, "max": governance.MaxUserDailyLimit}
		case "anonymousDailyLimit":
			violation.Code = "validation.range"
			violation.Params = problem.Parameters{"min": 0, "max": governance.MaxAnonymousDailyLimit}
		}
	}
	if err.Field == "reason" && err.Message == "is too long" {
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": governance.MaxReasonRunes}
	}
	return violation
}

func siteValidationViolation(err site.ValidationError) problem.Violation {
	violation := problem.Violation{
		Pointer: "/" + err.Field,
		Code:    "validation.invalid",
	}
	if err.Message == "is too long" && err.Field == "name" {
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": site.MaxNameRunes}
	}
	if err.Message == "is too long" && err.Field == "description" {
		violation.Code = "validation.max_length"
		violation.Params = problem.Parameters{"maxLength": site.MaxDescriptionRunes}
	}
	return violation
}

func validationViolation(err paste.ValidationError) problem.Violation {
	pointer := "/" + strings.ReplaceAll(strings.ReplaceAll(err.Field, "~", "~0"), "/", "~1")
	pointer = strings.ReplaceAll(pointer, ".", "/")
	violation := problem.Violation{
		Pointer: pointer,
		Code:    "validation.invalid",
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
