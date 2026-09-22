package httpapi

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/pasteauthz"
	"github.com/yueli-official/paste/internal/pasteerr"
)

var personalPermissions = []foundationauth.PersonalPermission{
	{Key: string(pasteauthz.CapabilityPasteCreate), Label: "创建片段", Description: "以当前账号身份创建多文件代码片段，仍受账号创建状态与每日额度限制。"},
	{Key: string(pasteauthz.CapabilityPasteRead), Label: "读取自己的片段", Description: "列出并读取当前账号拥有的片段，不包含其他用户或匿名片段的后台治理。"},
	{Key: string(pasteauthz.CapabilityPasteUpdate), Label: "编辑自己的片段", Description: "编辑当前账号拥有的片段，继续遵守修订号并发保护。"},
	{Key: string(pasteauthz.CapabilityPasteDelete), Label: "删除自己的片段", Description: "删除当前账号拥有的片段，继续遵守修订号并发保护。"},
	{Key: string(pasteauthz.CapabilityPasteModerate), Label: "治理站内片段", Description: "查询、修改和删除站内片段；仍需当前 Paste 管理员权限。"},
	{Key: string(pasteauthz.CapabilitySettingsManage), Label: "管理公开站点设置", Description: "修改 Paste 名称与公开说明；仍需当前 Paste 管理员权限。"},
}

type PersonalPermissions struct {
	core *Core
	site string
}

func (core *Core) PersonalPermissions(site string) *PersonalPermissions {
	return &PersonalPermissions{core: core, site: site}
}

func (controller *PersonalPermissions) GetPersonalPermissions(ctx context.Context, req *v1.PersonalPermissionsReq) (*v1.PersonalPermissionsRes, error) {
	principal, _ := foundationauth.FromContext(ctx)
	if principal == nil || principal.SubjectKind != foundationauth.SubjectClient || principal.ClientID != "identity-svc" || !principal.HasScope(foundationauth.PersonalPermissionsScope) || controller.site == "" {
		return nil, mustProblem(pasteerr.Forbidden)
	}
	if controller.core == nil || controller.core.authorization == nil {
		return nil, mustProblem(pasteerr.Internal)
	}
	userCtx := foundationauth.NewContext(ctx, &foundationauth.Principal{Subject: req.UserKey, SubjectKind: foundationauth.SubjectUser})
	items := make([]foundationauth.PersonalPermission, 0, len(personalPermissions))
	for _, permission := range personalPermissions {
		allowed, err := controller.core.authorization.CheckCapability(userCtx, authorization.CapabilityKey(permission.Key))
		if err != nil {
			return nil, pasteerr.Map(err)
		}
		if allowed {
			items = append(items, permission)
		}
	}
	return &v1.PersonalPermissionsRes{Site: controller.site, UserKey: req.UserKey, Items: items}, nil
}

type personalRoute struct {
	method, path string
	capability   authorization.CapabilityKey
}

var personalRoutes = []personalRoute{
	{"POST", "/api/v1/pastes", pasteauthz.CapabilityPasteCreate},
	{"GET", "/api/v1/pastes/{code}", pasteauthz.CapabilityPasteRead},
	{"POST", "/api/v1/pastes/{code}/access", pasteauthz.CapabilityPasteRead},
	{"GET", "/api/v1/settings", pasteauthz.CapabilityPasteRead},
	{"GET", "/api/v1/me/pastes", pasteauthz.CapabilityPasteRead},
	{"GET", "/api/v1/me/pastes/{code}", pasteauthz.CapabilityPasteRead},
	{"PATCH", "/api/v1/me/pastes/{code}", pasteauthz.CapabilityPasteUpdate},
	{"DELETE", "/api/v1/me/pastes/{code}", pasteauthz.CapabilityPasteDelete},
	{"GET", "/api/v1/admin/pastes", pasteauthz.CapabilityPasteModerate},
	{"PATCH", "/api/v1/admin/pastes/{code}", pasteauthz.CapabilityPasteModerate},
	{"DELETE", "/api/v1/admin/pastes/{code}", pasteauthz.CapabilityPasteModerate},
	{"PATCH", "/api/v1/admin/settings", pasteauthz.CapabilitySettingsManage},
}

func PersonalTokenRoutes(request *ghttp.Request) {
	principal, _ := foundationauth.FromContext(request.Context())
	if principal != nil && principal.IsPersonalToken() && !allowsPersonalRoute(request.Context(), request.Method, request.URL.Path) {
		request.SetError(mustProblem(pasteerr.Forbidden))
		return
	}
	request.Middleware.Next()
}

func allowsPersonalRoute(ctx context.Context, method, path string) bool {
	parts := strings.Split(path, "/")
	for _, route := range personalRoutes {
		if route.method != method {
			continue
		}
		pattern := strings.Split(route.path, "/")
		if len(pattern) != len(parts) {
			continue
		}
		matches := true
		for index, part := range pattern {
			if strings.HasPrefix(part, "{") {
				matches = matches && parts[index] != "" && parts[index] != "." && parts[index] != ".."
			} else {
				matches = matches && parts[index] == part
			}
		}
		if matches && foundationauth.AllowsPersonalCapability(ctx, string(route.capability)) {
			return true
		}
	}
	return false
}
