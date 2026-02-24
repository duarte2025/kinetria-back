package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	domainsessions "github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

// contextKey is the type for context keys to avoid collisions.
type contextKey string

// userIDKey is the context key for storing the authenticated user ID.
const userIDKey contextKey = "userID"

// SessionsHandler handles HTTP requests for session endpoints.
type SessionsHandler struct {
	startSessionUC *domainsessions.StartSessionUC
}

// NewSessionsHandler creates a new SessionsHandler with the required use case.
func NewSessionsHandler(startSessionUC *domainsessions.StartSessionUC) *SessionsHandler {
	return &SessionsHandler{
		startSessionUC: startSessionUC,
	}
}

// StartSession handles POST /sessions
func (h *SessionsHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	// Extract userID from context (injected by AuthMiddleware)
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	var req struct {
		WorkoutID string `json:"workoutId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	workoutID, err := uuid.Parse(req.WorkoutID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid workoutId format.")
		return
	}

	output, err := h.startSessionUC.Execute(r.Context(), domainsessions.StartSessionInput{
		UserID:    userID,
		WorkoutID: workoutID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrMalformedParameters):
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid parameters provided.")
		case errors.Is(err, domainerrors.ErrWorkoutNotFound):
			writeError(w, http.StatusNotFound, "WORKOUT_NOT_FOUND", "Workout not found.")
		case errors.Is(err, domainerrors.ErrActiveSessionExists):
			writeError(w, http.StatusConflict, "ACTIVE_SESSION_EXISTS", "User already has an active session.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":        output.Session.ID.String(),
		"workoutId": output.Session.WorkoutID.String(),
		"startedAt": output.Session.StartedAt,
		"status":    string(output.Session.Status),
	})
}
