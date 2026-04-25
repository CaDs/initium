package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

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

// userOutput wraps the User wire type so Huma generates the right
// response schema reference (User) instead of inlining anonymous fields.
type userOutput struct {
	Body User
}

// updateProfileInput carries the PATCH /api/me JSON body. Huma validates
// minLength/maxLength via the struct tags before the handler runs.
type updateProfileInput struct {
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"100" doc:"New display name (1-100 characters)"`
	}
}

// RegisterUser wires the user-profile endpoints onto the Huma API.
// Auth + role middleware come from the caller via the operation
// Middlewares slice — see app/api.go for how they're attached.
func (h *UserHandler) RegisterUser(api huma.API, authMW func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/api/me",
		Summary:     "Get current user profile",
		Tags:        []string{"user"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{authMW},
	}, h.getMe)

	huma.Register(api, huma.Operation{
		OperationID: "update-me",
		Method:      http.MethodPatch,
		Path:        "/api/me",
		Summary:     "Update current user's display name",
		Tags:        []string{"user"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{authMW},
	}, h.updateMe)
}

func (h *UserHandler) getMe(ctx context.Context, _ *struct{}) (*userOutput, error) {
	userID := middleware.GetUserID(ctx)
	user, err := h.users.GetProfile(ctx, userID)
	if err != nil {
		return nil, MapDomainErr(ctx, err)
	}
	return &userOutput{Body: toUserDTO(user)}, nil
}

func (h *UserHandler) updateMe(ctx context.Context, in *updateProfileInput) (*userOutput, error) {
	userID := middleware.GetUserID(ctx)
	user, err := h.users.UpdateProfile(ctx, userID, in.Body.Name)
	if err != nil {
		return nil, MapDomainErr(ctx, err)
	}
	return &userOutput{Body: toUserDTO(user)}, nil
}

// toUserDTO converts a domain.User into the wire-shape User. Centralized
// so every endpoint that returns a user uses identical field projection.
func toUserDTO(u *domain.User) User {
	return User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
