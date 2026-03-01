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

// UpdateWorkoutUC handles the business logic for updating a workout.
type UpdateWorkoutUC struct {
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.ExerciseRepository
}

// NewUpdateWorkoutUC creates a new UpdateWorkoutUC.
func NewUpdateWorkoutUC(workoutRepo ports.WorkoutRepository, exerciseRepo ports.ExerciseRepository) *UpdateWorkoutUC {
	return &UpdateWorkoutUC{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
	}
}

// Execute updates an existing workout owned by the given user.
func (uc *UpdateWorkoutUC) Execute(ctx context.Context, userID, workoutID uuid.UUID, input UpdateWorkoutInput) (*entities.Workout, error) {
	// Fetch workout
	workout, err := uc.workoutRepo.GetByIDOnly(ctx, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return nil, domerrors.ErrWorkoutNotFound
	}

	// Check if template (created_by IS NULL)
	if workout.CreatedBy == nil {
		return nil, fmt.Errorf("%w: cannot modify template workouts", domerrors.ErrCannotModifyTemplate)
	}

	// Check ownership
	if *workout.CreatedBy != userID {
		return nil, domerrors.ErrForbidden
	}

	// Apply updates (partial)
	if input.Name != nil {
		if len(*input.Name) < 3 || len(*input.Name) > 255 {
			return nil, fmt.Errorf("%w: name must be between 3 and 255 characters", domerrors.ErrMalformedParameters)
		}
		workout.Name = *input.Name
	}
	if input.Description != nil {
		workout.Description = *input.Description
	}
	if input.Type != nil {
		if !validWorkoutTypes[*input.Type] {
			return nil, fmt.Errorf("%w: type must be one of FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO", domerrors.ErrMalformedParameters)
		}
		workout.Type = *input.Type
	}
	if input.Intensity != nil {
		if !validWorkoutIntensities[*input.Intensity] {
			return nil, fmt.Errorf("%w: intensity must be one of BAIXA, MODERADA, ALTA", domerrors.ErrMalformedParameters)
		}
		workout.Intensity = *input.Intensity
	}
	if input.Duration != nil {
		if *input.Duration < 1 || *input.Duration > 300 {
			return nil, fmt.Errorf("%w: duration must be between 1 and 300 minutes", domerrors.ErrMalformedParameters)
		}
		workout.Duration = *input.Duration
	}
	if input.ImageURL != nil {
		workout.ImageURL = *input.ImageURL
	}

	// Validate and build exercises (if provided)
	var workoutExercises []entities.WorkoutExercise
	if len(input.Exercises) > 0 {
		if len(input.Exercises) > 20 {
			return nil, fmt.Errorf("%w: maximum of 20 exercises allowed", domerrors.ErrMalformedParameters)
		}

		orderIndexes := make(map[int]bool)
		workoutExercises = make([]entities.WorkoutExercise, len(input.Exercises))
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

			exercise, err := uc.exerciseRepo.GetByID(ctx, ex.ExerciseID)
			if err != nil {
				return nil, fmt.Errorf("failed to validate exercise: %w", err)
			}
			if exercise == nil {
				return nil, fmt.Errorf("%w: exercise with id '%s' not found", domerrors.ErrMalformedParameters, ex.ExerciseID)
			}

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
	}

	workout.UpdatedAt = time.Now().UTC()

	if err := uc.workoutRepo.Update(ctx, *workout, workoutExercises); err != nil {
		return nil, fmt.Errorf("failed to update workout: %w", err)
	}

	return workout, nil
}
