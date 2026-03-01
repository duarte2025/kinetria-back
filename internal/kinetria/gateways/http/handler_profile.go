package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/profile"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

// ProfileHandler handles HTTP requests for profile endpoints.
type ProfileHandler struct {
	getProfileUC    *profile.GetProfileUC
	updateProfileUC *profile.UpdateProfileUC
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(
	getProfileUC *profile.GetProfileUC,
	updateProfileUC *profile.UpdateProfileUC,
) *ProfileHandler {
	return &ProfileHandler{
		getProfileUC:    getProfileUC,
		updateProfileUC: updateProfileUC,
	}
}

// userPreferencesDTO is the DTO for user preferences in request/response.
type userPreferencesDTO struct {
	Theme    string `json:"theme"`
	Language string `json:"language"`
}

// profileResponse is the response DTO for profile endpoints.
type profileResponse struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Email           string             `json:"email"`
	ProfileImageURL *string            `json:"profileImageUrl"`
	Preferences     userPreferencesDTO `json:"preferences"`
}

// updateProfileRequest is the request DTO for PATCH /profile.
type updateProfileRequest struct {
	Name            *string             `json:"name"`
	ProfileImageURL *string             `json:"profileImageUrl"`
	Preferences     *userPreferencesDTO `json:"preferences"`
}

// HandleGetProfile godoc
// @Summary Get user profile
// @Description Get the authenticated user's profile
// @Tags profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse{data=profileResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/profile [get]
func (h *ProfileHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	out, err := h.getProfileUC.Execute(ctx, profile.GetProfileInput{UserID: userID})
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found.")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	u := out.User
	var profileImageURL *string
	if u.ProfileImageURL != "" {
		s := u.ProfileImageURL
		profileImageURL = &s
	}

	writeSuccess(w, http.StatusOK, profileResponse{
		ID:              u.ID.String(),
		Name:            u.Name,
		Email:           u.Email,
		ProfileImageURL: profileImageURL,
		Preferences: userPreferencesDTO{
			Theme:    string(u.Preferences.Theme),
			Language: string(u.Preferences.Language),
		},
	})
}

// HandleUpdateProfile godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile (partial update)
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body updateProfileRequest true "Profile update fields"
// @Success 200 {object} SuccessResponse{data=profileResponse}
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/profile [patch]
func (h *ProfileHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	// Map request DTO to use case input
	input := profile.UpdateProfileInput{
		Name:            req.Name,
		ProfileImageURL: req.ProfileImageURL,
	}
	if req.Preferences != nil {
		prefs := vos.UserPreferences{
			Theme:    vos.Theme(req.Preferences.Theme),
			Language: vos.Language(req.Preferences.Language),
		}
		input.Preferences = &prefs
	}

	u, err := h.updateProfileUC.Execute(ctx, userID, input)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found.")
			return
		}
		if errors.Is(err, domainerrors.ErrMalformedParameters) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	var profileImageURL *string
	if u.ProfileImageURL != "" {
		s := u.ProfileImageURL
		profileImageURL = &s
	}

	writeSuccess(w, http.StatusOK, profileResponse{
		ID:              u.ID.String(),
		Name:            u.Name,
		Email:           u.Email,
		ProfileImageURL: profileImageURL,
		Preferences: userPreferencesDTO{
			Theme:    string(u.Preferences.Theme),
			Language: string(u.Preferences.Language),
		},
	})
}
