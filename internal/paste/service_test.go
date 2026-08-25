package paste

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yueli-official/paste/internal/governance"
)

var fixedNow = time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)

func newTestService(t *testing.T) (*Service, *MemoryStore) {
	t.Helper()
	store := NewMemoryStore()
	service, err := New(store, Options{Now: func() time.Time { return fixedNow }})
	if err != nil {
		t.Fatal(err)
	}
	return service, store
}

func TestCreateAnonymousMultiFilePaste(t *testing.T) {
	service, store := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		Title: "HTTP retry example",
		Tags:  []string{"Go", "go", "network"},
		Files: []File{
			{Path: "main.go", Language: "Go", Content: "package main\n"},
			{Path: "README.md", Language: "Markdown", Content: "# Retry\n"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Code) != 8 || !created.Anonymous() {
		t.Fatalf("unexpected created Paste: %#v", created)
	}
	if created.Visibility != VisibilityUnlisted || created.Revision != 1 {
		t.Fatalf("unexpected lifecycle defaults: %#v", created)
	}
	if len(created.Tags) != 2 || created.Files[0].Language != "go" || created.Files[1].Order != 1 {
		t.Fatalf("normalization failed: %#v", created)
	}
	stored, err := store.GetByCode(context.Background(), created.Code)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ID == "" || stored.CreatedAt != fixedNow {
		t.Fatalf("storage facts missing: %#v", stored)
	}
}

func TestCreateKeepsEmptyCollectionsJSONAndDatabaseSafe(t *testing.T) {
	service, _ := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		Files: []File{{Path: "note.txt", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Tags == nil {
		t.Fatal("empty tags must be represented as an empty collection, not null")
	}
}

func TestCreateEnforcesSuccessfulDailyLimitsAndUserSuspension(t *testing.T) {
	service, store := newTestService(t)
	governanceService, err := governance.New(store, governance.Options{Now: func() time.Time { return fixedNow }})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := governanceService.UpdateSettings(context.Background(), governance.UpdateSettingsInput{
		UserDailyLimit: 1, AnonymousDailyLimit: 1, ExpectedRevision: 1, UpdatedBy: "usr_ADMIN",
	}); err != nil {
		t.Fatal(err)
	}
	input := CreateInput{OwnerUserKey: "usr_LIMITED", Files: []File{{Path: "note.txt", Content: "hello"}}}
	if _, err := service.Create(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), input); !errors.Is(err, governance.ErrDailyLimitReached) {
		t.Fatalf("expected authenticated daily limit, got %v", err)
	}
	input.OwnerUserKey = "usr_OTHER"
	if _, err := service.Create(context.Background(), input); err != nil {
		t.Fatalf("another user should have an independent allowance: %v", err)
	}
	input.OwnerUserKey = ""
	if _, err := service.Create(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), input); !errors.Is(err, governance.ErrDailyLimitReached) {
		t.Fatalf("expected anonymous site-wide limit, got %v", err)
	}
	if _, err := governanceService.UpdateUser(context.Background(), governance.UpdateUserInput{
		UserKey: "usr_SUSPENDED", State: governance.UserStateSuspended, ExpectedRevision: 0, UpdatedBy: "usr_ADMIN",
	}); err != nil {
		t.Fatal(err)
	}
	input.OwnerUserKey = "usr_SUSPENDED"
	if _, err := service.Create(context.Background(), input); !errors.Is(err, governance.ErrCreationSuspended) {
		t.Fatalf("expected suspended creation, got %v", err)
	}
}

func TestPrivatePasteRequiresAuthenticatedOwner(t *testing.T) {
	service, _ := newTestService(t)
	_, err := service.Create(context.Background(), CreateInput{
		Visibility: VisibilityPrivate,
		Files:      []File{{Path: "main.go", Content: "package main"}},
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestProtectedPasteNeverReturnsPasswordHash(t *testing.T) {
	service, _ := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		Password: "correct horse battery staple",
		Files:    []File{{Path: "secret.txt", Content: "shared secret example"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !created.PasswordGuard || len(created.PasswordHash) != 0 {
		t.Fatalf("public result leaked password state: %#v", created)
	}
	if _, err := service.Open(context.Background(), created.Code, Access{}); !errors.Is(err, ErrPasswordNeeded) {
		t.Fatalf("expected password challenge, got %v", err)
	}
	if _, err := service.Open(context.Background(), created.Code, Access{Password: "wrong password"}); !errors.Is(err, ErrPasswordInvalid) {
		t.Fatalf("expected invalid password, got %v", err)
	}
	opened, err := service.Open(context.Background(), created.Code, Access{Password: "correct horse battery staple"})
	if err != nil {
		t.Fatal(err)
	}
	if opened.Files[0].Content != "shared secret example" || len(opened.PasswordHash) != 0 {
		t.Fatalf("unexpected opened Paste: %#v", opened)
	}
}

func TestOpenEnforcesPrivateOwnerAndExpiry(t *testing.T) {
	service, _ := newTestService(t)
	expires := fixedNow.Add(time.Hour)
	created, err := service.Create(context.Background(), CreateInput{
		OwnerUserKey: "usr_TESTOWNER",
		Visibility:   VisibilityPrivate,
		ExpiresAt:    &expires,
		Files:        []File{{Path: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(context.Background(), created.Code, Access{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if _, err := service.Open(context.Background(), created.Code, Access{UserKey: "usr_TESTOWNER"}); err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return expires }
	if _, err := service.Open(context.Background(), created.Code, Access{UserKey: "usr_TESTOWNER"}); !errors.Is(err, ErrExpired) {
		t.Fatalf("expected expired, got %v", err)
	}
}

func TestCreateRejectsDuplicateOrEmptyFiles(t *testing.T) {
	service, _ := newTestService(t)
	tests := []CreateInput{
		{Files: []File{{Path: "main.go", Content: "package main"}, {Path: "MAIN.GO", Content: "duplicate"}}},
		{Files: []File{{Path: "empty.txt", Content: " \n"}}},
		{Files: []File{{Path: "../secret.txt", Content: "nope"}}},
	}
	for _, input := range tests {
		if _, err := service.Create(context.Background(), input); !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected invalid input for %#v, got %v", input, err)
		}
	}
}

func TestListMineRequiresAndFiltersOwner(t *testing.T) {
	service, _ := newTestService(t)
	var owned []Paste
	for _, owner := range []string{"usr_A", "usr_B", "usr_A"} {
		created, err := service.Create(context.Background(), CreateInput{
			OwnerUserKey: owner,
			Files:        []File{{Path: "main.go", Content: "package main"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if owner == "usr_A" {
			owned = append(owned, created)
		}
	}
	if _, err := service.ListMine(context.Background(), ""); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	values, err := service.ListMine(context.Background(), "usr_A")
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 {
		t.Fatalf("expected two Pastes, got %d", len(values))
	}
	if err := service.Delete(context.Background(), owned[0].Code, "usr_A", owned[0].Revision); err != nil {
		t.Fatal(err)
	}
	values, err = service.ListMine(context.Background(), "usr_A")
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || values[0].Code == owned[0].Code {
		t.Fatalf("deleted Paste must not reappear in management: %#v", values)
	}
}

func TestGetMineBypassesThePublicPasswordChallengeForTheOwner(t *testing.T) {
	service, _ := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		OwnerUserKey: "usr_OWNER",
		Password:     "safe password",
		Files:        []File{{Path: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(context.Background(), created.Code, Access{UserKey: "usr_OWNER"}); !errors.Is(err, ErrPasswordNeeded) {
		t.Fatalf("public open should still require a password, got %v", err)
	}
	managed, err := service.GetMine(context.Background(), created.Code, "usr_OWNER")
	if err != nil {
		t.Fatal(err)
	}
	if managed.Files[0].Content != "package main" || len(managed.PasswordHash) != 0 || !managed.PasswordGuard {
		t.Fatalf("unexpected managed Paste: %#v", managed)
	}
	if _, err := service.GetMine(context.Background(), created.Code, "usr_OTHER"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected owner boundary, got %v", err)
	}
}

func TestOwnerCanUpdateWithOptimisticRevision(t *testing.T) {
	service, _ := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		OwnerUserKey: "usr_OWNER",
		Title:        "Before",
		Files:        []File{{Path: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	title := "After"
	visibility := VisibilityPrivate
	password := "a better password"
	updated, err := service.Update(context.Background(), created.Code, UpdateInput{
		OwnerUserKey:     "usr_OWNER",
		ExpectedRevision: created.Revision,
		Title:            &title,
		Visibility:       &visibility,
		Password:         &password,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != title || updated.Visibility != VisibilityPrivate || !updated.PasswordGuard || updated.Revision != 2 {
		t.Fatalf("unexpected update: %#v", updated)
	}
	if _, err := service.Update(context.Background(), created.Code, UpdateInput{
		OwnerUserKey:     "usr_OWNER",
		ExpectedRevision: created.Revision,
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected revision conflict, got %v", err)
	}
}

func TestAnonymousOrDifferentOwnerCannotUpdate(t *testing.T) {
	service, _ := newTestService(t)
	anonymous, err := service.Create(context.Background(), CreateInput{
		Files: []File{{Path: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(context.Background(), anonymous.Code, UpdateInput{
		OwnerUserKey: "usr_OWNER", ExpectedRevision: 1,
	}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected anonymous Paste to remain immutable, got %v", err)
	}

	owned, err := service.Create(context.Background(), CreateInput{
		OwnerUserKey: "usr_OWNER",
		Files:        []File{{Path: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(context.Background(), owned.Code, UpdateInput{
		OwnerUserKey: "usr_OTHER", ExpectedRevision: 1,
	}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected different owner to be forbidden, got %v", err)
	}
}

func TestDeleteKeepsTerminalLocatorAndClearsContent(t *testing.T) {
	service, store := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		OwnerUserKey: "usr_OWNER",
		Password:     "safe password",
		Files:        []File{{Path: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(context.Background(), created.Code, "usr_OWNER", created.Revision); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(context.Background(), created.Code, Access{Password: "safe password"}); !errors.Is(err, ErrDeleted) {
		t.Fatalf("expected deleted terminal state, got %v", err)
	}
	stored, err := store.GetByCode(context.Background(), created.Code)
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != StateDeleted || stored.Revision != 2 || len(stored.Files) != 0 || len(stored.PasswordHash) != 0 {
		t.Fatalf("deleted Paste retained private content: %#v", stored)
	}
}
