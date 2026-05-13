-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
RETURNING *;

-- name: ResetChirps :exec
TRUNCATE TABLE chirps CASCADE;

-- sqlc dynamic queries workaround: https://dizzy.zone/2024/07/03/SQLC-dynamic-queries/
-- name: GetAllChirps :many
SELECT * FROM chirps
WHERE id IS NOT NULL
AND (NOT @has_user_id::boolean OR user_id = $1)
ORDER BY created_at ASC;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirp :one
DELETE FROM chirps WHERE user_id = $1 AND id = $2
RETURNING *;
