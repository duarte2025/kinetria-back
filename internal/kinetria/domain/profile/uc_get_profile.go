package profile

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"go.opentelemetry.io/otel/trace"
)

// GetProfileInput holds the input for the [GetProfileUC] use case.
type GetProfileInput struct {
	// UserID is the UUID of the user whose profile should be retrieved.
	UserID uuid.UUID
}

// GetProfileOutput holds the output of the [GetProfileUC] use case.
type GetProfileOutput struct {
	// User is the retrieved user entity. It is never nil on a successful execution.
	User *entities.User
}

// GetProfileUC implements the use case for retrieving a user's profile.
// It fetches the user by ID from the repository and returns the full entity.
type GetProfileUC struct {
	tracer   trace.Tracer
	userRepo ports.UserRepository
}

// NewGetProfileUC creates a new [GetProfileUC] wired with the given tracer and repository.
func NewGetProfileUC(tracer trace.Tracer, userRepo ports.UserRepository) *GetProfileUC {
	return &GetProfileUC{tracer: tracer, userRepo: userRepo}
}

// Execute retrieves the user profile identified by [GetProfileInput.UserID].
//
// It returns [domainerrors.ErrNotFound] (via the repository) when no user with the
// given ID exists, or any other repository error on infrastructure failures.
func (uc *GetProfileUC) Execute(ctx context.Context, input GetProfileInput) (*GetProfileOutput, error) {
	ctx, span := uc.tracer.Start(ctx, "GetProfileUC")
	defer span.End()

	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetProfileOutput{User: user}, nil
}
