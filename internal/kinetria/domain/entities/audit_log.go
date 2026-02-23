package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLogID = uuid.UUID

type AuditLog struct {
	ID         AuditLogID
	UserID     UserID
	EntityType string
	EntityID   uuid.UUID
	Action     string
	ActionData json.RawMessage
	OccurredAt time.Time
	IPAddress  string
	UserAgent  string
}
