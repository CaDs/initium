package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
)

// RequestID covers two contracts that downstream code depends on:
//   1. The ID is generated AND stored in context (handlers + slog
//      access log retrieve it via GetRequestID).
//   2. The same ID is mirrored to the X-Request-ID response header so
//      clients can quote it back when reporting bugs.
//
// If either contract regresses, MapDomainErr stops propagating
// request_id to the wire envelope — silent loss of correlation.

func TestRequestID_GeneratesUniqueIDPerRequest(t *testing.T) {
	t.Parallel()

	var captured []string
	chained := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = append(captured, middleware.GetRequestID(r.Context()))
	}))

	for range 3 {
		chained.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	}

	require.Len(t, captured, 3)
	for _, id := range captured {
		assert.NotEmpty(t, id, "request ID must be set in context")
	}
	// All three IDs must be different.
	assert.NotEqual(t, captured[0], captured[1])
	assert.NotEqual(t, captured[1], captured[2])
	assert.NotEqual(t, captured[0], captured[2])
}

func TestRequestID_MirrorsIDToResponseHeader(t *testing.T) {
	t.Parallel()

	var ctxID string
	chained := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctxID = middleware.GetRequestID(r.Context())
	}))

	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	headerID := rec.Header().Get("X-Request-ID")
	require.NotEmpty(t, headerID, "X-Request-ID response header must be set")
	assert.Equal(t, ctxID, headerID,
		"context ID and response header must match — clients quote the header when reporting bugs")
}

func TestGetRequestID_NoIDInContext_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	id := middleware.GetRequestID(httptest.NewRequest(http.MethodGet, "/x", nil).Context())
	assert.Empty(t, id, "GetRequestID must return empty string when context has no ID")
}
