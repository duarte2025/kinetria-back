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

// Mock inline do WorkoutRepository
type mockWorkoutRepo struct {
	listByUserIDFunc func(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error)
}

func (m *mockWorkoutRepo) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error) {
	if m.listByUserIDFunc != nil {
		return m.listByUserIDFunc(ctx, userID, offset, limit)
	}
	return nil, 0, nil
}

func TestListWorkoutsUC_Execute(t *testing.T) {
	validUserID := uuid.New()
	now := time.Now()

	tests := []struct {
		name          string
		input         workouts.ListWorkoutsInput
		mockReturn    []entities.Workout
		mockTotal     int
		mockError     error
		expectedOut   workouts.ListWorkoutsOutput
		expectedError string
	}{
		{
			name: "happy_path_with_workouts",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     1,
				PageSize: 10,
			},
			mockReturn: []entities.Workout{
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 1", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 2", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 3", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 4", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 5", CreatedAt: now, UpdatedAt: now},
			},
			mockTotal:     25,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{Page: 1, PageSize: 10, Total: 25, TotalPages: 3},
			expectedError: "",
		},
		{
			name: "user_without_workouts",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     1,
				PageSize: 20,
			},
			mockReturn:    []entities.Workout{},
			mockTotal:     0,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{Page: 1, PageSize: 20, Total: 0, TotalPages: 0},
			expectedError: "",
		},
		{
			name: "page_beyond_total",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     10,
				PageSize: 20,
			},
			mockReturn:    []entities.Workout{},
			mockTotal:     5,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{Page: 10, PageSize: 20, Total: 5, TotalPages: 1},
			expectedError: "",
		},
		{
			name: "validation_nil_userID",
			input: workouts.ListWorkoutsInput{
				UserID:   uuid.Nil,
				Page:     1,
				PageSize: 20,
			},
			mockReturn:    nil,
			mockTotal:     0,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{},
			expectedError: "userID is required",
		},
		{
			name: "validation_negative_page",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     -1,
				PageSize: 20,
			},
			mockReturn:    nil,
			mockTotal:     0,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{},
			expectedError: "page must be greater than or equal to 1",
		},
		{
			name: "validation_pageSize_over_100",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     1,
				PageSize: 101,
			},
			mockReturn:    nil,
			mockTotal:     0,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{},
			expectedError: "pageSize must be between 1 and 100",
		},
		{
			name: "default_page_zero_becomes_one",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     0,
				PageSize: 20,
			},
			mockReturn:    []entities.Workout{},
			mockTotal:     0,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{Page: 1, PageSize: 20, Total: 0, TotalPages: 0},
			expectedError: "",
		},
		{
			name: "default_pageSize_zero_becomes_twenty",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     1,
				PageSize: 0,
			},
			mockReturn:    []entities.Workout{},
			mockTotal:     0,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{Page: 1, PageSize: 20, Total: 0, TotalPages: 0},
			expectedError: "",
		},
		{
			name: "calculate_totalPages_correctly",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     1,
				PageSize: 7,
			},
			mockReturn: []entities.Workout{
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 1", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 2", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 3", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 4", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 5", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 6", CreatedAt: now, UpdatedAt: now},
				{ID: uuid.New(), UserID: validUserID, Name: "Workout 7", CreatedAt: now, UpdatedAt: now},
			},
			mockTotal:     20,
			mockError:     nil,
			expectedOut:   workouts.ListWorkoutsOutput{Page: 1, PageSize: 7, Total: 20, TotalPages: 3},
			expectedError: "",
		},
		{
			name: "repository_error",
			input: workouts.ListWorkoutsInput{
				UserID:   validUserID,
				Page:     1,
				PageSize: 20,
			},
			mockReturn:    nil,
			mockTotal:     0,
			mockError:     errors.New("database connection failed"),
			expectedOut:   workouts.ListWorkoutsOutput{},
			expectedError: "failed to list workouts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &mockWorkoutRepo{
				listByUserIDFunc: func(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error) {
					return tt.mockReturn, tt.mockTotal, tt.mockError
				},
			}

			// Create use case
			uc := workouts.NewListWorkoutsUC(mockRepo)

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

			// Validate output fields
			if output.Page != tt.expectedOut.Page {
				t.Errorf("expected Page=%d, got %d", tt.expectedOut.Page, output.Page)
			}
			if output.PageSize != tt.expectedOut.PageSize {
				t.Errorf("expected PageSize=%d, got %d", tt.expectedOut.PageSize, output.PageSize)
			}
			if output.Total != tt.expectedOut.Total {
				t.Errorf("expected Total=%d, got %d", tt.expectedOut.Total, output.Total)
			}
			if output.TotalPages != tt.expectedOut.TotalPages {
				t.Errorf("expected TotalPages=%d, got %d", tt.expectedOut.TotalPages, output.TotalPages)
			}

			// Validate workouts count
			if len(output.Workouts) != len(tt.mockReturn) {
				t.Errorf("expected %d workouts, got %d", len(tt.mockReturn), len(output.Workouts))
			}
		})
	}
}
