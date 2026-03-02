package statistics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// GetPersonalRecordsUC retrieves personal records for a user.
type GetPersonalRecordsUC struct {
	setRecordRepo ports.SetRecordRepository
}

// NewGetPersonalRecordsUC creates a new GetPersonalRecordsUC.
func NewGetPersonalRecordsUC(setRecordRepo ports.SetRecordRepository) *GetPersonalRecordsUC {
	return &GetPersonalRecordsUC{setRecordRepo: setRecordRepo}
}

// Execute returns personal records for the given user.
// Returns at most 15 PRs, one per muscle group, ordered by weight descending.
func (uc *GetPersonalRecordsUC) Execute(ctx context.Context, userID uuid.UUID) ([]PersonalRecord, error) {
	rows, err := uc.setRecordRepo.GetPersonalRecordsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get personal records: %w", err)
	}

	result := make([]PersonalRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, PersonalRecord{
			ExerciseID:   row.ExerciseID,
			ExerciseName: row.ExerciseName,
			Weight:       row.Weight,
			Reps:         row.Reps,
			Volume:       row.Volume,
			AchievedAt:   row.AchievedAt,
		})
	}
	return result, nil
}
