package httpapi

import (
	"context"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/problem"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/paste"
)

func testController(t *testing.T) *Core {
	t.Helper()
	service, err := paste.New(paste.NewMemoryStore(), paste.Options{})
	if err != nil {
		t.Fatal(err)
	}
	controller, err := New(service, "https://paste.example")
	if err != nil {
		t.Fatal(err)
	}
	return controller
}

func userContext(userKey string) context.Context {
	return foundationauth.NewContext(context.Background(), &foundationauth.Principal{
		Subject: userKey, SubjectKind: foundationauth.SubjectUser,
	})
}

func TestAnonymousCreateAndProtectedAccess(t *testing.T) {
	controller := testController(t)
	created, err := controller.Public().CreatePaste(context.Background(), &v1.CreatePasteReq{
		Title:    "Protected",
		Files:    []v1.FileInput{{Path: "main.go", Language: "go", Content: "package main"}},
		Password: "safe password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Paste.ShareURL != "https://paste.example/p/"+created.Paste.Code || !created.Paste.PasswordProtected {
		t.Fatalf("unexpected create response: %#v", created)
	}
	if created.Paste.Tags == nil {
		t.Fatal("API collections must serialize as arrays, not null")
	}
	if _, err := controller.Public().GetPaste(context.Background(), &v1.GetPasteReq{Code: created.Paste.Code}); problemCode(t, err) != "paste.password_required" {
		t.Fatalf("expected password challenge, got %v", err)
	}
	opened, err := controller.Public().UnlockPaste(context.Background(), &v1.UnlockPasteReq{
		Code: created.Paste.Code, Password: "safe password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opened.Paste.Files[0].Content != "package main" {
		t.Fatalf("unexpected opened Paste: %#v", opened)
	}
}

func TestGuestPrincipalCreatesAnAnonymousPaste(t *testing.T) {
	controller := testController(t)
	guestContext := foundationauth.NewContext(context.Background(), &foundationauth.Principal{
		Subject: "gst_SESSION", SubjectKind: foundationauth.SubjectGuest,
	})
	created, err := controller.Public().CreatePaste(guestContext, &v1.CreatePasteReq{
		Files: []v1.FileInput{{Path: "note.txt", Language: "text", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := controller.Managed().ListMyPastes(guestContext, &v1.ListMyPastesReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("guest Paste must remain anonymous, got managed access error %v", err)
	}
	opened, err := controller.Public().GetPaste(guestContext, &v1.GetPasteReq{Code: created.Paste.Code})
	if err != nil || opened.Paste.Files[0].Content != "hello" {
		t.Fatalf("guest-created Paste was not public: value=%#v error=%v", opened, err)
	}
}

func TestManagedLifecycleUsesAuthenticatedOwner(t *testing.T) {
	controller := testController(t)
	ctx := userContext("usr_OWNER")
	created, err := controller.Public().CreatePaste(ctx, &v1.CreatePasteReq{
		Visibility: "private",
		Files:      []v1.FileInput{{Path: "main.go", Language: "go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := controller.Managed().ListMyPastes(ctx, &v1.ListMyPastesReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Pastes) != 1 || listed.Pastes[0].Code != created.Paste.Code {
		t.Fatalf("unexpected owner list: %#v", listed)
	}
	managed, err := controller.Managed().GetMyPaste(ctx, &v1.GetMyPasteReq{Code: created.Paste.Code})
	if err != nil || managed.Paste.Code != created.Paste.Code {
		t.Fatalf("unexpected managed read: value=%#v error=%v", managed, err)
	}
	title := "Updated"
	updated, err := controller.Managed().UpdatePaste(ctx, &v1.UpdatePasteReq{
		Code: created.Paste.Code, ExpectedRevision: created.Paste.Revision, Title: &title,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Paste.Title != title || updated.Paste.Revision != 2 {
		t.Fatalf("unexpected update: %#v", updated)
	}
	if _, err := controller.Managed().DeletePaste(ctx, &v1.DeletePasteReq{
		Code: created.Paste.Code, ExpectedRevision: updated.Paste.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := controller.Public().GetPaste(ctx, &v1.GetPasteReq{Code: created.Paste.Code}); problemCode(t, err) != "paste.gone" {
		t.Fatalf("expected terminal Paste, got %v", err)
	}
}

func TestManagedRoutesRequireUserPrincipal(t *testing.T) {
	controller := testController(t)
	if _, err := controller.Managed().ListMyPastes(context.Background(), &v1.ListMyPastesReq{}); problemCode(t, err) != "paste.not_authenticated" {
		t.Fatalf("expected authentication failure, got %v", err)
	}
	clientContext := foundationauth.NewContext(context.Background(), &foundationauth.Principal{
		SubjectKind: foundationauth.SubjectClient, ClientID: "site-client",
	})
	if _, err := controller.Public().CreatePaste(clientContext, &v1.CreatePasteReq{
		Files: []v1.FileInput{{Path: "main.go", Content: "package main"}},
	}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("expected client principal to be forbidden, got %v", err)
	}
}

func problemCode(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		return ""
	}
	value, ok, resolveErr := problem.FromError(err, "test-trace")
	if resolveErr != nil {
		t.Fatal(resolveErr)
	}
	if !ok {
		t.Fatalf("error is not a public Problem: %v", err)
	}
	return value.Code
}
