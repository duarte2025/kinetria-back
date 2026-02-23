-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, token, expires_at, revoked_at, created_at;

-- name: GetRefreshTokenByToken :one
SELECT id, user_id, token, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token = $1
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
