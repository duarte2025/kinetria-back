package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

// RefreshTokenRepository defines persistence operations for refresh tokens.
// The token field stores a SHA-256 hash of the actual token (never the plain text).
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
	RevokeByToken(ctx context.Context, tokenHash string) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// SessionRepository defines persistence operations for workout sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error)
	FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error)
	UpdateStatus(ctx context.Context, sessionID uuid.UUID, status string, finishedAt *time.Time, notes string) (bool, error)
	// GetCompletedSessionsByUserAndDateRange retorna todas as sessões completed do usuário
	// no intervalo de datas (inclusive).
	// Datas devem estar em UTC. Usa DATE(started_at) para determinar o dia.
	GetCompletedSessionsByUserAndDateRange(
		ctx context.Context,
		userID uuid.UUID,
		startDate time.Time,
		endDate time.Time,
	) ([]entities.Session, error)
}

// SetRecordRepository defines persistence operations for set records.
type SetRecordRepository interface {
	Create(ctx context.Context, setRecord *entities.SetRecord) error
	FindBySessionExerciseSet(ctx context.Context, sessionID, exerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error)
}

// ExerciseRepository defines persistence operations for exercises.
type ExerciseRepository interface {
	ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error)
}

// AuditLogRepository defines persistence for audit log entries (append-only).
type AuditLogRepository interface {
	Append(ctx context.Context, entry *entities.AuditLog) error
}
