// gen-openapi builds the Huma API in-process and writes its generated
// OpenAPI 3.1 spec to backend/api/openapi.yaml. Wired into `make gen:openapi`.
//
// The spec produced by Huma is the source of truth for downstream
// codegen (web's openapi-typescript, mobile DTOs once codegen lands).
// The on-disk YAML is purely a tracked artifact; never hand-edit.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/app"
	"github.com/eridia/initium/backend/internal/domain"
)

func main() {
	out := flag.String("out", "api/openapi.yaml", "output path (relative to backend/)")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Build a router with stub deps. We only need route registration;
	// handlers are never invoked, so nil services are safe.
	r := chi.NewRouter()
	api := app.NewAPI(r)
	handler.InstallErrorEnvelope()

	// Register every Huma operation. The middleware functions don't run
	// during spec generation, but we need real instances for the
	// Operation.Middlewares slice to be valid.
	authMW := handler.HumaAuthMiddleware(api, stubTokenGen{}, false)
	requireAdmin := handler.HumaRequireRole(api, "admin", stubRoleLookup)

	handler.RegisterLanding(api)
	handler.NewUserHandler(nil).RegisterUser(api, authMW)
	handler.RegisterAdmin(api, authMW, requireAdmin)

	// TODO: register the remaining auth JSON operations once they migrate.
	// For now (step 2) the magic-link / refresh / logout / mobile auth
	// endpoints stay chi-native and are not in the generated spec.

	yaml, err := api.OpenAPI().YAML()
	if err != nil {
		slog.Error("marshalling openapi spec", "error", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*out, yaml, 0o644); err != nil {
		slog.Error("writing spec", "path", *out, "error", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d bytes)\n", *out, len(yaml))
}

// stubTokenGen / stubRoleLookup satisfy the constructor signatures so we
// can build the API graph without importing infra (would create import cycles).
type stubTokenGen struct{}

func (stubTokenGen) GenerateAccessToken(_, _ string) (string, error)      { return "", nil }
func (stubTokenGen) GenerateRefreshToken() (string, error)                { return "", nil }
func (stubTokenGen) ValidateAccessToken(_ string) (string, string, error) { return "", "", nil }
func (stubTokenGen) HashToken(_ string) string                            { return "" }

func stubRoleLookup(_ context.Context, _ string) (string, error) { return "", nil }

// Compile-time assertion that stubTokenGen implements domain.TokenGenerator.
var _ domain.TokenGenerator = stubTokenGen{}
