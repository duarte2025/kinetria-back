package entities

import (
	"time"

	"github.com/google/uuid"
)

type ExerciseID = uuid.UUID

type Exercise struct {
	ID           ExerciseID
	WorkoutID    WorkoutID
	Name         string
	ThumbnailURL string
	Sets         int
	Reps         string
	Muscles      []string
	RestTime     int
	Weight       int
	OrderIndex   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
