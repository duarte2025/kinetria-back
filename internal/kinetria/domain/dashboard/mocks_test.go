package dashboard_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
)

// mockUserRepository is a mock implementation of ports.UserRepository for testing.
type mockUserRepository struct {
	user       *entities.User
	getByIDErr error
}

func (m *mockUserRepository) Create(_ context.Context, _ *entities.User) error {
	return nil
}

func (m *mockUserRepository) GetByEmail(_ context.Context, _ string) (*entities.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.user, nil
}

func (m *mockUserRepository) Update(_ context.Context, _ *entities.User) error {
	return nil
}

// mockWorkoutRepository is a mock implementation of ports.WorkoutRepository for testing.
type mockWorkoutRepository struct {
	firstWorkout      *entities.Workout
	getFirstByUserErr error
}

func (m *mockWorkoutRepository) ExistsByIDAndUserID(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockWorkoutRepository) ListByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]entities.Workout, int, error) {
	return nil, 0, nil
}

func (m *mockWorkoutRepository) GetFirstByUserID(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	if m.getFirstByUserErr != nil {
		return nil, m.getFirstByUserErr
	}
	return m.firstWorkout, nil
}

func (m *mockWorkoutRepository) GetByID(_ context.Context, _, _ uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
	return nil, nil, nil
}

func (m *mockWorkoutRepository) GetByIDOnly(_ context.Context, _ uuid.UUID) (*entities.Workout, error) {
	return nil, nil
}

func (m *mockWorkoutRepository) Create(_ context.Context, _ entities.Workout, _ []entities.WorkoutExercise) error {
	return nil
}

func (m *mockWorkoutRepository) Update(_ context.Context, _ entities.Workout, _ []entities.WorkoutExercise) error {
	return nil
}

func (m *mockWorkoutRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockWorkoutRepository) HasActiveSessions(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

// mockSessionRepository is a mock implementation of ports.SessionRepository for testing.
type mockSessionRepository struct {
	completedSessions []entities.Session
	completedErr      error
}

func (m *mockSessionRepository) Create(_ context.Context, _ *entities.Session) error {
	return nil
}

func (m *mockSessionRepository) FindActiveByUserID(_ context.Context, _ uuid.UUID) (*entities.Session, error) {
	return nil, nil
}

func (m *mockSessionRepository) FindByID(_ context.Context, _ uuid.UUID) (*entities.Session, error) {
	return nil, nil
}

func (m *mockSessionRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _ string, _ *time.Time, _ string) (bool, error) {
	return true, nil
}

func (m *mockSessionRepository) GetCompletedSessionsByUserAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]entities.Session, error) {
	if m.completedErr != nil {
		return nil, m.completedErr
	}
	return m.completedSessions, nil
}
