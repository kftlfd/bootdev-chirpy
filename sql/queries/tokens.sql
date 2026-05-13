-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES ($1, now(), now(), $2, $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE
    revoked_at IS NULL
    AND expires_at > now()
    AND token = $1;

-- name: MarkTokenRevoked :one
UPDATE refresh_tokens
SET
    updated_at = now(),
    revoked_at = now()
WHERE token = $1
RETURNING *;
