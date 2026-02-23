package entities

import (
	"time"

	"github.com/google/uuid"
)

type SetRecordID = uuid.UUID

type SetRecord struct {
	ID         SetRecordID
	SessionID  SessionID
	ExerciseID ExerciseID
	SetNumber  int
	Weight     int
	Reps       int
	Status     string
	RecordedAt time.Time
}
