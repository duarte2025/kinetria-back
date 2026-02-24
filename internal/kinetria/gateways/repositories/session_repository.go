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
