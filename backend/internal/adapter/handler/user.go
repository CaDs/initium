package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/gen/api"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	users domain.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(users domain.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// GetProfile returns the current user's profile.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetProfile(r.Context(), userID)
	if err != nil {
		Error(w, r, err)
		return
	}

	writeUser(w, r, user)
}

// UpdateProfile updates the current user's name.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req api.UpdateProfileRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		Error(w, r, domain.ErrInvalidInput)
		return
	}
	if req.Name == nil || len(*req.Name) == 0 || len(*req.Name) > 100 {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	user, err := h.users.UpdateProfile(r.Context(), userID, *req.Name)
	if err != nil {
		Error(w, r, err)
		return
	}

	writeUser(w, r, user)
}

// writeUser serializes a domain.User into the generated api.User wire type.
// Used by every endpoint that returns a user (GET /api/me, PATCH /api/me).
// If the stored ID is not a valid UUID (should never happen), returns 500.
func writeUser(w http.ResponseWriter, r *http.Request, u *domain.User) {
	id, err := uuid.Parse(u.ID)
	if err != nil {
		slog.Error("invalid user uuid", "id", u.ID, "error", err)
		Error(w, r, err)
		return
	}
	JSON(w, r, http.StatusOK, api.User{
		Id:        openapi_types.UUID(id),
		Email:     openapi_types.Email(u.Email),
		Name:      u.Name,
		AvatarUrl: u.AvatarURL,
		Role:      api.UserRole(u.Role),
		CreatedAt: u.CreatedAt,
	})
}
