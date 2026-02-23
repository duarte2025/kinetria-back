-- Migration 005: Create set_records table
CREATE TABLE IF NOT EXISTS set_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    set_number INT NOT NULL CHECK (set_number >= 1),
    weight INT NOT NULL DEFAULT 0 CHECK (weight >= 0),
    reps INT NOT NULL DEFAULT 0 CHECK (reps >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'skipped')),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, exercise_id, set_number)
);

CREATE INDEX IF NOT EXISTS idx_set_records_session_id ON set_records(session_id);
CREATE INDEX IF NOT EXISTS idx_set_records_exercise_id ON set_records(exercise_id);
CREATE INDEX IF NOT EXISTS idx_set_records_session_exercise ON set_records(session_id, exercise_id);
