package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetWeekProgressInput struct {
	UserID uuid.UUID
}

type DayProgress struct {
	Day    string // "S", "T", "Q", "Q", "S", "S", "D"
	Date   string // "2026-02-17" (formato ISO)
	Status string // "completed", "missed", "future"
}

type GetWeekProgressOutput struct {
	Days []DayProgress // sempre 7 itens
}

type GetWeekProgressUC struct {
	sessionRepo ports.SessionRepository
}

func NewGetWeekProgressUC(sessionRepo ports.SessionRepository) *GetWeekProgressUC {
	return &GetWeekProgressUC{sessionRepo: sessionRepo}
}

func (uc *GetWeekProgressUC) Execute(ctx context.Context, input GetWeekProgressInput) (*GetWeekProgressOutput, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDate := today.AddDate(0, 0, -6) // 6 dias atrás
	endDate := today

	// Buscar sessões completed na semana
	sessions, err := uc.sessionRepo.GetCompletedSessionsByUserAndDateRange(ctx, input.UserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Mapear datas de sessões completed
	completedDates := make(map[string]bool)
	for _, s := range sessions {
		dateStr := s.StartedAt.Format("2006-01-02")
		completedDates[dateStr] = true
	}

	// Gerar array de 7 dias
	days := make([]DayProgress, 7)
	dayLabels := []string{"D", "S", "T", "Q", "Q", "S", "S"} // domingo=0, segunda=1, ...

	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		weekday := int(date.Weekday()) // 0=Sunday, 1=Monday, ...

		status := "missed"
		if date.After(today) {
			status = "future"
		} else if completedDates[dateStr] {
			status = "completed"
		}

		days[i] = DayProgress{
			Day:    dayLabels[weekday],
			Date:   dateStr,
			Status: status,
		}
	}

	return &GetWeekProgressOutput{Days: days}, nil
}
