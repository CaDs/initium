package handler

import (
	"encoding/json"
	"net/http"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
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

	JSON(w, r, http.StatusOK, userResponse(user))
}

// UpdateProfile updates the current user's name.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	if len(body.Name) == 0 || len(body.Name) > 100 {
		Error(w, r, domain.ErrInvalidInput)
		return
	}

	user, err := h.users.UpdateProfile(r.Context(), userID, body.Name)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, r, http.StatusOK, userResponse(user))
}

// userResponse is the single source of truth for how a user is serialized on the wire.
// Keeps GET /api/me and PATCH /api/me responses in lockstep.
func userResponse(user *domain.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
	}
}
