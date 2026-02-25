-- name: ExistsExerciseByIDAndWorkoutID :one
SELECT EXISTS(
    SELECT 1 FROM exercises WHERE id = $1 AND workout_id = $2
) AS exists;
