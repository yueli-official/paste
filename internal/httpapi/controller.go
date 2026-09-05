package httpapi

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/net/ghttp"
	"net/http"
	"strings"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/problem"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/paste"
	"github.com/yueli-official/paste/internal/pasteerr"
	"github.com/yueli-official/paste/internal/site"
)

type Core struct {
	pastes         *paste.Service
	settings       *site.Service
	governance     *governance.Service
	publicBase     string
	administrators map[string]struct{}
}

type Options struct {
	PublicBase            string
	AdministratorSubjects []string
}

func New(pastes *paste.Service, settings *site.Service, governanceService *governance.Service, options Options) (*Core, error) {
	if pastes == nil {
		return nil, errors.New("paste/httpapi: Paste service is required")
	}
	if settings == nil {
		return nil, errors.New("paste/httpapi: Site service is required")
	}
	if governanceService == nil {
		return nil, errors.New("paste/httpapi: Governance service is required")
	}
	administrators := make(map[string]struct{}, len(options.AdministratorSubjects))
	for _, raw := range options.AdministratorSubjects {
		if subject := strings.TrimSpace(raw); subject != "" {
			administrators[subject] = struct{}{}
		}
	}
	return &Core{
		pastes: pastes, settings: settings, governance: governanceService,
		publicBase:     strings.TrimRight(strings.TrimSpace(options.PublicBase), "/"),
		administrators: administrators,
	}, nil
}

type Public struct{ core *Core }
type Managed struct{ core *Core }
type Administrator struct{ core *Core }

func (core *Core) Public() *Public               { return &Public{core: core} }
func (core *Core) Managed() *Managed             { return &Managed{core: core} }
func (core *Core) Administrator() *Administrator { return &Administrator{core: core} }

