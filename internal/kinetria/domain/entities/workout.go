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
	CreatedBy   *uuid.UUID
	DeletedAt   *time.Time
}

// WorkoutExercise represents the association between a workout and an exercise.
type WorkoutExercise struct {
	ID         uuid.UUID
	WorkoutID  uuid.UUID
	ExerciseID uuid.UUID
	Sets       int
	Reps       string
	RestTime   int
	Weight     int
	OrderIndex int
}
