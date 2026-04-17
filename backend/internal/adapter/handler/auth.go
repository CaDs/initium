package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/infra/google"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	auth       domain.AuthService
	verifier   *google.OAuthVerifier
	appURL     string
	secureCookies bool
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth domain.AuthService, verifier *google.OAuthVerifier, appURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{auth: auth, verifier: verifier, appURL: appURL, secureCookies: secureCookies}
}

// GoogleRedirect redirects to Google's consent screen.
func (h *AuthHandler) GoogleRedirect(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.Error("failed to generate oauth state", "error", err)
		Error(w, r, fmt.Errorf("internal error"))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.verifier.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback.
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(r.URL.Query().Get("state"))) != 1 {
		Error(w, r, domain.ErrInvalidCredentials)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	code := r.URL.Query().Get("code")
	user, pair, err := h.auth.LoginWithGoogle(r.Context(), code)
	if err != nil {
		slog.Error("google login failed", "error", err)
		Error(w, r, err)
		return
	}

	h.setTokenCookies(w, pair)
	slog.Info("user logged in via google", "user_id", user.ID)
	http.Redirect(w, r, h.appURL+"/home", http.StatusTemporaryRedirect)
}

// RequestMagicLink sends a magic link email.
func (h *AuthHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, r, domain.ErrEmailRequired)
		return
	}

	if err := h.auth.RequestMagicLink(r.Context(), body.Email); err != nil {
		slog.Error("magic link request failed", "error", err)
		Error(w, r, err)
		return
	}

	JSON(w, r, http.StatusOK, map[string]string{"message": "magic link sent"})
}

// VerifyMagicLink validates a magic link token and sets session cookies.
func (h *AuthHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		Error(w, r, domain.ErrTokenInvalid)
		return
	}

	user, pair, err := h.auth.VerifyMagicLink(r.Context(), token)
	if err != nil {
		slog.Error("magic link verification failed", "error", err)
		Error(w, r, err)
		return
	}

	h.setTokenCookies(w, pair)
	slog.Info("user logged in via magic link", "user_id", user.ID)
	http.Redirect(w, r, h.appURL+"/home", http.StatusTemporaryRedirect)
}

// RefreshTokens issues a new token pair using a refresh token.
func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if c, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = c.Value
	}

	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			refreshToken = body.RefreshToken
		}
	}

	if refreshToken == "" {
		Error(w, r, domain.ErrSessionNotFound)
		return
	}

	pair, err := h.auth.RefreshTokens(r.Context(), refreshToken)
	if err != nil {
		Error(w, r, err)
		return
	}

	h.setTokenCookies(w, pair)
	JSON(w, r, http.StatusOK, map[string]string{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}

// Logout revokes the current session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if c, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = c.Value
	}

	if refreshToken != "" {
		if err := h.auth.Logout(r.Context(), refreshToken); err != nil {
			slog.Error("logout failed", "error", err)
		}
	}

	clearTokenCookies(w)
	JSON(w, r, http.StatusOK, map[string]string{"message": "logged out"})
}

// LogoutAll revokes all sessions for the current user.
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.auth.LogoutAll(r.Context(), userID); err != nil {
		slog.Error("logout all failed", "error", err)
		Error(w, r, err)
		return
	}

	clearTokenCookies(w)
	JSON(w, r, http.StatusOK, map[string]string{"message": "all sessions revoked"})
}

func (h *AuthHandler) setTokenCookies(w http.ResponseWriter, pair *domain.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    pair.AccessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		Path:     "/api/auth",
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/api/auth", MaxAge: -1})
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Healthz returns a simple liveness check (no dependencies).
func Healthz(w http.ResponseWriter, r *http.Request) {
	JSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz returns a readiness check that verifies DB connectivity.
// Returns 200 {"status":"ok"} on success, 503 {"status":"unready","error":"..."} on failure.
func Readyz(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := db.DB()
		if err != nil {
			slog.Error("readyz: failed to get sql.DB", "error", err)
			JSON(w, r, http.StatusServiceUnavailable, map[string]string{
				"status": "unready",
				"error":  "database unavailable",
			})
			return
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			slog.Error("readyz: database ping failed", "error", err)
			JSON(w, r, http.StatusServiceUnavailable, map[string]string{
				"status": "unready",
				"error":  err.Error(),
			})
			return
		}

		JSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	}
}
