// Package app wires the HTTP router. Extracted from cmd/server/main.go so
// tests can build a router with stub dependencies. Spec/route parity is
// enforced by `cmd/check-parity` and Huma's typed registration.
package app

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
)

// RouterDeps bundles everything NewRouter needs. main() builds these from the
// fully-initialized composition root; tests may supply stub values because
// NewRouter only registers routes — it does not invoke handlers.
type RouterDeps struct {
	Auth       *handler.AuthHandler
	MobileAuth *handler.MobileAuthHandler
	User       *handler.UserHandler
	TokenGen   domain.TokenGenerator
	RoleLookup func(ctx context.Context, userID string) (string, error)
	DB         *gorm.DB
	AppEnv     string
	AppURL     string
	DevBypass  bool
}

// NewRouter wires every HTTP route. Mounted routes must match the OpenAPI
// spec 1:1 — Huma registration is the source of truth, and `make check:parity`
// catches spec paths with no client consumer.
func NewRouter(d RouterDeps) chi.Router {
	r := chi.NewRouter()

	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.AccessLog)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestSize(1 << 20)) // 1 MiB body limit
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{d.AppURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health / readiness / metrics
	r.Get("/healthz", handler.Healthz)
	r.Get("/readyz", handler.Readyz(d.DB))
	r.Handle("/metrics", promhttp.Handler())

	// Dev-only route table introspection. Used by `make routes`.
	if d.AppEnv != "production" {
		r.Get("/_debug/routes", handler.RoutesDebug(r))
	}

	// Huma API — source of truth for the OpenAPI spec. All JSON-in /
	// JSON-out endpoints register here. Per-operation middleware (auth,
	// role, rate limit) is wired via huma.Operation.Middlewares and the
	// Huma-flavor adapters in handler/auth_huma.go +
	// handler/middleware_bridge.go.
	handler.InstallErrorEnvelope() // override huma.NewError once
	api := NewAPI(r)
	authMW := handler.HumaAuthMiddleware(api, d.TokenGen, d.DevBypass)
	requireAdmin := handler.HumaRequireRole(api, "admin", d.RoleLookup)
	rateLimitMW := handler.HumaFromHTTP(httprate.LimitByIP(10, time.Minute))

	handler.RegisterLanding(api)
	d.User.RegisterUser(api, authMW)
	handler.RegisterAdmin(api, authMW, requireAdmin)
	d.Auth.RegisterAuth(api, authMW, rateLimitMW)
	d.MobileAuth.RegisterMobileAuth(api, rateLimitMW)

	// Chi-native routes — browser-flow redirects with Set-Cookie. Same
	// rate limiter as the JSON auth endpoints, applied chi-style here.
	r.Route("/api/auth", func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, time.Minute))
		r.Get("/google", d.Auth.GoogleRedirect)
		r.Get("/google/callback", d.Auth.GoogleCallback)
		r.Get("/verify", d.Auth.VerifyMagicLink)
	})

	return r
}
