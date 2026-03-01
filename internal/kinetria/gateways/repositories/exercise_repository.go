package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
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
		ExerciseID: exerciseID,
		WorkoutID:  workoutID,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

// FindWorkoutExerciseID finds the workout_exercise ID for a given exercise and workout.
func (r *ExerciseRepository) FindWorkoutExerciseID(ctx context.Context, exerciseID, workoutID uuid.UUID) (uuid.UUID, error) {
	return r.q.FindWorkoutExerciseID(ctx, queries.FindWorkoutExerciseIDParams{
		ExerciseID: exerciseID,
		WorkoutID:  workoutID,
	})
}

// List returns a paginated list of exercises from the library, optionally filtered.
func (r *ExerciseRepository) List(ctx context.Context, filters ports.ExerciseFilters, page, pageSize int) ([]*entities.Exercise, int, error) {
	offset := (page - 1) * pageSize

	params := queries.ListExercisesParams{
		Search:      toNullString(filters.Search),
		MuscleGroup: toNullString(filters.MuscleGroup),
		Equipment:   toNullString(filters.Equipment),
		Difficulty:  toNullString(filters.Difficulty),
		Limit:       int32(pageSize),
		Offset:      int32(offset),
	}

	countParams := queries.CountExercisesParams{
		Search:      params.Search,
		MuscleGroup: params.MuscleGroup,
		Equipment:   params.Equipment,
		Difficulty:  params.Difficulty,
	}

	total, err := r.q.CountExercises(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.ListExercises(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	exercises := make([]*entities.Exercise, 0, len(rows))
	for _, row := range rows {
		e := mapSQLCLibraryExerciseToEntity(row)
		exercises = append(exercises, &e)
	}

	return exercises, int(total), nil
}

// GetByID returns a single exercise by ID, or nil if not found.
func (r *ExerciseRepository) GetByID(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error) {
	row, err := r.q.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	e := mapSQLCLibraryExerciseToEntity(row)
	return &e, nil
}

// GetUserStats returns performance statistics for a specific user on an exercise.
func (r *ExerciseRepository) GetUserStats(ctx context.Context, userID, exerciseID uuid.UUID) (*ports.ExerciseUserStats, error) {
	row, err := r.q.GetExerciseUserStats(ctx, queries.GetExerciseUserStatsParams{
		ExerciseID: exerciseID,
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	stats := &ports.ExerciseUserStats{
		TimesPerformed: int(row.TimesPerformed),
	}
	if row.LastPerformed.Valid {
		t := row.LastPerformed.Time
		stats.LastPerformed = &t
	}
	if row.BestWeight.Valid {
		w := int(row.BestWeight.Int64)
		stats.BestWeight = &w
	}
	if row.AverageWeight.Valid {
		stats.AverageWeight = &row.AverageWeight.Float64
	}

	return stats, nil
}

// GetHistory returns a paginated list of sessions in which the user performed the exercise.
func (r *ExerciseRepository) GetHistory(ctx context.Context, userID, exerciseID uuid.UUID, page, pageSize int) ([]*ports.ExerciseHistoryEntry, int, error) {
	offset := (page - 1) * pageSize

	total, err := r.q.CountExerciseHistory(ctx, queries.CountExerciseHistoryParams{
		ExerciseID: exerciseID,
		UserID:     userID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.GetExerciseHistory(ctx, queries.GetExerciseHistoryParams{
		ExerciseID: exerciseID,
		UserID:     userID,
		Limit:      int32(pageSize),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	// Group set rows by sessionID
	entries := make([]*ports.ExerciseHistoryEntry, 0)
	sessionIndex := make(map[uuid.UUID]int) // sessionID → index in entries

	for _, row := range rows {
		idx, exists := sessionIndex[row.SessionID]
		if !exists {
			entry := &ports.ExerciseHistoryEntry{
				SessionID:   row.SessionID,
				WorkoutName: row.WorkoutName,
				PerformedAt: row.PerformedAt,
				Sets:        []ports.SetDetail{},
			}
			entries = append(entries, entry)
			idx = len(entries) - 1
			sessionIndex[row.SessionID] = idx
		}
		w := int(row.Weight)
		entries[idx].Sets = append(entries[idx].Sets, ports.SetDetail{
			SetNumber: int(row.SetNumber),
			Reps:      int(row.Reps),
			Weight:    &w,
			Status:    row.Status,
		})
	}

	return entries, int(total), nil
}

// toNullString converts a *string to sql.NullString.
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// mapSQLCLibraryExerciseToEntity converts a queries.Exercise to entities.Exercise for library use.
func mapSQLCLibraryExerciseToEntity(row queries.Exercise) entities.Exercise {
	var muscles []string
	if len(row.Muscles) > 0 {
		_ = json.Unmarshal(row.Muscles, &muscles)
	}

	e := entities.Exercise{
		ID:           row.ID,
		Name:         row.Name,
		ThumbnailURL: row.ThumbnailUrl,
		Muscles:      muscles,
	}

	if row.Description != "" {
		e.Description = &row.Description
	}
	if row.Instructions.Valid && row.Instructions.String != "" {
		e.Instructions = &row.Instructions.String
	}
	if row.Tips.Valid && row.Tips.String != "" {
		e.Tips = &row.Tips.String
	}
	if row.Difficulty.Valid && row.Difficulty.String != "" {
		e.Difficulty = &row.Difficulty.String
	}
	if row.Equipment.Valid && row.Equipment.String != "" {
		e.Equipment = &row.Equipment.String
	}
	if row.VideoUrl.Valid && row.VideoUrl.String != "" {
		e.VideoURL = &row.VideoUrl.String
	}

	return e
}
