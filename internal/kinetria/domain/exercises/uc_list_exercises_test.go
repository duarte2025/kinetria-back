package exercises_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/exercises"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// mockExerciseRepo is an inline mock that only needs methods for ListExercisesUC tests.
type mockExerciseRepoForList struct {
	listFunc func(ctx context.Context, filters ports.ExerciseFilters, page, pageSize int) ([]*entities.Exercise, int, error)
}

func (m *mockExerciseRepoForList) List(ctx context.Context, filters ports.ExerciseFilters, page, pageSize int) ([]*entities.Exercise, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters, page, pageSize)
	}
	return []*entities.Exercise{}, 0, nil
}

func (m *mockExerciseRepoForList) ExistsByIDAndWorkoutID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockExerciseRepoForList) FindWorkoutExerciseID(_ context.Context, _, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockExerciseRepoForList) GetByID(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
	return nil, nil
}
func (m *mockExerciseRepoForList) GetUserStats(_ context.Context, _, _ uuid.UUID) (*ports.ExerciseUserStats, error) {
	return nil, nil
}
func (m *mockExerciseRepoForList) GetHistory(_ context.Context, _, _ uuid.UUID, _, _ int) ([]*ports.ExerciseHistoryEntry, int, error) {
	return nil, 0, nil
}

func strPtr(s string) *string { return &s }

func makeExercises(n int) []*entities.Exercise {
	out := make([]*entities.Exercise, n)
	for i := range out {
		out[i] = &entities.Exercise{ID: uuid.New(), Name: "Exercise"}
	}
	return out
}

func TestListExercisesUC_Execute(t *testing.T) {
	tests := []struct {
		name            string
		input           exercises.ListExercisesInput
		mockReturn      []*entities.Exercise
		mockTotal       int
		mockError       error
		wantTotal       int
		wantPage        int
		wantPageSize    int
		wantTotalPages  int
		wantErrContains string
		capturedFilters *ports.ExerciseFilters
	}{
		{
			name: "happy_path_paginated",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     1,
				PageSize: 10,
			},
			mockReturn:     makeExercises(10),
			mockTotal:      25,
			wantTotal:      25,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 3,
		},
		{
			name: "filter_by_muscle_group",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{MuscleGroup: strPtr("Peito")},
				Page:     1,
				PageSize: 20,
			},
			mockReturn:     makeExercises(5),
			mockTotal:      5,
			wantTotal:      5,
			wantPage:       1,
			wantPageSize:   20,
			wantTotalPages: 1,
		},
		{
			name: "filter_by_equipment",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{Equipment: strPtr("Barra")},
				Page:     1,
				PageSize: 20,
			},
			mockReturn:     makeExercises(3),
			mockTotal:      3,
			wantTotal:      3,
			wantTotalPages: 1,
		},
		{
			name: "filter_by_difficulty",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{Difficulty: strPtr("Intermediário")},
				Page:     1,
				PageSize: 20,
			},
			mockReturn:     makeExercises(7),
			mockTotal:      7,
			wantTotal:      7,
			wantTotalPages: 1,
		},
		{
			name: "filter_by_search",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{Search: strPtr("supino")},
				Page:     1,
				PageSize: 20,
			},
			mockReturn:     makeExercises(2),
			mockTotal:      2,
			wantTotal:      2,
			wantTotalPages: 1,
		},
		{
			name: "combine_multiple_filters",
			input: exercises.ListExercisesInput{
				Filters: ports.ExerciseFilters{
					MuscleGroup: strPtr("Peito"),
					Equipment:   strPtr("Barra"),
					Difficulty:  strPtr("Intermediário"),
				},
				Page:     1,
				PageSize: 20,
			},
			mockReturn:     makeExercises(1),
			mockTotal:      1,
			wantTotal:      1,
			wantTotalPages: 1,
		},
		{
			name: "empty_library",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     1,
				PageSize: 20,
			},
			mockReturn:     []*entities.Exercise{},
			mockTotal:      0,
			wantTotal:      0,
			wantTotalPages: 0,
		},
		{
			name: "page_zero_returns_error",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     0,
				PageSize: 20,
			},
			wantErrContains: "page must be >= 1",
		},
		{
			name: "negative_page_returns_error",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     -1,
				PageSize: 20,
			},
			wantErrContains: "page must be >= 1",
		},
		{
			name: "pageSize_over_limit_returns_error",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     1,
				PageSize: 200,
			},
			wantErrContains: "pageSize must be <= 100",
		},
		{
			name: "pageSize_zero_returns_error",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     1,
				PageSize: 0,
			},
			wantErrContains: "pageSize must be >= 1",
		},
		{
			name: "repository_error_propagates",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     1,
				PageSize: 20,
			},
			mockError:       errors.New("db connection failed"),
			wantErrContains: "failed to list exercises",
		},
		{
			name: "pagination_page_2",
			input: exercises.ListExercisesInput{
				Filters:  ports.ExerciseFilters{},
				Page:     2,
				PageSize: 10,
			},
			mockReturn:     makeExercises(10),
			mockTotal:      50,
			wantTotal:      50,
			wantPage:       2,
			wantPageSize:   10,
			wantTotalPages: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockExerciseRepoForList{
				listFunc: func(_ context.Context, _ ports.ExerciseFilters, _, _ int) ([]*entities.Exercise, int, error) {
					if tt.mockError != nil {
						return nil, 0, tt.mockError
					}
					return tt.mockReturn, tt.mockTotal, nil
				},
			}

			uc := exercises.NewListExercisesUC(mockRepo)
			output, err := uc.Execute(context.Background(), tt.input)

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
			if output.Total != tt.wantTotal {
				t.Errorf("Total: got %d, want %d", output.Total, tt.wantTotal)
			}
			if tt.wantPage != 0 && output.Page != tt.wantPage {
				t.Errorf("Page: got %d, want %d", output.Page, tt.wantPage)
			}
			if tt.wantPageSize != 0 && output.PageSize != tt.wantPageSize {
				t.Errorf("PageSize: got %d, want %d", output.PageSize, tt.wantPageSize)
			}
			if output.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages: got %d, want %d", output.TotalPages, tt.wantTotalPages)
			}
			if tt.mockReturn != nil {
				if len(output.Exercises) != len(tt.mockReturn) {
					t.Errorf("Exercises count: got %d, want %d", len(output.Exercises), len(tt.mockReturn))
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
