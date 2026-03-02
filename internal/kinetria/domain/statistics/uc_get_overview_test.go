package statistics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks for GetOverviewUC ---

type mockSessionRepoOverview struct {
	statsResult     *ports.SessionStats
	statsErr        error
	frequencyResult []ports.FrequencyData
	frequencyErr    error
	streakResult    []time.Time
	streakErr       error
}

func (m *mockSessionRepoOverview) Create(_ context.Context, _ *entities.Session) error {
	return nil
}
func (m *mockSessionRepoOverview) FindActiveByUserID(_ context.Context, _ uuid.UUID) (*entities.Session, error) {
	return nil, nil
}
func (m *mockSessionRepoOverview) FindByID(_ context.Context, _ uuid.UUID) (*entities.Session, error) {
	return nil, nil
}
func (m *mockSessionRepoOverview) UpdateStatus(_ context.Context, _ uuid.UUID, _ string, _ *time.Time, _ string) (bool, error) {
	return false, nil
}
func (m *mockSessionRepoOverview) GetCompletedSessionsByUserAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]entities.Session, error) {
	return nil, nil
}
func (m *mockSessionRepoOverview) GetStatsByUserAndPeriod(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SessionStats, error) {
	return m.statsResult, m.statsErr
}
func (m *mockSessionRepoOverview) GetFrequencyByUserAndPeriod(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]ports.FrequencyData, error) {
	return m.frequencyResult, m.frequencyErr
}
func (m *mockSessionRepoOverview) GetSessionsForStreak(_ context.Context, _ uuid.UUID) ([]time.Time, error) {
	return m.streakResult, m.streakErr
}

type mockSetRecordRepoOverview struct {
	statsResult *ports.SetRecordStats
	statsErr    error
}

func (m *mockSetRecordRepoOverview) Create(_ context.Context, _ *entities.SetRecord) error {
	return nil
}
func (m *mockSetRecordRepoOverview) FindBySessionExerciseSet(_ context.Context, _, _ uuid.UUID, _ int) (*entities.SetRecord, error) {
	return nil, nil
}
func (m *mockSetRecordRepoOverview) GetTotalSetsRepsVolume(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SetRecordStats, error) {
	return m.statsResult, m.statsErr
}
func (m *mockSetRecordRepoOverview) GetPersonalRecordsByUser(_ context.Context, _ uuid.UUID) ([]ports.PersonalRecord, error) {
	return nil, nil
}
func (m *mockSetRecordRepoOverview) GetProgressionByUserAndExercise(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _, _ time.Time) ([]ports.ProgressionPoint, error) {
	return nil, nil
}

// --- Tests ---

func TestGetOverviewUC_Execute(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()

	start := now.AddDate(0, 0, -7)
	end := now

	t.Run("happy path: returns stats correctly", func(t *testing.T) {
		sessRepo := &mockSessionRepoOverview{
			statsResult: &ports.SessionStats{TotalWorkouts: 5, TotalTime: 120},
			streakResult: []time.Time{
				now,
				now.AddDate(0, 0, -1),
				now.AddDate(0, 0, -2),
			},
		}
		setRepo := &mockSetRecordRepoOverview{
			statsResult: &ports.SetRecordStats{TotalSets: 20, TotalReps: 100, TotalVolume: 50000},
		}

		uc := NewGetOverviewUC(sessRepo, setRepo)
		result, err := uc.Execute(context.Background(), GetOverviewInput{
			UserID:    userID,
			StartDate: &start,
			EndDate:   &end,
		})

		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalWorkouts)
		assert.Equal(t, 120, result.TotalTimeMinutes)
		assert.Equal(t, 20, result.TotalSets)
		assert.Equal(t, 100, result.TotalReps)
		assert.Equal(t, int64(50000), result.TotalVolume)
		assert.Greater(t, result.AveragePerWeek, 0.0)
	})

	t.Run("invalid period: startDate > endDate returns error", func(t *testing.T) {
		sessRepo := &mockSessionRepoOverview{}
		setRepo := &mockSetRecordRepoOverview{}

		uc := NewGetOverviewUC(sessRepo, setRepo)
		badStart := end
		badEnd := start
		_, err := uc.Execute(context.Background(), GetOverviewInput{
			UserID:    userID,
			StartDate: &badStart,
			EndDate:   &badEnd,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "startDate must be before or equal to endDate")
	})

	t.Run("period too long (>730 days) returns error", func(t *testing.T) {
		sessRepo := &mockSessionRepoOverview{}
		setRepo := &mockSetRecordRepoOverview{}

		uc := NewGetOverviewUC(sessRepo, setRepo)
		longStart := now.AddDate(-3, 0, 0) // 3 years ago
		_, err := uc.Execute(context.Background(), GetOverviewInput{
			UserID:    userID,
			StartDate: &longStart,
			EndDate:   &end,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "period must not exceed")
	})

	t.Run("user with no workouts returns zeros", func(t *testing.T) {
		sessRepo := &mockSessionRepoOverview{
			statsResult:  &ports.SessionStats{TotalWorkouts: 0, TotalTime: 0},
			streakResult: []time.Time{},
		}
		setRepo := &mockSetRecordRepoOverview{
			statsResult: &ports.SetRecordStats{TotalSets: 0, TotalReps: 0, TotalVolume: 0},
		}

		uc := NewGetOverviewUC(sessRepo, setRepo)
		result, err := uc.Execute(context.Background(), GetOverviewInput{
			UserID:    userID,
			StartDate: &start,
			EndDate:   &end,
		})

		require.NoError(t, err)
		assert.Equal(t, 0, result.TotalWorkouts)
		assert.Equal(t, 0, result.TotalSets)
		assert.Equal(t, int64(0), result.TotalVolume)
		assert.Equal(t, 0, result.CurrentStreak)
		assert.Equal(t, 0, result.LongestStreak)
	})

	t.Run("streak calculated correctly: 3 consecutive days", func(t *testing.T) {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		sessRepo := &mockSessionRepoOverview{
			statsResult: &ports.SessionStats{TotalWorkouts: 3, TotalTime: 90},
			streakResult: []time.Time{
				today,
				today.AddDate(0, 0, -1),
				today.AddDate(0, 0, -2),
			},
		}
		setRepo := &mockSetRecordRepoOverview{
			statsResult: &ports.SetRecordStats{},
		}

		uc := NewGetOverviewUC(sessRepo, setRepo)
		result, err := uc.Execute(context.Background(), GetOverviewInput{
			UserID:    userID,
			StartDate: &start,
			EndDate:   &end,
		})

		require.NoError(t, err)
		assert.Equal(t, 3, result.CurrentStreak)
		assert.Equal(t, 3, result.LongestStreak)
	})

	t.Run("sessionRepo error returns error", func(t *testing.T) {
		sessRepo := &mockSessionRepoOverview{
			statsErr: errors.New("db connection error"),
		}
		setRepo := &mockSetRecordRepoOverview{}

		uc := NewGetOverviewUC(sessRepo, setRepo)
		_, err := uc.Execute(context.Background(), GetOverviewInput{
			UserID:    userID,
			StartDate: &start,
			EndDate:   &end,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db connection error")
	})
}
