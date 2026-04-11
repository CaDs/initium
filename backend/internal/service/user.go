package service

import (
	"context"
	"fmt"

	"github.com/eridia/initium/backend/internal/domain"
)

// UserService implements domain.UserService.
type UserService struct {
	users domain.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

// GetProfile returns the user profile for the given user ID.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user profile: %w", err)
	}
	return user, nil
}

// UpdateProfile updates the user's display name.
func (s *UserService) UpdateProfile(ctx context.Context, userID string, name string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user for update: %w", err)
	}

	user.Name = name

	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user profile: %w", err)
	}

	return user, nil
}
