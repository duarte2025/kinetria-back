package entities

import (
	"github.com/google/uuid"
)

type ExerciseID = uuid.UUID

// Exercise represents a shared exercise from the library
type Exercise struct {
	ID           ExerciseID
	Name         string
	ThumbnailURL string
	Muscles      []string
	// Workout-specific configuration (from workout_exercises)
	Sets       int
	Reps       string
	RestTime   int
	Weight     int
	OrderIndex int
}
