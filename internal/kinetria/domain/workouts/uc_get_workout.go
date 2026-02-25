package workouts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// GetWorkoutInput represents the input for getting a specific workout
type GetWorkoutInput struct {
	WorkoutID uuid.UUID
	UserID    uuid.UUID
}

// GetWorkoutOutput represents the output containing workout details and exercises
type GetWorkoutOutput struct {
	Workout   entities.Workout
	Exercises []entities.Exercise
}

// GetWorkoutUC is the use case for retrieving a specific workout with its exercises
type GetWorkoutUC struct {
	repo ports.WorkoutRepository
}

// NewGetWorkoutUC creates a new instance of GetWorkoutUC
func NewGetWorkoutUC(repo ports.WorkoutRepository) *GetWorkoutUC {
	return &GetWorkoutUC{repo: repo}
}

// Execute retrieves a workout by ID, validating ownership and input parameters
func (uc *GetWorkoutUC) Execute(ctx context.Context, input GetWorkoutInput) (GetWorkoutOutput, error) {
	// Validate input
	if input.WorkoutID == uuid.Nil {
		return GetWorkoutOutput{}, fmt.Errorf("workoutId cannot be empty")
	}
	if input.UserID == uuid.Nil {
		return GetWorkoutOutput{}, fmt.Errorf("userId cannot be empty")
	}

	// Fetch workout and exercises (ownership verified by repository)
	workout, exercises, err := uc.repo.GetByID(ctx, input.WorkoutID, input.UserID)
	if err != nil {
		return GetWorkoutOutput{}, fmt.Errorf("failed to get workout: %w", err)
	}

	// Check if workout was found
	if workout == nil {
		return GetWorkoutOutput{}, fmt.Errorf("workout with id '%s' not found", input.WorkoutID.String())
	}

	return GetWorkoutOutput{
		Workout:   *workout,
		Exercises: exercises,
	}, nil
}
