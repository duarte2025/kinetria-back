package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// SetRecordRepository implements ports.SetRecordRepository using SQLC.
type SetRecordRepository struct {
	db *sql.DB
}

// NewSetRecordRepository creates a new SetRecordRepository.
func NewSetRecordRepository(db *sql.DB) *SetRecordRepository {
	return &SetRecordRepository{db: db}
}

// Create inserts a new set record.
func (r *SetRecordRepository) Create(ctx context.Context, setRecord *entities.SetRecord) error {
	q := queries.New(r.db)
	return q.CreateSetRecord(ctx, queries.CreateSetRecordParams{
		ID:         setRecord.ID,
		SessionID:  setRecord.SessionID,
		ExerciseID: setRecord.ExerciseID,
		SetNumber:  int32(setRecord.SetNumber),
		Weight:     int32(setRecord.Weight),
		Reps:       int32(setRecord.Reps),
		Status:     setRecord.Status,
		RecordedAt: setRecord.RecordedAt,
	})
}

// FindBySessionExerciseSet finds a set record by session, exercise and set number.
func (r *SetRecordRepository) FindBySessionExerciseSet(ctx context.Context, sessionID, exerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error) {
	q := queries.New(r.db)
	row, err := q.FindSetRecordBySessionExerciseSet(ctx, queries.FindSetRecordBySessionExerciseSetParams{
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
