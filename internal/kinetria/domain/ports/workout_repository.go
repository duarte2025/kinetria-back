package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
)

// WorkoutRepository defines persistence operations for workouts.
type WorkoutRepository interface {
	// ExistsByIDAndUserID checks if a workout exists for the given user.
	ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error)

	// ListByUserID returns paginated workouts for a user.
	// Parameters:
	//   - ctx: context for cancellation/timeout
	//   - userID: UUID of the authenticated user
	//   - offset: number of records to skip
	//   - limit: maximum number of records to return
	// Returns:
	//   - workouts slice
	//   - total count of workouts for the user
	//   - error if query fails
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error)

	// GetFirstByUserID retorna o primeiro workout do usuário (ordenado por created_at ASC).
	// Retorna nil se o usuário não tiver workouts.
	GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error)

	// GetByID returns a specific workout with its exercises.
	// Parameters:
	//   - ctx: context for cancellation/timeout
	//   - workoutID: UUID of the workout
	//   - userID: UUID of the authenticated user (validates ownership)
	// Returns:
	//   - *entities.Workout: the workout found (nil if not found or user doesn't own it)
	//   - []entities.Exercise: list of exercises for the workout
	//   - error: error if query fails
	GetByID(ctx context.Context, workoutID, userID uuid.UUID) (*entities.Workout, []entities.Exercise, error)

	// GetByIDOnly returns a workout by ID without user ownership validation.
	// Used for update/delete operations where ownership is checked in the use case.
	GetByIDOnly(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)

	// Create creates a new workout with exercises (transactional).
	Create(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error

	// Update updates an existing workout and replaces its exercises (transactional).
	Update(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error

	// Delete soft-deletes a workout by setting deleted_at.
	Delete(ctx context.Context, workoutID uuid.UUID) error

	// HasActiveSessions checks if a workout has any active sessions.
	HasActiveSessions(ctx context.Context, workoutID uuid.UUID) (bool, error)
}
