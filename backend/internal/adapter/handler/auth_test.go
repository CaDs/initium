package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/infra/google"
	"github.com/eridia/initium/backend/internal/testutil"
)

// AuthHandler is split across two transports:
//   - Huma JSON ops: requestMagicLink, refreshTokens, logout, logoutAll
//     (registered via RegisterAuth). Tested with humatest.
//   - chi-native redirects: GoogleRedirect, GoogleCallback,
//     VerifyMagicLink. Tested with httptest because they emit
//     307 redirects with Set-Cookie that humatest can't faithfully
//     model — they're browser chrome, not REST.
//
// Tests below cover both transports plus the ops endpoints (Healthz).

// newAuthAPI builds a humatest API with the JSON auth routes registered.
// rateLimitMW is no-op so tests don't hit real rate-limit state.
// authMW is also no-op so logout/logoutAll tests can drive userID via
// context directly when needed (they don't here — they verify the
// happy + error paths through the service).
func newAuthAPI(t *testing.T) (humatest.TestAPI, *testutil.MockAuthService) {
	t.Helper()
	handler.InstallErrorEnvelope()
	svc := &testutil.MockAuthService{}
	verifier := google.NewOAuthVerifier("test-client-id", "test-client-secret", "http://localhost/callback")
	h := handler.NewAuthHandler(svc, verifier, "http://localhost:3000", false)
	_, api := humatest.New(t)
	h.RegisterAuth(api, testutil.NoopMiddleware, testutil.NoopMiddleware)
	return api, svc
}

// ----------------------------------------------------------------------
// requestMagicLink
// ----------------------------------------------------------------------

func TestAuthHandler_RequestMagicLink_Success(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	var capturedEmail string
	svc.RequestMagicLinkFn = func(_ context.Context, email string) error {
		capturedEmail = email
		return nil
	}

	resp := api.Post("/api/auth/magic-link", map[string]string{"email": "user@example.com"})

	require.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "user@example.com", capturedEmail)
	body := testutil.MustDecodeJSON[handler.MessageResponse](t, resp.Body)
	assert.Equal(t, "magic link sent", body.Message)
}

func TestAuthHandler_RequestMagicLink_InvalidEmail_422(t *testing.T) {
	t.Parallel()

	api, _ := newAuthAPI(t)

	// Missing required `email` field — Huma validation rejects.
	resp := api.Post("/api/auth/magic-link", map[string]any{})

	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "INVALID_INPUT", env.Code)
}

func TestAuthHandler_RequestMagicLink_ServiceError_MapsDomainErr(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.RequestMagicLinkFn = func(_ context.Context, _ string) error {
		return domain.ErrEmailRequired
	}

	resp := api.Post("/api/auth/magic-link", map[string]string{"email": "user@example.com"})

	require.Equal(t, http.StatusBadRequest, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "EMAIL_REQUIRED", env.Code)
}

// ----------------------------------------------------------------------
// refreshTokens — cookie path, body path, missing, error
// ----------------------------------------------------------------------

func TestAuthHandler_RefreshTokens_CookieToken_Success(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.RefreshTokensFn = func(_ context.Context, token string) (*domain.TokenPair, error) {
		require.Equal(t, "cookie-refresh", token)
		return &testutil.ValidTokenPair, nil
	}

	resp := api.Post("/api/auth/refresh",
		"Cookie: refresh_token=cookie-refresh",
		map[string]any{})

	require.Equal(t, http.StatusOK, resp.Code)
	pair := testutil.MustDecodeJSON[handler.TokenPair](t, resp.Body)
	assert.Equal(t, testutil.ValidTokenPair.AccessToken, pair.AccessToken)

	// Set-Cookie should carry both access_token and refresh_token.
	cookies := resp.Result().Cookies()
	names := make(map[string]string, len(cookies))
	for _, c := range cookies {
		names[c.Name] = c.Value
	}
	assert.Equal(t, testutil.ValidTokenPair.AccessToken, names["access_token"])
	assert.Equal(t, testutil.ValidTokenPair.RefreshToken, names["refresh_token"])
}

