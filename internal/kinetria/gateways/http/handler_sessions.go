package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	domainsessions "github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

// contextKey is the type for context keys to avoid collisions.
type contextKey string

// userIDKey is the context key for storing the authenticated user ID.
const userIDKey contextKey = "userID"

// SessionsHandler handles HTTP requests for session endpoints.
type SessionsHandler struct {
	startSessionUC   *domainsessions.StartSessionUC
	recordSetUC      *domainsessions.RecordSetUseCase
	finishSessionUC  *domainsessions.FinishSessionUseCase
	abandonSessionUC *domainsessions.AbandonSessionUseCase
}

// NewSessionsHandler creates a new SessionsHandler with the required use cases.
func NewSessionsHandler(
	startSessionUC *domainsessions.StartSessionUC,
	recordSetUC *domainsessions.RecordSetUseCase,
	finishSessionUC *domainsessions.FinishSessionUseCase,
	abandonSessionUC *domainsessions.AbandonSessionUseCase,
) *SessionsHandler {
	return &SessionsHandler{
		startSessionUC:   startSessionUC,
		recordSetUC:      recordSetUC,
		finishSessionUC:  finishSessionUC,
		abandonSessionUC: abandonSessionUC,
	}
}

// StartSession godoc
// @Summary Start a workout session
// @Description Start a new workout session for a specific workout
// @Tags sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body StartSessionRequest true "Session details"
// @Success 201 {object} SuccessResponse{data=StartSessionResponse}
// @Failure 400 {object} ErrorResponse "Active session already exists"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/sessions [post]
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

// RecordSet godoc
// @Summary Record a set
// @Description Record a completed or skipped set for an exercise
// @Tags sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session ID"
// @Param request body RecordSetRequest true "Set details"
// @Success 201 {object} SuccessResponse{data=RecordSetResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Session not found"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/sessions/{sessionId}/sets [post]
func (h *SessionsHandler) RecordSet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid sessionId format.")
		return
	}

	var req struct {
		ExerciseID string `json:"exerciseId"`
		SetNumber  int    `json:"setNumber"`
		Weight     int    `json:"weight"` // grams
		Reps       int    `json:"reps"`
		Status     string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	exerciseID, err := uuid.Parse(req.ExerciseID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid exerciseId format.")
		return
	}

	output, err := h.recordSetUC.Execute(r.Context(), domainsessions.RecordSetInput{
		UserID:     userID,
		SessionID:  sessionID,
		ExerciseID: exerciseID,
		SetNumber:  req.SetNumber,
		Weight:     req.Weight,
		Reps:       req.Reps,
		Status:     vos.SetRecordStatus(req.Status),
	})
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrMalformedParameters):
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid parameters provided.")
		case errors.Is(err, domainerrors.ErrNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found.")
		case errors.Is(err, domainerrors.ErrSessionNotActive):
			writeError(w, http.StatusConflict, "SESSION_NOT_ACTIVE", "Session is not active.")
		case errors.Is(err, domainerrors.ErrExerciseNotFound):
			writeError(w, http.StatusNotFound, "EXERCISE_NOT_FOUND", "Exercise not found.")
		case errors.Is(err, domainerrors.ErrSetAlreadyRecorded):
			writeError(w, http.StatusConflict, "SET_ALREADY_RECORDED", "Set already recorded.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":         output.SetRecord.ID.String(),
		"sessionId":  output.SetRecord.SessionID.String(),
		"exerciseId": output.SetRecord.ExerciseID.String(),
		"setNumber":  output.SetRecord.SetNumber,
		"weight":     output.SetRecord.Weight,
		"reps":       output.SetRecord.Reps,
		"status":     output.SetRecord.Status,
		"recordedAt": output.SetRecord.RecordedAt,
	})
}

// FinishSession godoc
// @Summary Finish a workout session
// @Description Mark a workout session as completed
// @Tags sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session ID"
// @Param request body FinishSessionRequest true "Finish notes"
// @Success 200 {object} SuccessResponse{data=SessionStatusResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Session not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/sessions/{sessionId}/finish [patch]
func (h *SessionsHandler) FinishSession(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid sessionId format.")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
			return
		}
	}

	output, err := h.finishSessionUC.Execute(r.Context(), domainsessions.FinishSessionInput{
		UserID:    userID,
		SessionID: sessionID,
		Notes:     req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found.")
		case errors.Is(err, domainerrors.ErrSessionAlreadyClosed):
			writeError(w, http.StatusConflict, "SESSION_ALREADY_CLOSED", "Session is already closed.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"id":         output.Session.ID.String(),
		"workoutId":  output.Session.WorkoutID.String(),
		"startedAt":  output.Session.StartedAt,
		"finishedAt": output.Session.FinishedAt,
		"status":     string(output.Session.Status),
		"notes":      output.Session.Notes,
	})
}

// AbandonSession godoc
// @Summary Abandon a workout session
// @Description Mark a workout session as abandoned
// @Tags sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session ID"
// @Param request body AbandonSessionRequest true "Abandon notes"
// @Success 200 {object} SuccessResponse{data=SessionStatusResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Session not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/sessions/{sessionId}/abandon [patch]
func (h *SessionsHandler) AbandonSession(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid sessionId format.")
		return
	}

	output, err := h.abandonSessionUC.Execute(r.Context(), domainsessions.AbandonSessionInput{
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrNotFound):
			writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found.")
		case errors.Is(err, domainerrors.ErrSessionAlreadyClosed):
			writeError(w, http.StatusConflict, "SESSION_ALREADY_CLOSED", "Session is already closed.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"id":         output.Session.ID.String(),
		"workoutId":  output.Session.WorkoutID.String(),
		"startedAt":  output.Session.StartedAt,
		"finishedAt": output.Session.FinishedAt,
		"status":     string(output.Session.Status),
	})
}
