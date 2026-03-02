-- name: CreateSetRecord :exec
INSERT INTO set_records (id, session_id, workout_exercise_id, set_number, weight, reps, status, recorded_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindSetRecordBySessionExerciseSet :one
SELECT id, session_id, workout_exercise_id, set_number, weight, reps, status, recorded_at
FROM set_records
WHERE session_id = $1 AND workout_exercise_id = $2 AND set_number = $3;

-- name: GetTotalSetsRepsVolume :one
SELECT
    COUNT(sr.id)::bigint AS total_sets,
    COALESCE(SUM(sr.reps), 0)::bigint AS total_reps,
    COALESCE(SUM(sr.weight::bigint * sr.reps), 0)::bigint AS total_volume
FROM set_records sr
JOIN sessions s ON sr.session_id = s.id
WHERE s.user_id = $1
  AND s.status = 'completed'
  AND sr.status = 'completed'
  AND s.started_at >= $2
  AND s.started_at <= $3;

-- name: GetPersonalRecordsByUser :many
WITH best_sets AS (
    SELECT
        we.exercise_id,
        e.name                                                                          AS exercise_name,
        e.muscles ->> 0                                                                 AS primary_muscle,
        sr.weight,
        sr.reps,
        s.started_at,
        ROW_NUMBER() OVER (PARTITION BY we.exercise_id ORDER BY sr.weight DESC, sr.reps DESC, s.started_at DESC) AS rn_exercise
    FROM set_records sr
    JOIN sessions s ON sr.session_id = s.id
    JOIN workout_exercises we ON sr.workout_exercise_id = we.id
    JOIN exercises e ON we.exercise_id = e.id
    WHERE s.user_id = $1
      AND s.status = 'completed'
      AND sr.status = 'completed'
      AND sr.weight > 0
),
best_per_exercise AS (
    SELECT exercise_id, exercise_name, primary_muscle, weight, reps, started_at
    FROM best_sets
    WHERE rn_exercise = 1
),
exercise_frequency AS (
    SELECT
        we.exercise_id,
        COUNT(DISTINCT s.id) AS times_used
    FROM set_records sr
    JOIN sessions s ON sr.session_id = s.id
    JOIN workout_exercises we ON sr.workout_exercise_id = we.id
    WHERE s.user_id = $1
      AND s.status = 'completed'
    GROUP BY we.exercise_id
),
exercise_data AS (
    SELECT
        bpe.exercise_id,
        bpe.exercise_name,
        bpe.primary_muscle,
        bpe.weight,
        bpe.reps,
        bpe.started_at,
        COALESCE(ef.times_used, 0) AS times_used
    FROM best_per_exercise bpe
    LEFT JOIN exercise_frequency ef ON bpe.exercise_id = ef.exercise_id
),
ranked_by_muscle AS (
    SELECT *,
           ROW_NUMBER() OVER (
               PARTITION BY primary_muscle
               ORDER BY times_used DESC, weight DESC
               ) AS rank_in_muscle
    FROM exercise_data
)
SELECT
    exercise_id,
    exercise_name,
    weight::int    AS weight,
    reps::int      AS reps,
    (weight::bigint * reps) AS volume,
    started_at     AS achieved_at
FROM ranked_by_muscle
WHERE rank_in_muscle = 1
ORDER BY weight DESC
LIMIT 15;

-- name: GetProgressionByUserAndExercise :many
SELECT
    DATE(s.started_at)                  AS date,
    MAX(sr.weight)::bigint              AS max_weight,
    SUM(sr.weight::bigint * sr.reps)    AS total_volume
FROM set_records sr
JOIN sessions s ON sr.session_id = s.id
JOIN workout_exercises we ON sr.workout_exercise_id = we.id
WHERE s.user_id = $1
  AND s.status = 'completed'
  AND sr.status = 'completed'
  AND sr.weight > 0
  AND s.started_at >= $2
  AND s.started_at <= $3
  AND ($4::uuid IS NULL OR we.exercise_id = $4::uuid)
GROUP BY DATE(s.started_at)
ORDER BY date;