func TestAuthHandler_RefreshTokens_BodyToken_Success(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.RefreshTokensFn = func(_ context.Context, token string) (*domain.TokenPair, error) {
		require.Equal(t, "body-refresh", token)
		return &testutil.ValidTokenPair, nil
	}

	resp := api.Post("/api/auth/refresh", map[string]string{"refresh_token": "body-refresh"})

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestAuthHandler_RefreshTokens_CookieWinsOverBody(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.RefreshTokensFn = func(_ context.Context, token string) (*domain.TokenPair, error) {
		require.Equal(t, "cookie-token", token,
			"cookie must take precedence over body refresh_token")
		return &testutil.ValidTokenPair, nil
	}

	resp := api.Post("/api/auth/refresh",
		"Cookie: refresh_token=cookie-token",
		map[string]string{"refresh_token": "body-token"})

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestAuthHandler_RefreshTokens_NoToken_401(t *testing.T) {
	t.Parallel()

	api, _ := newAuthAPI(t)

	resp := api.Post("/api/auth/refresh", map[string]any{})

	require.Equal(t, http.StatusUnauthorized, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "SESSION_NOT_FOUND", env.Code)
}

func TestAuthHandler_RefreshTokens_ServiceError(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.RefreshTokensFn = func(_ context.Context, _ string) (*domain.TokenPair, error) {
		return nil, domain.ErrSessionExpired
	}

	resp := api.Post("/api/auth/refresh", map[string]string{"refresh_token": "expired"})

	require.Equal(t, http.StatusUnauthorized, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "SESSION_EXPIRED", env.Code)
}

// ----------------------------------------------------------------------
// logout / logoutAll
// ----------------------------------------------------------------------

func TestAuthHandler_Logout_Success_ClearsCookies(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	var revokedToken string
	svc.LogoutFn = func(_ context.Context, refreshToken string) error {
		revokedToken = refreshToken
		return nil
	}

	resp := api.Post("/api/auth/logout",
		"Cookie: refresh_token=to-revoke",
		map[string]any{})

	require.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "to-revoke", revokedToken,
		"refresh token from cookie must be passed to Logout")

	// Both clearing cookies must come back.
	cookies := resp.Result().Cookies()
	cleared := make(map[string]bool, len(cookies))
	for _, c := range cookies {
		if c.MaxAge == -1 {
			cleared[c.Name] = true
		}
	}
	assert.True(t, cleared["access_token"], "access_token must be cleared")
	assert.True(t, cleared["refresh_token"], "refresh_token must be cleared")
}

func TestAuthHandler_Logout_NoCookie_StillSucceeds(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	logoutCalled := false
	svc.LogoutFn = func(_ context.Context, _ string) error {
		logoutCalled = true
		return nil
	}

	// No refresh_token cookie — handler should still succeed and clear cookies.
	resp := api.Post("/api/auth/logout", map[string]any{})

	require.Equal(t, http.StatusOK, resp.Code)
	assert.False(t, logoutCalled,
		"Logout(svc) must NOT be called when there's no refresh token to revoke")
}

func TestAuthHandler_LogoutAll_Success(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.LogoutAllFn = func(_ context.Context, _ string) error {
		return nil
	}

	resp := api.Post("/api/auth/logout-all", map[string]any{})

	require.Equal(t, http.StatusOK, resp.Code)
	body := testutil.MustDecodeJSON[handler.MessageResponse](t, resp.Body)
	assert.Equal(t, "all sessions revoked", body.Message)
}

func TestAuthHandler_LogoutAll_ServiceError_500(t *testing.T) {
	t.Parallel()

	api, svc := newAuthAPI(t)
	svc.LogoutAllFn = func(_ context.Context, _ string) error {
		return errors.New("db down")
	}

	resp := api.Post("/api/auth/logout-all", map[string]any{})

	require.Equal(t, http.StatusInternalServerError, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "INTERNAL_ERROR", env.Code)
	assert.Equal(t, "internal error", env.Message,
		"internal error message must not leak details")
}

// ----------------------------------------------------------------------
// chi-native redirects: GoogleRedirect, GoogleCallback, VerifyMagicLink
// ----------------------------------------------------------------------

func newChiAuthHandler() (*handler.AuthHandler, *testutil.MockAuthService) {
	svc := &testutil.MockAuthService{}
	verifier := google.NewOAuthVerifier("test-client-id", "test-client-secret",
		"http://localhost/callback")
	h := handler.NewAuthHandler(svc, verifier, "http://localhost:3000", false)
	return h, svc
}

func TestAuthHandler_GoogleRedirect_SetsStateCookie_AndRedirects(t *testing.T) {
	t.Parallel()

	h, _ := newChiAuthHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google", nil)
	rec := httptest.NewRecorder()

	h.GoogleRedirect(rec, req)

	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	loc := rec.Result().Header.Get("Location")
	assert.Contains(t, loc, "accounts.google.com",
		"redirect target should be Google's consent screen")

	// State cookie must be set with HttpOnly + Lax.
	var stateCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
			break
		}
	}
	require.NotNil(t, stateCookie, "oauth_state cookie must be set")
	assert.NotEmpty(t, stateCookie.Value)
	assert.True(t, stateCookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, stateCookie.SameSite)
}

