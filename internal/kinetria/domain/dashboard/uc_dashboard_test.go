package dashboard_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestGetUserProfileUC_Execute(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	userID := uuid.New()

	tests := []struct {
		name        string
		userRepo    *mockUserRepository
		wantErr     bool
		checkOutput func(t *testing.T, out *dashboard.GetUserProfileOutput)
	}{
		{
			name: "success - returns user profile",
			userRepo: &mockUserRepository{
				user: &entities.User{
					ID:              userID,
					Name:            "Alice",
					Email:           "alice@example.com",
					ProfileImageURL: "https://example.com/avatar.png",
				},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetUserProfileOutput) {
				if out.ID != userID {
					t.Errorf("ID = %v, want %v", out.ID, userID)
				}
				if out.Name != "Alice" {
					t.Errorf("Name = %v, want Alice", out.Name)
				}
				if out.Email != "alice@example.com" {
					t.Errorf("Email = %v, want alice@example.com", out.Email)
				}
				if out.ProfileImageURL != "https://example.com/avatar.png" {
					t.Errorf("ProfileImageURL = %v, want https://example.com/avatar.png", out.ProfileImageURL)
				}
			},
		},
		{
			name: "error - user not found",
			userRepo: &mockUserRepository{
				getByIDErr: errors.New("not found"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := dashboard.NewGetUserProfileUC(tracer, tt.userRepo)
			out, err := uc.Execute(context.Background(), dashboard.GetUserProfileInput{UserID: userID})

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, out)
			}
		})
	}
}

func TestGetTodayWorkoutUC_Execute(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	userID := uuid.New()
	workoutID := uuid.New()

	tests := []struct {
		name        string
		workoutRepo *mockWorkoutRepository
		wantErr     bool
		checkOutput func(t *testing.T, out *dashboard.GetTodayWorkoutOutput)
	}{
		{
			name: "success - user has a workout",
			workoutRepo: &mockWorkoutRepository{
				firstWorkout: &entities.Workout{
					ID:   workoutID,
					Name: "Upper Body",
					Type: "strength",
				},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetTodayWorkoutOutput) {
				if out.Workout == nil {
					t.Fatal("Workout should not be nil")
				}
				if out.Workout.ID != workoutID {
					t.Errorf("Workout.ID = %v, want %v", out.Workout.ID, workoutID)
				}
			},
		},
		{
			name: "success - user has no workouts",
			workoutRepo: &mockWorkoutRepository{
				firstWorkout: nil,
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetTodayWorkoutOutput) {
				if out.Workout != nil {
					t.Error("Workout should be nil when user has no workouts")
				}
			},
		},
		{
			name: "error - workout repo fails",
			workoutRepo: &mockWorkoutRepository{
				getFirstByUserErr: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := dashboard.NewGetTodayWorkoutUC(tracer, tt.workoutRepo)
			out, err := uc.Execute(context.Background(), dashboard.GetTodayWorkoutInput{UserID: userID})

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, out)
			}
		})
	}
}

