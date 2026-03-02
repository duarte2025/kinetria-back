-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, workout_id, started_at, status, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindActiveSessionByUserID :one
SELECT id, user_id, workout_id, started_at, finished_at, status, notes, created_at, updated_at
FROM sessions
WHERE user_id = $1 AND status = 'active'
LIMIT 1;

-- name: FindSessionByID :one
SELECT id, user_id, workout_id, started_at, finished_at, status, notes, created_at, updated_at
FROM sessions
WHERE id = $1;

-- name: UpdateSessionStatus :execrows
UPDATE sessions
SET status = $2, finished_at = $3, notes = $4, updated_at = $5
WHERE id = $1 AND status = 'active';

-- name: GetCompletedSessionsByDateRange :many
SELECT 
    id, 
    user_id, 
    workout_id, 
    status, 
    notes,
    started_at, 
    finished_at, 
    created_at, 
    updated_at
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND DATE(started_at) BETWEEN $2 AND $3
ORDER BY started_at DESC;


-- name: GetStatsByUserAndPeriod :one
SELECT
    COUNT(*)::bigint AS total_workouts,
    COALESCE(SUM(EXTRACT(EPOCH FROM (finished_at - started_at)) / 60), 0)::bigint AS total_time_minutes
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND started_at >= $2
  AND started_at <= $3;

-- name: GetFrequencyByUserAndPeriod :many
SELECT
    DATE(started_at) AS date,
    COUNT(*)::bigint AS count
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND started_at >= $2
  AND started_at <= $3
GROUP BY DATE(started_at)
ORDER BY date;

-- name: GetSessionsForStreak :many
SELECT
    DATE(started_at)::text AS date
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND started_at >= NOW() - INTERVAL '365 days'
GROUP BY DATE(started_at)
ORDER BY date DESC;
