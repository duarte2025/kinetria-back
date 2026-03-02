package statistics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// GetFrequencyInput holds the input parameters for GetFrequencyUC.
type GetFrequencyInput struct {
	UserID    uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}

// GetFrequencyUC retrieves workout frequency data for a user.
type GetFrequencyUC struct {
	sessionRepo ports.SessionRepository
}

// NewGetFrequencyUC creates a new GetFrequencyUC.
func NewGetFrequencyUC(sessionRepo ports.SessionRepository) *GetFrequencyUC {
	return &GetFrequencyUC{sessionRepo: sessionRepo}
}

// Execute returns workout frequency (count per day) for the given user and period.
// If StartDate/EndDate are nil, defaults to the last 365 days.
// All days in the period are returned; days without workouts have Count=0.
func (uc *GetFrequencyUC) Execute(ctx context.Context, input GetFrequencyInput) ([]FrequencyData, error) {
	now := time.Now().UTC()

	// Apply defaults
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	start := end.AddDate(0, 0, -364) // 365 days total inclusive
	if input.EndDate != nil {
		d := input.EndDate.UTC()
		end = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	}
	if input.StartDate != nil {
		d := input.StartDate.UTC()
		start = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Validate period
	if start.After(end) {
		return nil, fmt.Errorf("startDate must be before or equal to endDate")
	}
	if end.Sub(start).Hours()/24 > maxPeriodDays {
		return nil, fmt.Errorf("period must not exceed %d days", maxPeriodDays)
	}

	// Fetch data from DB (only days with workouts)
	dbRows, err := uc.sessionRepo.GetFrequencyByUserAndPeriod(ctx, input.UserID, start, end.Add(24*time.Hour-time.Second))
	if err != nil {
		return nil, fmt.Errorf("get frequency: %w", err)
	}

	// Build a map for quick lookup
	countByDate := make(map[string]int, len(dbRows))
	for _, row := range dbRows {
		key := row.Date.Format("2006-01-02")
		countByDate[key] = row.Count
	}

	// Fill all days in period
	totalDays := int(end.Sub(start).Hours()/24) + 1
	result := make([]FrequencyData, 0, totalDays)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		result = append(result, FrequencyData{
			Date:  d,
			Count: countByDate[key],
		})
	}
	return result, nil
}