func TestGetWeekProgressUC_Execute(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	userID := uuid.New()

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	sessionOnYesterday := entities.Session{
		ID:        uuid.New(),
		UserID:    userID,
		StartedAt: yesterday.Add(10 * time.Hour),
	}
	sessionOnTwoDaysAgo := entities.Session{
		ID:        uuid.New(),
		UserID:    userID,
		StartedAt: twoDaysAgo.Add(10 * time.Hour),
	}

	tests := []struct {
		name        string
		sessionRepo *mockSessionRepository
		wantErr     bool
		checkOutput func(t *testing.T, out *dashboard.GetWeekProgressOutput)
	}{
		{
			name: "success - empty week returns 7 days all missed or future",
			sessionRepo: &mockSessionRepository{
				completedSessions: nil,
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetWeekProgressOutput) {
				if len(out.Days) != 7 {
					t.Errorf("Days count = %d, want 7", len(out.Days))
					return
				}
				for _, d := range out.Days {
					if d.Status != "missed" && d.Status != "future" {
						t.Errorf("Day %s has unexpected status %q (expected missed or future)", d.Date, d.Status)
					}
				}
			},
		},
		{
			name: "success - sessions on some days mark them as completed",
			sessionRepo: &mockSessionRepository{
				completedSessions: []entities.Session{sessionOnYesterday, sessionOnTwoDaysAgo},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetWeekProgressOutput) {
				if len(out.Days) != 7 {
					t.Fatalf("Days count = %d, want 7", len(out.Days))
				}
				completedCount := 0
				for _, d := range out.Days {
					if d.Status == "completed" {
						completedCount++
					}
				}
				if completedCount != 2 {
					t.Errorf("completed days = %d, want 2", completedCount)
				}
			},
		},
		{
			name: "success - today is included as missed if no session today",
			sessionRepo: &mockSessionRepository{
				completedSessions: nil,
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetWeekProgressOutput) {
				todayStr := today.Format("2006-01-02")
				found := false
				for _, d := range out.Days {
					if d.Date == todayStr {
						found = true
						if d.Status != "missed" {
							t.Errorf("today status = %q, want missed", d.Status)
						}
					}
				}
				if !found {
					t.Error("today not found in week progress days")
				}
			},
		},
		{
			name: "error - session repo fails",
			sessionRepo: &mockSessionRepository{
				completedErr: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := dashboard.NewGetWeekProgressUC(tracer, tt.sessionRepo)
			out, err := uc.Execute(context.Background(), dashboard.GetWeekProgressInput{UserID: userID})

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, out)
			}
		})
	}
}

func TestGetWeekStatsUC_Execute(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	userID := uuid.New()

	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)

	finishedAt60 := yesterday.Add(60 * time.Minute)
	finishedAt30 := yesterday.Add(30 * time.Minute)

	session60min := entities.Session{
		ID:         uuid.New(),
		UserID:     userID,
		StartedAt:  yesterday,
		FinishedAt: &finishedAt60,
	}
	session30min := entities.Session{
		ID:         uuid.New(),
		UserID:     userID,
		StartedAt:  yesterday,
		FinishedAt: &finishedAt30,
	}
	sessionNoFinish := entities.Session{
		ID:         uuid.New(),
		UserID:     userID,
		StartedAt:  yesterday,
		FinishedAt: nil,
	}

	tests := []struct {
		name        string
		sessionRepo *mockSessionRepository
		wantErr     bool
		checkOutput func(t *testing.T, out *dashboard.GetWeekStatsOutput)
	}{
		{
			name: "success - no sessions returns zero stats",
			sessionRepo: &mockSessionRepository{
				completedSessions: nil,
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetWeekStatsOutput) {
				if out.TotalTimeMinutes != 0 {
					t.Errorf("TotalTimeMinutes = %d, want 0", out.TotalTimeMinutes)
				}
				if out.Calories != 0 {
					t.Errorf("Calories = %d, want 0", out.Calories)
				}
			},
		},
		{
			name: "success - calculates total time and calories correctly",
			sessionRepo: &mockSessionRepository{
				completedSessions: []entities.Session{session60min, session30min},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetWeekStatsOutput) {
				if out.TotalTimeMinutes != 90 {
					t.Errorf("TotalTimeMinutes = %d, want 90", out.TotalTimeMinutes)
				}
				if out.Calories != 90*7 {
					t.Errorf("Calories = %d, want %d", out.Calories, 90*7)
				}
			},
		},
		{
			name: "success - sessions without finished_at are excluded from time calculation",
			sessionRepo: &mockSessionRepository{
				completedSessions: []entities.Session{session60min, sessionNoFinish},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, out *dashboard.GetWeekStatsOutput) {
				if out.TotalTimeMinutes != 60 {
					t.Errorf("TotalTimeMinutes = %d, want 60 (sessionNoFinish excluded)", out.TotalTimeMinutes)
				}
			},
		},
		{
			name: "error - session repo fails",
			sessionRepo: &mockSessionRepository{
				completedErr: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := dashboard.NewGetWeekStatsUC(tracer, tt.sessionRepo)
			out, err := uc.Execute(context.Background(), dashboard.GetWeekStatsInput{UserID: userID})

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, out)
			}
		})
	}
}
