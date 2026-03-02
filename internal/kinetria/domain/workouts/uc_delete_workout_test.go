package workouts_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
)

// mockDeleteWorkoutRepo implements ports.WorkoutRepository for DeleteWorkoutUC tests.
type mockDeleteWorkoutRepo struct {
	getByIDOnlyFn      func(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)
	hasActiveSessionsFn func(ctx context.Context, workoutID uuid.UUID) (bool, error)
	deleteFn            func(ctx context.Context, workoutID uuid.UUID) error
}

func (m *mockDeleteWorkoutRepo) ExistsByIDAndUserID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockDeleteWorkoutRepo) ListByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]entities.Workout, int, error) {
	return nil, 0, nil
}

func (m *mockDeleteWorkoutRepo) GetFirstByUserID(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	return nil, nil
}

func (m *mockDeleteWorkoutRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
	return nil, nil, nil
}

func (m *mockDeleteWorkoutRepo) GetByIDOnly(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error) {
	if m.getByIDOnlyFn != nil {
		return m.getByIDOnlyFn(ctx, workoutID)
	}
	return nil, nil
}

func (m *mockDeleteWorkoutRepo) Create(_ context.Context, _ entities.Workout, _ []entities.WorkoutExercise) error {
	return nil
}

func (m *mockDeleteWorkoutRepo) Update(_ context.Context, _ entities.Workout, _ []entities.WorkoutExercise) error {
	return nil
}

func (m *mockDeleteWorkoutRepo) Delete(ctx context.Context, workoutID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, workoutID)
	}
	return nil
}

func (m *mockDeleteWorkoutRepo) HasActiveSessions(ctx context.Context, workoutID uuid.UUID) (bool, error) {
	if m.hasActiveSessionsFn != nil {
		return m.hasActiveSessionsFn(ctx, workoutID)
	}
	return false, nil
}

func TestDeleteWorkoutUC_Execute(t *testing.T) {
	validUserID := uuid.New()
	otherUserID := uuid.New()
	validWorkoutID := uuid.New()

	// Helper to build base owned workout
	ownedWorkout := func() *entities.Workout {
		return &entities.Workout{
			ID:        validWorkoutID,
			UserID:    validUserID,
			Name:      "Treino Para Deletar",
			CreatedBy: &validUserID,
		}
	}

	tests := []struct {
		name                string
		userID              uuid.UUID
		workoutID           uuid.UUID
		getByIDOnlyFn       func(ctx context.Context, id uuid.UUID) (*entities.Workout, error)
		hasActiveSessionsFn func(ctx context.Context, id uuid.UUID) (bool, error)
		deleteFn            func(ctx context.Context, id uuid.UUID) error
		expectedError       string
		expectedDomErr      error
	}{
		{
			name:      "happy_path_deletes_successfully",
			userID:    validUserID,
			workoutID: validWorkoutID,
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return ownedWorkout(), nil
			},
			hasActiveSessionsFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return false, nil
			},
			expectedError: "",
		},
		{
			name:      "workout_not_found_returns_error",
			userID:    validUserID,
			workoutID: validWorkoutID,
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return nil, nil
			},
			expectedError:  "workout not found",
			expectedDomErr: domerrors.ErrWorkoutNotFound,
		},
		{
			name:      "template_workout_cannot_be_deleted",
			userID:    validUserID,
			workoutID: validWorkoutID,
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				w := ownedWorkout()
				w.CreatedBy = nil // template
				return w, nil
			},
			expectedError:  "cannot",
			expectedDomErr: domerrors.ErrCannotModifyTemplate,
		},
		{
			name:      "different_owner_returns_forbidden",
			userID:    otherUserID,
			workoutID: validWorkoutID,
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return ownedWorkout(), nil // owned by validUserID
			},
			expectedError:  "forbidden",
			expectedDomErr: domerrors.ErrForbidden,
		},
		{
			name:      "has_active_sessions_returns_error",
			userID:    validUserID,
			workoutID: validWorkoutID,
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return ownedWorkout(), nil
			},
			hasActiveSessionsFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return true, nil
			},
			expectedError:  "workout has active sessions",
			expectedDomErr: domerrors.ErrWorkoutHasActiveSessions,
		},
		{
			name:      "completed_sessions_only_deletes_successfully",
			userID:    validUserID,
			workoutID: validWorkoutID,
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return ownedWorkout(), nil
			},
			hasActiveSessionsFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return false, nil // only completed sessions
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDeleteWorkoutRepo{
				getByIDOnlyFn:       tt.getByIDOnlyFn,
				hasActiveSessionsFn: tt.hasActiveSessionsFn,
				deleteFn:            tt.deleteFn,
			}

			uc := workouts.NewDeleteWorkoutUC(repo)
			err := uc.Execute(context.Background(), tt.userID, tt.workoutID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				if tt.expectedDomErr != nil && !errors.Is(err, tt.expectedDomErr) {
					t.Errorf("expected domain error %v, but errors.Is returned false; got: %v", tt.expectedDomErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
