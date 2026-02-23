CREATE TYPE exercise_category AS ENUM ('strength', 'cardio', 'flexibility', 'balance');
CREATE TYPE muscle_group AS ENUM ('chest', 'back', 'legs', 'shoulders', 'arms', 'core', 'full_body');

CREATE TABLE IF NOT EXISTS exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category exercise_category NOT NULL,
    primary_muscle_group muscle_group NOT NULL,
    equipment_required VARCHAR(255),
    difficulty_level INT NOT NULL CHECK (difficulty_level BETWEEN 1 AND 5),
    video_url VARCHAR(500),
    thumbnail_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exercises_category ON exercises(category);
CREATE INDEX IF NOT EXISTS idx_exercises_muscle_group ON exercises(primary_muscle_group);
CREATE INDEX IF NOT EXISTS idx_exercises_difficulty ON exercises(difficulty_level);
