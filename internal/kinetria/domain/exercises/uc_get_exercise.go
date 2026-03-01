package exercises

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// GetExerciseUC is the use case for retrieving a single exercise by ID.
// If a userID is provided, it also fetches the user's performance stats for the exercise.
type GetExerciseUC struct {
	exerciseRepo ports.ExerciseRepository
}

// NewGetExerciseUC creates a new GetExerciseUC.
func NewGetExerciseUC(exerciseRepo ports.ExerciseRepository) *GetExerciseUC {
	return &GetExerciseUC{exerciseRepo: exerciseRepo}
}

// Execute retrieves the exercise and optionally its user stats.
// userID may be nil for unauthenticated requests — stats will be omitted in that case.
// Returns errors.ErrExerciseNotFound if the exercise does not exist.
func (uc *GetExerciseUC) Execute(ctx context.Context, exerciseID uuid.UUID, userID *uuid.UUID) (*ExerciseWithStats, error) {
	exercise, err := uc.exerciseRepo.GetByID(ctx, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}
	if exercise == nil {
		return nil, errors.ErrExerciseNotFound
	}

	result := &ExerciseWithStats{Exercise: exercise}

	if userID != nil {
		stats, err := uc.exerciseRepo.GetUserStats(ctx, *userID, exerciseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user stats: %w", err)
		}
		result.UserStats = stats
	}

	return result, nil
}
