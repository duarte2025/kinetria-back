package entities

import (
	"time"

	"github.com/google/uuid"
)

type RefreshTokenID = uuid.UUID

type RefreshToken struct {
	ID        RefreshTokenID
	UserID    UserID
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}
