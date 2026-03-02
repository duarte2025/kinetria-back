package workouts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// DeleteWorkoutUC handles the business logic for deleting a workout.
type DeleteWorkoutUC struct {
	workoutRepo ports.WorkoutRepository
}

// NewDeleteWorkoutUC creates a new DeleteWorkoutUC.
func NewDeleteWorkoutUC(workoutRepo ports.WorkoutRepository) *DeleteWorkoutUC {
	return &DeleteWorkoutUC{workoutRepo: workoutRepo}
}

// Execute soft-deletes a workout owned by the given user.
func (uc *DeleteWorkoutUC) Execute(ctx context.Context, userID, workoutID uuid.UUID) error {
	// Fetch workout
	workout, err := uc.workoutRepo.GetByIDOnly(ctx, workoutID)
	if err != nil {
		return fmt.Errorf("failed to get workout: %w", err)
	}
	if workout == nil {
		return domerrors.ErrWorkoutNotFound
	}

	// Check if template (created_by IS NULL)
	if workout.CreatedBy == nil {
		return fmt.Errorf("%w: cannot delete template workouts", domerrors.ErrCannotModifyTemplate)
	}

	// Check ownership
	if *workout.CreatedBy != userID {
		return domerrors.ErrForbidden
	}

	// Check for active sessions
	hasActive, err := uc.workoutRepo.HasActiveSessions(ctx, workoutID)
	if err != nil {
		return fmt.Errorf("failed to check active sessions: %w", err)
	}
	if hasActive {
		return domerrors.ErrWorkoutHasActiveSessions
	}

	// Soft delete
	if err := uc.workoutRepo.Delete(ctx, workoutID); err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	return nil
}
