package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// WorkoutRepository implements ports.WorkoutRepository using PostgreSQL via SQLC.
type WorkoutRepository struct {
	q *queries.Queries
}

// NewWorkoutRepository creates a new WorkoutRepository backed by the provided *sql.DB.
func NewWorkoutRepository(db *sql.DB) *WorkoutRepository {
	return &WorkoutRepository{q: queries.New(db)}
}

// ExistsByIDAndUserID checks if a workout exists for the given ID and user ID.
func (r *WorkoutRepository) ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error) {
	return r.q.ExistsWorkoutByIDAndUserID(ctx, queries.ExistsWorkoutByIDAndUserIDParams{
		ID:     workoutID,
		UserID: userID,
	})
}
