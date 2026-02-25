-- name: ExistsExerciseByIDAndWorkoutID :one
SELECT EXISTS(
    SELECT 1 FROM exercises WHERE id = $1 AND workout_id = $2
) AS exists;

-- name: ListExercisesByWorkoutID :many
SELECT 
    id, 
    workout_id, 
    name, 
    thumbnail_url, 
    sets, 
    reps, 
    muscles, 
    rest_time, 
    weight, 
    order_index
FROM exercises
WHERE workout_id = $1
ORDER BY order_index ASC;
