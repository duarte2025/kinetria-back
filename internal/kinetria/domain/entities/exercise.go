package entities

import (
	"github.com/google/uuid"
)

type ExerciseID = uuid.UUID

// Exercise represents a shared exercise from the library.
// Library fields (Description, Instructions, etc.) are populated when fetching from the library.
// Workout-specific fields (Sets, Reps, etc.) are populated when fetching exercises for a workout.
type Exercise struct {
	ID           ExerciseID
	Name         string
	ThumbnailURL string
	Muscles      []string

	// Library metadata fields (nullable — populated from exercise library endpoints)
	Description  *string
	Instructions *string
	Tips         *string
	Difficulty   *string
	Equipment    *string
	VideoURL     *string

	// Workout-specific configuration (from workout_exercises)
	Sets       int
	Reps       string
	RestTime   int
	Weight     int
	OrderIndex int
}
