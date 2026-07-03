-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByRefreshTokenHash :one
SELECT * FROM sessions WHERE refresh_token_hash = $1;

-- name: ClaimSessionByRefreshTokenHash :one
UPDATE sessions
SET revoked_at = now()
WHERE refresh_token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
RETURNING *;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = now() WHERE id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL;
