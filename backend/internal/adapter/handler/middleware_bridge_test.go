package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
)

// HumaFromHTTP bridges chi-style middleware (httprate, etc.) to Huma's
// per-operation Middlewares slot. Two cases that matter:
//   1. The chi middleware calls through → Huma proceeds to the handler.
//   2. The chi middleware short-circuits (writes a 429, doesn't call
//      next) → Huma stops; the handler never runs.
//
// HumaFromHTTP relies on humachi.Unwrap to access the underlying
// *http.Request, so these tests can't use humatest (which is built on
// humaflow). They build a real chi router + humachi adapter and drive
// it through net/http/httptest.
//
// We exercise both pass-through and short-circuit with synthetic chi
// middlewares so the test doesn't depend on httprate's internal state.

// newBridgeTestAPI builds a chi+humachi API ready for HumaFromHTTP
// middleware to attach to a registered operation.
func newBridgeTestAPI(t *testing.T) (*chi.Mux, huma.API) {
	t.Helper()
	handler.InstallErrorEnvelope()
	r := chi.NewRouter()
	cfg := huma.DefaultConfig("Bridge Test API", "0.0.1")
	cfg.SchemasPath = ""
	cfg.Transformers = nil
	cfg.CreateHooks = nil
	api := humachi.New(r, cfg)
	return r, api
}

func doRequest(t *testing.T, r *chi.Mux, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHumaFromHTTP_PassthroughMiddleware_ReachesHandler(t *testing.T) {
	t.Parallel()

	var handlerCalled atomic.Bool
	var passthroughCalled atomic.Bool

	passthrough := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			passthroughCalled.Store(true)
			next.ServeHTTP(w, r)
		})
	}

	mux, api := newBridgeTestAPI(t)
	huma.Register(api, huma.Operation{
		OperationID: "passthrough-test",
		Method:      http.MethodGet,
		Path:        "/probe",
		Middlewares: huma.Middlewares{handler.HumaFromHTTP(passthrough)},
	}, func(_ context.Context, _ *struct{}) (*struct{}, error) {
		handlerCalled.Store(true)
		return &struct{}{}, nil
	})

	resp := doRequest(t, mux, http.MethodGet, "/probe")

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.True(t, passthroughCalled.Load(), "chi middleware must run")
	assert.True(t, handlerCalled.Load(), "handler must run after middleware passes")
}

func TestHumaFromHTTP_BlockingMiddleware_StopsHandler(t *testing.T) {
	t.Parallel()

	var handlerCalled atomic.Bool

	// Simulates rate-limit exceeded: writes a 429 and does NOT call next.
	blocker := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":"RATE_LIMITED","message":"slow down"}`))
		})
	}

	mux, api := newBridgeTestAPI(t)
	huma.Register(api, huma.Operation{
		OperationID: "blocked-test",
		Method:      http.MethodGet,
		Path:        "/probe",
		Middlewares: huma.Middlewares{handler.HumaFromHTTP(blocker)},
	}, func(_ context.Context, _ *struct{}) (*struct{}, error) {
		handlerCalled.Store(true)
		return &struct{}{}, nil
	})

	resp := doRequest(t, mux, http.MethodGet, "/probe")

	require.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.False(t, handlerCalled.Load(),
		"handler must not run when middleware short-circuits")
	assert.Contains(t, resp.Body.String(), `"code":"RATE_LIMITED"`)
}

// HumaFromHTTP builds the wrapper once at startup. Calling the
// resulting Huma middleware many times must not rebuild it. We verify
// by tracking how many times the chi middleware factory function runs
// — should be exactly once.
func TestHumaFromHTTP_BuildsWrapperOnce(t *testing.T) {
	t.Parallel()

	var factoryCalls atomic.Int32
	factory := func(next http.Handler) http.Handler {
		factoryCalls.Add(1)
		return next
	}

	mux, api := newBridgeTestAPI(t)
	huma.Register(api, huma.Operation{
		OperationID: "wrapper-once-test",
		Method:      http.MethodGet,
		Path:        "/probe",
		Middlewares: huma.Middlewares{handler.HumaFromHTTP(factory)},
	}, func(_ context.Context, _ *struct{}) (*struct{}, error) {
		return &struct{}{}, nil
	})

	for range 5 {
		doRequest(t, mux, http.MethodGet, "/probe")
	}

	assert.Equal(t, int32(1), factoryCalls.Load(),
		"chi middleware factory must be invoked once at setup, not per-request")
}
