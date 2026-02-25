package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
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
		ID:         setRecord.ID,
		SessionID:  setRecord.SessionID,
		ExerciseID: setRecord.ExerciseID,
		SetNumber:  int32(setRecord.SetNumber),
		Weight:     int32(setRecord.Weight),
		Reps:       int32(setRecord.Reps),
		Status:     setRecord.Status,
		RecordedAt: setRecord.RecordedAt,
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
func (r *SetRecordRepository) FindBySessionExerciseSet(ctx context.Context, sessionID, exerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error) {
	row, err := r.q.FindSetRecordBySessionExerciseSet(ctx, queries.FindSetRecordBySessionExerciseSetParams{
		SessionID:  sessionID,
		ExerciseID: exerciseID,
		SetNumber:  int32(setNumber),
	})
	if err != nil {
		return nil, err
	}

	return &entities.SetRecord{
		ID:         row.ID,
		SessionID:  row.SessionID,
		ExerciseID: row.ExerciseID,
		SetNumber:  int(row.SetNumber),
		Weight:     int(row.Weight),
		Reps:       int(row.Reps),
		Status:     row.Status,
		RecordedAt: row.RecordedAt,
	}, nil
}
