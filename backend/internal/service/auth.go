package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/eridia/initium/backend/internal/domain"
)

// AuthService implements domain.AuthService.
type AuthService struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	oauth    domain.OAuthVerifier
	email    domain.EmailSender
	tokens   domain.TokenGenerator
}

// NewAuthService creates a new AuthService with all required dependencies.
func NewAuthService(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	oauth domain.OAuthVerifier,
	email domain.EmailSender,
	tokens domain.TokenGenerator,
) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		oauth:    oauth,
		email:    email,
		tokens:   tokens,
	}
}

// LoginWithGoogle exchanges an OAuth authorization code for user + tokens.
func (s *AuthService) LoginWithGoogle(ctx context.Context, code string) (*domain.User, *domain.TokenPair, error) {
	profile, err := s.oauth.ExchangeCode(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("exchanging oauth code: %w", err)
	}

	user, err := s.findOrCreateUser(ctx, profile.Email, profile.Name, profile.AvatarURL, "google")
	if err != nil {
		return nil, nil, fmt.Errorf("finding or creating user: %w", err)
	}

	pair, err := s.createSession(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("creating session: %w", err)
	}

	return user, pair, nil
}

// VerifyGoogleIDToken verifies a Google ID token from mobile clients.
func (s *AuthService) VerifyGoogleIDToken(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error) {
	profile, err := s.oauth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, nil, fmt.Errorf("verifying id token: %w", err)
	}

	user, err := s.findOrCreateUser(ctx, profile.Email, profile.Name, profile.AvatarURL, "google")
	if err != nil {
		return nil, nil, fmt.Errorf("finding or creating user: %w", err)
	}

	pair, err := s.createSession(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("creating session: %w", err)
	}

	return user, pair, nil
}

// RequestMagicLink generates a magic link token and sends it via email.
func (s *AuthService) RequestMagicLink(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return domain.ErrEmailRequired
	}

	rawToken, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return fmt.Errorf("generating magic link token: %w", err)
	}

	mlt := &domain.MagicLinkToken{
		Email:     email,
		TokenHash: s.tokens.HashToken(rawToken),
	}

	if err := s.sessions.CreateMagicLinkToken(ctx, mlt); err != nil {
		return fmt.Errorf("storing magic link token: %w", err)
	}

	if err := s.email.SendMagicLink(ctx, email, rawToken); err != nil {
		return fmt.Errorf("sending magic link email: %w", err)
	}

	return nil
}

// VerifyMagicLink validates a magic link token and returns user + session tokens.
func (s *AuthService) VerifyMagicLink(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error) {
	hash := s.tokens.HashToken(token)

	mlt, err := s.sessions.FindMagicLinkTokenByHash(ctx, hash)
	if err != nil {
		return nil, nil, fmt.Errorf("finding magic link token: %w", err)
	}

	if !mlt.IsValid() {
		if mlt.UsedAt != nil {
			return nil, nil, domain.ErrTokenUsed
		}
		return nil, nil, domain.ErrTokenExpired
	}

	if err := s.sessions.MarkMagicLinkTokenUsed(ctx, mlt.ID); err != nil {
		if errors.Is(err, domain.ErrTokenUsed) {
			return nil, nil, domain.ErrTokenUsed
		}
		return nil, nil, fmt.Errorf("marking token used: %w", err)
	}

	user, err := s.findOrCreateUser(ctx, mlt.Email, "", "", "magic_link")
	if err != nil {
		return nil, nil, fmt.Errorf("finding or creating user: %w", err)
	}

	pair, err := s.createSession(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("creating session: %w", err)
	}

	return user, pair, nil
}

// RefreshTokens validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	hash := s.tokens.HashToken(refreshToken)

	session, err := s.sessions.ClaimSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("claiming refresh session: %w", err)
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user for refresh: %w", err)
	}

	pair, err := s.createSession(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("creating new session: %w", err)
	}

	return pair, nil
}

// Logout revokes the session associated with the given refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := s.tokens.HashToken(refreshToken)

	session, err := s.sessions.FindSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("finding session for logout: %w", err)
	}

	if err := s.sessions.RevokeSession(ctx, session.ID); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}

	return nil
}

// LogoutAll revokes all sessions for the given user.
func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	if err := s.sessions.RevokeAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("revoking all sessions: %w", err)
	}
	return nil
}

func (s *AuthService) findOrCreateUser(ctx context.Context, email, name, avatarURL, provider string) (*domain.User, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	newUser := &domain.User{
		Email:        email,
		Name:         name,
		AvatarURL:    avatarURL,
		AuthProvider: provider,
	}

	if err := s.users.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return newUser, nil
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	accessToken, err := s.tokens.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	session := &domain.Session{
		UserID:           user.ID,
		RefreshTokenHash: s.tokens.HashToken(refreshToken),
	}

	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("persisting session: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
