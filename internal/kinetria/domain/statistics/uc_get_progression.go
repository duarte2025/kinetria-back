package statistics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// GetProgressionInput holds the input parameters for GetProgressionUC.
type GetProgressionInput struct {
	UserID     uuid.UUID
	ExerciseID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
}

// GetProgressionUC retrieves workout progression data for a user.
type GetProgressionUC struct {
	setRecordRepo ports.SetRecordRepository
}

// NewGetProgressionUC creates a new GetProgressionUC.
func NewGetProgressionUC(setRecordRepo ports.SetRecordRepository) *GetProgressionUC {
	return &GetProgressionUC{setRecordRepo: setRecordRepo}
}

// Execute computes progression data for the given user and period.
// If StartDate/EndDate are nil, defaults to the last 30 days.
func (uc *GetProgressionUC) Execute(ctx context.Context, input GetProgressionInput) (*ProgressionData, error) {
	now := time.Now().UTC()

	// Apply defaults
	end := now
	start := now.AddDate(0, 0, -30)
	if input.EndDate != nil {
		end = input.EndDate.UTC()
	}
	if input.StartDate != nil {
		start = input.StartDate.UTC()
	}

	// Validate period
	if start.After(end) {
		return nil, fmt.Errorf("startDate must be before or equal to endDate")
	}
	if end.Sub(start).Hours()/24 > maxPeriodDays {
		return nil, fmt.Errorf("period must not exceed %d days", maxPeriodDays)
	}

	rawPoints, err := uc.setRecordRepo.GetProgressionByUserAndExercise(ctx, input.UserID, input.ExerciseID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get progression: %w", err)
	}

	points := make([]ProgressionPoint, 0, len(rawPoints))
	for i, p := range rawPoints {
		var change float64
		if i > 0 && rawPoints[i-1].MaxWeight > 0 {
			prev := float64(rawPoints[i-1].MaxWeight)
			curr := float64(p.MaxWeight)
			change = (curr - prev) / prev * 100
		}
		points = append(points, ProgressionPoint{
			Date:        p.Date,
			MaxWeight:   p.MaxWeight,
			TotalVolume: p.TotalVolume,
			Change:      change,
		})
	}

	return &ProgressionData{
		ExerciseID: input.ExerciseID,
		Points:     points,
	}, nil
}
