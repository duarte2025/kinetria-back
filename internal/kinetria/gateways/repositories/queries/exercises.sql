-- name: ExistsExerciseByIDAndWorkoutID :one
SELECT EXISTS(
    SELECT 1 FROM workout_exercises WHERE exercise_id = $1 AND workout_id = $2
) AS exists;

-- name: FindWorkoutExerciseID :one
SELECT id FROM workout_exercises WHERE exercise_id = $1 AND workout_id = $2;

-- name: ListExercisesByWorkoutID :many
SELECT 
    e.id, 
    e.name, 
    e.thumbnail_url, 
    e.muscles,
    we.sets, 
    we.reps, 
    we.rest_time, 
    we.weight, 
    we.order_index
FROM exercises e
INNER JOIN workout_exercises we ON e.id = we.exercise_id
WHERE we.workout_id = $1
ORDER BY we.order_index ASC;
