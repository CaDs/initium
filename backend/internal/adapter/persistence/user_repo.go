package persistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/eridia/initium/backend/internal/domain"
)

// GormUserRepo implements domain.UserRepository using GORM.
type GormUserRepo struct {
	db *gorm.DB
}

// NewGormUserRepo creates a new GORM-backed user repository.
func NewGormUserRepo(db *gorm.DB) *GormUserRepo {
	return &GormUserRepo{db: db}
}

func (r *GormUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return m.ToDomain(), nil
}

func (r *GormUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return m.ToDomain(), nil
}

func (r *GormUserRepo) Create(ctx context.Context, user *domain.User) error {
	m := UserModelFromDomain(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *GormUserRepo) Update(ctx context.Context, user *domain.User) error {
	m := UserModelFromDomain(user)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	user.UpdatedAt = m.UpdatedAt
	return nil
}
