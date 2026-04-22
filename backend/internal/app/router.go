// Package app wires the HTTP router. Extracted from cmd/server/main.go so
// tests can build a router with stub dependencies and assert the registered
// routes match the OpenAPI contract (see contract_test.go).
package app

import (
	"context"
	"net/http"
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
// spec 1:1 — `contract_test.go` enforces this.
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

	r.Route("/api", func(r chi.Router) {
		r.Get("/landing", handler.Landing)

		// Auth routes (rate limited)
		r.Route("/auth", func(r chi.Router) {
			r.Use(httprate.LimitByIP(10, time.Minute))

			r.Get("/google", d.Auth.GoogleRedirect)
			r.Get("/google/callback", d.Auth.GoogleCallback)
			r.Post("/magic-link", d.Auth.RequestMagicLink)
			r.Get("/verify", d.Auth.VerifyMagicLink)
			r.Post("/refresh", d.Auth.RefreshTokens)
			r.Post("/mobile/google", d.MobileAuth.GoogleIDToken)
			r.Post("/mobile/verify", d.MobileAuth.VerifyMagicLink)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(d.TokenGen, d.DevBypass))

			r.Get("/me", d.User.GetProfile)
			r.Patch("/me", d.User.UpdateProfile)
			r.Post("/auth/logout", d.Auth.Logout)
			r.Post("/auth/logout-all", d.Auth.LogoutAll)
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(d.TokenGen, d.DevBypass))
			r.Use(middleware.RequireRole("admin", d.RoleLookup))

			r.Get("/admin/ping", func(w http.ResponseWriter, req *http.Request) {
				handler.JSON(w, req, http.StatusOK, map[string]string{"role": "admin"})
			})
		})
	})

	return r
}
