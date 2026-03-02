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

// --- Mocks for GetProgressionUC ---

type mockSetRecordRepoProgression struct {
	progressionResult []ports.ProgressionPoint
	progressionErr    error
}

func (m *mockSetRecordRepoProgression) Create(_ context.Context, _ *entities.SetRecord) error {
	return nil
}
func (m *mockSetRecordRepoProgression) FindBySessionExerciseSet(_ context.Context, _, _ uuid.UUID, _ int) (*entities.SetRecord, error) {
	return nil, nil
}
func (m *mockSetRecordRepoProgression) GetTotalSetsRepsVolume(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SetRecordStats, error) {
	return nil, nil
}
func (m *mockSetRecordRepoProgression) GetPersonalRecordsByUser(_ context.Context, _ uuid.UUID) ([]ports.PersonalRecord, error) {
	return nil, nil
}
func (m *mockSetRecordRepoProgression) GetProgressionByUserAndExercise(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _, _ time.Time) ([]ports.ProgressionPoint, error) {
	return m.progressionResult, m.progressionErr
}

// --- Tests ---

func TestGetProgressionUC_Execute(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()
	exerciseID := uuid.New()

	start := now.AddDate(0, 0, -30)
	end := now

	t.Run("happy path: returns points with change% calculated", func(t *testing.T) {
		day1 := now.AddDate(0, 0, -10)
		day2 := now.AddDate(0, 0, -5)
		day3 := now.AddDate(0, 0, -1)

		setRepo := &mockSetRecordRepoProgression{
			progressionResult: []ports.ProgressionPoint{
				{Date: day1, MaxWeight: 100000, TotalVolume: 500000},
				{Date: day2, MaxWeight: 110000, TotalVolume: 550000},
				{Date: day3, MaxWeight: 121000, TotalVolume: 605000},
			},
		}

		uc := NewGetProgressionUC(setRepo)
		result, err := uc.Execute(context.Background(), GetProgressionInput{
			UserID:     userID,
			ExerciseID: &exerciseID,
			StartDate:  &start,
			EndDate:    &end,
		})

		require.NoError(t, err)
		assert.Len(t, result.Points, 3)
		// First point has change=0
		assert.Equal(t, 0.0, result.Points[0].Change)
		// Second point: (110000 - 100000) / 100000 * 100 = 10%
		assert.InDelta(t, 10.0, result.Points[1].Change, 0.001)
		// Third point: (121000 - 110000) / 110000 * 100 = 10%
		assert.InDelta(t, 10.0, result.Points[2].Change, 0.001)
	})

	t.Run("invalid period: startDate > endDate returns error", func(t *testing.T) {
		setRepo := &mockSetRecordRepoProgression{}

		uc := NewGetProgressionUC(setRepo)
		badStart := end
		badEnd := start
		_, err := uc.Execute(context.Background(), GetProgressionInput{
			UserID:    userID,
			StartDate: &badStart,
			EndDate:   &badEnd,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "startDate must be before or equal to endDate")
	})

	t.Run("returns empty array when no data", func(t *testing.T) {
		setRepo := &mockSetRecordRepoProgression{
			progressionResult: []ports.ProgressionPoint{},
		}

		uc := NewGetProgressionUC(setRepo)
		result, err := uc.Execute(context.Background(), GetProgressionInput{
			UserID:    userID,
			StartDate: &start,
			EndDate:   &end,
		})

		require.NoError(t, err)
		assert.Empty(t, result.Points)
	})

	t.Run("first point has change=0, second point with calculated change", func(t *testing.T) {
		day1 := now.AddDate(0, 0, -5)
		day2 := now.AddDate(0, 0, -1)

		setRepo := &mockSetRecordRepoProgression{
			progressionResult: []ports.ProgressionPoint{
				{Date: day1, MaxWeight: 80000, TotalVolume: 320000},
				{Date: day2, MaxWeight: 100000, TotalVolume: 400000},
			},
		}

		uc := NewGetProgressionUC(setRepo)
		result, err := uc.Execute(context.Background(), GetProgressionInput{
			UserID:    userID,
			StartDate: &start,
			EndDate:   &end,
		})

		require.NoError(t, err)
		require.Len(t, result.Points, 2)
		assert.Equal(t, 0.0, result.Points[0].Change)
		// (100000 - 80000) / 80000 * 100 = 25%
		assert.InDelta(t, 25.0, result.Points[1].Change, 0.001)
	})
}
