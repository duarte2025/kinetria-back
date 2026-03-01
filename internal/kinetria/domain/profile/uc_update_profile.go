package profile

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
	"go.opentelemetry.io/otel/trace"
)

// UpdateProfileInput holds the optional fields for profile update.
// A nil pointer means "not provided" (field should not be changed).
type UpdateProfileInput struct {
	Name            *string
	ProfileImageURL *string
	Preferences     *vos.UserPreferences
}

// UpdateProfileUC implements the use case for updating a user's profile.
type UpdateProfileUC struct {
	tracer   trace.Tracer
	userRepo ports.UserRepository
}

// NewUpdateProfileUC creates a new UpdateProfileUC.
func NewUpdateProfileUC(tracer trace.Tracer, userRepo ports.UserRepository) *UpdateProfileUC {
	return &UpdateProfileUC{tracer: tracer, userRepo: userRepo}
}

// Execute updates the user profile with the provided fields.
// Only non-nil fields are updated. Returns the updated user.
func (uc *UpdateProfileUC) Execute(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*entities.User, error) {
	ctx, span := uc.tracer.Start(ctx, "UpdateProfileUC")
	defer span.End()

	// Validate that at least one field is provided
	if input.Name == nil && input.ProfileImageURL == nil && input.Preferences == nil {
		return nil, fmt.Errorf("%w: at least one field must be provided", domainerrors.ErrMalformedParameters)
	}

	// Validate name if provided
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if len(name) < 2 {
			return nil, fmt.Errorf("%w: name must be at least 2 characters", domainerrors.ErrMalformedParameters)
		}
		if len(name) > 100 {
			return nil, fmt.Errorf("%w: name must be at most 100 characters", domainerrors.ErrMalformedParameters)
		}
		*input.Name = name
	}

	// Validate preferences if provided
	if input.Preferences != nil {
		if err := input.Preferences.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", domainerrors.ErrMalformedParameters, err.Error())
		}
	}

	// Fetch current user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates (partial update pattern)
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.ProfileImageURL != nil {
		user.ProfileImageURL = *input.ProfileImageURL
	}
	if input.Preferences != nil {
		user.Preferences = *input.Preferences
	}

	// Persist
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
