package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/gen/api"
)

// MobileAuthHandler handles mobile-specific authentication.
type MobileAuthHandler struct {
	auth domain.AuthService
}

// NewMobileAuthHandler creates a new MobileAuthHandler.
func NewMobileAuthHandler(auth domain.AuthService) *MobileAuthHandler {
	return &MobileAuthHandler{auth: auth}
}

// VerifyMagicLink validates a magic link token and returns tokens in the response body (mobile flow).
func (h *MobileAuthHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req api.MobileVerifyRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || req.Token == "" {
		Error(w, r, domain.ErrTokenInvalid)
		return
	}

	user, pair, err := h.auth.VerifyMagicLink(r.Context(), req.Token)
	if err != nil {
		slog.Error("mobile magic link verification failed", "error", err)
		Error(w, r, err)
		return
	}

	slog.Info("user logged in via mobile magic link", "user_id", user.ID)
	JSON(w, r, http.StatusOK, api.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// GoogleIDToken verifies a Google ID token from mobile and returns tokens in the response body.
func (h *MobileAuthHandler) GoogleIDToken(w http.ResponseWriter, r *http.Request) {
	var req api.MobileGoogleRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || req.IdToken == "" {
		Error(w, r, domain.ErrInvalidOAuthToken)
		return
	}

	user, pair, err := h.auth.VerifyGoogleIDToken(r.Context(), req.IdToken)
	if err != nil {
		slog.Error("mobile google login failed", "error", err)
		Error(w, r, err)
		return
	}

	slog.Info("user logged in via mobile google", "user_id", user.ID)
	JSON(w, r, http.StatusOK, api.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}
