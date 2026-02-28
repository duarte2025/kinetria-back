-- Migration 009: Refactor exercises to shared library with N:N relationship
-- This migration restructures the database to support exercise sharing across workouts

-- Step 1: Create new exercises table as a shared library
CREATE TABLE IF NOT EXISTS exercises_new (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    thumbnail_url VARCHAR(500) NOT NULL DEFAULT '/assets/exercises/generic.png',
    muscles JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exercises_new_name ON exercises_new(name);
CREATE INDEX IF NOT EXISTS idx_exercises_new_muscles ON exercises_new USING GIN(muscles);

-- Step 2: Create workout_exercises junction table
CREATE TABLE IF NOT EXISTS workout_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises_new(id) ON DELETE RESTRICT,
    sets INT NOT NULL DEFAULT 1 CHECK (sets >= 1),
    reps VARCHAR(20) NOT NULL DEFAULT '',
    rest_time INT NOT NULL DEFAULT 60 CHECK (rest_time >= 0),
    weight INT NOT NULL DEFAULT 0 CHECK (weight >= 0),
    order_index INT NOT NULL DEFAULT 0 CHECK (order_index >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (workout_id, exercise_id)
);

CREATE INDEX IF NOT EXISTS idx_workout_exercises_workout_id ON workout_exercises(workout_id);
CREATE INDEX IF NOT EXISTS idx_workout_exercises_exercise_id ON workout_exercises(exercise_id);
CREATE INDEX IF NOT EXISTS idx_workout_exercises_order ON workout_exercises(workout_id, order_index);

-- Step 3: Migrate existing exercises data
-- This preserves the relationship in a way that respects the original data
INSERT INTO exercises_new (id, name, description, thumbnail_url, muscles, created_at, updated_at)
SELECT 
    gen_random_uuid() as id,
    name,
    '' as description,
    thumbnail_url,
    muscles,
    created_at,
    updated_at
FROM exercises
ON CONFLICT DO NOTHING;

-- Step 4: Create junction table entries with specific configurations
-- Each exercise from a workout becomes an entry in workout_exercises
INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, rest_time, weight, order_index, created_at, updated_at)
SELECT 
    e.workout_id,
    en.id,
    e.sets,
    e.reps,
    e.rest_time,
    e.weight,
    e.order_index,
    e.created_at,
    e.updated_at
FROM exercises e
JOIN exercises_new en ON e.name = en.name AND e.thumbnail_url = en.thumbnail_url
ON CONFLICT (workout_id, exercise_id) DO NOTHING;

-- Step 5: Update set_records to reference workout_exercises
ALTER TABLE set_records ADD COLUMN workout_exercise_id UUID REFERENCES workout_exercises(id) ON DELETE RESTRICT;

-- Migrate set_records data
UPDATE set_records sr
SET workout_exercise_id = we.id
FROM workout_exercises we
JOIN sessions s ON we.workout_id = s.workout_id
WHERE sr.session_id = s.id
AND sr.exercise_id = we.exercise_id;

-- Step 6: Drop old constraint and column from set_records
ALTER TABLE set_records DROP CONSTRAINT set_records_exercise_id_fkey;
ALTER TABLE set_records DROP COLUMN exercise_id;

-- Step 7: Drop old exercises table and rename
DROP TABLE exercises CASCADE;
ALTER TABLE exercises_new RENAME TO exercises;
ALTER INDEX idx_exercises_new_name RENAME TO idx_exercises_name;
ALTER INDEX idx_exercises_new_muscles RENAME TO idx_exercises_muscles;

-- Step 8: Update UNIQUE constraint on set_records
ALTER TABLE set_records DROP CONSTRAINT IF EXISTS set_records_session_id_exercise_id_set_number_key;
ALTER TABLE set_records ADD CONSTRAINT set_records_session_exercise_set_unique 
    UNIQUE (session_id, workout_exercise_id, set_number);
