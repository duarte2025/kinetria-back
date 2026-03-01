package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// WorkoutRepository implements ports.WorkoutRepository using PostgreSQL via SQLC.
type WorkoutRepository struct {
	q  *queries.Queries
	db *sql.DB
}

// NewWorkoutRepository creates a new WorkoutRepository backed by the provided *sql.DB.
func NewWorkoutRepository(db *sql.DB) *WorkoutRepository {
	return &WorkoutRepository{q: queries.New(db), db: db}
}

// ExistsByIDAndUserID checks if a workout exists for the given ID and user ID.
func (r *WorkoutRepository) ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error) {
	return r.q.ExistsWorkoutByIDAndUserID(ctx, queries.ExistsWorkoutByIDAndUserIDParams{
		ID:     workoutID,
		UserID: userID,
	})
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

// GetFirstByUserID retorna o primeiro workout do usuário (ordenado por created_at ASC).
// Retorna nil se o usuário não tiver workouts.
func (r *WorkoutRepository) GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error) {
	row, err := r.q.GetFirstWorkoutByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	workout := mapSQLCWorkoutToEntity(row)
	return &workout, nil
}

// GetByID retorna um workout com seus exercises pelo ID, validando ownership do usuário.
// Retorna (nil, nil, nil) se o workout não for encontrado ou não pertencer ao usuário.
func (r *WorkoutRepository) GetByID(ctx context.Context, workoutID, userID uuid.UUID) (*entities.Workout, []entities.Exercise, error) {
	// 1. Buscar workout com validação de ownership
	workoutRow, err := r.q.GetWorkoutByID(ctx, queries.GetWorkoutByIDParams{
		ID:     workoutID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil // Workout não encontrado ou não pertence ao usuário
		}
		return nil, nil, fmt.Errorf("failed to get workout: %w", err)
	}

	// 2. Buscar exercises do workout
	exerciseRows, err := r.q.ListExercisesByWorkoutID(ctx, workoutID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list exercises: %w", err)
	}

	// 3. Mapear para entidades de domínio
	workout := mapSQLCWorkoutToEntity(workoutRow)
	exercises := make([]entities.Exercise, len(exerciseRows))
	for i, row := range exerciseRows {
		exercises[i] = mapSQLCExerciseToEntity(row)
	}

	return &workout, exercises, nil
}

