package server

import (
	"database/sql"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	goframeauth "github.com/yueli-official/foundation/go/goframe/auth"
	"github.com/yueli-official/paste/internal/httpapi"
	pasteruntime "github.com/yueli-official/paste/internal/runtime"
)

type Dependencies struct {
	Controller       *httpapi.Core
	Verifier         *foundationauth.Verifier
	PersonalVerifier *foundationauth.PersonalTokenVerifier
	PersonalSite     string
	Database         *sql.DB
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
	var verifier goframeauth.TokenVerifier
	if dependencies.Verifier != nil {
		verifier = foundationauth.CompositeVerifier{JWT: dependencies.Verifier, Personal: dependencies.PersonalVerifier}
	}
	server.Group("/", func(group *ghttp.RouterGroup) {
		middlewares := []ghttp.HandlerFunc{apiMiddleware}
		if verifier != nil {
			middlewares = append(middlewares, pasteruntime.OptionalAuth(verifier), httpapi.PersonalTokenRoutes)
		}
		group.Middleware(middlewares...)
		group.Bind(dependencies.Controller.Public())
	})
	server.Group("/", func(group *ghttp.RouterGroup) {
		middlewares := []ghttp.HandlerFunc{apiMiddleware}
		if verifier != nil {
			middlewares = append(middlewares, pasteruntime.RequiredAuth(verifier), httpapi.PersonalTokenRoutes)
		}
		group.Middleware(middlewares...)
		group.Bind(dependencies.Controller.Managed())
		group.Bind(dependencies.Controller.PersonalPermissions(dependencies.PersonalSite))
	})
	server.Group("/", func(group *ghttp.RouterGroup) {
		middlewares := []ghttp.HandlerFunc{apiMiddleware}
		if verifier != nil {
			middlewares = append(middlewares, pasteruntime.RequiredAuth(verifier), httpapi.PersonalTokenRoutes)
		}
		group.Middleware(middlewares...)
		group.Bind(dependencies.Controller.Administrator())
	})
}
