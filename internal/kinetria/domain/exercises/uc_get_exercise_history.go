package exercises

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// GetExerciseHistoryInput holds parameters for fetching a user's exercise history.
type GetExerciseHistoryInput struct {
	ExerciseID uuid.UUID
	UserID     uuid.UUID
	Page       int
	PageSize   int
}

// GetExerciseHistoryOutput holds the result of the history query.
type GetExerciseHistoryOutput struct {
	Entries    []*ports.ExerciseHistoryEntry
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// GetExerciseHistoryUC is the use case for retrieving a user's history for a specific exercise.
type GetExerciseHistoryUC struct {
	exerciseRepo ports.ExerciseRepository
}

// NewGetExerciseHistoryUC creates a new GetExerciseHistoryUC.
func NewGetExerciseHistoryUC(exerciseRepo ports.ExerciseRepository) *GetExerciseHistoryUC {
	return &GetExerciseHistoryUC{exerciseRepo: exerciseRepo}
}

// Execute retrieves the paginated history of a user performing a specific exercise.
// Entries are ordered from most recent to oldest.
// Returns errors.ErrExerciseNotFound if the exercise does not exist.
func (uc *GetExerciseHistoryUC) Execute(ctx context.Context, input GetExerciseHistoryInput) (GetExerciseHistoryOutput, error) {
	// Validate
	if input.Page < 1 {
		return GetExerciseHistoryOutput{}, fmt.Errorf("page must be >= 1")
	}
	if input.PageSize < 1 {
		return GetExerciseHistoryOutput{}, fmt.Errorf("pageSize must be >= 1")
	}
	if input.PageSize > 100 {
		return GetExerciseHistoryOutput{}, fmt.Errorf("pageSize must be <= 100")
	}

	// Verify exercise exists
	exercise, err := uc.exerciseRepo.GetByID(ctx, input.ExerciseID)
	if err != nil {
		return GetExerciseHistoryOutput{}, fmt.Errorf("failed to check exercise: %w", err)
	}
	if exercise == nil {
		return GetExerciseHistoryOutput{}, errors.ErrExerciseNotFound
	}

	entries, total, err := uc.exerciseRepo.GetHistory(ctx, input.UserID, input.ExerciseID, input.Page, input.PageSize)
	if err != nil {
		return GetExerciseHistoryOutput{}, fmt.Errorf("failed to get exercise history: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(input.PageSize)))
	}

	return GetExerciseHistoryOutput{
		Entries:    entries,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
