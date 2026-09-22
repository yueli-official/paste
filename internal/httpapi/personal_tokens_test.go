package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/pasteauthz"
)

func pastePersonalContext(t *testing.T, user string, capabilities ...authorization.CapabilityKey) context.Context {
	t.Helper()
	scopes := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		scope, err := foundationauth.PersonalScope("paste-yueli-web", string(capability))
		if err != nil {
			t.Fatal(err)
		}
		scopes = append(scopes, scope)
	}
	endpoint := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(response).Encode(map[string]any{"userKey": user, "scopes": scopes}); err != nil {
			t.Error(err)
		}
	}))
	defer endpoint.Close()
	verifier, err := foundationauth.NewPersonalTokenVerifier(endpoint.URL, "paste-yueli-web", nil)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := verifier.Verify(context.Background(), "pat_test")
	if err != nil {
		t.Fatal(err)
	}
	return foundationauth.NewContext(context.Background(), principal)
}

func pastePersonalCore(t *testing.T) (*Core, *authorization.Memory) {
	t.Helper()
	runtime, err := authorization.NewMemory(authorization.MustCompile(pasteauthz.Definition()), authorization.MemoryOptions{
		RootScopeID:       pasteauthz.RootScopeID,
		ProtectedSubjects: []authorization.SubjectRef{{Kind: authorization.SubjectUser, ID: "admin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	core := testController(t)
	core.authorization = pasteauthz.New(runtime)
	return core, runtime
}

func TestPersonalScopeIntersectsCurrentPasteRights(t *testing.T) {
	core, runtime := pastePersonalCore(t)
	for _, test := range []struct {
		name, user         string
		selected, required authorization.CapabilityKey
		allowed            bool
	}{
		{"ordinary create", "user", pasteauthz.CapabilityPasteCreate, pasteauthz.CapabilityPasteCreate, true},
		{"ordinary read", "user", pasteauthz.CapabilityPasteRead, pasteauthz.CapabilityPasteRead, true},
		{"admin moderate", "admin", pasteauthz.CapabilityPasteModerate, pasteauthz.CapabilityPasteModerate, true},
		{"admin unselected", "admin", pasteauthz.CapabilityPasteRead, pasteauthz.CapabilityPasteModerate, false},
		{"scope cannot grant moderation", "user", pasteauthz.CapabilityPasteModerate, pasteauthz.CapabilityPasteModerate, false},
		{"scope cannot grant settings", "user", pasteauthz.CapabilitySettingsManage, pasteauthz.CapabilitySettingsManage, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			allowed, err := core.authorization.CheckCapability(pastePersonalContext(t, test.user, test.selected), test.required)
			if err != nil || allowed != test.allowed {
				t.Fatalf("allowed=%v error=%v want=%v", allowed, err, test.allowed)
			}
		})
	}

	admin := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "admin"}
	grant, err := runtime.Grant(context.Background(), authorization.GrantCommand{
		Actor: admin, Target: authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "temporary"},
		Role: pasteauthz.RoleAdministrator, ScopeID: pasteauthz.RootScopeID, Source: authorization.GrantSourceDirect,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := pastePersonalContext(t, "temporary", pasteauthz.CapabilityPasteModerate)
	if allowed, err := core.authorization.CheckCapability(ctx, pasteauthz.CapabilityPasteModerate); err != nil || !allowed {
		t.Fatalf("granted administrator denied: allowed=%v error=%v", allowed, err)
	}
	if _, err := runtime.Revoke(context.Background(), authorization.RevokeCommand{Actor: admin, GrantID: grant.ID}); err != nil {
		t.Fatal(err)
	}
	if allowed, err := core.authorization.CheckCapability(ctx, pasteauthz.CapabilityPasteModerate); err != nil || allowed {
		t.Fatalf("revoked administrator retained access: allowed=%v error=%v", allowed, err)
	}
}

func TestPersonalDirectoryRequiresTrustedIdentityAndCurrentRights(t *testing.T) {
	core, _ := pastePersonalCore(t)
	controller := core.PersonalPermissions("paste-yueli-web")
	trusted := foundationauth.NewContext(context.Background(), &foundationauth.Principal{
		SubjectKind: foundationauth.SubjectClient, ClientID: "identity-svc", Scopes: []string{foundationauth.PersonalPermissionsScope},
	})
	for _, test := range []struct {
		user string
		keys []string
	}{
		{"user", []string{"paste.paste.create", "paste.paste.read", "paste.paste.update", "paste.paste.delete"}},
		{"admin", []string{"paste.paste.create", "paste.paste.read", "paste.paste.update", "paste.paste.delete", "paste.paste.moderate", "paste.settings.manage"}},
	} {
		result, err := controller.GetPersonalPermissions(trusted, &v1.PersonalPermissionsReq{UserKey: test.user})
		if err != nil {
			t.Fatal(err)
		}
		keys := make([]string, 0, len(result.Items))
		for _, permission := range result.Items {
			keys = append(keys, permission.Key)
		}
		if result.Site != "paste-yueli-web" || result.UserKey != test.user || !reflect.DeepEqual(keys, test.keys) {
			t.Fatalf("directory for %s = %+v", test.user, result)
		}
	}
	for _, principal := range []*foundationauth.Principal{
		{SubjectKind: foundationauth.SubjectUser, Subject: "admin", ClientID: "identity-svc", Scopes: []string{foundationauth.PersonalPermissionsScope}},
		{SubjectKind: foundationauth.SubjectClient, ClientID: "another-service", Scopes: []string{foundationauth.PersonalPermissionsScope}},
		{SubjectKind: foundationauth.SubjectClient, ClientID: "identity-svc"},
	} {
		ctx := foundationauth.NewContext(context.Background(), principal)
		if _, err := controller.GetPersonalPermissions(ctx, &v1.PersonalPermissionsReq{UserKey: "admin"}); err == nil {
			t.Fatalf("untrusted directory request accepted: %+v", principal)
		}
	}
}

func TestPersonalRoutesAreExplicitAndCapabilityBound(t *testing.T) {
	for _, test := range []struct {
		method, path string
		capability   authorization.CapabilityKey
	}{
		{"POST", "/api/v1/pastes", pasteauthz.CapabilityPasteCreate},
		{"GET", "/api/v1/pastes/ABC123", pasteauthz.CapabilityPasteRead},
		{"POST", "/api/v1/pastes/ABC123/access", pasteauthz.CapabilityPasteRead},
		{"GET", "/api/v1/settings", pasteauthz.CapabilityPasteRead},
		{"GET", "/api/v1/me/pastes", pasteauthz.CapabilityPasteRead},
		{"GET", "/api/v1/me/pastes/ABC123", pasteauthz.CapabilityPasteRead},
		{"PATCH", "/api/v1/me/pastes/ABC123", pasteauthz.CapabilityPasteUpdate},
		{"DELETE", "/api/v1/me/pastes/ABC123", pasteauthz.CapabilityPasteDelete},
		{"GET", "/api/v1/admin/pastes", pasteauthz.CapabilityPasteModerate},
		{"PATCH", "/api/v1/admin/pastes/ABC123", pasteauthz.CapabilityPasteModerate},
		{"DELETE", "/api/v1/admin/pastes/ABC123", pasteauthz.CapabilityPasteModerate},
		{"PATCH", "/api/v1/admin/settings", pasteauthz.CapabilitySettingsManage},
	} {
		ctx := pastePersonalContext(t, "admin", test.capability)
		if !allowsPersonalRoute(ctx, test.method, test.path) {
			t.Errorf("selected capability denied: %s %s", test.method, test.path)
		}
		if allowsPersonalRoute(pastePersonalContext(t, "admin", pasteauthz.CapabilityPasteCreate), test.method, test.path) && test.capability != pasteauthz.CapabilityPasteCreate {
			t.Errorf("unselected capability accepted: %s %s", test.method, test.path)
		}
	}

	all := pastePersonalContext(t, "admin",
		pasteauthz.CapabilityPasteCreate,
		pasteauthz.CapabilityPasteRead,
		pasteauthz.CapabilityPasteUpdate,
		pasteauthz.CapabilityPasteDelete,
		pasteauthz.CapabilityPasteModerate,
		pasteauthz.CapabilitySettingsManage,
	)
	for _, path := range []string{
		"/api/v1/authorization/setup/claim",
		"/api/v1/admin/session",
		"/api/v1/admin/users",
		"/api/v1/admin/users/user",
		"/api/v1/admin/governance-settings",
		"/api/v1/internal/personal-token/permissions",
		"/api/v1/me/pastes//extra",
		"/api/v1/admin/pastes/../extra",
	} {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE", "PUT"} {
			if allowsPersonalRoute(all, method, path) {
				t.Errorf("unexpected route allowed: %s %s", method, path)
			}
		}
	}
}
