package httpapi

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	v1 "github.com/yueli-official/paste/api/v1"
	"github.com/yueli-official/paste/internal/pasteerr"
	"github.com/yueli-official/paste/internal/site"
)

func (controller *Public) GetSiteBranding(ctx context.Context, _ *v1.GetSiteBrandingReq) (*v1.GetSiteBrandingRes, error) {
	settings, err := controller.core.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	header := g.RequestFromCtx(ctx).Response.Header()
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Cache-Control", "public, no-cache")
	return &v1.GetSiteBrandingRes{BrandImageURLs: settings.BrandImageURLs.WithDefaults(site.BrandingDefaults)}, nil
}

func (controller *Administrator) PatchSiteBranding(ctx context.Context, request *v1.PatchSiteBrandingReq) (*v1.PatchSiteBrandingRes, error) {
	administrator, err := controller.core.requiredAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := controller.core.settings.PatchBranding(ctx, request.BrandImagePatch, administrator)
	if err != nil {
		return nil, pasteerr.Map(err)
	}
	g.RequestFromCtx(ctx).Response.Header().Set("Cache-Control", "private, no-store")
	return &v1.PatchSiteBrandingRes{BrandImageURLs: settings.BrandImageURLs.WithDefaults(site.BrandingDefaults)}, nil
}
