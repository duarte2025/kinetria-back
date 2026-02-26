package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"go.opentelemetry.io/otel/trace"
)

type GetWeekStatsInput struct {
	UserID uuid.UUID
}

type GetWeekStatsOutput struct {
	Calories         int
	TotalTimeMinutes int
}

type GetWeekStatsUC struct {
	tracer      trace.Tracer
	sessionRepo ports.SessionRepository
}

func NewGetWeekStatsUC(tracer trace.Tracer, sessionRepo ports.SessionRepository) *GetWeekStatsUC {
	return &GetWeekStatsUC{tracer: tracer, sessionRepo: sessionRepo}
}

func (uc *GetWeekStatsUC) Execute(ctx context.Context, input GetWeekStatsInput) (*GetWeekStatsOutput, error) {
	ctx, span := uc.tracer.Start(ctx, "GetWeekStatsUC")
	defer span.End()

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
