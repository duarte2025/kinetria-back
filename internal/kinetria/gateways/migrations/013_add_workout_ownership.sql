ALTER TABLE workouts 
ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE CASCADE,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_workouts_created_by ON workouts(created_by);
CREATE INDEX IF NOT EXISTS idx_workouts_deleted_at ON workouts(deleted_at) WHERE deleted_at IS NULL;
