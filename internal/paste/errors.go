package paste

import "errors"

var (
	ErrNotFound        = errors.New("paste: not found")
	ErrCodeCollision   = errors.New("paste: public code collision")
	ErrInvalid         = errors.New("paste: invalid input")
	ErrForbidden       = errors.New("paste: access forbidden")
	ErrPasswordNeeded  = errors.New("paste: password required")
	ErrPasswordInvalid = errors.New("paste: password invalid")
	ErrExpired         = errors.New("paste: expired")
	ErrDeleted         = errors.New("paste: deleted")
	ErrConflict        = errors.New("paste: revision conflict")
)

type ValidationError struct {
	Field   string
	Message string
}

func (err ValidationError) Error() string {
	if err.Field == "" {
		return ErrInvalid.Error() + ": " + err.Message
	}
	return ErrInvalid.Error() + ": " + err.Field + " " + err.Message
}

func (err ValidationError) Unwrap() error { return ErrInvalid }
