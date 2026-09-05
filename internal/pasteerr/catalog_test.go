package pasteerr

import (
	"errors"
	"testing"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/paste"
)

func TestMapValidationPublishesTheInvalidField(t *testing.T) {
	mapped := Map(paste.ValidationError{Field: "files.content", Message: "exceeds the per-file limit"})
	value, ok, err := problem.FromError(mapped, "test-trace")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("validation error was not mapped to a public Problem: %v", mapped)
	}
	if len(value.Violations) != 1 {
		t.Fatalf("violations = %#v, want one field violation", value.Violations)
	}
	violation := value.Violations[0]
	if violation.Pointer != "/files/content" || violation.Code != "validation.max_bytes" {
		t.Fatalf("violation = %#v, want files/content max-bytes", violation)
	}
	if violation.Params["maxBytes"] != paste.MaxFileBytes {
		t.Fatalf("maxBytes = %#v, want %d", violation.Params["maxBytes"], paste.MaxFileBytes)
	}
}

func TestMapCreationGovernanceErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		code   string
		status int
	}{
		{name: "daily limit", err: governance.ErrDailyLimitReached, code: "paste.daily_limit_reached", status: 429},
		{name: "creation suspended", err: governance.ErrCreationSuspended, code: "paste.creation_suspended", status: 403},
		{name: "anonymous creation disabled", err: governance.ErrAnonymousCreationOff, code: "paste.anonymous_creation_disabled", status: 403},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped := Map(test.err)
			value, ok, err := problem.FromError(mapped, "test-trace")
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatalf("governance error was not mapped to a public Problem: %v", mapped)
			}
			if value.Code != test.code || value.Status != test.status {
				t.Fatalf("problem = %s/%d, want %s/%d", value.Code, value.Status, test.code, test.status)
			}
			if !errors.Is(mapped, test.err) {
				t.Fatalf("mapped error no longer wraps %v", test.err)
			}
		})
	}
}

func TestValidationNeverPublishesInternalMessages(t *testing.T) {
	cause := paste.ValidationError{Field: "title", Message: "SQL password=secret /private/path"}
	value, ok, err := problem.FromError(Map(cause), "safe-trace")
	if err != nil || !ok {
		t.Fatal(err)
	}
	if value.Code != "common.validation_failed" || len(value.Violations) != 1 || len(value.Violations[0].Params) != 0 {
		t.Fatalf("unsafe validation: %#v", value)
	}
	if !errors.Is(Map(cause), paste.ErrInvalid) {
		t.Fatal("typed cause lost")
	}
}
