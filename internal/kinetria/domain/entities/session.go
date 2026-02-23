package entities

import (
	"time"

	"github.com/google/uuid"
)

type SessionID = uuid.UUID

type Session struct {
	ID         SessionID
	UserID     UserID
	WorkoutID  WorkoutID
	Status     string
	Notes      string
	StartedAt  time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
