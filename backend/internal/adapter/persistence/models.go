package persistence

import (
	"github.com/eridia/initium/backend/internal/adapter/persistence/db"
	"github.com/eridia/initium/backend/internal/domain"
)

// Mappers from sqlc row structs (backend/internal/adapter/persistence/db) to
// domain entities. Domain types stay free of persistence concerns — the only
// impedance is nullable columns, which sqlc emits as pointers.

func userToDomain(m db.User) *domain.User {
	avatarURL := ""
	if m.AvatarUrl != nil {
		avatarURL = *m.AvatarUrl
	}
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		Name:         m.Name,
		AvatarURL:    avatarURL,
		AuthProvider: m.AuthProvider,
		Role:         m.Role,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func avatarPtr(avatarURL string) *string {
	if avatarURL == "" {
		return nil
	}
	return &avatarURL
}

func sessionToDomain(m db.Session) *domain.Session {
	return &domain.Session{
		ID:               m.ID,
		UserID:           m.UserID,
		RefreshTokenHash: m.RefreshTokenHash,
		ExpiresAt:        m.ExpiresAt,
		RevokedAt:        m.RevokedAt,
		CreatedAt:        m.CreatedAt,
	}
}

func magicLinkTokenToDomain(m db.MagicLinkToken) *domain.MagicLinkToken {
	return &domain.MagicLinkToken{
		ID:        m.ID,
		Email:     m.Email,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		CreatedAt: m.CreatedAt,
	}
}
