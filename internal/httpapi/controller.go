package httpapi

import (
	"context"
	"errors"
	"strings"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/problem"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/paste"
	"github.com/yueli-official/paste/internal/pasteerr"
)

type Core struct {
	pastes     *paste.Service
	publicBase string
}

func New(pastes *paste.Service, publicBase string) (*Core, error) {
	if pastes == nil {
		return nil, errors.New("paste/httpapi: Paste service is required")
	}
	return &Core{pastes: pastes, publicBase: strings.TrimRight(strings.TrimSpace(publicBase), "/")}, nil
}

type Public struct{ core *Core }
type Managed struct{ core *Core }

func (core *Core) Public() *Public   { return &Public{core: core} }
func (core *Core) Managed() *Managed { return &Managed{core: core} }

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

func (controller *Managed) ListMyPastes(ctx context.Context, _ *v1.ListMyPastesReq) (*v1.ListMyPastesRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	values, err := controller.core.pastes.ListMine(ctx, userKey)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	result := make([]v1.PasteSummaryView, len(values))
	for index, value := range values {
		result[index] = controller.core.summary(value)
	}
	return &v1.ListMyPastesRes{Pastes: result}, nil
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
	return &v1.DeletePasteRes{}, nil
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

func mustProblem(descriptor problem.Descriptor) error {
	mapped, err := problem.NewError(descriptor, nil)
	if err != nil {
		return err
	}
	return mapped
}
