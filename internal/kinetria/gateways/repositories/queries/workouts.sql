-- name: ExistsWorkoutByIDAndUserID :one
SELECT EXISTS(
    SELECT 1 FROM workouts WHERE id = $1 AND user_id = $2
) AS "exists";
