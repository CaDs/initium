package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/testutil"
)

// auth_test.go covers the chi-native middleware that protects the
// remaining browser-chrome endpoints (Google callback flow, magic-link
// verify). The Huma-side equivalent lives in adapter/handler — this
// test file is the safety net for the chi pipeline.

// okHandler200 is the dummy "next" handler used by tests that expect
// the middleware to call through. Tests that expect a 4xx short-circuit
// use a t.Fatal handler so an unexpected pass-through fails loudly.
func okHandler200() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// ----------------------------------------------------------------------
// Auth middleware
// ----------------------------------------------------------------------

func TestAuth_DevBypass_InjectsTestUser(t *testing.T) {
	t.Parallel()

	var gotUserID, gotEmail string
	chained := middleware.Auth(nil, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = middleware.GetUserID(r.Context())
		gotEmail, _ = r.Context().Value(middleware.EmailKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", gotUserID)
	assert.Equal(t, "dev@initium.local", gotEmail)
}

func TestAuth_ValidBearerToken_AuthenticatesUser(t *testing.T) {
	t.Parallel()

	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(token string) (string, string, error) {
			require.Equal(t, "good-token", token)
			return "user-1", "u@test.com", nil
		},
	}
	var gotUserID string
	chained := middleware.Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-1", gotUserID)
}

func TestAuth_ValidCookieToken_AuthenticatesUser(t *testing.T) {
	t.Parallel()

	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(token string) (string, string, error) {
			require.Equal(t, "cookie-token", token)
			return "user-2", "c@test.com", nil
		},
	}
	chained := middleware.Auth(tokens, false)(okHandler200())

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_MissingToken_Returns401(t *testing.T) {
	t.Parallel()

	chained := middleware.Auth(&testutil.MockTokenGenerator{}, false)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler must not be called when token is missing")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	t.Parallel()

	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(_ string) (string, string, error) {
			return "", "", errors.New("signature invalid")
		},
	}
	chained := middleware.Auth(tokens, false)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler must not be called when token is invalid")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_BearerOverridesCookie(t *testing.T) {
	t.Parallel()

	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(token string) (string, string, error) {
			require.Equal(t, "bearer-token", token, "Authorization header must take precedence")
			return "bearer-user", "b@test.com", nil
		},
	}
	var gotUserID string
	chained := middleware.Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "bearer-user", gotUserID)
}

// ----------------------------------------------------------------------
// RequireRole middleware
// ----------------------------------------------------------------------

func TestRequireRole_AdminRole_AllowsThrough(t *testing.T) {
	t.Parallel()

	var lookupCalledFor string
	roleMW := middleware.RequireRole("admin", func(_ context.Context, userID string) (string, error) {
		lookupCalledFor = userID
		return "admin", nil
	})
	chained := roleMW(okHandler200())

	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, testutil.AdminUser.ID)
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req.WithContext(ctx))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, testutil.AdminUser.ID, lookupCalledFor)
}

func TestRequireRole_WrongRole_Returns403(t *testing.T) {
	t.Parallel()

	roleMW := middleware.RequireRole("admin", func(_ context.Context, _ string) (string, error) {
		return "user", nil
	})
	chained := roleMW(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler must not be called when role doesn't match")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, testutil.RegularUser.ID)
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req.WithContext(ctx))

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "FORBIDDEN")
}

func TestRequireRole_LookupError_Returns500(t *testing.T) {
	t.Parallel()

	roleMW := middleware.RequireRole("admin", func(_ context.Context, _ string) (string, error) {
		return "", errors.New("database down")
	})
	chained := roleMW(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler must not be called when role lookup fails")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "u")
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req.WithContext(ctx))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "INTERNAL_ERROR")
}

func TestRequireRole_MissingUserID_Returns401(t *testing.T) {
	t.Parallel()

	roleMW := middleware.RequireRole("admin", func(_ context.Context, _ string) (string, error) {
		t.Fatal("lookup must not be called when userID is missing")
		return "", nil
	})
	chained := roleMW(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler must not be called when userID is missing")
	}))

	// No userID in context — middleware should 401 before invoking lookup.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/ping", nil)
	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
