package httpapi

import (
	"context"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	v1 "github.com/yueli-official/paste/api/v1"
)

func TestGuestManagementIsolation(t *testing.T) {
	core := testController(t)
	guest := func(id string) context.Context {
		return foundationauth.NewContext(context.Background(), &foundationauth.Principal{Subject: id, SubjectKind: foundationauth.SubjectGuest})
	}
	ctx := guest("guest-a")
	value, err := core.Public().CreatePaste(ctx, &v1.CreatePasteReq{Files: []v1.FileInput{{Path: "a", Content: "secret"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, other := range []context.Context{guest("guest-b"), userContext("guest-a")} {
		if _, err := core.Managed().GetMyPaste(other, &v1.GetMyPasteReq{Code: value.Paste.Code}); problemCode(t, err) != "paste.forbidden" {
			t.Fatalf("other owner read: %v", err)
		}
		if _, err := core.Managed().DeletePaste(other, &v1.DeletePasteReq{Code: value.Paste.Code, ExpectedRevision: 1}); problemCode(t, err) != "paste.forbidden" {
			t.Fatalf("other owner deleted: %v", err)
		}
	}
	if _, err := core.Administrator().ListPastes(ctx, &v1.ListAdministrationPastesReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("guest admin: %v", err)
	}
	for _, other := range []context.Context{ctx, userContext("account"), context.Background()} {
		if _, err := core.Managed().ClaimGuestPastes(other, &v1.ClaimGuestPastesReq{}); problemCode(t, err) != "paste.forbidden" {
			t.Fatalf("unscoped claim: %v", err)
		}
	}
	claim := foundationauth.NewContext(context.Background(), &foundationauth.Principal{Subject: "account", SubjectKind: foundationauth.SubjectUser, Scopes: []string{"guest:claim"}})
	if _, err := core.Managed().ListMyPastes(claim, &v1.ListMyPastesReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("claim token used for management: %v", err)
	}
	if _, err := core.Public().CreatePaste(claim, &v1.CreatePasteReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("claim token used for creation: %v", err)
	}
	if _, err := core.Managed().ClaimGuestPastes(claim, &v1.ClaimGuestPastesReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("claim missing guest subject accepted: %v", err)
	}
	if _, err := core.Managed().DeletePaste(ctx, &v1.DeletePasteReq{Code: value.Paste.Code, ExpectedRevision: 1}); err != nil {
		t.Fatal(err)
	}
}
