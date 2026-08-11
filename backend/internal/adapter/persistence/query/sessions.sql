-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByRefreshTokenHash :one
SELECT * FROM sessions WHERE refresh_token_hash = $1;

-- name: ClaimSessionByRefreshTokenHash :one
-- Reuse-grace rotation: a claim within 60s of revocation still succeeds, so
-- web clients that legitimately race refreshes (per-request middleware,
-- parallel tabs, prefetches) don't log the race loser out. COALESCE keeps
-- the original revocation anchor so re-claims can never extend the window.
-- Theft reuse-detection still applies beyond the grace (Auth0-style).
UPDATE sessions
SET revoked_at = COALESCE(revoked_at, now())
WHERE refresh_token_hash = $1
  AND (revoked_at IS NULL OR revoked_at > now() - interval '60 seconds')
  AND expires_at > now()
RETURNING *;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = now() WHERE id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL;
