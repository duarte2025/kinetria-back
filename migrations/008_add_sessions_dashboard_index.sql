-- Migration 008: Add composite index to optimize dashboard sessions query
-- Covers user_id + status + started_at to avoid table scan in GetCompletedSessionsByDateRange
CREATE INDEX IF NOT EXISTS idx_sessions_user_status_started
ON sessions(user_id, status, started_at DESC);
