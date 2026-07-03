package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/eridia/initium/backend/internal/adapter/persistence/db"
	"github.com/eridia/initium/backend/internal/domain"
)

// UserRepo implements domain.UserRepository using sqlc-generated queries.
type UserRepo struct {
	q *db.Queries
}

// NewUserRepo creates a new sqlc-backed user repository.
func NewUserRepo(q *db.Queries) *UserRepo {
	return &UserRepo{q: q}
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	m, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return userToDomain(m), nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return userToDomain(m), nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	m, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:        user.Email,
		Name:         user.Name,
		AvatarUrl:    avatarPtr(user.AvatarURL),
		AuthProvider: user.AuthProvider,
	})
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	*user = *userToDomain(m)
	return nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	if _, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:        user.ID,
		Name:      user.Name,
		AvatarUrl: avatarPtr(user.AvatarURL),
	}); err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}
