package workouts

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// validWorkoutTypes defines the allowed workout types.
var validWorkoutTypes = map[string]bool{
	"FORÇA":           true,
	"HIPERTROFIA":     true,
	"MOBILIDADE":      true,
	"CONDICIONAMENTO": true,
}

// validWorkoutIntensities defines the allowed workout intensities.
var validWorkoutIntensities = map[string]bool{
	"BAIXA":    true,
	"MODERADA": true,
	"ALTA":     true,
}

// CreateWorkoutUC handles the business logic for creating a workout.
type CreateWorkoutUC struct {
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.ExerciseRepository
}

// NewCreateWorkoutUC creates a new CreateWorkoutUC.
func NewCreateWorkoutUC(workoutRepo ports.WorkoutRepository, exerciseRepo ports.ExerciseRepository) *CreateWorkoutUC {
	return &CreateWorkoutUC{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
	}
}

// Execute creates a new workout for the given user.
func (uc *CreateWorkoutUC) Execute(ctx context.Context, userID uuid.UUID, input CreateWorkoutInput) (*entities.Workout, error) {
	// Validate name
	if len(input.Name) < 3 || len(input.Name) > 255 {
		return nil, fmt.Errorf("%w: name must be between 3 and 255 characters", domerrors.ErrMalformedParameters)
	}

	// Validate type
	if !validWorkoutTypes[input.Type] {
		return nil, fmt.Errorf("%w: type must be one of FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO", domerrors.ErrMalformedParameters)
	}

	// Validate intensity
	if !validWorkoutIntensities[input.Intensity] {
		return nil, fmt.Errorf("%w: intensity must be one of BAIXA, MODERADA, ALTA", domerrors.ErrMalformedParameters)
	}

	// Validate duration
	if input.Duration < 1 || input.Duration > 300 {
		return nil, fmt.Errorf("%w: duration must be between 1 and 300 minutes", domerrors.ErrMalformedParameters)
	}

	// Validate exercises count
	if len(input.Exercises) == 0 {
		return nil, fmt.Errorf("%w: at least one exercise is required", domerrors.ErrMalformedParameters)
	}
	if len(input.Exercises) > 20 {
		return nil, fmt.Errorf("%w: maximum of 20 exercises allowed", domerrors.ErrMalformedParameters)
	}

	// Validate each exercise + check for duplicate orderIndex
	orderIndexes := make(map[int]bool)
	for i, ex := range input.Exercises {
		if ex.Sets < 1 || ex.Sets > 10 {
			return nil, fmt.Errorf("%w: exercise %d: sets must be between 1 and 10", domerrors.ErrMalformedParameters, i+1)
		}
		if ex.RestTime < 0 || ex.RestTime > 600 {
			return nil, fmt.Errorf("%w: exercise %d: restTime must be between 0 and 600 seconds", domerrors.ErrMalformedParameters, i+1)
		}
		if orderIndexes[ex.OrderIndex] {
			return nil, fmt.Errorf("%w: duplicate orderIndex %d", domerrors.ErrMalformedParameters, ex.OrderIndex)
		}
		orderIndexes[ex.OrderIndex] = true

		// Validate exercise exists in library
		exercise, err := uc.exerciseRepo.GetByID(ctx, ex.ExerciseID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate exercise: %w", err)
		}
		if exercise == nil {
			return nil, fmt.Errorf("%w: exercise with id '%s' not found", domerrors.ErrMalformedParameters, ex.ExerciseID)
		}
	}

	// Build workout entity
	now := time.Now().UTC()
	workoutID := uuid.New()

	description := ""
	if input.Description != nil {
		description = *input.Description
	}
	imageURL := ""
	if input.ImageURL != nil {
		imageURL = *input.ImageURL
	}

	workout := entities.Workout{
		ID:          workoutID,
		UserID:      userID,
		Name:        input.Name,
		Description: description,
		Type:        input.Type,
		Intensity:   input.Intensity,
		Duration:    input.Duration,
		ImageURL:    imageURL,
		CreatedBy:   &userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Build workout exercises
	workoutExercises := make([]entities.WorkoutExercise, len(input.Exercises))
	for i, ex := range input.Exercises {
		weight := 0
		if ex.Weight != nil {
			weight = *ex.Weight
		}
		workoutExercises[i] = entities.WorkoutExercise{
			ID:         uuid.New(),
			WorkoutID:  workoutID,
			ExerciseID: ex.ExerciseID,
			Sets:       ex.Sets,
			Reps:       ex.Reps,
			RestTime:   ex.RestTime,
			Weight:     weight,
			OrderIndex: ex.OrderIndex,
		}
	}

	// Persist
	if err := uc.workoutRepo.Create(ctx, workout, workoutExercises); err != nil {
		return nil, fmt.Errorf("failed to create workout: %w", err)
	}

	return &workout, nil
}
