package entities

import (
	"time"

	"github.com/google/uuid"
)

type RefreshTokenID = uuid.UUID

type RefreshToken struct {
	ID        RefreshTokenID
	UserID    UserID
	Token     string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
