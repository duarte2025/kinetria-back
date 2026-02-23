package entities

import (
	"time"

	"github.com/google/uuid"
)

type WorkoutID = uuid.UUID

type Workout struct {
	ID          WorkoutID
	UserID      UserID
	Name        string
	Description string
	Type        string
	Intensity   string
	Duration    int
	ImageURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
