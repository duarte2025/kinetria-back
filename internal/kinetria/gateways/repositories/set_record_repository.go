package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// SetRecordRepository implements ports.SetRecordRepository using SQLC.
type SetRecordRepository struct {
	db *sql.DB
	q  *queries.Queries
}

// NewSetRecordRepository creates a new SetRecordRepository.
func NewSetRecordRepository(db *sql.DB) *SetRecordRepository {
	return &SetRecordRepository{db: db, q: queries.New(db)}
}

// Create inserts a new set record.
// Returns ErrSetAlreadyRecorded if a unique constraint violation occurs (concurrent insert).
func (r *SetRecordRepository) Create(ctx context.Context, setRecord *entities.SetRecord) error {
	err := r.q.CreateSetRecord(ctx, queries.CreateSetRecordParams{
		ID:                setRecord.ID,
		SessionID:         setRecord.SessionID,
		WorkoutExerciseID: uuid.NullUUID{UUID: setRecord.WorkoutExerciseID, Valid: true},
		SetNumber:         int32(setRecord.SetNumber),
		Weight:            int32(setRecord.Weight),
		Reps:              int32(setRecord.Reps),
		Status:            setRecord.Status,
		RecordedAt:        setRecord.RecordedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.ErrSetAlreadyRecorded
		}
		return err
	}
	return nil
}

// FindBySessionExerciseSet finds a set record by session, exercise and set number.
func (r *SetRecordRepository) FindBySessionExerciseSet(ctx context.Context, sessionID, workoutExerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error) {
	row, err := r.q.FindSetRecordBySessionExerciseSet(ctx, queries.FindSetRecordBySessionExerciseSetParams{
		SessionID:         sessionID,
		WorkoutExerciseID: uuid.NullUUID{UUID: workoutExerciseID, Valid: true},
		SetNumber:         int32(setNumber),
	})
	if err != nil {
		return nil, err
	}

	return &entities.SetRecord{
		ID:                row.ID,
		SessionID:         row.SessionID,
		WorkoutExerciseID: row.WorkoutExerciseID.UUID,
		SetNumber:         int(row.SetNumber),
		Weight:            int(row.Weight),
		Reps:              int(row.Reps),
		Status:            row.Status,
		RecordedAt:        row.RecordedAt,
	}, nil
}

// GetTotalSetsRepsVolume retorna estatísticas agregadas de sets do usuário no período.
func (r *SetRecordRepository) GetTotalSetsRepsVolume(ctx context.Context, userID uuid.UUID, start, end time.Time) (*ports.SetRecordStats, error) {
	row, err := r.q.GetTotalSetsRepsVolume(ctx, queries.GetTotalSetsRepsVolumeParams{
		UserID:      userID,
		StartedAt:   start,
		StartedAt_2: end,
	})
	if err != nil {
		return nil, err
	}
	return &ports.SetRecordStats{
		TotalSets:   int(row.TotalSets),
		TotalReps:   int(row.TotalReps),
		TotalVolume: row.TotalVolume,
	}, nil
}

// GetPersonalRecordsByUser retorna os recordes pessoais do usuário por grupo muscular.
func (r *SetRecordRepository) GetPersonalRecordsByUser(ctx context.Context, userID uuid.UUID) ([]ports.PersonalRecord, error) {
	rows, err := r.q.GetPersonalRecordsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]ports.PersonalRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, ports.PersonalRecord{
			ExerciseID:   row.ExerciseID,
			ExerciseName: row.ExerciseName,
			Weight:       int(row.Weight),
			Reps:         int(row.Reps),
			Volume:       row.Volume,
			AchievedAt:   row.AchievedAt,
		})
	}
	return result, nil
}

// GetProgressionByUserAndExercise retorna a progressão de treinos do usuário no período.
func (r *SetRecordRepository) GetProgressionByUserAndExercise(ctx context.Context, userID uuid.UUID, exerciseID *uuid.UUID, start, end time.Time) ([]ports.ProgressionPoint, error) {
	var nullExID uuid.NullUUID
	if exerciseID != nil {
		nullExID = uuid.NullUUID{UUID: *exerciseID, Valid: true}
	}
	rows, err := r.q.GetProgressionByUserAndExercise(ctx, queries.GetProgressionByUserAndExerciseParams{
		UserID:      userID,
		StartedAt:   start,
		StartedAt_2: end,
		ExerciseID:  nullExID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]ports.ProgressionPoint, 0, len(rows))
	for _, row := range rows {
		result = append(result, ports.ProgressionPoint{
			Date:        row.Date,
			MaxWeight:   row.MaxWeight,
			TotalVolume: row.TotalVolume,
		})
	}
	return result, nil
}
