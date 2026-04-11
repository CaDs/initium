package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
)

// ErrorResponse is the standardized error format.
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Error writes a standardized error response, mapping domain errors to HTTP status codes.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetRequestID(r.Context())
	code, status := mapError(err)

	JSON(w, r, status, ErrorResponse{
		Code:      code,
		Message:   err.Error(),
		RequestID: reqID,
	})
}

func mapError(err error) (code string, status int) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return "USER_NOT_FOUND", http.StatusNotFound
	case errors.Is(err, domain.ErrSessionNotFound):
		return "SESSION_NOT_FOUND", http.StatusUnauthorized
	case errors.Is(err, domain.ErrSessionExpired):
		return "SESSION_EXPIRED", http.StatusUnauthorized
	case errors.Is(err, domain.ErrSessionRevoked):
		return "SESSION_REVOKED", http.StatusUnauthorized
	case errors.Is(err, domain.ErrTokenExpired):
		return "TOKEN_EXPIRED", http.StatusUnauthorized
	case errors.Is(err, domain.ErrTokenUsed):
		return "TOKEN_USED", http.StatusConflict
	case errors.Is(err, domain.ErrTokenInvalid):
		return "TOKEN_INVALID", http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidOAuthToken):
		return "INVALID_OAUTH_TOKEN", http.StatusUnauthorized
	case errors.Is(err, domain.ErrInvalidCredentials):
		return "INVALID_CREDENTIALS", http.StatusUnauthorized
	case errors.Is(err, domain.ErrEmailRequired):
		return "EMAIL_REQUIRED", http.StatusBadRequest
	case errors.Is(err, domain.ErrRateLimited):
		return "RATE_LIMITED", http.StatusTooManyRequests
	default:
		return "INTERNAL_ERROR", http.StatusInternalServerError
	}
}
