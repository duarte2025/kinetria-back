package repositories

import (
	"context"
	"database/sql"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
	"github.com/sqlc-dev/pqtype"
)

// AuditLogRepository implements ports.AuditLogRepository using PostgreSQL via SQLC.
type AuditLogRepository struct {
	q *queries.Queries
}

// NewAuditLogRepository creates a new AuditLogRepository backed by the provided *sql.DB.
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{q: queries.New(db)}
}

// Append inserts a new audit log entry into the database.
func (r *AuditLogRepository) Append(ctx context.Context, entry *entities.AuditLog) error {
	return r.q.AppendAuditLog(ctx, queries.AppendAuditLogParams{
		ID:         entry.ID,
		UserID:     entry.UserID,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Action:     entry.Action,
		ActionData: pqtype.NullRawMessage{
			RawMessage: entry.ActionData,
			Valid:      len(entry.ActionData) > 0,
		},
		OccurredAt: entry.OccurredAt,
		IpAddress: sql.NullString{
			String: entry.IPAddress,
			Valid:  entry.IPAddress != "",
		},
		UserAgent: sql.NullString{
			String: entry.UserAgent,
			Valid:  entry.UserAgent != "",
		},
	})
}
