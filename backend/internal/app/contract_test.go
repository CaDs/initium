package app

import (
	"context"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
)

// excludedPaths is the set of operational route paths that are intentionally
// not documented in the OpenAPI spec. Keep this list short and justified.
// Chi registers r.Handle() for all HTTP methods, so we match by path only.
var excludedPaths = map[string]bool{
	"/metrics":            true, // Prometheus scrape endpoint, not an API contract
	"/docs":               true, // Huma runtime — auto-rendered docs UI
	"/openapi.yaml":       true, // Huma runtime spec endpoint
	"/openapi.json":       true, // Huma runtime spec endpoint
	"/openapi-3.0.yaml":   true, // Huma runtime spec endpoint
	"/openapi-3.0.json":   true, // Huma runtime spec endpoint
	"/schemas/{schema}":   true, // Huma runtime — per-schema JSON Schema
}

// TestRouter_MatchesOpenAPISpec asserts that every route registered on the
// chi router has a corresponding path in backend/api/openapi.yaml, and vice
// versa. Deleted in the next migration step — once the spec is generated
// from Huma, route↔spec parity is structurally guaranteed.
func TestRouter_MatchesOpenAPISpec(t *testing.T) {
	t.Parallel()

	router := NewRouter(RouterDeps{
		Auth:       handler.NewAuthHandler(nil, nil, "", false),
		MobileAuth: handler.NewMobileAuthHandler(nil),
		User:       handler.NewUserHandler(nil),
		TokenGen:   nil,
		RoleLookup: func(context.Context, string) (string, error) { return "", nil },
		DB:         nil,
		AppEnv:     "development", // mount /_debug/routes so we can check it
		AppURL:     "http://localhost:3000",
		DevBypass:  true,
	})

	routerRoutes := walkRouter(t, router)
	specRoutes := loadSpecRoutes(t)

	onlyInRouter := diff(routerRoutes, specRoutes)
	onlyInSpec := diff(specRoutes, routerRoutes)

	if len(onlyInRouter) > 0 {
		t.Errorf("routes registered in router but missing from OpenAPI spec:\n  %s",
			strings.Join(onlyInRouter, "\n  "))
	}
	if len(onlyInSpec) > 0 {
		t.Errorf("operations in OpenAPI spec but not registered in router:\n  %s",
			strings.Join(onlyInSpec, "\n  "))
	}
}

// walkRouter returns a sorted set of "METHOD /path" strings for every route
// registered on the given router, minus the operational exceptions.
func walkRouter(t *testing.T, router chi.Router) []string {
	t.Helper()

	seen := map[string]bool{}
	err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		// chi reports trailing "/*" on subrouters and wildcard paths; strip for comparison.
		route = strings.TrimSuffix(route, "/*")
		if excludedPaths[route] {
			return nil
		}
		seen[method+" "+route] = true
		return nil
	})
	require.NoError(t, err)

	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// loadSpecRoutes returns a sorted set of "METHOD /path" strings for every
// operation declared in backend/api/openapi.yaml.
func loadSpecRoutes(t *testing.T) []string {
	t.Helper()

	specPath, err := filepath.Abs(filepath.Join("..", "..", "api", "openapi.yaml"))
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))

	out := []string{}
	for path, item := range doc.Paths.Map() {
		for method := range item.Operations() {
			out = append(out, method+" "+path)
		}
	}
	sort.Strings(out)
	return out
}

// diff returns elements of a that are not in b.
func diff(a, b []string) []string {
	bset := map[string]bool{}
	for _, s := range b {
		bset[s] = true
	}
	var out []string
	for _, s := range a {
		if !bset[s] {
			out = append(out, s)
		}
	}
	return out
}

func TestRouter_NoDebugRoutesInProduction(t *testing.T) {
	t.Parallel()

	router := NewRouter(RouterDeps{
		Auth:       handler.NewAuthHandler(nil, nil, "", false),
		MobileAuth: handler.NewMobileAuthHandler(nil),
		User:       handler.NewUserHandler(nil),
		RoleLookup: func(context.Context, string) (string, error) { return "", nil },
		AppEnv:     "production",
		AppURL:     "https://example.com",
		DevBypass:  false,
	})

	routes := walkRouter(t, router)
	for _, r := range routes {
		assert.NotContains(t, r, "/_debug/", "debug route leaked into production router: %s", r)
	}
}
