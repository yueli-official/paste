package pasteauthz

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/yueli-official/foundation/go/authorization"
	authpostgres "github.com/yueli-official/foundation/go/authorization/postgres"
	"github.com/yueli-official/paste/internal/postgres"
)

func TestPersistentInitialClaim(t *testing.T) {
	dsn := os.Getenv("PASTE_AUTHORIZATION_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PASTE_AUTHORIZATION_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := fmt.Sprintf("paste_claim_%d", time.Now().UnixNano())
	if _, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
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
	if err = postgres.ApplySchema(ctx, isolated); err != nil {
		t.Fatal(err)
	}
	catalog, err := authorization.Compile(Definition())
	if err != nil {
		t.Fatal(err)
	}
	options := authpostgres.Options{DB: isolated, InstanceKey: "paste:claim-test", Memory: authorization.MemoryOptions{RootScopeID: RootScopeID, AllowUnclaimed: true}}
	first, err := authpostgres.New(ctx, catalog, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := authpostgres.New(ctx, catalog, options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = first.ClaimInitialAdministrator(ctx, authorization.ClaimInitialAdministratorCommand{Actor: authorization.SubjectRef{Kind: authorization.SubjectAnonymous}}); !authorization.Is(err, authorization.ErrorInvalidInput) {
		t.Fatalf("anonymous: %v", err)
	}
	var wg sync.WaitGroup
	winners := make(chan string, 20)
	failures := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			adapter := first
			if i%2 == 1 {
				adapter = second
			}
			id := fmt.Sprintf("claimant-%d", i)
			result, err := adapter.ClaimInitialAdministrator(ctx, authorization.ClaimInitialAdministratorCommand{Actor: authorization.SubjectRef{Kind: authorization.SubjectUser, ID: id}})
			if err == nil && result.Created {
				winners <- id
			} else if err != nil && !authorization.Is(err, authorization.ErrorConflict) {
				failures <- err
			}
		}(i)
	}
	wg.Wait()
	close(winners)
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	if len(winners) != 1 {
		t.Fatalf("winners=%d", len(winners))
	}
	owner := <-winners
	restarted, err := authpostgres.New(ctx, catalog, options)
	if err != nil {
		t.Fatal(err)
	}
	status, err := restarted.AdministratorClaimStatus(ctx)
	if err != nil || !status.Claimed {
		t.Fatalf("restart status=%+v err=%v", status, err)
	}
	for _, id := range []string{owner, "another-user"} {
		decision, err := restarted.Decide(ctx, authorization.DecisionRequest{Subject: authorization.SubjectRef{Kind: authorization.SubjectUser, ID: id}, ScopeID: RootScopeID, Capability: CapabilityContentManage})
		if err != nil || decision.Allowed != (id == owner) {
			t.Fatalf("access %s: %+v %v", id, decision, err)
		}
	}
	options.InstanceKey = "paste:bootstrap-test"
	options.Memory.ProtectedSubjects = []authorization.SubjectRef{{Kind: authorization.SubjectUser, ID: "bootstrap"}}
	bootstrap, err := authpostgres.New(ctx, catalog, options)
	if err != nil {
		t.Fatal(err)
	}
	status, err = bootstrap.AdministratorClaimStatus(ctx)
	if err != nil || !status.Claimed {
		t.Fatalf("bootstrap status: %+v %v", status, err)
	}
	if _, err = bootstrap.ClaimInitialAdministrator(ctx, authorization.ClaimInitialAdministratorCommand{Actor: authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "other"}}); !authorization.Is(err, authorization.ErrorConflict) {
		t.Fatalf("bootstrap reopened: %v", err)
	}
}
