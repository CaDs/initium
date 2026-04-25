package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/infra/google"
)

// AuthHandler handles authentication endpoints. Mixed Huma + chi:
// the JSON-in/JSON-out endpoints (magic link request, refresh,
// logout, logout-all) register through Huma; the OAuth + magic-link
// redirect flows stay chi-native because they're browser chrome
// (307 redirects + Set-Cookie), not REST.
type AuthHandler struct {
	auth          domain.AuthService
	verifier      *google.OAuthVerifier
	appURL        string
	secureCookies bool
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth domain.AuthService, verifier *google.OAuthVerifier, appURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{auth: auth, verifier: verifier, appURL: appURL, secureCookies: secureCookies}
}

// ----------------------------------------------------------------------
// Huma JSON endpoints
// ----------------------------------------------------------------------

type magicLinkInput struct {
	Body struct {
		Email string `json:"email" required:"true" format:"email" doc:"User email"`
	}
}

type messageOutput struct {
	Body MessageResponse
}

type refreshInput struct {
	RefreshTokenCookie string `cookie:"refresh_token" doc:"Refresh token (web cookie path /api/auth)"`
	Body               struct {
		RefreshToken string `json:"refresh_token,omitempty" doc:"Refresh token (mobile JSON body)"`
	}
}

// tokenPairOutput is shared by the refresh + mobile auth endpoints. Marshals
// the existing TokenPair shape; also sets cookies via header tags so the
// web flow keeps working unchanged after migration.
type tokenPairOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      TokenPair
}

// RegisterAuth wires the JSON auth endpoints onto the Huma API.
func (h *AuthHandler) RegisterAuth(
	api huma.API,
	authMW func(huma.Context, func(huma.Context)),
	rateLimitMW func(huma.Context, func(huma.Context)),
) {
	huma.Register(api, huma.Operation{
		OperationID: "request-magic-link",
		Method:      http.MethodPost,
		Path:        "/api/auth/magic-link",
		Summary:     "Email a one-use magic link",
		Tags:        []string{"auth"},
		Middlewares: huma.Middlewares{rateLimitMW},
	}, h.requestMagicLink)

	huma.Register(api, huma.Operation{
		OperationID: "refresh-tokens",
		Method:      http.MethodPost,
		Path:        "/api/auth/refresh",
		Summary:     "Exchange a refresh token for a new access + refresh pair",
		Tags:        []string{"auth"},
		Middlewares: huma.Middlewares{rateLimitMW},
	}, h.refreshTokens)

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/api/auth/logout",
		Summary:     "Revoke the current session",
		Tags:        []string{"auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{rateLimitMW, authMW},
	}, h.logout)

	huma.Register(api, huma.Operation{
		OperationID: "logout-all",
		Method:      http.MethodPost,
		Path:        "/api/auth/logout-all",
		Summary:     "Revoke every session for the current user",
		Tags:        []string{"auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{rateLimitMW, authMW},
	}, h.logoutAll)
}

func (h *AuthHandler) requestMagicLink(ctx context.Context, in *magicLinkInput) (*messageOutput, error) {
	if err := h.auth.RequestMagicLink(ctx, in.Body.Email); err != nil {
		slog.Error("magic link request failed", "error", err)
		return nil, MapDomainErr(ctx, err)
	}
	return &messageOutput{Body: MessageResponse{Message: "magic link sent"}}, nil
}

func (h *AuthHandler) refreshTokens(ctx context.Context, in *refreshInput) (*tokenPairOutput, error) {
	refreshToken := in.RefreshTokenCookie
	if refreshToken == "" {
		refreshToken = in.Body.RefreshToken
	}
	if refreshToken == "" {
		return nil, MapDomainErr(ctx, domain.ErrSessionNotFound)
	}

	pair, err := h.auth.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return nil, MapDomainErr(ctx, err)
	}

	return &tokenPairOutput{
		SetCookie: h.tokenCookies(pair),
		Body:      TokenPair{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken},
	}, nil
}

type logoutInput struct {
	RefreshTokenCookie string `cookie:"refresh_token" doc:"Refresh token to revoke (web cookie)"`
}

func (h *AuthHandler) logout(ctx context.Context, in *logoutInput) (*logoutOutput, error) {
	// Auth middleware already validated the access token; we use the
	// refresh-token cookie purely to revoke the matching session.
	if in.RefreshTokenCookie != "" {
		if err := h.auth.Logout(ctx, in.RefreshTokenCookie); err != nil {
			slog.Error("logout failed", "error", err)
		}
	}
	return &logoutOutput{
		SetCookie: clearTokenCookies(),
		Body:      MessageResponse{Message: "logged out"},
	}, nil
}

