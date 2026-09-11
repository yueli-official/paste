package httpapi

import (
	"context"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/pasteauthz"
	"testing"
)

func TestInitialClaimProtectsAdministrativeAccess(t *testing.T) {
	controller := testController(t)
	runtime, err := authorization.NewMemory(authorization.MustCompile(pasteauthz.Definition()), authorization.MemoryOptions{RootScopeID: pasteauthz.RootScopeID, AllowUnclaimed: true})
	if err != nil {
		t.Fatal(err)
	}
	controller.authorization = pasteauthz.New(runtime)
	status, err := controller.Public().GetAuthorizationSetup(context.Background(), &v1.GetAuthorizationSetupReq{})
	if err != nil || status.Claimed || !status.CanClaim {
		t.Fatalf("initial status: %#v %v", status, err)
	}
	if _, err := controller.Managed().ClaimAdministrator(context.Background(), &v1.ClaimAdministratorReq{}); problemCode(t, err) != "paste.not_authenticated" {
		t.Fatalf("anonymous: %v", err)
	}
	serviceContext := foundationauth.NewContext(context.Background(), &foundationauth.Principal{Subject: "service", SubjectKind: foundationauth.SubjectClient})
	if _, err := controller.Managed().ClaimAdministrator(serviceContext, &v1.ClaimAdministratorReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("service: %v", err)
	}
	owner := userContext("owner")
	result, err := controller.Managed().ClaimAdministrator(owner, &v1.ClaimAdministratorReq{})
	if err != nil || !result.Claimed {
		t.Fatalf("claim: %#v %v", result, err)
	}
	if _, err := controller.Administrator().GetSession(owner, &v1.GetAdministrationSessionReq{}); err != nil {
		t.Fatal(err)
	}
	if _, err := controller.Managed().ClaimAdministrator(userContext("other"), &v1.ClaimAdministratorReq{}); problemCode(t, err) != "paste.conflict" {
		t.Fatalf("second claimant: %v", err)
	}
	if _, err := controller.Administrator().GetSession(userContext("other"), &v1.GetAdministrationSessionReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("other admin access: %v", err)
	}
	status, err = controller.Public().GetAuthorizationSetup(context.Background(), &v1.GetAuthorizationSetupReq{})
	if err != nil || !status.Claimed || status.CanClaim {
		t.Fatalf("claimed status: %#v %v", status, err)
	}
}
