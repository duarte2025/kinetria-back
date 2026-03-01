package exercises

import (
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// ExerciseWithStats combines an exercise with optional user performance statistics.
// UserStats is nil when the request is unauthenticated or the user has never performed the exercise.
type ExerciseWithStats struct {
	Exercise  *entities.Exercise
	UserStats *ports.ExerciseUserStats
}
