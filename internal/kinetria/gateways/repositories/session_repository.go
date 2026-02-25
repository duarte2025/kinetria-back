package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// SessionRepository implements ports.SessionRepository using PostgreSQL via SQLC.
type SessionRepository struct {
	q *queries.Queries
}

// NewSessionRepository creates a new SessionRepository backed by the provided *sql.DB.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{q: queries.New(db)}
}

// Create inserts a new session into the database.
func (r *SessionRepository) Create(ctx context.Context, session *entities.Session) error {
	return r.q.CreateSession(ctx, queries.CreateSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		WorkoutID: session.WorkoutID,
		StartedAt: session.StartedAt,
		Status:    string(session.Status),
		Notes:     session.Notes,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	})
}

// FindActiveByUserID retrieves the active session for a user, if one exists.
// Returns (nil, nil) if no active session is found.
func (r *SessionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error) {
	row, err := r.q.FindActiveSessionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var finishedAt *time.Time
	if row.FinishedAt.Valid {
		finishedAt = &row.FinishedAt.Time
	}

	return &entities.Session{
		ID:         row.ID,
		UserID:     row.UserID,
		WorkoutID:  row.WorkoutID,
		Status:     vos.SessionStatus(row.Status),
		Notes:      row.Notes,
		StartedAt:  row.StartedAt,
		FinishedAt: finishedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

// FindByID retrieves a session by its ID.
func (r *SessionRepository) FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error) {
	row, err := r.q.FindSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var finishedAt *time.Time
	if row.FinishedAt.Valid {
		finishedAt = &row.FinishedAt.Time
	}

	return &entities.Session{
		ID:         row.ID,
		UserID:     row.UserID,
		WorkoutID:  row.WorkoutID,
		Status:     vos.SessionStatus(row.Status),
		Notes:      row.Notes,
		StartedAt:  row.StartedAt,
		FinishedAt: finishedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

// UpdateStatus updates the status, finishedAt and notes of a session.
// Returns (true, nil) if the session was updated, (false, nil) if the session was not active.
func (r *SessionRepository) UpdateStatus(ctx context.Context, sessionID uuid.UUID, status string, finishedAt *time.Time, notes string) (bool, error) {
	var finishedAtSQL sql.NullTime
	if finishedAt != nil {
		finishedAtSQL = sql.NullTime{Time: *finishedAt, Valid: true}
	}

	rowsAffected, err := r.q.UpdateSessionStatus(ctx, queries.UpdateSessionStatusParams{
		ID:         sessionID,
		Status:     status,
		FinishedAt: finishedAtSQL,
		Notes:      notes,
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// GetCompletedSessionsByUserAndDateRange retorna todas as sessões completed do usuário
// no intervalo de datas (inclusive).
func (r *SessionRepository) GetCompletedSessionsByUserAndDateRange(
	ctx context.Context,
	userID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) ([]entities.Session, error) {
	rows, err := r.q.GetCompletedSessionsByDateRange(ctx, queries.GetCompletedSessionsByDateRangeParams{
		UserID:      userID,
		StartedAt:   startDate,
		StartedAt_2: endDate,
	})
	if err != nil {
		return nil, err
	}

	sessions := make([]entities.Session, 0, len(rows))
	for _, row := range rows {
		var finishedAt *time.Time
		if row.FinishedAt.Valid {
			finishedAt = &row.FinishedAt.Time
		}

		sessions = append(sessions, entities.Session{
			ID:         row.ID,
			UserID:     row.UserID,
			WorkoutID:  row.WorkoutID,
			Status:     vos.SessionStatus(row.Status),
			Notes:      row.Notes,
			StartedAt:  row.StartedAt,
			FinishedAt: finishedAt,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}

	return sessions, nil
}
