-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, workout_id, started_at, status, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindActiveSessionByUserID :one
SELECT id, user_id, workout_id, started_at, finished_at, status, notes, created_at, updated_at
FROM sessions
WHERE user_id = $1 AND status = 'active'
LIMIT 1;