// mapSQLCWorkoutToEntity converts a queries.Workout (SQLC) to entities.Workout (domain).
func mapSQLCWorkoutToEntity(sqlcWorkout queries.Workout) entities.Workout {
	workout := entities.Workout{
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
	if sqlcWorkout.CreatedBy.Valid {
		workout.CreatedBy = &sqlcWorkout.CreatedBy.UUID
	}
	if sqlcWorkout.DeletedAt.Valid {
		workout.DeletedAt = &sqlcWorkout.DeletedAt.Time
	}
	return workout
}

// GetByIDOnly returns a workout by ID without user ownership validation.
func (r *WorkoutRepository) GetByIDOnly(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error) {
	row, err := r.q.GetWorkoutByIDOnly(ctx, workoutID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}
	workout := mapSQLCWorkoutToEntity(row)
	return &workout, nil
}

// Create creates a new workout with exercises in a single transaction.
func (r *WorkoutRepository) Create(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	createdBy := uuid.NullUUID{}
	if workout.CreatedBy != nil {
		createdBy = uuid.NullUUID{UUID: *workout.CreatedBy, Valid: true}
	}

	err = qtx.CreateWorkout(ctx, queries.CreateWorkoutParams{
		ID:          workout.ID,
		UserID:      workout.UserID,
		Name:        workout.Name,
		Description: workout.Description,
		Type:        workout.Type,
		Intensity:   workout.Intensity,
		Duration:    int32(workout.Duration),
		ImageUrl:    workout.ImageURL,
		CreatedBy:   createdBy,
		CreatedAt:   workout.CreatedAt,
		UpdatedAt:   workout.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}

	for _, ex := range exercises {
		weight := int32(0)
		if ex.Weight > 0 {
			weight = int32(ex.Weight)
		}
		err = qtx.CreateWorkoutExercise(ctx, queries.CreateWorkoutExerciseParams{
			ID:         ex.ID,
			WorkoutID:  ex.WorkoutID,
			ExerciseID: ex.ExerciseID,
			Sets:       int32(ex.Sets),
			Reps:       ex.Reps,
			RestTime:   int32(ex.RestTime),
			Weight:     weight,
			OrderIndex: int32(ex.OrderIndex),
			CreatedAt:  workout.CreatedAt,
			UpdatedAt:  workout.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create workout exercise: %w", err)
		}
	}

	return tx.Commit()
}

// Update updates an existing workout and replaces its exercises in a single transaction.
func (r *WorkoutRepository) Update(ctx context.Context, workout entities.Workout, exercises []entities.WorkoutExercise) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	err = qtx.UpdateWorkout(ctx, queries.UpdateWorkoutParams{
		ID:          workout.ID,
		Name:        workout.Name,
		Description: workout.Description,
		Type:        workout.Type,
		Intensity:   workout.Intensity,
		Duration:    int32(workout.Duration),
		ImageUrl:    workout.ImageURL,
		UpdatedAt:   workout.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	err = qtx.DeleteWorkoutExercises(ctx, workout.ID)
	if err != nil {
		return fmt.Errorf("failed to delete workout exercises: %w", err)
	}

	for _, ex := range exercises {
		weight := int32(0)
		if ex.Weight > 0 {
			weight = int32(ex.Weight)
		}
		err = qtx.CreateWorkoutExercise(ctx, queries.CreateWorkoutExerciseParams{
			ID:         ex.ID,
			WorkoutID:  ex.WorkoutID,
			ExerciseID: ex.ExerciseID,
			Sets:       int32(ex.Sets),
			Reps:       ex.Reps,
			RestTime:   int32(ex.RestTime),
			Weight:     weight,
			OrderIndex: int32(ex.OrderIndex),
			CreatedAt:  workout.UpdatedAt,
			UpdatedAt:  workout.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create workout exercise: %w", err)
		}
	}

	return tx.Commit()
}

// Delete soft-deletes a workout by setting deleted_at.
func (r *WorkoutRepository) Delete(ctx context.Context, workoutID uuid.UUID) error {
	now := time.Now().UTC()
	err := r.q.SoftDeleteWorkout(ctx, queries.SoftDeleteWorkoutParams{
		ID:        workoutID,
		DeletedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
	})
	if err != nil {
		return fmt.Errorf("failed to soft delete workout: %w", err)
	}
	return nil
}

// HasActiveSessions checks if a workout has any active sessions.
func (r *WorkoutRepository) HasActiveSessions(ctx context.Context, workoutID uuid.UUID) (bool, error) {
	has, err := r.q.HasActiveSessions(ctx, workoutID)
	if err != nil {
		return false, fmt.Errorf("failed to check active sessions: %w", err)
	}
	return has, nil
}

// mapSQLCExerciseToEntity converts queries.ListExercisesByWorkoutIDRow to entities.Exercise.
func mapSQLCExerciseToEntity(row queries.ListExercisesByWorkoutIDRow) entities.Exercise {
	// Deserializar muscles (JSONB → []string)
	var muscles []string
	if len(row.Muscles) > 0 {
		_ = json.Unmarshal(row.Muscles, &muscles) // Ignora erro, usa slice vazio se falhar
	}

	return entities.Exercise{
		ID:           row.ID,
		Name:         row.Name,
		ThumbnailURL: row.ThumbnailUrl,
		Muscles:      muscles,
		Sets:         int(row.Sets),
		Reps:         row.Reps,
		RestTime:     int(row.RestTime),
		Weight:       int(row.Weight),
		OrderIndex:   int(row.OrderIndex),
	}
}
