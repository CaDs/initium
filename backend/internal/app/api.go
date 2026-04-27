package app

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// NewAPI wraps the chi router with Huma. Huma is the source of truth
// for every JSON-in / JSON-out endpoint under /api/. Chi-native handlers
// (OAuth redirects, /healthz, /readyz, /metrics, /_debug/routes) stay
// registered directly on the router and are not part of the spec.
//
// All middleware applied to the chi router — CORS, RealIP, RequestID,
// AccessLog, Recoverer, RequestSize — flows through to Huma operations
// because humachi just registers HTTP handlers on the underlying mux.
// Per-operation middleware (rate limiting, auth gates) is applied via
// huma.Operation.Middlewares at registration time.
func NewAPI(r chi.Router) huma.API {
	cfg := huma.DefaultConfig("Initium API", "0.1.0")
	cfg.Info.Description = "POC starter template API. Generated from Go code by Huma — never hand-edit."
	cfg.Servers = []*huma.Server{
		{URL: "http://localhost:8000", Description: "Local development"},
	}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	// Disable Huma's `$schema` linker — by default it injects a
	// `$schema` URL into every JSON response body AND adds the field
	// to every component schema in the spec. Three settings remove it:
	//   - Transformers: nil drops the body-write hook.
	//   - CreateHooks: nil drops the spec-mutation hook that adds the
	//     $schema field to every component schema.
	//   - SchemasPath: "" disables the runtime /schemas/{name} endpoint.
	// Web Zod and mobile Codable parsers see only the declared shape.
	cfg.SchemasPath = ""
	cfg.Transformers = nil
	cfg.CreateHooks = nil

	return humachi.New(r, cfg)
}
