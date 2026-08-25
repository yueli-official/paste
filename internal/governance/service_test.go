package governance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/paste"
)

func TestSettingsAndUserPolicyValidation(t *testing.T) {
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	store := paste.NewMemoryStore()
	service, err := governance.New(store, governance.Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	settings, err := service.UpdateSettings(context.Background(), governance.UpdateSettingsInput{
		UserDailyLimit: 25, AnonymousDailyLimit: 80, ExpectedRevision: 1, UpdatedBy: "usr_ADMIN",
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.Revision != 2 || settings.UserDailyLimit != 25 || settings.AnonymousDailyLimit != 80 {
		t.Fatalf("unexpected settings: %#v", settings)
	}
	if _, err := service.UpdateSettings(context.Background(), governance.UpdateSettingsInput{
		UserDailyLimit: 0, AnonymousDailyLimit: 80, ExpectedRevision: 2,
	}); !errors.Is(err, governance.ErrInvalid) {
		t.Fatalf("expected invalid user limit, got %v", err)
	}
	limit := 3
	policy, err := service.UpdateUser(context.Background(), governance.UpdateUserInput{
		UserKey: "usr_TARGET", State: governance.UserStateSuspended, DailyLimitOverride: &limit,
		Reason: "异常批量创建", ExpectedRevision: 0, UpdatedBy: "usr_ADMIN",
	})
	if err != nil {
		t.Fatal(err)
	}
	if policy.Revision != 1 || policy.DailyLimitOverride == nil || *policy.DailyLimitOverride != 3 {
		t.Fatalf("unexpected user policy: %#v", policy)
	}
	page, err := service.ListUsers(context.Background(), governance.UserQuery{State: governance.UserStateSuspended})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Users[0].UserKey != "usr_TARGET" || page.Users[0].EffectiveDailyLimit != 3 {
		t.Fatalf("unexpected user page: %#v", page)
	}
}
