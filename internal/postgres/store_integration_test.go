package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/yueli-official/paste/internal/paste"
)

func TestStoreLifecycleIntegration(t *testing.T) {
	databaseURL := os.Getenv("PASTE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PASTE_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err := ApplySchema(ctx, database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `TRUNCATE paste_daily_creation_usage, paste_user_policies, paste_files, pastes`); err != nil {
		t.Fatal(err)
	}
	store, err := New(database)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	value := paste.Paste{
		ID: "019c5d62-4c00-7000-8000-000000000001", Code: "Ab3Def7K", OwnerUserKey: "usr_OWNER",
		Title: "Integration", Tags: []string{"go"}, Visibility: paste.VisibilityUnlisted,
		PasswordHash: []byte("hash"), State: paste.StateActive, Revision: 1, CreatedAt: now, UpdatedAt: now,
		Files: []paste.File{{Path: "main.go", Language: "go", Content: "package main", Order: 0}},
	}
	if _, err := store.Insert(ctx, value); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Insert(ctx, value); !errors.Is(err, paste.ErrCodeCollision) {
		t.Fatalf("expected code collision, got %v", err)
	}
	stored, err := store.GetByCode(ctx, value.Code)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Files[0].Content != "package main" || !stored.PasswordGuard {
		t.Fatalf("unexpected stored Paste: %#v", stored)
	}
	stored.Title = "Updated"
	stored.Revision = 2
	stored.UpdatedAt = now.Add(time.Minute)
	if _, err := store.Update(ctx, stored, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(ctx, stored, 1); !errors.Is(err, paste.ErrConflict) {
		t.Fatalf("expected revision conflict, got %v", err)
	}
	stored.Tags = nil
	stored.Revision = 3
	stored.UpdatedAt = now.Add(2 * time.Minute)
	if _, err := store.Update(ctx, stored, 2); err != nil {
		t.Fatalf("nil collection must persist as an empty array: %v", err)
	}
	owned, err := store.ListByOwner(ctx, "usr_OWNER")
	if err != nil {
		t.Fatal(err)
	}
	if len(owned) != 1 || owned[0].Title != "Updated" || owned[0].Tags == nil {
		t.Fatalf("unexpected owner list: %#v", owned)
	}
}
