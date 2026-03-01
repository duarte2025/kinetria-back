package workouts

import "github.com/google/uuid"

// WorkoutExerciseInput represents a single exercise in a workout for create/update operations.
type WorkoutExerciseInput struct {
	ExerciseID uuid.UUID
	Sets       int
	Reps       string
	RestTime   int
	Weight     *int
	OrderIndex int
}

// CreateWorkoutInput contains the data needed to create a new workout.
type CreateWorkoutInput struct {
	Name        string
	Description *string
	Type        string
	Intensity   string
	Duration    int
	ImageURL    *string
	Exercises   []WorkoutExerciseInput
}

// UpdateWorkoutInput contains the data needed to update an existing workout.
// Nil pointer fields are left unchanged.
type UpdateWorkoutInput struct {
	Name        *string
	Description *string
	Type        *string
	Intensity   *string
	Duration    *int
	ImageURL    *string
	Exercises   []WorkoutExerciseInput
}
