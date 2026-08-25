package httpapi

import (
	"context"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/problem"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/paste"
	"github.com/yueli-official/paste/internal/site"
)

func testController(t *testing.T) *Core {
	t.Helper()
	store := paste.NewMemoryStore()
	service, err := paste.New(store, paste.Options{})
	if err != nil {
		t.Fatal(err)
	}
	settings, err := site.New(site.NewMemoryStore(), site.Options{})
	if err != nil {
		t.Fatal(err)
	}
	governanceService, err := governance.New(store, governance.Options{})
	if err != nil {
		t.Fatal(err)
	}
	controller, err := New(service, settings, governanceService, Options{
		PublicBase: "https://paste.example", AdministratorSubjects: []string{"usr_ADMIN"},
	})
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

func TestPublicSettingsAndAdministratorUpdate(t *testing.T) {
	controller := testController(t)
	read, err := controller.Public().GetSiteSettings(context.Background(), &v1.GetSiteSettingsReq{})
	if err != nil {
		t.Fatal(err)
	}
	if read.Settings.Name != site.DefaultName || read.Settings.Revision != 1 {
		t.Fatalf("unexpected public settings: %#v", read)
	}
	if _, err := controller.Administrator().UpdateSiteSettings(userContext("usr_MEMBER"), &v1.UpdateSiteSettingsReq{
		Name: "不能修改", ExpectedRevision: 1,
	}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("expected non-administrator to be forbidden, got %v", err)
	}
	updated, err := controller.Administrator().UpdateSiteSettings(userContext("usr_ADMIN"), &v1.UpdateSiteSettingsReq{
		Name: "月离代码", Description: "分享调试现场", ExpectedRevision: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Settings.Name != "月离代码" || updated.Settings.Revision != 2 {
		t.Fatalf("unexpected settings update: %#v", updated)
	}
}

func TestAdministratorCanListGovernAndDeleteAcrossOwners(t *testing.T) {
	controller := testController(t)
	owner := userContext("usr_OWNER")
	created, err := controller.Public().CreatePaste(owner, &v1.CreatePasteReq{
		Title: "Needs review", Files: []v1.FileInput{{Path: "main.go", Language: "go", Content: "package main"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	administrator := userContext("usr_ADMIN")
	listed, err := controller.Administrator().ListPastes(administrator, &v1.ListAdministrationPastesReq{
		Query: "review", Ownership: "owned",
	})
	if err != nil {
		t.Fatal(err)
	}
	if listed.Total != 1 || len(listed.Pastes) != 1 || listed.Pastes[0].OwnerUserKey != "usr_OWNER" {
		t.Fatalf("unexpected administration list: %#v", listed)
	}
	visibility := "private"
	governed, err := controller.Administrator().GovernPaste(administrator, &v1.GovernPasteReq{
		Code: created.Paste.Code, ExpectedRevision: created.Paste.Revision, Visibility: &visibility,
	})
	if err != nil {
		t.Fatal(err)
	}
	if governed.Paste.Visibility != "private" || governed.Paste.Revision != 2 {
		t.Fatalf("unexpected governed Paste: %#v", governed)
	}
	if _, err := controller.Administrator().DeletePaste(administrator, &v1.AdministrationDeletePasteReq{
		Code: created.Paste.Code, ExpectedRevision: governed.Paste.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	deleted, err := controller.Administrator().ListPastes(administrator, &v1.ListAdministrationPastesReq{State: "deleted"})
	if err != nil {
		t.Fatal(err)
	}
	if deleted.Total != 1 || deleted.Pastes[0].State != "deleted" {
		t.Fatalf("deleted Paste was not retained for governance: %#v", deleted)
	}
}

func TestAdministratorGovernsUsersAndCreationLimits(t *testing.T) {
	controller := testController(t)
	owner := userContext("usr_ABUSER")
	create := func(ctx context.Context) error {
		_, err := controller.Public().CreatePaste(ctx, &v1.CreatePasteReq{
			Files: []v1.FileInput{{Path: "main.go", Language: "go", Content: "package main"}},
		})
		return err
	}
	if err := create(owner); err != nil {
		t.Fatal(err)
	}
	administrator := userContext("usr_ADMIN")
	users, err := controller.Administrator().ListUsers(administrator, &v1.ListAdministrationUsersReq{Query: "ABUSER"})
	if err != nil {
		t.Fatal(err)
	}
	if users.Total != 1 || users.Users[0].UsedToday != 1 || users.Users[0].EffectiveDailyLimit != governance.DefaultUserDailyLimit {
		t.Fatalf("unexpected user governance list: %#v", users)
	}
	suspended, err := controller.Administrator().UpdateUser(administrator, &v1.UpdateAdministrationUserReq{
		UserKey: "usr_ABUSER", State: "suspended", Reason: "abuse", ExpectedRevision: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if suspended.User.State != "suspended" || suspended.User.Revision != 1 {
		t.Fatalf("unexpected suspended user: %#v", suspended)
	}
	if code := problemCode(t, create(owner)); code != "paste.creation_suspended" {
		t.Fatalf("expected creation suspension, got %s", code)
	}
	settings, err := controller.Administrator().GetGovernanceSettings(administrator, &v1.GetGovernanceSettingsReq{})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := controller.Administrator().UpdateGovernanceSettings(administrator, &v1.UpdateGovernanceSettingsReq{
		UserDailyLimit: 1, AnonymousDailyLimit: 0, ExpectedRevision: settings.Settings.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Settings.UserDailyLimit != 1 || updated.Settings.AnonymousDailyLimit != 0 {
		t.Fatalf("unexpected governance settings: %#v", updated)
	}
	if code := problemCode(t, create(context.Background())); code != "paste.anonymous_creation_disabled" {
		t.Fatalf("expected anonymous creation to be disabled, got %s", code)
	}
	if _, err := controller.Administrator().ListUsers(userContext("usr_MEMBER"), &v1.ListAdministrationUsersReq{}); problemCode(t, err) != "paste.forbidden" {
		t.Fatalf("expected non-administrator denial, got %v", err)
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
