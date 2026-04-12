package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockTokenGenerator implements domain.TokenGenerator for testing.
type mockTokenGenerator struct {
	validToken string
	userID     string
	email      string
}

func (m *mockTokenGenerator) GenerateAccessToken(userID, email string) (string, error) {
	return "mock-token", nil
}

func (m *mockTokenGenerator) GenerateRefreshToken() (string, error) {
	return "mock-refresh", nil
}

func (m *mockTokenGenerator) ValidateAccessToken(token string) (string, string, error) {
	if token == m.validToken {
		return m.userID, m.email, nil
	}
	return "", "", http.ErrNoCookie
}

func (m *mockTokenGenerator) HashToken(token string) string {
	return "hashed-" + token
}

func TestAuth_DevBypass_InjectsTestUser(t *testing.T) {
	t.Parallel()

	handler := Auth(nil, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "00000000-0000-0000-0000-000000000001" {
			t.Errorf("expected dev user ID, got %q", userID)
		}
		email, _ := r.Context().Value(EmailKey).(string)
		if email != "dev@initium.local" {
			t.Errorf("expected dev email, got %q", email)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAuth_ValidBearerToken(t *testing.T) {
	t.Parallel()

	tokens := &mockTokenGenerator{validToken: "good-token", userID: "user-1", email: "u@test.com"}

	handler := Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "user-1" {
			t.Errorf("expected user-1, got %q", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAuth_ValidCookieToken(t *testing.T) {
	t.Parallel()

	tokens := &mockTokenGenerator{validToken: "cookie-token", userID: "user-2", email: "c@test.com"}

	handler := Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "user-2" {
			t.Errorf("expected user-2, got %q", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAuth_MissingToken_Returns401(t *testing.T) {
	t.Parallel()

	tokens := &mockTokenGenerator{}
	handler := Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	t.Parallel()

	tokens := &mockTokenGenerator{validToken: "real-token"}
	handler := Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_BearerTakesPriorityOverCookie(t *testing.T) {
	t.Parallel()

	tokens := &mockTokenGenerator{validToken: "bearer-token", userID: "bearer-user", email: "b@test.com"}

	handler := Auth(tokens, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID != "bearer-user" {
			t.Errorf("expected bearer-user, got %q", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
