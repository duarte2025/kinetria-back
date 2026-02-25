package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// ExerciseRepository implements ports.ExerciseRepository using SQLC.
type ExerciseRepository struct {
	db *sql.DB
	q  *queries.Queries
}

// NewExerciseRepository creates a new ExerciseRepository.
func NewExerciseRepository(db *sql.DB) *ExerciseRepository {
	return &ExerciseRepository{db: db, q: queries.New(db)}
}

// ExistsByIDAndWorkoutID checks if an exercise exists and belongs to a workout.
func (r *ExerciseRepository) ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error) {
	result, err := r.q.ExistsExerciseByIDAndWorkoutID(ctx, queries.ExistsExerciseByIDAndWorkoutIDParams{
		ID:        exerciseID,
		WorkoutID: workoutID,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}
