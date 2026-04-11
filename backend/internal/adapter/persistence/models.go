package persistence

import (
	"time"

	"github.com/eridia/initium/backend/internal/domain"
)

// UserModel is the GORM representation of a user.
type UserModel struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string `gorm:"not null;default:''"`
	AvatarURL    *string
	AuthProvider string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}

func (UserModel) TableName() string { return "users" }

func (m *UserModel) ToDomain() *domain.User {
	avatarURL := ""
	if m.AvatarURL != nil {
		avatarURL = *m.AvatarURL
	}
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		Name:         m.Name,
		AvatarURL:    avatarURL,
		AuthProvider: m.AuthProvider,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func UserModelFromDomain(u *domain.User) *UserModel {
	var avatarURL *string
	if u.AvatarURL != "" {
		avatarURL = &u.AvatarURL
	}
	return &UserModel{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		AvatarURL:    avatarURL,
		AuthProvider: u.AuthProvider,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// SessionModel is the GORM representation of a session.
type SessionModel struct {
	ID               string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID           string     `gorm:"type:uuid;not null;index"`
	RefreshTokenHash string     `gorm:"uniqueIndex;not null"`
	ExpiresAt        time.Time  `gorm:"not null"`
	RevokedAt        *time.Time
	CreatedAt        time.Time  `gorm:"not null;default:now()"`
}

func (SessionModel) TableName() string { return "sessions" }

func (m *SessionModel) ToDomain() *domain.Session {
	return &domain.Session{
		ID:               m.ID,
		UserID:           m.UserID,
		RefreshTokenHash: m.RefreshTokenHash,
		ExpiresAt:        m.ExpiresAt,
		RevokedAt:        m.RevokedAt,
		CreatedAt:        m.CreatedAt,
	}
}

func SessionModelFromDomain(s *domain.Session) *SessionModel {
	return &SessionModel{
		ID:               s.ID,
		UserID:           s.UserID,
		RefreshTokenHash: s.RefreshTokenHash,
		ExpiresAt:        s.ExpiresAt,
		RevokedAt:        s.RevokedAt,
		CreatedAt:        s.CreatedAt,
	}
}

// MagicLinkTokenModel is the GORM representation of a magic link token.
type MagicLinkTokenModel struct {
	ID        string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string     `gorm:"not null"`
	TokenHash string     `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time  `gorm:"not null;default:now()"`
}

func (MagicLinkTokenModel) TableName() string { return "magic_link_tokens" }

func (m *MagicLinkTokenModel) ToDomain() *domain.MagicLinkToken {
	return &domain.MagicLinkToken{
		ID:        m.ID,
		Email:     m.Email,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		CreatedAt: m.CreatedAt,
	}
}