func (controller *Public) CreatePaste(ctx context.Context, request *v1.CreatePasteReq) (*v1.CreatePasteRes, error) {
	owner, err := optionalUser(ctx)
	if err != nil {
		return nil, err
	}
	files := make([]paste.File, len(request.Files))
	for index, file := range request.Files {
		files[index] = paste.File{Path: file.Path, Language: file.Language, Content: file.Content, Order: index}
	}
	created, err := controller.core.pastes.Create(ctx, paste.CreateInput{
		OwnerUserKey: owner,
		Title:        request.Title,
		Description:  request.Description,
		Tags:         request.Tags,
		Files:        files,
		Visibility:   paste.Visibility(request.Visibility),
		Password:     request.Password,
		ExpiresAt:    request.ExpiresAt,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	writeSuccess(ctx, http.StatusCreated, "/api/v1/pastes/"+created.Code)
	return &v1.CreatePasteRes{Paste: controller.core.view(created)}, nil
}

func (controller *Public) GetPaste(ctx context.Context, request *v1.GetPasteReq) (*v1.GetPasteRes, error) {
	userKey, err := optionalUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := controller.core.pastes.Open(ctx, request.Code, paste.Access{UserKey: userKey})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.GetPasteRes{Paste: controller.core.view(value)}, nil
}

func (controller *Public) UnlockPaste(ctx context.Context, request *v1.UnlockPasteReq) (*v1.UnlockPasteRes, error) {
	userKey, err := optionalUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := controller.core.pastes.Open(ctx, request.Code, paste.Access{UserKey: userKey, Password: request.Password})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.UnlockPasteRes{Paste: controller.core.view(value)}, nil
}

func (controller *Public) GetSiteSettings(ctx context.Context, _ *v1.GetSiteSettingsReq) (*v1.GetSiteSettingsRes, error) {
	value, err := controller.core.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSiteSettingsRes{Settings: siteSettingsView(value)}, nil
}

func (controller *Managed) ListMyPastes(ctx context.Context, request *v1.ListMyPastesReq) (*v1.ListMyPastesRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	pageNumber, size := pagination(request.Page, request.Size)
	page, err := controller.core.pastes.ListMinePage(ctx, userKey, request.Query, size, (pageNumber-1)*size)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	items := make([]v1.PasteSummaryView, 0, len(page.Items))
	for _, value := range page.Items {
		items = append(items, controller.core.summaryItem(value))
	}
	return &v1.ListMyPastesRes{Items: items, Total: page.Total, Page: pageNumber, Size: size}, nil
}

func (controller *Managed) GetMyPaste(ctx context.Context, request *v1.GetMyPasteReq) (*v1.GetMyPasteRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := controller.core.pastes.GetMine(ctx, request.Code, userKey)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.GetMyPasteRes{Paste: controller.core.view(value)}, nil
}

func (controller *Managed) UpdatePaste(ctx context.Context, request *v1.UpdatePasteReq) (*v1.UpdatePasteRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	var visibility *paste.Visibility
	if request.Visibility != nil {
		value := paste.Visibility(*request.Visibility)
		visibility = &value
	}
	var files *[]paste.File
	if request.Files != nil {
		values := make([]paste.File, len(*request.Files))
		for index, file := range *request.Files {
			values[index] = paste.File{Path: file.Path, Language: file.Language, Content: file.Content, Order: index}
		}
		files = &values
	}
	updated, err := controller.core.pastes.Update(ctx, request.Code, paste.UpdateInput{
		OwnerUserKey:     userKey,
		ExpectedRevision: request.ExpectedRevision,
		Title:            request.Title,
		Description:      request.Description,
		Tags:             request.Tags,
		Files:            files,
		Visibility:       visibility,
		Password:         request.Password,
		ExpiresAt:        request.ExpiresAt,
		ClearExpiry:      request.ClearExpiry,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.UpdatePasteRes{Paste: controller.core.view(updated)}, nil
}

func (controller *Managed) DeletePaste(ctx context.Context, request *v1.DeletePasteReq) (*v1.DeletePasteRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := controller.core.pastes.Delete(ctx, request.Code, userKey, request.ExpectedRevision); err != nil {
		return nil, pasteerr.Map(err)
	}
	writeSuccess(ctx, http.StatusNoContent, "")
	return &v1.DeletePasteRes{}, nil
}

func (controller *Administrator) ListPastes(ctx context.Context, request *v1.ListAdministrationPastesReq) (*v1.ListAdministrationPastesRes, error) {
	pageNumber, size := pagination(request.Page, request.Size)
	if _, err := controller.core.requiredAdministrator(ctx); err != nil {
		return nil, err
	}
	page, err := controller.core.pastes.ListForAdministration(ctx, paste.AdministrationQuery{
		Query: request.Query, Visibility: paste.Visibility(request.Visibility), State: paste.State(request.State),
		Ownership: request.Ownership, Limit: size, Offset: (pageNumber - 1) * size,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	values := make([]v1.AdministrationPasteView, len(page.Items))
	for index, value := range page.Items {
		values[index] = controller.core.administrationView(value)
	}
	return &v1.ListAdministrationPastesRes{Items: values, Total: page.Total, Size: size, Page: pageNumber}, nil
}

func (controller *Administrator) GetSession(ctx context.Context, _ *v1.GetAdministrationSessionReq) (*v1.GetAdministrationSessionRes, error) {
	userKey, err := controller.core.requiredAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetAdministrationSessionRes{Allowed: true, UserKey: userKey}, nil
}

func (controller *Administrator) GovernPaste(ctx context.Context, request *v1.GovernPasteReq) (*v1.GovernPasteRes, error) {
	if _, err := controller.core.requiredAdministrator(ctx); err != nil {
		return nil, err
	}
	var visibility *paste.Visibility
	if request.Visibility != nil {
		value := paste.Visibility(*request.Visibility)
		visibility = &value
	}
	updated, err := controller.core.pastes.Govern(ctx, request.Code, paste.GovernanceInput{
		ExpectedRevision: request.ExpectedRevision, Visibility: visibility,
		ExpiresAt: request.ExpiresAt, ClearExpiry: request.ClearExpiry,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.GovernPasteRes{Paste: controller.core.administrationView(administrationItem(updated))}, nil
}

func (controller *Administrator) DeletePaste(ctx context.Context, request *v1.AdministrationDeletePasteReq) (*v1.AdministrationDeletePasteRes, error) {
	if _, err := controller.core.requiredAdministrator(ctx); err != nil {
		return nil, err
	}
	if err := controller.core.pastes.DeleteAsAdministrator(ctx, request.Code, request.ExpectedRevision); err != nil {
		return nil, pasteerr.Map(err)
	}
	writeSuccess(ctx, http.StatusNoContent, "")
	return &v1.AdministrationDeletePasteRes{}, nil
}

func (controller *Administrator) ListUsers(ctx context.Context, request *v1.ListAdministrationUsersReq) (*v1.ListAdministrationUsersRes, error) {
	pageNumber, size := pagination(request.Page, request.Size)
	if _, err := controller.core.requiredAdministrator(ctx); err != nil {
		return nil, err
	}
	page, err := controller.core.governance.ListUsers(ctx, governance.UserQuery{
		Query: request.Query, State: governance.UserState(request.State), Limit: size, Offset: (pageNumber - 1) * size,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	users := make([]v1.AdministrationUserView, 0, len(page.Users))
	for _, user := range page.Users {
		users = append(users, administrationUserView(user))
	}
	return &v1.ListAdministrationUsersRes{Items: users, Total: page.Total, Size: size, Page: pageNumber}, nil
}

func (controller *Administrator) UpdateUser(ctx context.Context, request *v1.UpdateAdministrationUserReq) (*v1.UpdateAdministrationUserRes, error) {
	administrator, err := controller.core.requiredAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := controller.core.governance.UpdateUser(ctx, governance.UpdateUserInput{
		UserKey: request.UserKey, State: governance.UserState(request.State),
		DailyLimitOverride: request.DailyLimitOverride, ClearDailyLimit: request.ClearDailyLimit,
		Reason: request.Reason, ExpectedRevision: request.ExpectedRevision, UpdatedBy: administrator,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.UpdateAdministrationUserRes{User: administrationUserPolicyView(updated)}, nil
}

func (controller *Administrator) GetGovernanceSettings(ctx context.Context, _ *v1.GetGovernanceSettingsReq) (*v1.GetGovernanceSettingsRes, error) {
	if _, err := controller.core.requiredAdministrator(ctx); err != nil {
		return nil, err
	}
	value, err := controller.core.governance.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetGovernanceSettingsRes{Settings: governanceSettingsView(value)}, nil
}

func (controller *Administrator) UpdateGovernanceSettings(ctx context.Context, request *v1.UpdateGovernanceSettingsReq) (*v1.UpdateGovernanceSettingsRes, error) {
	administrator, err := controller.core.requiredAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := controller.core.governance.UpdateSettings(ctx, governance.UpdateSettingsInput{
		UserDailyLimit: request.UserDailyLimit, AnonymousDailyLimit: request.AnonymousDailyLimit,
		ExpectedRevision: request.ExpectedRevision, UpdatedBy: administrator,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.UpdateGovernanceSettingsRes{Settings: governanceSettingsView(updated)}, nil
}

func (controller *Administrator) UpdateSiteSettings(ctx context.Context, request *v1.UpdateSiteSettingsReq) (*v1.UpdateSiteSettingsRes, error) {
	administrator, err := controller.core.requiredAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := controller.core.settings.Update(ctx, site.UpdateInput{
		Name: request.Name, Description: request.Description,
		ExpectedRevision: request.ExpectedRevision, UpdatedBy: administrator,
	})
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	return &v1.UpdateSiteSettingsRes{Settings: siteSettingsView(updated)}, nil
}

func (core *Core) view(value paste.Paste) v1.PasteView {
	files := make([]v1.FileView, len(value.Files))
	for index, file := range value.Files {
		files[index] = v1.FileView{Path: file.Path, Language: file.Language, Content: file.Content, Order: file.Order}
	}
	return v1.PasteView{
		ID: value.ID, Code: value.Code, ShareURL: core.shareURL(value.Code), Title: value.Title,
		Description: value.Description, Tags: append([]string{}, value.Tags...), Files: files,
		Visibility: string(value.Visibility), PasswordProtected: value.PasswordGuard, State: string(value.State),
		Revision: value.Revision, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ExpiresAt: value.ExpiresAt,
	}
}

func (core *Core) summary(value paste.Paste) v1.PasteSummaryView {
	language := "text"
	if len(value.Files) > 0 {
		language = value.Files[0].Language
	}
	return v1.PasteSummaryView{
		Code: value.Code, ShareURL: core.shareURL(value.Code), Title: value.Title,
		Tags: append([]string{}, value.Tags...), FileCount: len(value.Files), PrimaryLanguage: language,
		Visibility: string(value.Visibility), PasswordProtected: value.PasswordGuard, State: string(value.State),
		Revision: value.Revision, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ExpiresAt: value.ExpiresAt,
	}
}

func (core *Core) administrationView(value paste.AdministrationItem) v1.AdministrationPasteView {
	return v1.AdministrationPasteView{
		Code: value.Code, ShareURL: core.shareURL(value.Code), OwnerUserKey: value.OwnerUserKey,
		Title: value.Title, Tags: append([]string{}, value.Tags...), FileCount: value.FileCount,
		PrimaryLanguage: value.PrimaryLanguage, Visibility: string(value.Visibility),
		PasswordProtected: value.PasswordProtected, State: string(value.State), Revision: value.Revision,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ExpiresAt: value.ExpiresAt,
	}
}

func administrationItem(value paste.Paste) paste.AdministrationItem {
	language := "text"
	if len(value.Files) > 0 {
		language = value.Files[0].Language
	}
	return paste.AdministrationItem{
		Code: value.Code, OwnerUserKey: value.OwnerUserKey, Title: value.Title,
		Tags: append([]string{}, value.Tags...), FileCount: len(value.Files), PrimaryLanguage: language,
		Visibility: value.Visibility, PasswordProtected: value.PasswordGuard, State: value.State,
		Revision: value.Revision, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ExpiresAt: value.ExpiresAt,
	}
}

func siteSettingsView(value site.Settings) v1.SiteSettingsView {
	return v1.SiteSettingsView{
		Name: value.Name, Description: value.Description, Revision: value.Revision, UpdatedAt: value.UpdatedAt,
	}
}

func governanceSettingsView(value governance.Settings) v1.GovernanceSettingsView {
	return v1.GovernanceSettingsView{
		UserDailyLimit: value.UserDailyLimit, AnonymousDailyLimit: value.AnonymousDailyLimit,
		Revision: value.Revision, UpdatedAt: value.UpdatedAt,
	}
}

func administrationUserView(value governance.User) v1.AdministrationUserView {
	return v1.AdministrationUserView{
		UserKey: value.UserKey, State: string(value.State), DailyLimitOverride: value.DailyLimitOverride,
		EffectiveDailyLimit: value.EffectiveDailyLimit, UsedToday: value.UsedToday,
		TotalPastes: value.TotalPastes, ActivePastes: value.ActivePastes, LastCreatedAt: value.LastCreatedAt,
		Reason: value.Reason, Revision: value.Revision, UpdatedAt: value.UpdatedAt,
	}
}

func administrationUserPolicyView(value governance.UserPolicy) v1.AdministrationUserPolicyView {
	return v1.AdministrationUserPolicyView{
		UserKey: value.UserKey, State: string(value.State), DailyLimitOverride: value.DailyLimitOverride,
		Reason: value.Reason, Revision: value.Revision, UpdatedAt: value.UpdatedAt,
	}
}

func (core *Core) shareURL(code string) string {
	if core.publicBase == "" {
		return "/p/" + code
	}
	return core.publicBase + "/p/" + code
}

func optionalUser(ctx context.Context) (string, error) {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok {
		return "", nil
	}
	if principal.SubjectKind == foundationauth.SubjectGuest {
		return "", nil
	}
	if !principal.IsUser() {
		return "", mustProblem(pasteerr.Forbidden)
	}
	return principal.Subject, nil
}

func requiredUser(ctx context.Context) (string, error) {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok {
		return "", mustProblem(pasteerr.Unauthorized)
	}
	if !principal.IsUser() {
		return "", mustProblem(pasteerr.Forbidden)
	}
	return principal.Subject, nil
}

func (core *Core) requiredAdministrator(ctx context.Context) (string, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return "", err
	}
	if _, ok := core.administrators[userKey]; !ok {
		return "", mustProblem(pasteerr.Forbidden)
	}
	return userKey, nil
}

func mustProblem(descriptor problem.Descriptor) error {
	mapped, err := problem.NewError(descriptor, nil)
	if err != nil {
		return err
	}
	return mapped
}

func pagination(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}
	if size > 100 {
		size = 100
	}
	return page, size
}
func writeSuccess(ctx context.Context, status int, location string) {
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		if location != "" {
			r.Response.Header().Set("Location", location)
		}
		r.Response.WriteHeader(status)
	}
}
func (core *Core) summaryItem(value paste.AdministrationItem) v1.PasteSummaryView {
	return v1.PasteSummaryView{Code: value.Code, ShareURL: core.shareURL(value.Code), Title: value.Title, Tags: append([]string{}, value.Tags...), FileCount: value.FileCount, PrimaryLanguage: value.PrimaryLanguage, Visibility: string(value.Visibility), PasswordProtected: value.PasswordProtected, State: string(value.State), Revision: value.Revision, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ExpiresAt: value.ExpiresAt}
}
