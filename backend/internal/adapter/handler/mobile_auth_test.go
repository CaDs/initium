package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/testutil"
)

// newMobileAuthAPI builds a humatest API with only the mobile-auth
// routes registered. Returns the test API + a pointer to the mock
// service so each test can rewire its Fn fields without rebuilding.
func newMobileAuthAPI(t *testing.T) (humatest.TestAPI, *testutil.MockAuthService) {
	t.Helper()
	handler.InstallErrorEnvelope()
	svc := &testutil.MockAuthService{}
	_, api := humatest.New(t)
	handler.NewMobileAuthHandler(svc).RegisterMobileAuth(api, testutil.NoopMiddleware)
	return api, svc
}

func TestMobileAuthHandler_VerifyMagicLink_ValidToken(t *testing.T) {
	t.Parallel()

	api, svc := newMobileAuthAPI(t)
	svc.VerifyMagicLinkFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return &testutil.RegularUser, &testutil.ValidTokenPair, nil
	}

	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": "valid-token"})

	require.Equal(t, http.StatusOK, resp.Code)

	pair := testutil.MustDecodeJSON[handler.TokenPair](t, resp.Body)
	assert.Equal(t, testutil.ValidTokenPair.AccessToken, pair.AccessToken)
	assert.Equal(t, testutil.ValidTokenPair.RefreshToken, pair.RefreshToken)
}

func TestMobileAuthHandler_VerifyMagicLink_MissingToken(t *testing.T) {
	t.Parallel()

	api, _ := newMobileAuthAPI(t)

	// Empty token — Huma's validation tag (required + minLength=1) rejects.
	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": ""})

	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_EmptyBody(t *testing.T) {
	t.Parallel()

	api, _ := newMobileAuthAPI(t)

	resp := api.Post("/api/auth/mobile/verify", map[string]any{})

	// Missing required field "token" → Huma 422.
	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_ExpiredToken(t *testing.T) {
	t.Parallel()

	api, svc := newMobileAuthAPI(t)
	svc.VerifyMagicLinkFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return nil, nil, domain.ErrTokenExpired
	}

	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": "expired-token"})

	require.Equal(t, http.StatusUnauthorized, resp.Code)

	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "TOKEN_EXPIRED", env.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_UsedToken(t *testing.T) {
	t.Parallel()

	api, svc := newMobileAuthAPI(t)
	svc.VerifyMagicLinkFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return nil, nil, domain.ErrTokenUsed
	}

	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": "used-token"})

	require.Equal(t, http.StatusConflict, resp.Code)

	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "TOKEN_USED", env.Code)
}

func TestMobileAuthHandler_GoogleIDToken_ValidToken(t *testing.T) {
	t.Parallel()

	api, svc := newMobileAuthAPI(t)
	svc.VerifyGoogleIDTokenFn = func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
		return &testutil.RegularUser, &testutil.ValidTokenPair, nil
	}

	resp := api.Post("/api/auth/mobile/google", map[string]string{"id_token": "valid-google-token"})

	require.Equal(t, http.StatusOK, resp.Code)

	pair := testutil.MustDecodeJSON[handler.TokenPair](t, resp.Body)
	assert.Equal(t, testutil.ValidTokenPair.AccessToken, pair.AccessToken)
	assert.Equal(t, testutil.ValidTokenPair.RefreshToken, pair.RefreshToken)
}
