package workouts

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type ListWorkoutsInput struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
}

type ListWorkoutsOutput struct {
	Workouts   []entities.Workout
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListWorkoutsUC struct {
	repo ports.WorkoutRepository
}

func NewListWorkoutsUC(repo ports.WorkoutRepository) *ListWorkoutsUC {
	return &ListWorkoutsUC{repo: repo}
}

func (uc *ListWorkoutsUC) Execute(ctx context.Context, input ListWorkoutsInput) (ListWorkoutsOutput, error) {
	// Apply defaults
	page := input.Page
	if page <= 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	// Validate
	if input.UserID == uuid.Nil {
		return ListWorkoutsOutput{}, fmt.Errorf("userID is required")
	}
	if input.Page < 0 {
		return ListWorkoutsOutput{}, fmt.Errorf("page must be greater than or equal to 1")
	}
	if input.PageSize < 0 || input.PageSize > 100 {
		return ListWorkoutsOutput{}, fmt.Errorf("pageSize must be between 1 and 100")
	}

	offset := (page - 1) * pageSize

	workouts, total, err := uc.repo.ListByUserID(ctx, input.UserID, offset, pageSize)
	if err != nil {
		return ListWorkoutsOutput{}, fmt.Errorf("failed to list workouts: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return ListWorkoutsOutput{
		Workouts:   workouts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
