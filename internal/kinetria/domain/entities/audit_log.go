package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

type AuditLogID = uuid.UUID

type AuditLog struct {
	ID         AuditLogID
	UserID     *UserID
	Action     vos.AuditAction
	EntityType string
	EntityID   *uuid.UUID
	Metadata   json.RawMessage
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}
