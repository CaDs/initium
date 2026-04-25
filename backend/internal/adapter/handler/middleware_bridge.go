package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
)

// HumaFromHTTP adapts a stdlib `func(http.Handler) http.Handler` middleware
// (chi / httprate / standard library) into a Huma middleware. Used to apply
// chi middleware (e.g. httprate.LimitByIP) to specific Huma operations
// without forcing every operation onto a chi sub-router.
//
// Mechanism: humachi.Unwrap exposes the underlying *http.Request and
// http.ResponseWriter. We wrap a sentinel handler that records whether
// the chi middleware called through. If it did, we proceed to next(ctx).
// If the middleware short-circuited (e.g. wrote a 429 + returned), we
// don't proceed — Huma's response is whatever the chi middleware wrote.
//
// Caveat: any response the chi middleware writes (rate-limit error body,
// CORS preflight) skips Huma's error-envelope formatter. Acceptable for
// rate-limit-style middleware where the body shape is well-known.
func HumaFromHTTP(mw func(http.Handler) http.Handler) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		w, r := unwrap(ctx)
		passed := false
		mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			passed = true
		})).ServeHTTP(w, r)
		if passed {
			next(ctx)
		}
	}
}

// unwrap returns the underlying writer + request from a humachi context.
// Wraps humachi.Unwrap to flip the return order so it reads more naturally
// at call sites (writer first, matching http.HandlerFunc convention).
func unwrap(ctx huma.Context) (http.ResponseWriter, *http.Request) {
	r, w := humachi.Unwrap(ctx)
	return w, r
}
