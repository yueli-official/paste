package main

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	_ "github.com/lib/pq"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	authorizationpostgres "github.com/yueli-official/foundation/go/authorization/postgres"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/httpapi"
	"github.com/yueli-official/paste/internal/paste"
	"github.com/yueli-official/paste/internal/pasteauthz"
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
	definition, err := authorization.Compile(pasteauthz.Definition())
	must(err)
	subjects := []authorization.SubjectRef{}
	for _, id := range commaSeparated(os.Getenv("PASTE_ADMIN_SUBS")) {
		subjects = append(subjects, authorization.SubjectRef{Kind: authorization.SubjectUser, ID: id})
	}
	runtime, err := authorizationpostgres.New(ctx, definition, authorizationpostgres.Options{DB: database, InstanceKey: "paste:" + environment("PASTE_INSTANCE_ID", environment("PASTE_JWKS_AUDIENCE", "paste-api")), Memory: authorization.MemoryOptions{RootScopeID: pasteauthz.RootScopeID, ProtectedSubjects: subjects, AllowUnclaimed: len(subjects) == 0}})
	must(err)
	controller, err := httpapi.New(pastes, settings, governanceService, httpapi.Options{
		PublicBase:    environment("PASTE_PUBLIC_BASE_URL", "http://localhost:3010"),
		Authorization: runtime,
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
	personalSite := strings.TrimSpace(os.Getenv("PASTE_PERSONAL_TOKEN_SITE_ID"))
	var personalVerifier *foundationauth.PersonalTokenVerifier
	if personalSite != "" {
		personalVerifier, err = foundationauth.NewPersonalTokenVerifier(requiredEnvironment("PASTE_PERSONAL_TOKEN_VERIFY_URL"), personalSite, nil,
			foundationauth.PersonalTransportOptions{AllowHTTP: environment("PASTE_PERSONAL_TOKEN_ALLOW_HTTP", "false") == "true"})
		must(err)
	}
	httpServer.SetAddr(environment("PASTE_API_ADDRESS", "127.0.0.1:8091"))
	server.Configure(httpServer, server.Dependencies{Controller: controller, Verifier: verifier, PersonalVerifier: personalVerifier, PersonalSite: personalSite, Database: database})
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
