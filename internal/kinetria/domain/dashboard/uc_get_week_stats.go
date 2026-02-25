package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetWeekStatsInput struct {
	UserID uuid.UUID
}

type GetWeekStatsOutput struct {
	Calories         int
	TotalTimeMinutes int
}

type GetWeekStatsUC struct {
	sessionRepo ports.SessionRepository
}

func NewGetWeekStatsUC(sessionRepo ports.SessionRepository) *GetWeekStatsUC {
	return &GetWeekStatsUC{sessionRepo: sessionRepo}
}

func (uc *GetWeekStatsUC) Execute(ctx context.Context, input GetWeekStatsInput) (*GetWeekStatsOutput, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDate := today.AddDate(0, 0, -6)
	endDate := today

	sessions, err := uc.sessionRepo.GetCompletedSessionsByUserAndDateRange(ctx, input.UserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	totalMinutes := 0
	for _, s := range sessions {
		if s.FinishedAt != nil {
			duration := s.FinishedAt.Sub(s.StartedAt)
			totalMinutes += int(duration.Minutes())
		}
	}

	// Calorias estimadas: 7 kcal/min (ACSM guideline para exercício moderado)
	calories := totalMinutes * 7

	return &GetWeekStatsOutput{
		Calories:         calories,
		TotalTimeMinutes: totalMinutes,
	}, nil
}
