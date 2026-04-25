package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
)

// HumaFromHTTP adapts a stdlib `func(http.Handler) http.Handler` middleware
// (chi / httprate / standard library) into a Huma middleware. Used to apply
// chi middleware (e.g. httprate.LimitByIP) to specific Huma operations
// without forcing every operation onto a chi sub-router.
//
// Mechanism: the chi middleware factory `mw` is invoked ONCE at setup and
// produces a wrapper handler that internally calls our sentinel. The
// sentinel records "the middleware called through" by flipping a *bool
// stored in the per-request request context. Huma's per-request work is
// then a single context.WithValue + r.WithContext + the chi wrapper's
// own logic — no per-request middleware-chain rebuild.
//
// Caveat: any response the chi middleware writes (rate-limit error body,
// CORS preflight) skips Huma's error-envelope formatter. Acceptable for
// rate-limit-style middleware where the body shape is well-known.
func HumaFromHTTP(mw func(http.Handler) http.Handler) func(huma.Context, func(huma.Context)) {
	sentinel := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if p, ok := r.Context().Value(passedKey{}).(*bool); ok {
			*p = true
		}
	})
	wrapped := mw(sentinel) // built ONCE at startup

	return func(ctx huma.Context, next func(huma.Context)) {
		w, r := unwrap(ctx)
		passed := false
		r = r.WithContext(context.WithValue(r.Context(), passedKey{}, &passed))
		wrapped.ServeHTTP(w, r)
		if passed {
			next(ctx)
		}
	}
}

// passedKey is the typed context key used by HumaFromHTTP to thread the
// "middleware called through" signal from the sentinel back to the Huma
// middleware. Defined as a struct (not a string) so it never collides with
// other context keys.
type passedKey struct{}

// unwrap returns the underlying writer + request from a humachi context.
// Wraps humachi.Unwrap to flip the return order so it reads more naturally
// at call sites (writer first, matching http.HandlerFunc convention).
func unwrap(ctx huma.Context) (http.ResponseWriter, *http.Request) {
	r, w := humachi.Unwrap(ctx)
	return w, r
}
