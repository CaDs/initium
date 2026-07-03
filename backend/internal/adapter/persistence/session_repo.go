package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/eridia/initium/backend/internal/adapter/persistence/db"
	"github.com/eridia/initium/backend/internal/domain"
)

const refreshTokenExpiry = 7 * 24 * time.Hour

// SessionRepo implements domain.SessionRepository using sqlc-generated queries.
type SessionRepo struct {
	q *db.Queries
}

// NewSessionRepo creates a new sqlc-backed session repository.
func NewSessionRepo(q *db.Queries) *SessionRepo {
	return &SessionRepo{q: q}
}

func (r *SessionRepo) CreateSession(ctx context.Context, session *domain.Session) error {
	session.ExpiresAt = time.Now().Add(refreshTokenExpiry)
	m, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		ExpiresAt:        session.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	*session = *sessionToDomain(m)
	return nil
}

func (r *SessionRepo) FindSessionByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	m, err := r.q.GetSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("querying session: %w", err)
	}
	return sessionToDomain(m), nil
}

func (r *SessionRepo) ClaimSessionByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	m, err := r.q.ClaimSessionByRefreshTokenHash(ctx, hash)
	if err == nil {
		return sessionToDomain(m), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("claiming refresh session: %w", err)
	}

	// No row was claimed — the token is unknown, revoked, or expired.
	// Re-read to report the precise reason.
	session, err := r.FindSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil {
		return nil, domain.ErrSessionRevoked
	}
	return nil, domain.ErrSessionExpired
}

func (r *SessionRepo) RevokeSession(ctx context.Context, sessionID string) error {
	if err := r.q.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

func (r *SessionRepo) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if err := r.q.RevokeAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("revoking all user sessions: %w", err)
	}
	return nil
}

func (r *SessionRepo) CreateMagicLinkToken(ctx context.Context, token *domain.MagicLinkToken) error {
	token.ExpiresAt = time.Now().Add(15 * time.Minute)
	m, err := r.q.CreateMagicLinkToken(ctx, db.CreateMagicLinkTokenParams{
		Email:     token.Email,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("creating magic link token: %w", err)
	}
	*token = *magicLinkTokenToDomain(m)
	return nil
}

func (r *SessionRepo) FindMagicLinkTokenByHash(ctx context.Context, hash string) (*domain.MagicLinkToken, error) {
	m, err := r.q.GetMagicLinkTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, fmt.Errorf("querying magic link token: %w", err)
	}
	return magicLinkTokenToDomain(m), nil
}

// DeleteExpiredMagicLinks removes magic_link_tokens that are expired or already used.
// Returns the number of rows deleted.
func (r *SessionRepo) DeleteExpiredMagicLinks(ctx context.Context) (int, error) {
	n, err := r.q.DeleteExpiredMagicLinks(ctx)
	if err != nil {
		return 0, fmt.Errorf("deleting expired magic links: %w", err)
	}
	return int(n), nil
}

// MarkMagicLinkTokenUsed atomically claims a magic-link token as used.
// Returns domain.ErrTokenUsed if the token was already consumed by a concurrent
// request — prevents the TOCTOU between FindMagicLinkTokenByHash and this call.
func (r *SessionRepo) MarkMagicLinkTokenUsed(ctx context.Context, tokenID string) error {
	n, err := r.q.MarkMagicLinkTokenUsed(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("marking magic link token used: %w", err)
	}
	if n == 0 {
		return domain.ErrTokenUsed
	}
	return nil
}
