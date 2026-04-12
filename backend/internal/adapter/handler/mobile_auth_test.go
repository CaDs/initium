package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eridia/initium/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthService implements domain.AuthService for handler tests.
type mockAuthService struct {
	verifyMagicLinkFn func(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error)
	verifyGoogleIDFn  func(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error)
}

func (m *mockAuthService) LoginWithGoogle(ctx context.Context, code string) (*domain.User, *domain.TokenPair, error) {
	return nil, nil, nil
}

func (m *mockAuthService) VerifyGoogleIDToken(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error) {
	if m.verifyGoogleIDFn != nil {
		return m.verifyGoogleIDFn(ctx, idToken)
	}
	return nil, nil, nil
}

func (m *mockAuthService) RequestMagicLink(ctx context.Context, email string) error {
	return nil
}

func (m *mockAuthService) VerifyMagicLink(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error) {
	if m.verifyMagicLinkFn != nil {
		return m.verifyMagicLinkFn(ctx, token)
	}
	return nil, nil, nil
}

func (m *mockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

func (m *mockAuthService) LogoutAll(ctx context.Context, userID string) error {
	return nil
}

func TestMobileAuthHandler_VerifyMagicLink_ValidToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyMagicLinkFn: func(_ context.Context, token string) (*domain.User, *domain.TokenPair, error) {
			return &domain.User{ID: "user-1", Email: "test@example.com"},
				&domain.TokenPair{AccessToken: "access-123", RefreshToken: "refresh-456"},
				nil
		},
	}
	h := NewMobileAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"token": "valid-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "access-123", resp["access_token"])
	assert.Equal(t, "refresh-456", resp["refresh_token"])
}

func TestMobileAuthHandler_VerifyMagicLink_MissingToken(t *testing.T) {
	t.Parallel()

	h := NewMobileAuthHandler(&mockAuthService{})

	body, _ := json.Marshal(map[string]string{"token": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "TOKEN_INVALID", resp.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_EmptyBody(t *testing.T) {
	t.Parallel()

	h := NewMobileAuthHandler(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/verify", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_ExpiredToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyMagicLinkFn: func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
			return nil, nil, domain.ErrTokenExpired
		},
	}
	h := NewMobileAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"token": "expired-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "TOKEN_EXPIRED", resp.Code)
}

func TestMobileAuthHandler_VerifyMagicLink_UsedToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyMagicLinkFn: func(_ context.Context, _ string) (*domain.User, *domain.TokenPair, error) {
			return nil, nil, domain.ErrTokenUsed
		},
	}
	h := NewMobileAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"token": "used-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.VerifyMagicLink(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "TOKEN_USED", resp.Code)
}

func TestMobileAuthHandler_GoogleIDToken_ValidToken(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		verifyGoogleIDFn: func(_ context.Context, idToken string) (*domain.User, *domain.TokenPair, error) {
			return &domain.User{ID: "user-2", Email: "google@example.com"},
				&domain.TokenPair{AccessToken: "access-g", RefreshToken: "refresh-g"},
				nil
		},
	}
	h := NewMobileAuthHandler(svc)

	body, _ := json.Marshal(map[string]string{"id_token": "valid-google-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/google", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.GoogleIDToken(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "access-g", resp["access_token"])
	assert.Equal(t, "refresh-g", resp["refresh_token"])
}
