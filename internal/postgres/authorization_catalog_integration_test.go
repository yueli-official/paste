package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/yueli-official/foundation/go/authorization"
	authorizationpostgres "github.com/yueli-official/foundation/go/authorization/postgres"
	"github.com/yueli-official/paste/internal/pasteauthz"
)

func TestAuthorizationCatalogV1UpgradesToV2(t *testing.T) {
	databaseURL := os.Getenv("PASTE_AUTHORIZATION_UPGRADE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PASTE_AUTHORIZATION_UPGRADE_TEST_DATABASE_URL is not set")
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

	// This test requires a dedicated disposable database. Rebuild it at the real
	// pre-v2 state so the upgrade path cannot be hidden by a fresh bootstrap.
	if _, err := database.ExecContext(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, schemaV1+"\n"+schemaV2+"\n"+schemaV3+"\n"+schemaV4+"\n"+schemaV5); err != nil {
		t.Fatal(err)
	}

	legacy := authorization.MustCompile(legacyAuthorizationDefinition())
	if legacy.Version() != 1 || legacy.Digest() != "9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb" {
		t.Fatalf("unexpected legacy catalog: version=%d digest=%s", legacy.Version(), legacy.Digest())
	}
	if _, err := authorizationpostgres.New(ctx, legacy, authorizationpostgres.Options{
		DB: database, InstanceKey: "paste:upgrade-test",
		Memory: authorization.MemoryOptions{
			RootScopeID:       pasteauthz.RootScopeID,
			ProtectedSubjects: []authorization.SubjectRef{{Kind: authorization.SubjectUser, ID: "admin"}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := database.ExecContext(ctx, schemaV6); err != nil {
		t.Fatal(err)
	}
	current := authorization.MustCompile(pasteauthz.Definition())
	runtime, err := authorizationpostgres.New(ctx, current, authorizationpostgres.Options{
		DB: database, InstanceKey: "paste:upgrade-test",
		Memory: authorization.MemoryOptions{RootScopeID: pasteauthz.RootScopeID},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name       string
		subject    string
		capability authorization.CapabilityKey
		allowed    bool
	}{
		{"authenticated create", "user", pasteauthz.CapabilityPasteCreate, true},
		{"administrator moderation", "admin", pasteauthz.CapabilityPasteModerate, true},
		{"administrator settings", "admin", pasteauthz.CapabilitySettingsManage, true},
		{"ordinary moderation denied", "user", pasteauthz.CapabilityPasteModerate, false},
	} {
		decision, err := runtime.Decide(ctx, authorization.DecisionRequest{
			Subject:    authorization.SubjectRef{Kind: authorization.SubjectUser, ID: test.subject},
			Capability: test.capability,
			ScopeID:    pasteauthz.RootScopeID,
		})
		if err != nil || decision.Allowed != test.allowed {
			t.Fatalf("%s: allowed=%v error=%v", test.name, decision.Allowed, err)
		}
	}
}

func legacyAuthorizationDefinition() authorization.Definition {
	return authorization.Definition{
		Consumer: "paste",
		Version:  1,
		Capabilities: []authorization.CapabilityDefinition{
			{
				Key: pasteauthz.CapabilityPublicRead, Version: 1,
				Binding: authorization.BindingAccessLayerEligible, AllowedScopes: []authorization.ScopeType{pasteauthz.ScopeSite},
			},
			{
				Key: pasteauthz.CapabilityContentManage, Version: 1,
				Binding: authorization.BindingProtectedOnly, Risk: authorization.RiskHigh, Audit: authorization.AuditFull,
				AllowedScopes: []authorization.ScopeType{pasteauthz.ScopeSite}, EligibleSubjects: []authorization.SubjectKind{authorization.SubjectUser},
			},
		},
		Scopes: authorization.ScopeSchema{Types: []authorization.ScopeTypeDefinition{{Key: pasteauthz.ScopeSite, Root: true}}},
		AccessLayers: []authorization.AccessLayerDefinition{
			{Key: authorization.AccessLayerVisitor, Capabilities: []authorization.CapabilityKey{pasteauthz.CapabilityPublicRead}},
			{Key: authorization.AccessLayerAuthenticated, Capabilities: []authorization.CapabilityKey{
				authorization.CapabilityApplicationCreate,
				authorization.CapabilityApplicationReadOwn,
				authorization.CapabilityApplicationWithdraw,
				authorization.CapabilityInvitationAccept,
			}},
		},
		Roles: []authorization.RoleDefinition{{
			Key: pasteauthz.RoleAdministrator, DisplayName: "管理员", Protected: true,
			Capabilities: []authorization.CapabilityKey{
				authorization.CapabilityManage,
				authorization.CapabilityAuditRead,
				pasteauthz.CapabilityContentManage,
			},
		}},
	}
}
