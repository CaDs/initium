package domain

import "context"

// --- Repository interfaces ---

// UserRepository defines persistence operations for users.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

// SessionRepository defines persistence operations for sessions and magic link tokens.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session) error
	FindSessionByRefreshTokenHash(ctx context.Context, hash string) (*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	CreateMagicLinkToken(ctx context.Context, token *MagicLinkToken) error
	FindMagicLinkTokenByHash(ctx context.Context, hash string) (*MagicLinkToken, error)
	MarkMagicLinkTokenUsed(ctx context.Context, tokenID string) error
}

// --- Service interfaces ---

// AuthService defines authentication and session management operations.
type AuthService interface {
	LoginWithGoogle(ctx context.Context, code string) (*User, *TokenPair, error)
	VerifyGoogleIDToken(ctx context.Context, idToken string) (*User, *TokenPair, error)
	RequestMagicLink(ctx context.Context, email string) error
	VerifyMagicLink(ctx context.Context, token string) (*User, *TokenPair, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID string) error
}

// UserService defines user profile operations.
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, userID string, name string) (*User, error)
}

// --- Infrastructure port interfaces ---

// OAuthVerifier verifies OAuth tokens and returns user profile info.
type OAuthVerifier interface {
	ExchangeCode(ctx context.Context, code string) (*OAuthProfile, error)
	VerifyIDToken(ctx context.Context, idToken string) (*OAuthProfile, error)
}

// OAuthProfile holds the profile data returned from an OAuth provider.
type OAuthProfile struct {
	Email     string
	Name      string
	AvatarURL string
}

// EmailSender sends emails.
type EmailSender interface {
	SendMagicLink(ctx context.Context, to string, token string) error
}

// TokenGenerator creates and validates JWTs.
type TokenGenerator interface {
	GenerateAccessToken(userID string, email string) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (userID string, email string, err error)
	HashToken(token string) string
}
