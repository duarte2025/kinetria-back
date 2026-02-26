package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"go.opentelemetry.io/otel/trace"
)

type GetUserProfileInput struct {
	UserID uuid.UUID
}

type GetUserProfileOutput struct {
	ID              uuid.UUID
	Name            string
	Email           string
	ProfileImageURL string
}

type GetUserProfileUC struct {
	tracer   trace.Tracer
	userRepo ports.UserRepository
}

func NewGetUserProfileUC(tracer trace.Tracer, userRepo ports.UserRepository) *GetUserProfileUC {
	return &GetUserProfileUC{tracer: tracer, userRepo: userRepo}
}

func (uc *GetUserProfileUC) Execute(ctx context.Context, input GetUserProfileInput) (*GetUserProfileOutput, error) {
	ctx, span := uc.tracer.Start(ctx, "GetUserProfileUC")
	defer span.End()

	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetUserProfileOutput{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		ProfileImageURL: user.ProfileImageURL,
	}, nil
}
