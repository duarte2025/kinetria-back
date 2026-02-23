CREATE TABLE IF NOT EXISTS set_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    set_number INT NOT NULL CHECK (set_number > 0),
    reps INT CHECK (reps >= 0),
    weight_kg DECIMAL(6,2) CHECK (weight_kg >= 0),
    duration_seconds INT CHECK (duration_seconds >= 0),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_set_records_session_id ON set_records(session_id);
CREATE INDEX IF NOT EXISTS idx_set_records_exercise_id ON set_records(exercise_id);
CREATE INDEX IF NOT EXISTS idx_set_records_session_exercise ON set_records(session_id, exercise_id);
