package pasteerr

import (
	"testing"

	"github.com/yueli-official/foundation/go/problem"
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
