package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

type UserID = uuid.UUID

type User struct {
	ID              UserID
	Email           string
	Name            string
	PasswordHash    string
	ProfileImageURL string
	Preferences     vos.UserPreferences
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
