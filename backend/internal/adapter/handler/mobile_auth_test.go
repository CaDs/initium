package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/domain"
)

// mockAuthService implements domain.AuthService for handler tests.
type mockAuthService struct {
	verifyMagicLinkFn func(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error)
	verifyGoogleIDFn  func(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error)
}

func (m *mockAuthService) LoginWithGoogle(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
	return nil, nil, nil
}

func (m *mockAuthService) VerifyGoogleIDToken(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error) {
	if m.verifyGoogleIDFn != nil {
		return m.verifyGoogleIDFn(ctx, idToken)
	}
	return nil, nil, nil
}

func (m *mockAuthService) RequestMagicLink(_ context.Context, _ string) error {
	return nil
}

func (m *mockAuthService) VerifyMagicLink(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error) {
	if m.verifyMagicLinkFn != nil {
		return m.verifyMagicLinkFn(ctx, token)
	}
	return nil, nil, nil
}

func (m *mockAuthService) RefreshTokens(_ context.Context, _ string) (*domain.TokenPair, error) {
	return nil, nil
}

func (m *mockAuthService) Logout(_ context.Context, _ string) error {
	return nil
}

func (m *mockAuthService) LogoutAll(_ context.Context, _ string) error {
	return nil
}

// noopMW is a Huma middleware that just calls next — used as the
// rate-limit slot in tests so we don't need an httprate instance.
func noopMW(ctx huma.Context, next func(huma.Context)) { next(ctx) }

// newMobileAuthTestAPI builds a humatest API with the mobile auth handler
// registered. Returns the test API for issuing requests.
func newMobileAuthTestAPI(t *testing.T, svc domain.AuthService) humatest.TestAPI {
	t.Helper()
	InstallErrorEnvelope()
	_, api := humatest.New(t)
	h := NewMobileAuthHandler(svc)
	h.RegisterMobileAuth(api, noopMW)
	return api
}

func TestMobileAuthHandler_VerifyMagicLink_ValidToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyMagicLinkFn: func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
			return &domain.User{ID: "user-1", Email: "test@example.com"},
				&domain.TokenPair{AccessToken: "access-123", RefreshToken: "refresh-456"},
				nil
		},
	}
	api := newMobileAuthTestAPI(t, svc)

	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": "valid-token"})

	require.Equal(t, http.StatusOK, resp.Code)

	var pair TokenPair
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pair))
	assert.Equal(t, "access-123", pair.AccessToken)
	assert.Equal(t, "refresh-456", pair.RefreshToken)
}

func TestMobileAuthHandler_VerifyMagicLink_MissingToken(t *testing.T) {
	t.Parallel()

	api := newMobileAuthTestAPI(t, &mockAuthService{})

	// Empty token — Huma's validation tag (required + minLength=1) rejects.
	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": ""})

	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_EmptyBody(t *testing.T) {
	t.Parallel()

	api := newMobileAuthTestAPI(t, &mockAuthService{})

	resp := api.Post("/api/auth/mobile/verify", map[string]any{})

	// Missing required field "token" → Huma 422.
	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_ExpiredToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyMagicLinkFn: func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
			return nil, nil, domain.ErrTokenExpired
		},
	}
	api := newMobileAuthTestAPI(t, svc)

	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": "expired-token"})

	require.Equal(t, http.StatusUnauthorized, resp.Code)

	var env APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Equal(t, "TOKEN_EXPIRED", env.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_UsedToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyMagicLinkFn: func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
			return nil, nil, domain.ErrTokenUsed
		},
	}
	api := newMobileAuthTestAPI(t, svc)

	resp := api.Post("/api/auth/mobile/verify", map[string]string{"token": "used-token"})

	require.Equal(t, http.StatusConflict, resp.Code)

	var env APIError
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Equal(t, "TOKEN_USED", env.Code)
}

func TestMobileAuthHandler_GoogleIDToken_ValidToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyGoogleIDFn: func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
			return &domain.User{ID: "user-2", Email: "google@example.com"},
				&domain.TokenPair{AccessToken: "access-g", RefreshToken: "refresh-g"},
				nil
		},
	}
	api := newMobileAuthTestAPI(t, svc)

	resp := api.Post("/api/auth/mobile/google", map[string]string{"id_token": "valid-google-token"})

	require.Equal(t, http.StatusOK, resp.Code)

	var pair TokenPair
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pair))
	assert.Equal(t, "access-g", pair.AccessToken)
	assert.Equal(t, "refresh-g", pair.RefreshToken)
}
