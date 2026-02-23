package ports

import (
	"context"

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
}

// WorkoutRepository defines persistence operations for workouts.
type WorkoutRepository interface {
	ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error)
}

// AuditLogRepository defines persistence for audit log entries (append-only).
type AuditLogRepository interface {
	Append(ctx context.Context, entry *entities.AuditLog) error
}
