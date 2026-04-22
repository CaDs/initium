package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// RouteEntry mirrors the OpenAPI RouteEntry schema.
type RouteEntry struct {
	Method  string `json:"method"`
	Pattern string `json:"pattern"`
}

// RouteList mirrors the OpenAPI RouteList schema.
type RouteList struct {
	Routes []RouteEntry `json:"routes"`
}

// RoutesDebug walks the chi router and returns the full route table as JSON.
// Mounted only when APP_ENV != "production" — production wiring must return 404.
// Used by `make routes` for local discovery.
func RoutesDebug(router chi.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entries := make([]RouteEntry, 0, 32)
		walk := func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
			entries = append(entries, RouteEntry{
				Method:  method,
				Pattern: strings.TrimSuffix(route, "/*"),
			})
			return nil
		}
		if err := chi.Walk(router, walk); err != nil {
			Error(w, r, err)
			return
		}
		JSON(w, r, http.StatusOK, RouteList{Routes: entries})
	}
}
