-- name: CreateChirps :one
INSERT INTO chirps (
  id,
  created_at,
  updated_at,
  body,
  user_id
) VALUES (
  gen_random_uuid(), 
  now(),
  now(),
  $1,
  $2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;

-- name: GetChirpByAuthor :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: DeleteChirp :execrows
DELETE FROM chirps
WHERE id = $1
AND user_id = $2;

-- name: DeleteChirps :exec
DELETE FROM chirps;
