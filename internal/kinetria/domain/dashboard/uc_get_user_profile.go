package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
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
	userRepo ports.UserRepository
}

func NewGetUserProfileUC(userRepo ports.UserRepository) *GetUserProfileUC {
	return &GetUserProfileUC{userRepo: userRepo}
}

func (uc *GetUserProfileUC) Execute(ctx context.Context, input GetUserProfileInput) (*GetUserProfileOutput, error) {
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
