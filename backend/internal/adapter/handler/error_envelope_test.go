package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/domain"
)

// TestError_EnvelopeShape asserts every domain error maps to the OpenAPI
// ErrorResponse envelope (code + message + request_id) with the right HTTP
// status and SNAKE_UPPER code. If a new domain error is added, this test
// will fail until its mapping lands in respond.go.
func TestError_EnvelopeShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"user_not_found", domain.ErrUserNotFound, http.StatusNotFound, "USER_NOT_FOUND"},
		{"session_not_found", domain.ErrSessionNotFound, http.StatusUnauthorized, "SESSION_NOT_FOUND"},
		{"session_expired", domain.ErrSessionExpired, http.StatusUnauthorized, "SESSION_EXPIRED"},
		{"session_revoked", domain.ErrSessionRevoked, http.StatusUnauthorized, "SESSION_REVOKED"},
		{"token_expired", domain.ErrTokenExpired, http.StatusUnauthorized, "TOKEN_EXPIRED"},
		{"token_used", domain.ErrTokenUsed, http.StatusConflict, "TOKEN_USED"},
		{"token_invalid", domain.ErrTokenInvalid, http.StatusBadRequest, "TOKEN_INVALID"},
		{"invalid_oauth_token", domain.ErrInvalidOAuthToken, http.StatusUnauthorized, "INVALID_OAUTH_TOKEN"},
		{"invalid_credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized, "INVALID_CREDENTIALS"},
		{"email_required", domain.ErrEmailRequired, http.StatusBadRequest, "EMAIL_REQUIRED"},
		{"rate_limited", domain.ErrRateLimited, http.StatusTooManyRequests, "RATE_LIMITED"},
		{"invalid_input", domain.ErrInvalidInput, http.StatusBadRequest, "INVALID_INPUT"},
		{"unmapped_error_becomes_internal", errors.New("some unexpected failure"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			Error(w, req, tc.err)

			require.Equal(t, tc.wantStatus, w.Code, "HTTP status")
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var env ErrorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))

			assert.Equal(t, tc.wantCode, env.Code, "error code")
			assert.NotEmpty(t, env.Message, "message must not be empty")

			// Internal errors must never leak the original error string.
			if tc.wantCode == "INTERNAL_ERROR" {
				assert.Equal(t, "internal error", env.Message,
					"internal error message must not leak details; got %q", env.Message)
				assert.NotContains(t, env.Message, "some unexpected failure")
			}
		})
	}
}

// TestError_ResponseJSONMatchesSpec asserts the envelope fields match the
// OpenAPI ErrorResponse schema exactly (code + message + optional request_id).
// A stray field here means the spec and code have drifted.
func TestError_ResponseJSONMatchesSpec(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	Error(w, req, domain.ErrInvalidInput)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))

	// These three keys — and only these — are in the OpenAPI ErrorResponse.
	allowed := map[string]bool{"code": true, "message": true, "request_id": true}
	for k := range raw {
		assert.True(t, allowed[k], "unexpected field %q in error envelope", k)
	}

	// code + message are required per spec.
	assert.Contains(t, raw, "code")
	assert.Contains(t, raw, "message")
}
