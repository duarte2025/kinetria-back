package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"go.opentelemetry.io/otel/trace"
)

type GetTodayWorkoutInput struct {
	UserID uuid.UUID
}

type GetTodayWorkoutOutput struct {
	Workout *entities.Workout // null se usuário não tem workouts
}

type GetTodayWorkoutUC struct {
	tracer      trace.Tracer
	workoutRepo ports.WorkoutRepository
}

func NewGetTodayWorkoutUC(tracer trace.Tracer, workoutRepo ports.WorkoutRepository) *GetTodayWorkoutUC {
	return &GetTodayWorkoutUC{tracer: tracer, workoutRepo: workoutRepo}
}

func (uc *GetTodayWorkoutUC) Execute(ctx context.Context, input GetTodayWorkoutInput) (*GetTodayWorkoutOutput, error) {
	ctx, span := uc.tracer.Start(ctx, "GetTodayWorkoutUC")
	defer span.End()

	workout, err := uc.workoutRepo.GetFirstByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetTodayWorkoutOutput{Workout: workout}, nil
}
