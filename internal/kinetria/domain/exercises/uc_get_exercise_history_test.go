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

// mockExerciseRepoForHistory is an inline mock for GetExerciseHistoryUC tests.
type mockExerciseRepoForHistory struct {
	getByIDFunc    func(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error)
	getHistoryFunc func(ctx context.Context, userID, exerciseID uuid.UUID, page, pageSize int) ([]*ports.ExerciseHistoryEntry, int, error)
}

func (m *mockExerciseRepoForHistory) GetByID(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, exerciseID)
	}
	return nil, nil
}

func (m *mockExerciseRepoForHistory) GetHistory(ctx context.Context, userID, exerciseID uuid.UUID, page, pageSize int) ([]*ports.ExerciseHistoryEntry, int, error) {
	if m.getHistoryFunc != nil {
		return m.getHistoryFunc(ctx, userID, exerciseID, page, pageSize)
	}
	return []*ports.ExerciseHistoryEntry{}, 0, nil
}

func (m *mockExerciseRepoForHistory) ExistsByIDAndWorkoutID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockExerciseRepoForHistory) FindWorkoutExerciseID(_ context.Context, _, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockExerciseRepoForHistory) List(_ context.Context, _ ports.ExerciseFilters, _, _ int) ([]*entities.Exercise, int, error) {
	return nil, 0, nil
}
func (m *mockExerciseRepoForHistory) GetUserStats(_ context.Context, _, _ uuid.UUID) (*ports.ExerciseUserStats, error) {
	return nil, nil
}

func makeHistoryEntries(sessionIDs []uuid.UUID, performedAts []time.Time) []*ports.ExerciseHistoryEntry {
	entries := make([]*ports.ExerciseHistoryEntry, len(sessionIDs))
	w := 80000
	for i, sid := range sessionIDs {
		entries[i] = &ports.ExerciseHistoryEntry{
			SessionID:   sid,
			WorkoutName: "Treino A",
			PerformedAt: performedAts[i],
			Sets: []ports.SetDetail{
				{SetNumber: 1, Reps: 12, Weight: &w, Status: "completed"},
				{SetNumber: 2, Reps: 10, Weight: &w, Status: "completed"},
			},
		}
	}
	return entries
}

func TestGetExerciseHistoryUC_Execute(t *testing.T) {
	exerciseID := uuid.New()
	userID := uuid.New()
	sampleExercise := &entities.Exercise{ID: exerciseID, Name: "Supino Reto"}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	session1 := uuid.New()
	session2 := uuid.New()
	session3 := uuid.New()

	tests := []struct {
		name            string
		input           exercises.GetExerciseHistoryInput
		mockExercise    *entities.Exercise
		mockExerciseErr error
		mockEntries     []*ports.ExerciseHistoryEntry
		mockTotal       int
		mockHistoryErr  error
		wantTotal       int
		wantPage        int
		wantPageSize    int
		wantTotalPages  int
		wantEntriesLen  int
		wantErrIs       error
		wantErrContains string
	}{
		{
			name: "happy_path_with_history",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   20,
			},
			mockExercise: sampleExercise,
			mockEntries: makeHistoryEntries(
				[]uuid.UUID{session1, session2, session3},
				[]time.Time{now, yesterday, twoDaysAgo},
			),
			mockTotal:      3,
			wantTotal:      3,
			wantPage:       1,
			wantPageSize:   20,
			wantTotalPages: 1,
			wantEntriesLen: 3,
		},
		{
			name: "sets_grouped_by_session",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   20,
			},
			mockExercise: sampleExercise,
			mockEntries: makeHistoryEntries(
				[]uuid.UUID{session1},
				[]time.Time{now},
			),
			mockTotal:      1,
			wantTotal:      1,
			wantTotalPages: 1,
			wantEntriesLen: 1,
		},
		{
			name: "user_without_history_returns_empty",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   20,
			},
			mockExercise:   sampleExercise,
			mockEntries:    []*ports.ExerciseHistoryEntry{},
			mockTotal:      0,
			wantTotal:      0,
			wantTotalPages: 0,
			wantEntriesLen: 0,
		},
		{
			name: "pagination_page_2",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       2,
				PageSize:   10,
			},
			mockExercise: sampleExercise,
			mockEntries:  makeHistoryEntries([]uuid.UUID{session1}, []time.Time{now}),
			mockTotal:    50,
			wantTotal:    50,
			wantPage:     2,
			wantPageSize: 10,
			wantTotalPages: 5,
			wantEntriesLen: 1,
		},
		{
			name: "exercise_not_found",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: uuid.New(),
				UserID:     userID,
				Page:       1,
				PageSize:   20,
			},
			mockExercise: nil,
			wantErrIs:    domainerrors.ErrExerciseNotFound,
		},
		{
			name: "page_zero_returns_error",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       0,
				PageSize:   20,
			},
			wantErrContains: "page must be >= 1",
		},
		{
			name: "pageSize_over_limit_returns_error",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   200,
			},
			wantErrContains: "pageSize must be <= 100",
		},
		{
			name: "pageSize_zero_returns_error",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   0,
			},
			wantErrContains: "pageSize must be >= 1",
		},
		{
			name: "exercise_repo_error_propagates",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   20,
			},
			mockExerciseErr: errors.New("db error"),
			wantErrContains: "failed to check exercise",
		},
		{
			name: "history_repo_error_propagates",
			input: exercises.GetExerciseHistoryInput{
				ExerciseID: exerciseID,
				UserID:     userID,
				Page:       1,
				PageSize:   20,
			},
			mockExercise:   sampleExercise,
			mockHistoryErr: errors.New("history db error"),
			wantErrContains: "failed to get exercise history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockExerciseRepoForHistory{
				getByIDFunc: func(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
					return tt.mockExercise, tt.mockExerciseErr
				},
				getHistoryFunc: func(_ context.Context, _, _ uuid.UUID, _, _ int) ([]*ports.ExerciseHistoryEntry, int, error) {
					if tt.mockHistoryErr != nil {
						return nil, 0, tt.mockHistoryErr
					}
					return tt.mockEntries, tt.mockTotal, nil
				},
			}

			uc := exercises.NewGetExerciseHistoryUC(mockRepo)
			output, err := uc.Execute(context.Background(), tt.input)

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
			if len(output.Entries) != tt.wantEntriesLen {
				t.Errorf("Entries count: got %d, want %d", len(output.Entries), tt.wantEntriesLen)
			}

			// Verify sets are present when entries are returned
			for i, entry := range output.Entries {
				if len(entry.Sets) == 0 {
					t.Errorf("Entry %d has no sets", i)
				}
			}
		})
	}
}
