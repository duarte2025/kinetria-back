package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

type SessionID = uuid.UUID

type Session struct {
	ID         SessionID
	UserID     UserID
	WorkoutID  WorkoutID
	Status     vos.SessionStatus
	Notes      string
	StartedAt  time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