func TestAuthHandler_GoogleCallback_StateMismatch_401(t *testing.T) {
	t.Parallel()

	h, _ := newChiAuthHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/api/auth/google/callback?state=client-state&code=abc", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "different-state"})
	rec := httptest.NewRecorder()

	h.GoogleCallback(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_CREDENTIALS")
}

func TestAuthHandler_GoogleCallback_MissingStateCookie_401(t *testing.T) {
	t.Parallel()

	h, _ := newChiAuthHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/api/auth/google/callback?state=anything&code=abc", nil)
	rec := httptest.NewRecorder()

	h.GoogleCallback(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_GoogleCallback_LoginError_MapsToHTTP(t *testing.T) {
	t.Parallel()

	h, svc := newChiAuthHandler()
	svc.LoginWithGoogleFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return nil, nil, domain.ErrInvalidOAuthToken
	}
	req := httptest.NewRequest(http.MethodGet,
		"/api/auth/google/callback?state=ok&code=abc", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "ok"})
	rec := httptest.NewRecorder()

	h.GoogleCallback(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_OAUTH_TOKEN")
}

func TestAuthHandler_GoogleCallback_Success_RedirectsToHome(t *testing.T) {
	t.Parallel()

	h, svc := newChiAuthHandler()
	svc.LoginWithGoogleFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return &testutil.RegularUser, &testutil.ValidTokenPair, nil
	}
	req := httptest.NewRequest(http.MethodGet,
		"/api/auth/google/callback?state=ok&code=abc", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "ok"})
	rec := httptest.NewRecorder()

	h.GoogleCallback(rec, req)

	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "http://localhost:3000/home", rec.Result().Header.Get("Location"))

	// Token cookies must be set on success.
	cookies := rec.Result().Cookies()
	got := make(map[string]string, len(cookies))
	for _, c := range cookies {
		got[c.Name] = c.Value
	}
	assert.Equal(t, testutil.ValidTokenPair.AccessToken, got["access_token"])
	assert.Equal(t, testutil.ValidTokenPair.RefreshToken, got["refresh_token"])
}

func TestAuthHandler_VerifyMagicLink_MissingToken_400(t *testing.T) {
	t.Parallel()

	h, _ := newChiAuthHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify", nil)
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "TOKEN_INVALID")
}

func TestAuthHandler_VerifyMagicLink_ServiceError_MapsToHTTP(t *testing.T) {
	t.Parallel()

	h, svc := newChiAuthHandler()
	svc.VerifyMagicLinkFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return nil, nil, domain.ErrTokenExpired
	}
	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify?token=expired", nil)
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "TOKEN_EXPIRED")
}

func TestAuthHandler_VerifyMagicLink_Success_RedirectsToHome(t *testing.T) {
	t.Parallel()

	h, svc := newChiAuthHandler()
	svc.VerifyMagicLinkFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return &testutil.RegularUser, &testutil.ValidTokenPair, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify?token=valid", nil)
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "http://localhost:3000/home", rec.Result().Header.Get("Location"))
}

// ----------------------------------------------------------------------
// Healthz — no dependencies, just shape.
// ----------------------------------------------------------------------

func TestHealthz_ReturnsOK(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.Healthz(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.True(t, strings.Contains(rec.Body.String(), `"status":"ok"`))
}
