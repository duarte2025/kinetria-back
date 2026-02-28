-- Migration 003: Create exercises table
CREATE TABLE IF NOT EXISTS exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    thumbnail_url VARCHAR(500) NOT NULL DEFAULT '/assets/exercises/generic.png',
    sets INT NOT NULL DEFAULT 1 CHECK (sets >= 1),
    reps VARCHAR(20) NOT NULL DEFAULT '',
    muscles JSONB NOT NULL DEFAULT '[]',
    rest_time INT NOT NULL DEFAULT 60 CHECK (rest_time >= 0),
    weight INT NOT NULL DEFAULT 0 CHECK (weight >= 0),
    order_index INT NOT NULL DEFAULT 0 CHECK (order_index >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exercises_workout_id ON exercises(workout_id);
CREATE INDEX IF NOT EXISTS idx_exercises_order ON exercises(workout_id, order_index);
CREATE INDEX IF NOT EXISTS idx_exercises_muscles ON exercises USING GIN(muscles);
