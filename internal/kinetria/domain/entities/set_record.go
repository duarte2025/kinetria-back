package entities

import (
	"time"

	"github.com/google/uuid"
)

type SetRecordID = uuid.UUID

type SetRecord struct {
	ID              SetRecordID
	SessionID       SessionID
	ExerciseID      ExerciseID
	SetNumber       int
	Reps            *int
	WeightKg        *float64
	DurationSeconds *int
	Notes           string
	CreatedAt       time.Time
}
