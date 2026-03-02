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

// --- Mocks for GetPersonalRecordsUC ---

type mockSetRecordRepoPR struct {
	prResult []ports.PersonalRecord
	prErr    error
}

func (m *mockSetRecordRepoPR) Create(_ context.Context, _ *entities.SetRecord) error {
	return nil
}
func (m *mockSetRecordRepoPR) FindBySessionExerciseSet(_ context.Context, _, _ uuid.UUID, _ int) (*entities.SetRecord, error) {
	return nil, nil
}
func (m *mockSetRecordRepoPR) GetTotalSetsRepsVolume(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SetRecordStats, error) {
	return nil, nil
}
func (m *mockSetRecordRepoPR) GetPersonalRecordsByUser(_ context.Context, _ uuid.UUID) ([]ports.PersonalRecord, error) {
	return m.prResult, m.prErr
}
func (m *mockSetRecordRepoPR) GetProgressionByUserAndExercise(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _, _ time.Time) ([]ports.ProgressionPoint, error) {
	return nil, nil
}

// --- Tests ---

func TestGetPersonalRecordsUC_Execute(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()

	t.Run("happy path: returns PRs correctly", func(t *testing.T) {
		exID1 := uuid.New()
		exID2 := uuid.New()

		setRepo := &mockSetRecordRepoPR{
			prResult: []ports.PersonalRecord{
				{
					ExerciseID:   exID1,
					ExerciseName: "Bench Press",
					Weight:       100000,
					Reps:         5,
					Volume:       500000,
					AchievedAt:   now.AddDate(0, 0, -3),
				},
				{
					ExerciseID:   exID2,
					ExerciseName: "Squat",
					Weight:       140000,
					Reps:         3,
					Volume:       420000,
					AchievedAt:   now.AddDate(0, 0, -1),
				},
			},
		}

		uc := NewGetPersonalRecordsUC(setRepo)
		result, err := uc.Execute(context.Background(), userID)

		require.NoError(t, err)
		require.Len(t, result, 2)

		assert.Equal(t, exID1, result[0].ExerciseID)
		assert.Equal(t, "Bench Press", result[0].ExerciseName)
		assert.Equal(t, 100000, result[0].Weight)
		assert.Equal(t, 5, result[0].Reps)
		assert.Equal(t, int64(500000), result[0].Volume)

		assert.Equal(t, exID2, result[1].ExerciseID)
		assert.Equal(t, "Squat", result[1].ExerciseName)
		assert.Equal(t, 140000, result[1].Weight)
	})

	t.Run("user without PRs: returns empty slice", func(t *testing.T) {
		setRepo := &mockSetRecordRepoPR{
			prResult: []ports.PersonalRecord{},
		}

		uc := NewGetPersonalRecordsUC(setRepo)
		result, err := uc.Execute(context.Background(), userID)

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
