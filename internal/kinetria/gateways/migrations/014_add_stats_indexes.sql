-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_sessions_user_stats
    ON sessions (user_id, started_at)
    WHERE status = 'completed';

CREATE INDEX IF NOT EXISTS idx_set_records_session_status
    ON set_records (session_id, status)
    WHERE status = 'completed';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sessions_user_stats;
DROP INDEX IF EXISTS idx_set_records_session_status;
-- +goose StatementEnd
