package repositories

import (
	"database/sql"
	"testing"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// TestWorkoutRepositoryImplementsInterface verifies that WorkoutRepository
// implements the ports.WorkoutRepository interface at compile time.
func TestWorkoutRepositoryImplementsInterface(t *testing.T) {
	var db *sql.DB
	var _ ports.WorkoutRepository = NewWorkoutRepository(db)
}
