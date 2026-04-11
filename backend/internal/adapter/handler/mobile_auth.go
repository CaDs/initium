package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eridia/initium/backend/internal/domain"
)

// MobileAuthHandler handles mobile-specific authentication.
type MobileAuthHandler struct {
	auth domain.AuthService
}

// NewMobileAuthHandler creates a new MobileAuthHandler.
func NewMobileAuthHandler(auth domain.AuthService) *MobileAuthHandler {
	return &MobileAuthHandler{auth: auth}
}

// GoogleIDToken verifies a Google ID token from mobile and returns tokens in the response body.
func (h *MobileAuthHandler) GoogleIDToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IDToken == "" {
		Error(w, r, domain.ErrInvalidOAuthToken)
		return
	}

	user, pair, err := h.auth.VerifyGoogleIDToken(r.Context(), body.IDToken)
	if err != nil {
		slog.Error("mobile google login failed", "error", err)
		Error(w, r, err)
		return
	}

	slog.Info("user logged in via mobile google", "user_id", user.ID, "email", user.Email)
	JSON(w, r, http.StatusOK, map[string]string{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}
