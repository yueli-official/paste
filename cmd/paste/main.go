package main

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	_ "github.com/lib/pq"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/httpapi"
	"github.com/yueli-official/paste/internal/paste"
	pastepostgres "github.com/yueli-official/paste/internal/postgres"
	pasteruntime "github.com/yueli-official/paste/internal/runtime"
	"github.com/yueli-official/paste/internal/server"
	"github.com/yueli-official/paste/internal/site"
)

func main() {
	if output := os.Getenv("PASTE_OPENAPI_OUTPUT"); output != "" {
		must(exportOpenAPI(output))
		return
	}
	ctx := context.Background()
	databaseURL := requiredEnvironment("PASTE_DATABASE_URL")
	database, err := sql.Open("postgres", databaseURL)
	must(err)
	defer database.Close()
	must(database.PingContext(ctx))

	store, err := pastepostgres.New(database)
	must(err)
	pastes, err := paste.New(store, paste.Options{})
	must(err)
	settings, err := site.New(store, site.Options{})
	must(err)
	governanceService, err := governance.New(store, governance.Options{})
	must(err)
	controller, err := httpapi.New(pastes, settings, governanceService, httpapi.Options{
		PublicBase:            environment("PASTE_PUBLIC_BASE_URL", "http://localhost:3010"),
		AdministratorSubjects: commaSeparated(os.Getenv("PASTE_ADMIN_SUBS")),
	})
	must(err)

	var verifier *foundationauth.Verifier
	if jwksURL := strings.TrimSpace(os.Getenv("PASTE_JWKS_URL")); jwksURL != "" {
		verifier, err = pasteruntime.NewRemoteVerifier(pasteruntime.RemoteVerifierConfig{
			JWKSURL: jwksURL, Issuer: requiredEnvironment("PASTE_JWKS_ISSUER"),
			Audience:          environment("PASTE_JWKS_AUDIENCE", "paste-api"),
			AllowLoopbackHTTP: environment("PASTE_JWKS_ALLOW_LOOPBACK_HTTP", "false") == "true",
		})
		must(err)
	}

	httpServer := g.Server()
	httpServer.SetAddr(environment("PASTE_API_ADDRESS", "127.0.0.1:8091"))
	server.Configure(httpServer, server.Dependencies{Controller: controller, Verifier: verifier, Database: database})
	httpServer.Run()
}

func environment(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func requiredEnvironment(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		panic(name + " is required")
	}
	return value
}

func commaSeparated(raw string) []string {
	values := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
