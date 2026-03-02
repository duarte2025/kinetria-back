package statistics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// maxPeriodDays is the maximum allowed period for statistics queries (2 years).
const maxPeriodDays = 730

// GetOverviewInput holds the input parameters for GetOverviewUC.
type GetOverviewInput struct {
	UserID    uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}

// GetOverviewUC retrieves aggregated workout statistics for a user.
type GetOverviewUC struct {
	sessionRepo   ports.SessionRepository
	setRecordRepo ports.SetRecordRepository
}

// NewGetOverviewUC creates a new GetOverviewUC.
func NewGetOverviewUC(sessionRepo ports.SessionRepository, setRecordRepo ports.SetRecordRepository) *GetOverviewUC {
	return &GetOverviewUC{sessionRepo: sessionRepo, setRecordRepo: setRecordRepo}
}

// Execute computes overview statistics for the given user and period.
// If StartDate/EndDate are nil, defaults to the last 30 days.
// Returns an error if the period is invalid or exceeds 2 years.
func (uc *GetOverviewUC) Execute(ctx context.Context, input GetOverviewInput) (*OverviewStats, error) {
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

	// Fetch session stats
	sessionStats, err := uc.sessionRepo.GetStatsByUserAndPeriod(ctx, input.UserID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get session stats: %w", err)
	}

	// Fetch set record stats
	setStats, err := uc.setRecordRepo.GetTotalSetsRepsVolume(ctx, input.UserID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get set record stats: %w", err)
	}

	// Fetch streak data (last 365 days)
	streakDates, err := uc.sessionRepo.GetSessionsForStreak(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get streak data: %w", err)
	}

	// Calculate streaks
	currentStreak, longestStreak := calculateStreaks(streakDates, now)

	// Calculate average per week
	days := end.Sub(start).Hours() / 24
	weeks := days / 7.0
	var avgPerWeek float64
	if weeks > 0 {
		avgPerWeek = float64(sessionStats.TotalWorkouts) / weeks
	}

	return &OverviewStats{
		StartDate:        start,
		EndDate:          end,
		TotalWorkouts:    sessionStats.TotalWorkouts,
		AveragePerWeek:   avgPerWeek,
		TotalTimeMinutes: sessionStats.TotalTime,
		TotalSets:        setStats.TotalSets,
		TotalReps:        setStats.TotalReps,
		TotalVolume:      setStats.TotalVolume,
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
	}, nil
}

// calculateStreaks computes currentStreak and longestStreak from a list of workout dates.
// dates must be sorted in descending order (most recent first), deduplicated by day.
// now is used as the reference point for currentStreak.
func calculateStreaks(dates []time.Time, now time.Time) (current, longest int) {
	if len(dates) == 0 {
		return 0, 0
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Build a set of dates for quick lookup
	dateSet := make(map[string]bool, len(dates))
	for _, d := range dates {
		day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		dateSet[day.Format("2006-01-02")] = true
	}

	// Current streak: count consecutive days back from today (or yesterday)
	currentStreak := 0
	checkDay := today
	// If no workout today, start from yesterday
	if !dateSet[checkDay.Format("2006-01-02")] {
		checkDay = today.AddDate(0, 0, -1)
	}
	for dateSet[checkDay.Format("2006-01-02")] {
		currentStreak++
		checkDay = checkDay.AddDate(0, 0, -1)
	}

	// Longest streak: iterate all dates in descending order
	longestStreak := 0
	streak := 0
	var prevDay time.Time
	for _, d := range dates {
		day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		if prevDay.IsZero() {
			streak = 1
		} else {
			expected := prevDay.AddDate(0, 0, -1)
			if day.Equal(expected) {
				streak++
			} else {
				streak = 1
			}
		}
		prevDay = day
		if streak > longestStreak {
			longestStreak = streak
		}
	}

	return currentStreak, longestStreak
}
