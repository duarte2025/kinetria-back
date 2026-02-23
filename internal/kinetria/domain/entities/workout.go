package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

type WorkoutID = uuid.UUID

type Workout struct {
	ID          WorkoutID
	UserID      UserID
	Name        string
	Description string
	Status      vos.WorkoutStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
