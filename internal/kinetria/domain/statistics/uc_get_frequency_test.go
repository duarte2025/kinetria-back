package statistics

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks for GetFrequencyUC ---

type mockSessionRepoFreq struct {
	frequencyResult []ports.FrequencyData
	frequencyErr    error
}

func (m *mockSessionRepoFreq) Create(_ context.Context, _ *entities.Session) error {
	return nil
}
func (m *mockSessionRepoFreq) FindActiveByUserID(_ context.Context, _ uuid.UUID) (*entities.Session, error) {
	return nil, nil
}
func (m *mockSessionRepoFreq) FindByID(_ context.Context, _ uuid.UUID) (*entities.Session, error) {
	return nil, nil
}
func (m *mockSessionRepoFreq) UpdateStatus(_ context.Context, _ uuid.UUID, _ string, _ *time.Time, _ string) (bool, error) {
	return false, nil
}
func (m *mockSessionRepoFreq) GetCompletedSessionsByUserAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]entities.Session, error) {
	return nil, nil
}
func (m *mockSessionRepoFreq) GetStatsByUserAndPeriod(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SessionStats, error) {
	return nil, nil
}
func (m *mockSessionRepoFreq) GetFrequencyByUserAndPeriod(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]ports.FrequencyData, error) {
	return m.frequencyResult, m.frequencyErr
}
func (m *mockSessionRepoFreq) GetSessionsForStreak(_ context.Context, _ uuid.UUID) ([]time.Time, error) {
	return nil, nil
}

// --- Tests ---

func TestGetFrequencyUC_Execute(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()

	t.Run("happy path: returns all days including empty ones", func(t *testing.T) {
		// 7-day period
		startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -6)
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		// Only 2 of 7 days have workouts
		workoutDay1 := startDate.AddDate(0, 0, 2)
		workoutDay2 := startDate.AddDate(0, 0, 5)

		sessRepo := &mockSessionRepoFreq{
			frequencyResult: []ports.FrequencyData{
				{Date: workoutDay1, Count: 1},
				{Date: workoutDay2, Count: 2},
			},
		}

		uc := NewGetFrequencyUC(sessRepo)
		result, err := uc.Execute(context.Background(), GetFrequencyInput{
			UserID:    userID,
			StartDate: &startDate,
			EndDate:   &endDate,
		})

		require.NoError(t, err)
		// Should return all 7 days
		assert.Len(t, result, 7)

		// Verify non-workout days have Count=0
		zeroCount := 0
		for _, d := range result {
			if d.Count == 0 {
				zeroCount++
			}
		}
		assert.Equal(t, 5, zeroCount)

		// Check workout days have correct counts
		for _, d := range result {
			dayStr := d.Date.Format("2006-01-02")
			if dayStr == workoutDay1.Format("2006-01-02") {
				assert.Equal(t, 1, d.Count)
			}
			if dayStr == workoutDay2.Format("2006-01-02") {
				assert.Equal(t, 2, d.Count)
			}
		}
	})

	t.Run("invalid period: startDate > endDate returns error", func(t *testing.T) {
		startDate := now
		endDate := now.AddDate(0, 0, -7)

		sessRepo := &mockSessionRepoFreq{}
		uc := NewGetFrequencyUC(sessRepo)
		_, err := uc.Execute(context.Background(), GetFrequencyInput{
			UserID:    userID,
			StartDate: &startDate,
			EndDate:   &endDate,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "startDate must be before or equal to endDate")
	})

	t.Run("user without workouts: all days have count=0", func(t *testing.T) {
		startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -4)
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		sessRepo := &mockSessionRepoFreq{
			frequencyResult: []ports.FrequencyData{},
		}

		uc := NewGetFrequencyUC(sessRepo)
		result, err := uc.Execute(context.Background(), GetFrequencyInput{
			UserID:    userID,
			StartDate: &startDate,
			EndDate:   &endDate,
		})

		require.NoError(t, err)
		// 5 days (inclusive)
		assert.Len(t, result, 5)
		for _, d := range result {
			assert.Equal(t, 0, d.Count, "expected count=0 for day %s", d.Date.Format("2006-01-02"))
		}
	})
}
