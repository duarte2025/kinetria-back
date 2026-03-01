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

// UpdateProfileInput holds the optional fields for a profile update.
// A nil pointer means "not provided" — the corresponding field will not be changed.
//
// Validation rules (enforced by [UpdateProfileUC.Execute]):
//   - Name: 2–100 characters after whitespace trimming.
//   - Preferences: [vos.UserPreferences.Validate] must pass (theme/language must be
//     one of their allowed values).
type UpdateProfileInput struct {
	// Name, when non-nil, replaces the user's display name.
	Name *string
	// ProfileImageURL, when non-nil, replaces the user's profile image URL.
	ProfileImageURL *string
	// Preferences, when non-nil, replaces the user's preferences entirely.
	Preferences *vos.UserPreferences
}

// UpdateProfileUC implements the use case for updating a user's profile.
// It performs a partial update: only fields explicitly set (non-nil) in
// [UpdateProfileInput] are written to the repository.
type UpdateProfileUC struct {
	tracer   trace.Tracer
	userRepo ports.UserRepository
}

// NewUpdateProfileUC creates a new [UpdateProfileUC] wired with the given tracer and repository.
func NewUpdateProfileUC(tracer trace.Tracer, userRepo ports.UserRepository) *UpdateProfileUC {
	return &UpdateProfileUC{tracer: tracer, userRepo: userRepo}
}

// Execute updates the profile of the user identified by userID.
//
// Only non-nil fields in input are applied. At least one field must be provided;
// otherwise the call returns [domainerrors.ErrMalformedParameters].
//
// Possible errors:
//   - [domainerrors.ErrMalformedParameters] — no fields provided, name too short/long,
//     or preferences contain invalid theme/language values.
//   - [domainerrors.ErrNotFound] — no user with the given userID exists.
//   - Any repository error on infrastructure failures.
//
// On success the returned [entities.User] reflects all applied changes.
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
