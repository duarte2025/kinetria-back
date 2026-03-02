package workouts_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
)

// mockCreateWorkoutRepo implements ports.WorkoutRepository for CreateWorkoutUC tests.
type mockCreateWorkoutRepo struct {
	createFn func(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error
}

func (m *mockCreateWorkoutRepo) ExistsByIDAndUserID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockCreateWorkoutRepo) ListByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]entities.Workout, int, error) {
	return nil, 0, nil
}

func (m *mockCreateWorkoutRepo) GetFirstByUserID(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	return nil, nil
}

func (m *mockCreateWorkoutRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
	return nil, nil, nil
}

func (m *mockCreateWorkoutRepo) GetByIDOnly(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	return nil, nil
}

func (m *mockCreateWorkoutRepo) Create(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error {
	if m.createFn != nil {
		return m.createFn(ctx, workout, exercises)
	}
	return nil
}

func (m *mockCreateWorkoutRepo) Update(_ context.Context, _ entities.Workout, _ []entities.WorkoutExercise) error {
	return nil
}

func (m *mockCreateWorkoutRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCreateWorkoutRepo) HasActiveSessions(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

// mockCreateExerciseRepo implements ports.ExerciseRepository for CreateWorkoutUC tests.
type mockCreateExerciseRepo struct {
	getByIDFn func(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error)
}

func (m *mockCreateExerciseRepo) ExistsByIDAndWorkoutID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockCreateExerciseRepo) FindWorkoutExerciseID(_ context.Context, _, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockCreateExerciseRepo) List(_ context.Context, _ ports.ExerciseFilters, _, _ int) ([]*entities.Exercise, int, error) {
	return nil, 0, nil
}

func (m *mockCreateExerciseRepo) GetByID(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, exerciseID)
	}
	return nil, nil
}

func (m *mockCreateExerciseRepo) GetUserStats(_ context.Context, _, _ uuid.UUID) (*ports.ExerciseUserStats, error) {
	return nil, nil
}

func (m *mockCreateExerciseRepo) GetHistory(_ context.Context, _, _ uuid.UUID, _, _ int) ([]*ports.ExerciseHistoryEntry, int, error) {
	return nil, 0, nil
}

func TestCreateWorkoutUC_Execute(t *testing.T) {
	validUserID := uuid.New()
	validExerciseID := uuid.New()

	validExercise := &entities.Exercise{
		ID:   validExerciseID,
		Name: "Agachamento",
	}

	validInput := workouts.CreateWorkoutInput{
		Name:      "Treino A",
		Type:      "FORÇA",
		Intensity: "MODERADA",
		Duration:  45,
		Exercises: []workouts.WorkoutExerciseInput{
			{
				ExerciseID: validExerciseID,
				Sets:       3,
				Reps:       "10",
				RestTime:   60,
				OrderIndex: 1,
			},
		},
	}

	tests := []struct {
		name          string
		input         workouts.CreateWorkoutInput
		exerciseFn    func(ctx context.Context, id uuid.UUID) (*entities.Exercise, error)
		expectedError string
		expectedDomErr error
	}{
		{
			name:  "happy_path_valid_input",
			input: validInput,
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError: "",
		},
		{
			name: "name_too_short",
			input: workouts.CreateWorkoutInput{
				Name:      "AB",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: validInput.Exercises,
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "name must be between 3 and 255 characters",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "name_too_long",
			input: workouts.CreateWorkoutInput{
				Name:      strings.Repeat("A", 256),
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: validInput.Exercises,
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "name must be between 3 and 255 characters",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "invalid_type",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "INVALID",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: validInput.Exercises,
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "type must be one of",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "invalid_intensity",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "EXTREMA",
				Duration:  45,
				Exercises: validInput.Exercises,
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "intensity must be one of",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "duration_too_low",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  0,
				Exercises: validInput.Exercises,
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "duration must be between 1 and 300",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "duration_too_high",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  301,
				Exercises: validInput.Exercises,
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "duration must be between 1 and 300",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "no_exercises",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{},
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "at least one exercise is required",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "too_many_exercises",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: func() []workouts.WorkoutExerciseInput {
					ex := make([]workouts.WorkoutExerciseInput, 21)
					for i := range ex {
						ex[i] = workouts.WorkoutExerciseInput{
							ExerciseID: validExerciseID,
							Sets:       3,
							Reps:       "10",
							RestTime:   60,
							OrderIndex: i + 1,
						}
					}
					return ex
				}(),
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "maximum of 20 exercises allowed",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "invalid_sets_zero",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: validExerciseID, Sets: 0, Reps: "10", RestTime: 60, OrderIndex: 1},
				},
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "sets must be between 1 and 10",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "invalid_sets_eleven",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: validExerciseID, Sets: 11, Reps: "10", RestTime: 60, OrderIndex: 1},
				},
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "sets must be between 1 and 10",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "invalid_rest_time_negative",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: validExerciseID, Sets: 3, Reps: "10", RestTime: -1, OrderIndex: 1},
				},
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "restTime must be between 0 and 600",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "invalid_rest_time_over_600",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: validExerciseID, Sets: 3, Reps: "10", RestTime: 601, OrderIndex: 1},
				},
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "restTime must be between 0 and 600",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "duplicate_order_index",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: validExerciseID, Sets: 3, Reps: "10", RestTime: 60, OrderIndex: 1},
					{ExerciseID: validExerciseID, Sets: 3, Reps: "10", RestTime: 60, OrderIndex: 1},
				},
			},
			exerciseFn: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
				return validExercise, nil
			},
			expectedError:  "duplicate orderIndex",
			expectedDomErr: domerrors.ErrMalformedParameters,
		},
		{
			name: "exercise_not_found",
			input: workouts.CreateWorkoutInput{
				Name:      "Treino A",
				Type:      "FORÇA",
				Intensity: "MODERADA",
				Duration:  45,
				Exercises: []workouts.WorkoutExerciseInput{
					{ExerciseID: uuid.New(), Sets: 3, Reps: "10", RestTime: 60, OrderIndex: 1},
				},
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
			workoutRepo := &mockCreateWorkoutRepo{}
			exerciseRepo := &mockCreateExerciseRepo{
				getByIDFn: tt.exerciseFn,
			}

			uc := workouts.NewCreateWorkoutUC(workoutRepo, exerciseRepo)
			result, err := uc.Execute(context.Background(), validUserID, tt.input)

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
				return
			}

			if result == nil {
				t.Error("expected non-nil workout result")
				return
			}

			if result.Name != tt.input.Name {
				t.Errorf("expected workout name %q, got %q", tt.input.Name, result.Name)
			}
		})
	}
}
