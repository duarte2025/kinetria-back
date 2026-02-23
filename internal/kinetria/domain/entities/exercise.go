package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

type ExerciseID = uuid.UUID

type Exercise struct {
	ID                 ExerciseID
	Name               string
	Description        string
	Category           vos.ExerciseCategory
	PrimaryMuscleGroup vos.MuscleGroup
	EquipmentRequired  string
	DifficultyLevel    int
	VideoURL           string
	ThumbnailURL       string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
