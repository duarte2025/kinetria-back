-- Migration 011: Expand exercises table with library metadata fields

ALTER TABLE exercises ADD COLUMN IF NOT EXISTS instructions TEXT;
ALTER TABLE exercises ADD COLUMN IF NOT EXISTS tips TEXT;
ALTER TABLE exercises ADD COLUMN IF NOT EXISTS difficulty VARCHAR(50);
ALTER TABLE exercises ADD COLUMN IF NOT EXISTS equipment VARCHAR(100);
ALTER TABLE exercises ADD COLUMN IF NOT EXISTS video_url TEXT;

CREATE INDEX IF NOT EXISTS idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX IF NOT EXISTS idx_exercises_equipment ON exercises(equipment);
