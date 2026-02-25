-- name: ExistsWorkoutByIDAndUserID :one
SELECT EXISTS(
    SELECT 1 FROM workouts WHERE id = $1 AND user_id = $2
) AS "exists";

-- name: ListWorkoutsByUserID :many
SELECT id, user_id, name, description, type, intensity, duration, image_url, created_at, updated_at
FROM workouts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWorkoutsByUserID :one
SELECT COUNT(*)
FROM workouts
WHERE user_id = $1;

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
    updated_at
FROM workouts
WHERE user_id = $1
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
    updated_at
FROM workouts
WHERE id = $1 AND user_id = $2;
