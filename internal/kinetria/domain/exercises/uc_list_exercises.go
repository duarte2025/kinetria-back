package exercises

import (
	"context"
	"fmt"
	"math"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// ListExercisesInput holds parameters for listing exercises from the library.
type ListExercisesInput struct {
	Filters  ports.ExerciseFilters
	Page     int
	PageSize int
}

// ListExercisesOutput holds the result of listing exercises.
type ListExercisesOutput struct {
	Exercises  []*entities.Exercise
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// ListExercisesUC is the use case for listing exercises from the library with optional filters.
type ListExercisesUC struct {
	exerciseRepo ports.ExerciseRepository
}

// NewListExercisesUC creates a new ListExercisesUC.
func NewListExercisesUC(exerciseRepo ports.ExerciseRepository) *ListExercisesUC {
	return &ListExercisesUC{exerciseRepo: exerciseRepo}
}

// Execute retrieves a paginated list of exercises matching the provided filters.
// Returns an error if pagination parameters are invalid.
func (uc *ListExercisesUC) Execute(ctx context.Context, input ListExercisesInput) (ListExercisesOutput, error) {
	// Validate
	if input.Page < 1 {
		return ListExercisesOutput{}, fmt.Errorf("page must be >= 1")
	}
	if input.PageSize < 1 {
		return ListExercisesOutput{}, fmt.Errorf("pageSize must be >= 1")
	}
	if input.PageSize > 100 {
		return ListExercisesOutput{}, fmt.Errorf("pageSize must be <= 100")
	}

	exercises, total, err := uc.exerciseRepo.List(ctx, input.Filters, input.Page, input.PageSize)
	if err != nil {
		return ListExercisesOutput{}, fmt.Errorf("failed to list exercises: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(input.PageSize)))
	}

	return ListExercisesOutput{
		Exercises:  exercises,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
