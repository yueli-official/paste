package paste

import (
	"context"
	"errors"
	"testing"

	"github.com/yueli-official/paste/internal/governance"
)

func TestGuestOwnershipAndClaim(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	service, _ := New(store, Options{})
	input := CreateInput{OwnerGuestID: "same-id", Files: []File{{Path: "a.go", Content: "package main"}}, Password: "password123"}
	created, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	guest := Owner{GuestID: "same-id"}
	for _, owner := range []Owner{{}, {GuestID: "other"}, {UserKey: "same-id"}, {UserKey: "same-id", GuestID: "same-id"}} {
		if _, err := service.GetMine(ctx, created.Code, owner); !errors.Is(err, ErrForbidden) {
			t.Fatalf("owner %+v accessed guest: %v", owner, err)
		}
		if err := service.Delete(ctx, created.Code, owner, 1); !errors.Is(err, ErrForbidden) {
			t.Fatalf("owner %+v deleted guest: %v", owner, err)
		}
	}
	if _, err := service.GetMine(ctx, created.Code, guest); err != nil {
		t.Fatal(err)
	}
	page, err := service.ListMinePage(ctx, guest, "", 20, 0)
	if err != nil || page.Total != 1 {
		t.Fatalf("guest listing: %+v %v", page, err)
	}
	private := VisibilityPrivate
	if _, err := service.Update(ctx, created.Code, UpdateInput{OwnerGuestID: guest.GuestID, ExpectedRevision: 1, Visibility: &private}); err == nil {
		t.Fatal("guest made private paste")
	}
	stale, _ := store.GetByCode(ctx, created.Code)
	count, err := service.ClaimGuest(ctx, guest.GuestID, "account")
	if err != nil || count != 1 {
		t.Fatalf("claim: %d %v", count, err)
	}
	if count, err := service.ClaimGuest(ctx, guest.GuestID, "account"); err != nil || count != 0 {
		t.Fatalf("retry: %d %v", count, err)
	}
	if _, err := service.ClaimGuest(ctx, guest.GuestID, "other-account"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("reassignment: %v", err)
	}
	if _, err := service.Create(ctx, input); !errors.Is(err, ErrForbidden) {
		t.Fatalf("stale guest credential created content: %v", err)
	}
	if _, err := service.GetMine(ctx, created.Code, guest); !errors.Is(err, ErrForbidden) {
		t.Fatalf("guest retained ownership: %v", err)
	}
	stale.Revision++
	if _, err := store.Update(ctx, stale, 1); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale update overwrote claim: %v", err)
	}
	owned, err := service.GetMine(ctx, created.Code, Owner{UserKey: "account"})
	if err != nil || owned.Revision != 2 || !owned.PasswordGuard || len(owned.Files) != 1 {
		t.Fatalf("claimed value: %+v %v", owned, err)
	}
	if err := service.Delete(ctx, created.Code, Owner{UserKey: "account"}, owned.Revision); err != nil {
		t.Fatal(err)
	}
}

func TestGuestCreationKeepsAnonymousQuota(t *testing.T) {
	store := NewMemoryStore()
	store.governanceSettings.AnonymousDailyLimit = 1
	service, _ := New(store, Options{})
	input := CreateInput{OwnerGuestID: "a", Files: []File{{Path: "a", Content: "test"}}}
	if value, err := service.Create(context.Background(), input); err != nil || !value.Anonymous() {
		t.Fatalf("guest should be anonymous: %+v %v", value, err)
	}
	input.OwnerGuestID = "b"
	if _, err := service.Create(context.Background(), input); !errors.Is(err, governance.ErrDailyLimitReached) {
		t.Fatalf("changing guest bypassed anonymous quota: %v", err)
	}
}
