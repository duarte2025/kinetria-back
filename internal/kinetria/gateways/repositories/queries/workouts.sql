-- name: ExistsWorkoutByIDAndUserID :one
SELECT EXISTS(
    SELECT 1 FROM workouts WHERE id = $1 AND user_id = $2
) AS "exists";

-- name: ListWorkoutsByUserID :many
SELECT id, user_id, name, description, type, intensity, duration, image_url, created_at, updated_at, created_by, deleted_at
FROM workouts
WHERE (created_by = $1 OR created_by IS NULL) AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWorkoutsByUserID :one
SELECT COUNT(*)
FROM workouts
WHERE (created_by = $1 OR created_by IS NULL) AND deleted_at IS NULL;

-- name: GetFirstWorkoutByUserID :one
SELECT 
    id, 
    user_id, 
    name, 
    description, 
    type, 
    intensity, 
    duration, 
    image_url, 
    created_at, 
    updated_at,
    created_by,
    deleted_at
FROM workouts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT 1;

-- name: GetWorkoutByID :one
SELECT 
    id, 
    user_id, 
    name, 
    description, 
    type, 
    intensity, 
    duration, 
    image_url, 
    created_at, 
    updated_at,
    created_by,
    deleted_at
FROM workouts
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetWorkoutByIDOnly :one
SELECT id, user_id, name, description, type, intensity, duration, image_url, created_at, updated_at, created_by, deleted_at
FROM workouts
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkout :exec
INSERT INTO workouts (id, user_id, name, description, type, intensity, duration, image_url, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: UpdateWorkout :exec
UPDATE workouts SET name=$2, description=$3, type=$4, intensity=$5, duration=$6, image_url=$7, updated_at=$8
WHERE id=$1 AND deleted_at IS NULL;

-- name: SoftDeleteWorkout :exec
UPDATE workouts SET deleted_at=$2, updated_at=$3 WHERE id=$1 AND deleted_at IS NULL;

-- name: HasActiveSessions :one
SELECT EXISTS(SELECT 1 FROM sessions WHERE workout_id=$1 AND status='active') AS "exists";

-- name: CreateWorkoutExercise :exec
INSERT INTO workout_exercises (id, workout_id, exercise_id, sets, reps, rest_time, weight, order_index, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: DeleteWorkoutExercises :exec
DELETE FROM workout_exercises WHERE workout_id=$1;
