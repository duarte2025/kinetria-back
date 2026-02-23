package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserID = uuid.UUID

type User struct {
	ID              UserID
	Email           string
	Name            string
	PasswordHash    string
	ProfileImageURL string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
