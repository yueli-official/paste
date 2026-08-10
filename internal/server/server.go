package server

import (
	"database/sql"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/paste/internal/httpapi"
	pasteruntime "github.com/yueli-official/paste/internal/runtime"
)

type Dependencies struct {
	Controller *httpapi.Core
	Verifier   *foundationauth.Verifier
	Database   *sql.DB
}

func Configure(server *ghttp.Server, dependencies Dependencies) {
	apiMiddleware := pasteruntime.MustAPIMiddleware().Handle
	server.Use(pasteruntime.TraceRoute)

	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(apiMiddleware)
		group.GET("/healthz", httpapi.Healthz)
		group.GET("/readyz", pasteruntime.ReadinessHandler(dependencies.Database))
	})
	if dependencies.Controller == nil {
		return
	}
	server.Group("/", func(group *ghttp.RouterGroup) {
		middlewares := []ghttp.HandlerFunc{apiMiddleware}
		if dependencies.Verifier != nil {
			middlewares = append(middlewares, pasteruntime.OptionalAuth(dependencies.Verifier))
		}
		group.Middleware(middlewares...)
		group.Bind(dependencies.Controller.Public())
	})
	server.Group("/", func(group *ghttp.RouterGroup) {
		middlewares := []ghttp.HandlerFunc{apiMiddleware}
		if dependencies.Verifier != nil {
			middlewares = append(middlewares, pasteruntime.RequiredAuth(dependencies.Verifier))
		}
		group.Middleware(middlewares...)
		group.Bind(dependencies.Controller.Managed())
	})
}
