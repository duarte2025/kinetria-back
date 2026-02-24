package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
)

// WorkoutRepository defines persistence operations for workouts.
type WorkoutRepository interface {
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
}
