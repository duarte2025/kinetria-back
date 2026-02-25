package workouts_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
)

// Mock inline do WorkoutRepository para GetWorkoutUC
type mockGetWorkoutRepo struct {
	getByIDFunc func(ctx context.Context, workoutID, userID uuid.UUID) (*entities.Workout, []entities.Exercise, error)
}

func (m *mockGetWorkoutRepo) GetByID(ctx context.Context, workoutID, userID uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, workoutID, userID)
	}
	return nil, nil, nil
}

func (m *mockGetWorkoutRepo) ListByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]entities.Workout, int, error) {
	return nil, 0, nil
}

func (m *mockGetWorkoutRepo) ExistsByIDAndUserID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockGetWorkoutRepo) GetFirstByUserID(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	return nil, nil
}

func TestGetWorkoutUC_Execute(t *testing.T) {
	validUserID := uuid.New()
	validWorkoutID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		input          workouts.GetWorkoutInput
		mockWorkout    *entities.Workout
		mockExercises  []entities.Exercise
		mockError      error
		expectedError  string
		validateOutput func(t *testing.T, output workouts.GetWorkoutOutput)
	}{
		{
			name: "success_workout_with_exercises",
			input: workouts.GetWorkoutInput{
				WorkoutID: validWorkoutID,
				UserID:    validUserID,
			},
			mockWorkout: &entities.Workout{
				ID:        validWorkoutID,
				UserID:    validUserID,
				Name:      "Full Body A",
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockExercises: []entities.Exercise{
				{
					ID:        uuid.New(),
					WorkoutID: validWorkoutID,
					Name:      "Agachamento",
					Sets:      3,
					Reps:      "10",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        uuid.New(),
					WorkoutID: validWorkoutID,
					Name:      "Supino",
					Sets:      3,
					Reps:      "8",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        uuid.New(),
					WorkoutID: validWorkoutID,
					Name:      "Remada",
					Sets:      3,
					Reps:      "10",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			mockError:     nil,
			expectedError: "",
			validateOutput: func(t *testing.T, output workouts.GetWorkoutOutput) {
				if output.Workout.ID != validWorkoutID {
					t.Errorf("expected workout ID %s, got %s", validWorkoutID, output.Workout.ID)
				}
				if output.Workout.Name != "Full Body A" {
					t.Errorf("expected workout name 'Full Body A', got %s", output.Workout.Name)
				}
				if len(output.Exercises) != 3 {
					t.Errorf("expected 3 exercises, got %d", len(output.Exercises))
				}
				if len(output.Exercises) > 0 && output.Exercises[0].Name != "Agachamento" {
					t.Errorf("expected first exercise 'Agachamento', got %s", output.Exercises[0].Name)
				}
			},
		},
		{
			name: "success_workout_without_exercises",
			input: workouts.GetWorkoutInput{
				WorkoutID: validWorkoutID,
				UserID:    validUserID,
			},
			mockWorkout: &entities.Workout{
				ID:        validWorkoutID,
				UserID:    validUserID,
				Name:      "Empty Workout",
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockExercises: []entities.Exercise{},
			mockError:     nil,
			expectedError: "",
			validateOutput: func(t *testing.T, output workouts.GetWorkoutOutput) {
				if output.Workout.ID != validWorkoutID {
					t.Errorf("expected workout ID %s, got %s", validWorkoutID, output.Workout.ID)
				}
				if output.Workout.Name != "Empty Workout" {
					t.Errorf("expected workout name 'Empty Workout', got %s", output.Workout.Name)
				}
				if len(output.Exercises) != 0 {
					t.Errorf("expected 0 exercises, got %d", len(output.Exercises))
				}
			},
		},
		{
			name: "error_workout_not_found",
			input: workouts.GetWorkoutInput{
				WorkoutID: validWorkoutID,
				UserID:    validUserID,
			},
			mockWorkout:   nil,
			mockExercises: nil,
			mockError:     nil,
			expectedError: "not found",
		},
		{
			name: "validation_error_nil_workoutID",
			input: workouts.GetWorkoutInput{
				WorkoutID: uuid.Nil,
				UserID:    validUserID,
			},
			mockWorkout:   nil,
			mockExercises: nil,
			mockError:     nil,
			expectedError: "workoutId cannot be empty",
		},
		{
			name: "validation_error_nil_userID",
			input: workouts.GetWorkoutInput{
				WorkoutID: validWorkoutID,
				UserID:    uuid.Nil,
			},
			mockWorkout:   nil,
			mockExercises: nil,
			mockError:     nil,
			expectedError: "userId cannot be empty",
		},
		{
			name: "repository_error_database_failure",
			input: workouts.GetWorkoutInput{
				WorkoutID: validWorkoutID,
				UserID:    validUserID,
			},
			mockWorkout:   nil,
			mockExercises: nil,
			mockError:     errors.New("database connection failed"),
			expectedError: "failed to get workout",
		},
		{
			name: "repository_error_timeout",
			input: workouts.GetWorkoutInput{
				WorkoutID: validWorkoutID,
				UserID:    validUserID,
			},
			mockWorkout:   nil,
			mockExercises: nil,
			mockError:     errors.New("context deadline exceeded"),
			expectedError: "failed to get workout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &mockGetWorkoutRepo{
				getByIDFunc: func(ctx context.Context, workoutID, userID uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
					return tt.mockWorkout, tt.mockExercises, tt.mockError
				},
			}

			// Create use case
			uc := workouts.NewGetWorkoutUC(mockRepo)

			// Execute
			output, err := uc.Execute(context.Background(), tt.input)

			// Validate error
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			// Validate no error when not expected
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Run custom validation if provided
			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}
		})
	}
}
