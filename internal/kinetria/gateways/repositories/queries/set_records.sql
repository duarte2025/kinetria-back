-- name: CreateSetRecord :exec
INSERT INTO set_records (id, session_id, workout_exercise_id, set_number, weight, reps, status, recorded_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindSetRecordBySessionExerciseSet :one
SELECT id, session_id, workout_exercise_id, set_number, weight, reps, status, recorded_at
FROM set_records
WHERE session_id = $1 AND workout_exercise_id = $2 AND set_number = $3;
