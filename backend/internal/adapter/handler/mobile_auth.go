package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/eridia/initium/backend/internal/domain"
)

// MobileAuthHandler handles mobile-specific authentication. Both endpoints
// are JSON-in / JSON-out, registered through Huma. The wire shape mirrors
// the iOS / Android hand-written DTOs (Models.swift / Models.kt).
type MobileAuthHandler struct {
	auth domain.AuthService
}

// NewMobileAuthHandler creates a new MobileAuthHandler.
func NewMobileAuthHandler(auth domain.AuthService) *MobileAuthHandler {
	return &MobileAuthHandler{auth: auth}
}

type mobileGoogleInput struct {
	Body struct {
		IDToken string `json:"id_token" required:"true" minLength:"1" doc:"Google ID token from the mobile SDK"`
	}
}

type mobileVerifyInput struct {
	Body struct {
		Token string `json:"token" required:"true" minLength:"1" doc:"Magic-link token from the deep link"`
	}
}

// mobileTokenOutput wraps the TokenPair without setting cookies — mobile
// clients store tokens via Keychain / EncryptedSharedPreferences, not
// cookies. The web flows that DO set cookies live on AuthHandler.
type mobileTokenOutput struct {
	Body TokenPair
}

// RegisterMobileAuth wires the mobile-specific auth endpoints onto the
// Huma API. Both inherit the shared rate limiter applied to /api/auth/*.
func (h *MobileAuthHandler) RegisterMobileAuth(
	api huma.API,
	rateLimitMW func(huma.Context, func(huma.Context)),
) {
	huma.Register(api, huma.Operation{
		OperationID: "mobile-google-id-token",
		Method:      http.MethodPost,
		Path:        "/api/auth/mobile/google",
		Summary:     "Exchange a Google ID token for a session (mobile flow)",
		Tags:        []string{"auth", "mobile"},
		Middlewares: huma.Middlewares{rateLimitMW},
	}, h.googleIDToken)

	huma.Register(api, huma.Operation{
		OperationID: "mobile-verify-magic-link",
		Method:      http.MethodPost,
		Path:        "/api/auth/mobile/verify",
		Summary:     "Exchange a magic-link token for a session (mobile flow)",
		Tags:        []string{"auth", "mobile"},
		Middlewares: huma.Middlewares{rateLimitMW},
	}, h.verifyMagicLink)
}

func (h *MobileAuthHandler) googleIDToken(ctx context.Context, in *mobileGoogleInput) (*mobileTokenOutput, error) {
	user, pair, err := h.auth.VerifyGoogleIDToken(ctx, in.Body.IDToken)
	if err != nil {
		slog.Error("mobile google login failed", "error", err)
		return nil, MapDomainErr(ctx, err)
	}
	slog.Info("user logged in via mobile google", "user_id", user.ID)
	return &mobileTokenOutput{Body: TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}}, nil
}

func (h *MobileAuthHandler) verifyMagicLink(ctx context.Context, in *mobileVerifyInput) (*mobileTokenOutput, error) {
	user, pair, err := h.auth.VerifyMagicLink(ctx, in.Body.Token)
	if err != nil {
		slog.Error("mobile magic link verification failed", "error", err)
		return nil, MapDomainErr(ctx, err)
	}
	slog.Info("user logged in via mobile magic link", "user_id", user.ID)
	return &mobileTokenOutput{Body: TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}}, nil
}