func (h *AuthHandler) logoutAll(ctx context.Context, _ *struct{}) (*logoutOutput, error) {
	userID := middleware.GetUserID(ctx)
	if err := h.auth.LogoutAll(ctx, userID); err != nil {
		slog.Error("logout all failed", "error", err)
		return nil, MapDomainErr(ctx, err)
	}
	return &logoutOutput{
		SetCookie: clearTokenCookies(),
		Body:      MessageResponse{Message: "all sessions revoked"},
	}, nil
}

type logoutOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      MessageResponse
}

// ----------------------------------------------------------------------
// chi-native redirect handlers — stay as http.HandlerFunc
// ----------------------------------------------------------------------

// GoogleRedirect redirects to Google's consent screen.
func (h *AuthHandler) GoogleRedirect(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.Error("failed to generate oauth state", "error", err)
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"internal error"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"code":"INVALID_CREDENTIALS","message":"invalid oauth state"}`, http.StatusUnauthorized)
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
		writeChiError(w, err)
		return
	}

	for _, c := range h.tokenCookies(pair) {
		http.SetCookie(w, &c)
	}
	slog.Info("user logged in via google", "user_id", user.ID)
	http.Redirect(w, r, h.appURL+"/home", http.StatusTemporaryRedirect)
}

// VerifyMagicLink validates a magic link token and sets session cookies (web flow).
func (h *AuthHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"code":"TOKEN_INVALID","message":"missing token"}`, http.StatusBadRequest)
		return
	}

	user, pair, err := h.auth.VerifyMagicLink(r.Context(), token)
	if err != nil {
		slog.Error("magic link verification failed", "error", err)
		writeChiError(w, err)
		return
	}

	for _, c := range h.tokenCookies(pair) {
		http.SetCookie(w, &c)
	}
	slog.Info("user logged in via magic link", "user_id", user.ID)
	http.Redirect(w, r, h.appURL+"/home", http.StatusTemporaryRedirect)
}

// ----------------------------------------------------------------------
// Cookie helpers — shared by Huma + chi-native handlers
// ----------------------------------------------------------------------

func (h *AuthHandler) tokenCookies(pair *domain.TokenPair) []http.Cookie {
	return []http.Cookie{
		{
			Name:     "access_token",
			Value:    pair.AccessToken,
			Path:     "/",
			MaxAge:   900,
			HttpOnly: true,
			Secure:   h.secureCookies,
			SameSite: http.SameSiteLaxMode,
		},
		{
			Name:     "refresh_token",
			Value:    pair.RefreshToken,
			Path:     "/api/auth",
			MaxAge:   604800,
			HttpOnly: true,
			Secure:   h.secureCookies,
			SameSite: http.SameSiteLaxMode,
		},
	}
}

// clearedCookieJar is a precomputed slice returned by clearTokenCookies.
// The values are static (empty token, MaxAge=-1) so we share one slice
// rather than allocating per logout / refresh-failure call. Callers must
// not mutate the returned slice.
var clearedCookieJar = []http.Cookie{
	{Name: "access_token", Value: "", Path: "/", MaxAge: -1},
	{Name: "refresh_token", Value: "", Path: "/api/auth", MaxAge: -1},
}

func clearTokenCookies() []http.Cookie {
	return clearedCookieJar
}

func writeChiError(w http.ResponseWriter, err error) {
	code, status := mapError(err)
	msg := err.Error()
	if code == "INTERNAL_ERROR" {
		msg = "internal error"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"code":%q,"message":%q}`, code, msg)
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ----------------------------------------------------------------------
// chi-native ops handlers (Healthz / Readyz). Stay here per plan —
// not user-facing API contracts, no benefit from Huma's typed pattern.
// ----------------------------------------------------------------------

// Healthz returns a simple liveness check (no dependencies).
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// Readyz returns a readiness check that verifies DB connectivity.
// Returns 200 {"status":"ok"} on success, 503 {"status":"unready","error":"..."} on failure.
func Readyz(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		sqlDB, err := db.DB()
		if err != nil {
			slog.Error("readyz: failed to get sql.DB", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unready","error":"database unavailable"}`))
			return
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			slog.Error("readyz: database ping failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"unready","error":%q}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
