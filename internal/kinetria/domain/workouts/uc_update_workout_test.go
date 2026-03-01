package workouts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
)

// mockUpdateWorkoutRepo implements ports.WorkoutRepository for UpdateWorkoutUC tests.
type mockUpdateWorkoutRepo struct {
	getByIDOnlyFn func(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)
	updateFn      func(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error
}

func (m *mockUpdateWorkoutRepo) ExistsByIDAndUserID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockUpdateWorkoutRepo) ListByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]entities.Workout, int, error) {
	return nil, 0, nil
}

func (m *mockUpdateWorkoutRepo) GetFirstByUserID(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	return nil, nil
}

func (m *mockUpdateWorkoutRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
	return nil, nil, nil
}

func (m *mockUpdateWorkoutRepo) GetByIDOnly(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error) {
	if m.getByIDOnlyFn != nil {
		return m.getByIDOnlyFn(ctx, workoutID)
	}
	return nil, nil
}

func (m *mockUpdateWorkoutRepo) Create(_ context.Context, _ entities.Workout, _ []entities.WorkoutExercise) error {
	return nil
}

func (m *mockUpdateWorkoutRepo) Update(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, workout, exercises)
	}
	return nil
}

func (m *mockUpdateWorkoutRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockUpdateWorkoutRepo) HasActiveSessions(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

// mockUpdateExerciseRepo implements ports.ExerciseRepository for UpdateWorkoutUC tests.
type mockUpdateExerciseRepo struct {
	getByIDFn func(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error)
}

func (m *mockUpdateExerciseRepo) ExistsByIDAndWorkoutID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockUpdateExerciseRepo) FindWorkoutExerciseID(_ context.Context, _, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockUpdateExerciseRepo) List(_ context.Context, _ ports.ExerciseFilters, _, _ int) ([]*entities.Exercise, int, error) {
	return nil, 0, nil
}

func (m *mockUpdateExerciseRepo) GetByID(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, exerciseID)
	}
	return nil, nil
}

func (m *mockUpdateExerciseRepo) GetUserStats(_ context.Context, _, _ uuid.UUID) (*ports.ExerciseUserStats, error) {
	return nil, nil
}

func (m *mockUpdateExerciseRepo) GetHistory(_ context.Context, _, _ uuid.UUID, _, _ int) ([]*ports.ExerciseHistoryEntry, int, error) {
	return nil, 0, nil
}

func strPtr(s string) *string { return &s }

func TestUpdateWorkoutUC_Execute(t *testing.T) {
	validUserID := uuid.New()
	otherUserID := uuid.New()
	validWorkoutID := uuid.New()
	validExerciseID := uuid.New()

	validExercise := &entities.Exercise{
		ID:   validExerciseID,
		Name: "Supino",
	}

	// Base workout owned by validUserID
	baseWorkout := func() *entities.Workout {
		return &entities.Workout{
			ID:        validWorkoutID,
			UserID:    validUserID,
			Name:      "Treino Original",
			Type:      "FORÇA",
			Intensity: "MODERADA",
			Duration:  45,
			CreatedBy: &validUserID,
		}
	}

	tests := []struct {
		name           string
		userID         uuid.UUID
		workoutID      uuid.UUID
		input          workouts.UpdateWorkoutInput
		getByIDOnlyFn  func(ctx context.Context, id uuid.UUID) (*entities.Workout, error)
		exerciseFn     func(ctx context.Context, id uuid.UUID) (*entities.Exercise, error)
		expectedError  string
		expectedDomErr error
	}{
		{
			name:      "happy_path_updates_successfully",
			userID:    validUserID,
			workoutID: validWorkoutID,
			input: workouts.UpdateWorkoutInput{
				Name: strPtr("Treino Atualizado"),
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: validExerciseID, Sets: 4, Reps: "8", RestTime: 90, OrderIndex: 1},
				},
			},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return baseWorkout(), nil
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError: "",
		},
		{
			name:      "workout_not_found_returns_error",
			userID:    validUserID,
			workoutID: validWorkoutID,
			input:     workouts.UpdateWorkoutInput{Name: strPtr("Novo Nome")},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return nil, nil
			},
			expectedError:  "workout not found",
			expectedDomErr: domerrors.ErrWorkoutNotFound,
		},
		{
			name:      "template_workout_returns_cannot_modify",
			userID:    validUserID,
			workoutID: validWorkoutID,
			input:     workouts.UpdateWorkoutInput{Name: strPtr("Novo Nome")},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				w := baseWorkout()
				w.CreatedBy = nil // template: no owner
				return w, nil
			},
			expectedError:  "cannot modify template workouts",
			expectedDomErr: domerrors.ErrCannotModifyTemplate,
		},
		{
			name:      "different_owner_returns_forbidden",
			userID:    otherUserID,
			workoutID: validWorkoutID,
			input:     workouts.UpdateWorkoutInput{Name: strPtr("Novo Nome")},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return baseWorkout(), nil // owned by validUserID
			},
			expectedError:  "forbidden",
			expectedDomErr: domerrors.ErrForbidden,
		},
		{
			name:      "invalid_name_too_short",
			userID:    validUserID,
			workoutID: validWorkoutID,
			input:     workouts.UpdateWorkoutInput{Name: strPtr("AB")},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return baseWorkout(), nil
			},
			expectedError:  "name must be between 3 and 255 characters",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name:      "invalid_type_returns_error",
			userID:    validUserID,
			workoutID: validWorkoutID,
			input:     workouts.UpdateWorkoutInput{Type: strPtr("INVALIDO")},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return baseWorkout(), nil
			},
			expectedError:  "type must be one of",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name:      "exercise_not_found_returns_error",
			userID:    validUserID,
			workoutID: validWorkoutID,
			input: workouts.UpdateWorkoutInput{
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: uuid.New(), Sets: 3, Reps: "10", RestTime: 60, OrderIndex: 1},
				},
			},
			getByIDOnlyFn: func(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
				return baseWorkout(), nil
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return nil, nil
			},
			expectedError:  "not found",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutRepo := &mockUpdateWorkoutRepo{
				getByIDOnlyFn: tt.getByIDOnlyFn,
			}
			exerciseRepo := &mockUpdateExerciseRepo{
				getByIDFn: tt.exerciseFn,
			}

			uc := workouts.NewUpdateWorkoutUC(workoutRepo, exerciseRepo)
			result, err := uc.Execute(context.Background(), tt.userID, tt.workoutID, tt.input)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !containsString(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				if tt.expectedDomErr != nil && !errors.Is(err, tt.expectedDomErr) {
					t.Errorf("expected domain error %v, but errors.Is returned false; got: %v", tt.expectedDomErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected non-nil workout result")
			}
		})
	}
}

// containsString checks if s contains substr (avoids importing strings in multiple test files).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
