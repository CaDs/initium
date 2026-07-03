-- name: CreateMagicLinkToken :one
INSERT INTO magic_link_tokens (email, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMagicLinkTokenByHash :one
SELECT * FROM magic_link_tokens WHERE token_hash = $1;

-- name: DeleteExpiredMagicLinks :execrows
DELETE FROM magic_link_tokens WHERE expires_at < now() OR used_at IS NOT NULL;

-- name: MarkMagicLinkTokenUsed :execrows
UPDATE magic_link_tokens SET used_at = now() WHERE id = $1 AND used_at IS NULL;
