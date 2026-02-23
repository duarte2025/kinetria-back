package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
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

// ListByUserID returns paginated workouts for a user.
func (r *WorkoutRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error) {
	// Count total workouts for the user
	total, err := r.q.CountWorkoutsByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// List workouts with pagination
	sqlcWorkouts, err := r.q.ListWorkoutsByUserID(ctx, queries.ListWorkoutsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	// Map SQLC workouts to domain entities
	workouts := make([]entities.Workout, len(sqlcWorkouts))
	for i, sqlcWorkout := range sqlcWorkouts {
		workouts[i] = mapSQLCWorkoutToEntity(sqlcWorkout)
	}

	return workouts, int(total), nil
}

// mapSQLCWorkoutToEntity converts a queries.Workout (SQLC) to entities.Workout (domain).
func mapSQLCWorkoutToEntity(sqlcWorkout queries.Workout) entities.Workout {
	return entities.Workout{
		ID:          sqlcWorkout.ID,
		UserID:      sqlcWorkout.UserID,
		Name:        sqlcWorkout.Name,
		Description: sqlcWorkout.Description,
		Type:        sqlcWorkout.Type,
		Intensity:   sqlcWorkout.Intensity,
		Duration:    int(sqlcWorkout.Duration),
		ImageURL:    sqlcWorkout.ImageUrl,
		CreatedAt:   sqlcWorkout.CreatedAt,
		UpdatedAt:   sqlcWorkout.UpdatedAt,
	}
}
