package exercises_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/exercises"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// mockExerciseRepoForGet is an inline mock for GetExerciseUC tests.
type mockExerciseRepoForGet struct {
	getByIDFunc      func(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error)
	getUserStatsFunc func(ctx context.Context, userID, exerciseID uuid.UUID) (*ports.ExerciseUserStats, error)
}

func (m *mockExerciseRepoForGet) GetByID(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, exerciseID)
	}
	return nil, nil
}

func (m *mockExerciseRepoForGet) GetUserStats(ctx context.Context, userID, exerciseID uuid.UUID) (*ports.ExerciseUserStats, error) {
	if m.getUserStatsFunc != nil {
		return m.getUserStatsFunc(ctx, userID, exerciseID)
	}
	return &ports.ExerciseUserStats{}, nil
}

func (m *mockExerciseRepoForGet) ExistsByIDAndWorkoutID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockExerciseRepoForGet) FindWorkoutExerciseID(_ context.Context, _, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockExerciseRepoForGet) List(_ context.Context, _ ports.ExerciseFilters, _, _ int) ([]*entities.Exercise, int, error) {
	return nil, 0, nil
}
func (m *mockExerciseRepoForGet) GetHistory(_ context.Context, _, _ uuid.UUID, _, _ int) ([]*ports.ExerciseHistoryEntry, int, error) {
	return nil, 0, nil
}

func TestGetExerciseUC_Execute(t *testing.T) {
	exerciseID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	bestWeight := 80000
	avgWeight := 75000.5
	timesPerformed := 5

	sampleExercise := &entities.Exercise{
		ID:   exerciseID,
		Name: "Supino Reto",
	}

	tests := []struct {
		name            string
		exerciseID      uuid.UUID
		userID          *uuid.UUID
		mockExercise    *entities.Exercise
		mockExerciseErr error
		mockStats       *ports.ExerciseUserStats
		mockStatsErr    error
		wantNil         bool
		wantStats       bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:         "happy_path_without_auth",
			exerciseID:   exerciseID,
			userID:       nil,
			mockExercise: sampleExercise,
			wantStats:    false,
		},
		{
			name:       "happy_path_with_auth_and_stats",
			exerciseID: exerciseID,
			userID:     &userID,
			mockExercise: sampleExercise,
			mockStats: &ports.ExerciseUserStats{
				LastPerformed:  &now,
				BestWeight:     &bestWeight,
				TimesPerformed: timesPerformed,
				AverageWeight:  &avgWeight,
			},
			wantStats: true,
		},
		{
			name:       "user_never_executed_exercise",
			exerciseID: exerciseID,
			userID:     &userID,
			mockExercise: sampleExercise,
			mockStats: &ports.ExerciseUserStats{
				TimesPerformed: 0,
			},
			wantStats: true,
		},
		{
			name:         "exercise_not_found",
			exerciseID:   uuid.New(),
			userID:       nil,
			mockExercise: nil,
			wantErrIs:    domainerrors.ErrExerciseNotFound,
		},
		{
			name:            "repository_error_on_get",
			exerciseID:      exerciseID,
			userID:          nil,
			mockExerciseErr: errors.New("db error"),
			wantErrContains: "failed to get exercise",
		},
		{
			name:         "stats_error_propagates",
			exerciseID:   exerciseID,
			userID:       &userID,
			mockExercise: sampleExercise,
			mockStatsErr: errors.New("stats db error"),
			wantErrContains: "failed to get user stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockExerciseRepoForGet{
				getByIDFunc: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
					return tt.mockExercise, tt.mockExerciseErr
				},
				getUserStatsFunc: func(_ context.Context, _, _ uuid.UUID) (*ports.ExerciseUserStats, error) {
					return tt.mockStats, tt.mockStatsErr
				},
			}

			uc := exercises.NewGetExerciseUC(mockRepo)
			result, err := uc.Execute(context.Background(), tt.exerciseID, tt.userID)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrContains)
				}
				if !contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Exercise == nil {
				t.Fatal("expected non-nil Exercise in result")
			}
			if result.Exercise.ID != tt.exerciseID {
				t.Errorf("Exercise.ID: got %v, want %v", result.Exercise.ID, tt.exerciseID)
			}

			if tt.wantStats {
				if result.UserStats == nil {
					t.Error("expected UserStats to be non-nil when authenticated")
				}
			} else {
				if result.UserStats != nil {
					t.Error("expected UserStats to be nil when unauthenticated")
				}
			}
		})
	}
}
