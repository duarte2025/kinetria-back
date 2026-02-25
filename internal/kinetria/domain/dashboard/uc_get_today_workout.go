package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetTodayWorkoutInput struct {
	UserID uuid.UUID
}

type GetTodayWorkoutOutput struct {
	Workout *entities.Workout // null se usuário não tem workouts
}

type GetTodayWorkoutUC struct {
	workoutRepo ports.WorkoutRepository
}

func NewGetTodayWorkoutUC(workoutRepo ports.WorkoutRepository) *GetTodayWorkoutUC {
	return &GetTodayWorkoutUC{workoutRepo: workoutRepo}
}

func (uc *GetTodayWorkoutUC) Execute(ctx context.Context, input GetTodayWorkoutInput) (*GetTodayWorkoutOutput, error) {
	workout, err := uc.workoutRepo.GetFirstByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetTodayWorkoutOutput{Workout: workout}, nil
}
