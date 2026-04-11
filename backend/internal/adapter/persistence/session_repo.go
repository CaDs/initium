package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/eridia/initium/backend/internal/domain"
)

const refreshTokenExpiry = 7 * 24 * time.Hour

// GormSessionRepo implements domain.SessionRepository using GORM.
type GormSessionRepo struct {
	db *gorm.DB
}

// NewGormSessionRepo creates a new GORM-backed session repository.
func NewGormSessionRepo(db *gorm.DB) *GormSessionRepo {
	return &GormSessionRepo{db: db}
}

func (r *GormSessionRepo) CreateSession(ctx context.Context, session *domain.Session) error {
	session.ExpiresAt = time.Now().Add(refreshTokenExpiry)
	m := SessionModelFromDomain(session)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	session.ID = m.ID
	session.CreatedAt = m.CreatedAt
	return nil
}

func (r *GormSessionRepo) FindSessionByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	var m SessionModel
	if err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("querying session: %w", err)
	}
	return m.ToDomain(), nil
}

func (r *GormSessionRepo) RevokeSession(ctx context.Context, sessionID string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&SessionModel{}).
		Where("id = ?", sessionID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

func (r *GormSessionRepo) RevokeAllUserSessions(ctx context.Context, userID string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&SessionModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("revoking all user sessions: %w", err)
	}
	return nil
}

func (r *GormSessionRepo) CreateMagicLinkToken(ctx context.Context, token *domain.MagicLinkToken) error {
	token.ExpiresAt = time.Now().Add(15 * time.Minute)
	m := &MagicLinkTokenModel{
		Email:     token.Email,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("creating magic link token: %w", err)
	}
	token.ID = m.ID
	token.CreatedAt = m.CreatedAt
	return nil
}

func (r *GormSessionRepo) FindMagicLinkTokenByHash(ctx context.Context, hash string) (*domain.MagicLinkToken, error) {
	var m MagicLinkTokenModel
	if err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, fmt.Errorf("querying magic link token: %w", err)
	}
	return m.ToDomain(), nil
}

func (r *GormSessionRepo) MarkMagicLinkTokenUsed(ctx context.Context, tokenID string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&MagicLinkTokenModel{}).
		Where("id = ?", tokenID).
		Update("used_at", now).Error; err != nil {
		return fmt.Errorf("marking magic link token used: %w", err)
	}
	return nil
}
