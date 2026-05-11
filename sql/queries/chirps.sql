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

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE user_id = $1 AND id = $2;
