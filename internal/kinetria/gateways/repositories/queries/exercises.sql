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

-- name: ListExercises :many
SELECT
    id, name, description, thumbnail_url, muscles,
    instructions, tips, difficulty, equipment, video_url,
    created_at, updated_at
FROM exercises
WHERE
    ($1::text IS NULL OR name ILIKE '%' || $1::text || '%')
    AND ($2::text IS NULL OR muscles @> jsonb_build_array($2::text))
    AND ($3::text IS NULL OR equipment = $3::text)
    AND ($4::text IS NULL OR difficulty = $4::text)
ORDER BY name ASC
LIMIT $5 OFFSET $6;

-- name: CountExercises :one
SELECT COUNT(*)
FROM exercises
WHERE
    ($1::text IS NULL OR name ILIKE '%' || $1::text || '%')
    AND ($2::text IS NULL OR muscles @> jsonb_build_array($2::text))
    AND ($3::text IS NULL OR equipment = $3::text)
    AND ($4::text IS NULL OR difficulty = $4::text);

-- name: GetExerciseByID :one
SELECT
    id, name, description, thumbnail_url, muscles,
    instructions, tips, difficulty, equipment, video_url,
    created_at, updated_at
FROM exercises
WHERE id = $1;

-- name: GetExerciseUserStats :one
SELECT
    MAX(s.started_at)        AS last_performed,
    MAX(sr.weight)           AS best_weight,
    COUNT(DISTINCT s.id)     AS times_performed,
    AVG(sr.weight::float)    AS average_weight
FROM sessions s
JOIN workout_exercises we ON we.workout_id = s.workout_id AND we.exercise_id = $1
JOIN set_records sr ON sr.session_id = s.id AND sr.workout_exercise_id = we.id
WHERE s.user_id = $2 AND s.status = 'completed';

-- name: GetExerciseHistory :many
SELECT
    s.id           AS session_id,
    w.name         AS workout_name,
    s.started_at   AS performed_at,
    sr.set_number,
    sr.reps,
    sr.weight,
    sr.status
FROM sessions s
JOIN workouts w ON s.workout_id = w.id
JOIN workout_exercises we ON we.workout_id = s.workout_id AND we.exercise_id = $1
JOIN set_records sr ON sr.session_id = s.id AND sr.workout_exercise_id = we.id
WHERE s.user_id = $2 AND s.status = 'completed'
ORDER BY s.started_at DESC, sr.set_number ASC
LIMIT $3 OFFSET $4;

-- name: CountExerciseHistory :one
SELECT COUNT(DISTINCT s.id)
FROM sessions s
JOIN workout_exercises we ON we.workout_id = s.workout_id AND we.exercise_id = $1
JOIN set_records sr ON sr.session_id = s.id AND sr.workout_exercise_id = we.id
WHERE s.user_id = $2 AND s.status = 'completed';
