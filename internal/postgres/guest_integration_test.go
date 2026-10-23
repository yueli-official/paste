package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/yueli-official/paste/internal/paste"
)

func TestGuestClaimIntegration(t *testing.T) {
	dsn := os.Getenv("PASTE_GUEST_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PASTE_GUEST_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := fmt.Sprintf("paste_guest_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE")
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	isolated, err := sql.Open("postgres", parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer isolated.Close()
	if err := ApplySchema(ctx, isolated); err != nil {
		t.Fatal(err)
	}
	if _, err := isolated.ExecContext(ctx, schemaV7); err != nil {
		t.Fatalf("repeat schema apply: %v", err)
	}
	store, _ := New(isolated)
	service, _ := paste.New(store, paste.Options{})
	input := paste.CreateInput{OwnerGuestID: "guest", Files: []paste.File{{Path: "a", Content: "preserve"}}}
	created, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if page, err := service.ListMinePage(ctx, paste.Owner{GuestID: "guest"}, "", 10, 0); err != nil || page.Total != 1 {
		t.Fatalf("guest filter: %+v %v", page, err)
	}
	stale, err := store.GetByCode(ctx, created.Code)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	failures := make(chan error, 9)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Create(ctx, input)
			if err != nil && !errors.Is(err, paste.ErrForbidden) {
				failures <- err
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := service.ClaimGuest(ctx, "guest", "user")
		if err != nil {
			failures <- err
		}
	}()
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	page, err := service.ListMinePage(ctx, paste.Owner{GuestID: "guest"}, "", 10, 0)
	if err != nil || page.Total != 0 {
		t.Fatalf("rows escaped concurrent transfer: %+v %v", page, err)
	}
	if _, err := service.Create(ctx, input); !errors.Is(err, paste.ErrForbidden) {
		t.Fatalf("post-claim insert: %v", err)
	}
	if count, err := service.ClaimGuest(ctx, "guest", "user"); err != nil || count != 0 {
		t.Fatalf("repeated claim: %d %v", count, err)
	}
	if _, err := service.ClaimGuest(ctx, "guest", "other"); !errors.Is(err, paste.ErrForbidden) {
		t.Fatalf("claim reassigned: %v", err)
	}
	stale.Revision++
	if _, err := store.Update(ctx, stale, 1); !errors.Is(err, paste.ErrConflict) {
		t.Fatalf("stale guest update: %v", err)
	}
	owned, err := service.GetMine(ctx, created.Code, paste.Owner{UserKey: "user"})
	if err != nil || owned.Files[0].Content != "preserve" {
		t.Fatalf("claimed content: %+v %v", owned, err)
	}
	if err := service.Delete(ctx, created.Code, paste.Owner{UserKey: "user"}, owned.Revision); err != nil {
		t.Fatal(err)
	}
}
