package domain

import "time"

// Session represents an active user session backed by a refresh token.
type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

// IsValid returns true if the session is not expired and not revoked.
func (s Session) IsValid() bool {
	if s.RevokedAt != nil {
		return false
	}
	return time.Now().Before(s.ExpiresAt)
}

// MagicLinkToken represents a single-use magic link for passwordless auth.
type MagicLinkToken struct {
	ID        string
	Email     string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// IsValid returns true if the token is unused and not expired.
func (t MagicLinkToken) IsValid() bool {
	if t.UsedAt != nil {
		return false
	}
	return time.Now().Before(t.ExpiresAt)
}

// TokenPair holds an access token and refresh token issued after authentication.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
