package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	domainworkouts "github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// DTOs
type WorkoutSummaryDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	Intensity   *string `json:"intensity"`
	Duration    int     `json:"duration"`
	ImageURL    *string `json:"imageUrl"`
}

type ExerciseDTO struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	ThumbnailURL *string  `json:"thumbnailUrl"`
	Sets         int      `json:"sets"`
	Reps         string   `json:"reps"`
	Muscles      []string `json:"muscles"`
	RestTime     int      `json:"restTime"`
	Weight       *int     `json:"weight"`
}

type WorkoutDTO struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Type        *string       `json:"type"`
	Intensity   *string       `json:"intensity"`
	Duration    int           `json:"duration"`
	ImageURL    *string       `json:"imageUrl"`
	Exercises   []ExerciseDTO `json:"exercises"`
}

type PaginationMetaDTO struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type ApiResponseDTO struct {
	Data interface{}        `json:"data"`
	Meta *PaginationMetaDTO `json:"meta,omitempty"`
}

func mapWorkoutToSummaryDTO(w entities.Workout) WorkoutSummaryDTO {
	dto := WorkoutSummaryDTO{
		ID:       w.ID.String(),
		Name:     w.Name,
		Duration: w.Duration,
	}
	if w.Description != "" {
		dto.Description = &w.Description
	}
	if w.Type != "" {
		dto.Type = &w.Type
	}
	if w.Intensity != "" {
		dto.Intensity = &w.Intensity
	}
	if w.ImageURL != "" {
		dto.ImageURL = &w.ImageURL
	}
	return dto
}

func mapExerciseToDTO(e entities.Exercise) ExerciseDTO {
	dto := ExerciseDTO{
		ID:       e.ID.String(),
		Name:     e.Name,
		Sets:     e.Sets,
		Reps:     e.Reps,
		Muscles:  e.Muscles,
		RestTime: e.RestTime,
	}

	if e.ThumbnailURL != "" {
		dto.ThumbnailURL = &e.ThumbnailURL
	}
	if e.Weight > 0 {
		dto.Weight = &e.Weight
	}

	return dto
}

func mapWorkoutToFullDTO(w entities.Workout, exercises []entities.Exercise) WorkoutDTO {
	dto := WorkoutDTO{
		ID:        w.ID.String(),
		Name:      w.Name,
		Duration:  w.Duration,
		Exercises: make([]ExerciseDTO, len(exercises)),
	}

	// Mapear campos opcionais do workout
	if w.Description != "" {
		dto.Description = &w.Description
	}
	if w.Type != "" {
		dto.Type = &w.Type
	}
	if w.Intensity != "" {
		dto.Intensity = &w.Intensity
	}
	if w.ImageURL != "" {
		dto.ImageURL = &w.ImageURL
	}

	// Mapear exercises
	for i, exercise := range exercises {
		dto.Exercises[i] = mapExerciseToDTO(exercise)
	}

	return dto
}

// Request DTOs

type WorkoutExerciseRequest struct {
	ExerciseID string `json:"exerciseId"`
	Sets       int    `json:"sets"`
	Reps       string `json:"reps"`
	RestTime   int    `json:"restTime"`
	Weight     *int   `json:"weight"`
	OrderIndex int    `json:"orderIndex"`
}

type CreateWorkoutRequest struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description"`
	Type        string                   `json:"type"`
	Intensity   string                   `json:"intensity"`
	Duration    int                      `json:"duration"`
	ImageURL    *string                  `json:"imageUrl"`
	Exercises   []WorkoutExerciseRequest `json:"exercises"`
}

type UpdateWorkoutRequest struct {
	Name        *string                  `json:"name"`
	Description *string                  `json:"description"`
	Type        *string                  `json:"type"`
	Intensity   *string                  `json:"intensity"`
	Duration    *int                     `json:"duration"`
	ImageURL    *string                  `json:"imageUrl"`
	Exercises   []WorkoutExerciseRequest `json:"exercises"`
}

// WorkoutsHandler handles HTTP requests for workouts endpoints.
type WorkoutsHandler struct {
	listWorkoutsUC  *domainworkouts.ListWorkoutsUC
	getWorkoutUC    *domainworkouts.GetWorkoutUC
	createWorkoutUC *domainworkouts.CreateWorkoutUC
	updateWorkoutUC *domainworkouts.UpdateWorkoutUC
	deleteWorkoutUC *domainworkouts.DeleteWorkoutUC
	jwtManager      *gatewayauth.JWTManager
}

// NewWorkoutsHandler creates a new WorkoutsHandler.
func NewWorkoutsHandler(
	listWorkoutsUC *domainworkouts.ListWorkoutsUC,
	getWorkoutUC *domainworkouts.GetWorkoutUC,
	createWorkoutUC *domainworkouts.CreateWorkoutUC,
	updateWorkoutUC *domainworkouts.UpdateWorkoutUC,
	deleteWorkoutUC *domainworkouts.DeleteWorkoutUC,
	jwtManager *gatewayauth.JWTManager,
) *WorkoutsHandler {
	return &WorkoutsHandler{
		listWorkoutsUC:  listWorkoutsUC,
		getWorkoutUC:    getWorkoutUC,
		createWorkoutUC: createWorkoutUC,
		updateWorkoutUC: updateWorkoutUC,
		deleteWorkoutUC: deleteWorkoutUC,
		jwtManager:      jwtManager,
	}
}

func mapExerciseRequestToInput(req WorkoutExerciseRequest) (domainworkouts.WorkoutExerciseInput, error) {
	exerciseID, err := uuid.Parse(req.ExerciseID)
	if err != nil {
		return domainworkouts.WorkoutExerciseInput{}, fmt.Errorf("invalid exerciseId '%s': must be a valid UUID", req.ExerciseID)
	}
	return domainworkouts.WorkoutExerciseInput{
		ExerciseID: exerciseID,
		Sets:       req.Sets,
		Reps:       req.Reps,
		RestTime:   req.RestTime,
		Weight:     req.Weight,
		OrderIndex: req.OrderIndex,
	}, nil
}

func mapDomainErrorToHTTP(err error) (int, string, string) {
	switch {
	case errors.Is(err, domerrors.ErrWorkoutNotFound):
		return http.StatusNotFound, "WORKOUT_NOT_FOUND", "Workout not found."
	case errors.Is(err, domerrors.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action."
	case errors.Is(err, domerrors.ErrCannotModifyTemplate):
		return http.StatusForbidden, "CANNOT_MODIFY_TEMPLATE", "Cannot modify or delete template workouts."
	case errors.Is(err, domerrors.ErrWorkoutHasActiveSessions):
		return http.StatusConflict, "WORKOUT_HAS_ACTIVE_SESSIONS", "Cannot delete workout with active sessions."
	case errors.Is(err, domerrors.ErrMalformedParameters):
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred."
	}
}

// ListWorkouts godoc
// @Summary List user workouts
// @Description Get paginated list of user's workouts
// @Tags workouts
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} SuccessResponse{data=WorkoutListResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/workouts [get]
func (h *WorkoutsHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract userID from JWT
	userID, err := h.extractUserIDFromJWT(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	// Parse query params
	page, err := parseIntQueryParam(r, "page", 1)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "page must be a valid integer")
		return
	}
	pageSize, err := parseIntQueryParam(r, "pageSize", 20)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "pageSize must be a valid integer")
		return
	}

	output, err := h.listWorkoutsUC.Execute(ctx, domainworkouts.ListWorkoutsInput{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	dtos := make([]WorkoutSummaryDTO, len(output.Workouts))
	for i, wo := range output.Workouts {
		dtos[i] = mapWorkoutToSummaryDTO(wo)
	}

	resp := ApiResponseDTO{
		Data: dtos,
		Meta: &PaginationMetaDTO{
			Page:       output.Page,
			PageSize:   output.PageSize,
			Total:      output.Total,
			TotalPages: output.TotalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetWorkout godoc
// @Summary Get workout by ID
// @Description Get detailed workout information with exercises
// @Tags workouts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workout ID (UUID)"
// @Success 200 {object} SuccessResponse{data=WorkoutDTO}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Workout not found"
// @Failure 422 {object} ErrorResponse "Invalid workout ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/workouts/{id} [get]
func (h *WorkoutsHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extrair userID do JWT
	userID, err := h.extractUserIDFromJWT(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	// 2. Extrair workoutID do path parameter
	workoutIDStr := chi.URLParam(r, "id")
	if workoutIDStr == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "workout id is required")
		return
	}

	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "workoutId must be a valid UUID")
		return
	}

	// 3. Chamar use case
	output, err := h.getWorkoutUC.Execute(ctx, domainworkouts.GetWorkoutInput{
		WorkoutID: workoutID,
		UserID:    userID,
	})
	if err != nil {
		// Workout not found (ou ownership fail)
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "WORKOUT_NOT_FOUND",
				fmt.Sprintf("Workout with id '%s' was not found.", workoutID.String()))
			return
		}
		// Erro interno
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
		return
	}

	// 4. Mapear para DTO
	dto := mapWorkoutToFullDTO(output.Workout, output.Exercises)

	// 5. Responder com sucesso
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ApiResponseDTO{
		Data: dto,
	})
}

// CreateWorkout godoc
// @Summary Create a new workout
// @Description Creates a new workout with exercises for the authenticated user
// @Tags workouts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateWorkoutRequest true "Workout data"
// @Success 201 {object} ApiResponseDTO{data=WorkoutSummaryDTO}
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/workouts [post]
func (h *WorkoutsHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.extractUserIDFromJWT(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body.")
		return
	}

	exercises := make([]domainworkouts.WorkoutExerciseInput, len(req.Exercises))
	for i, ex := range req.Exercises {
		exInput, err := mapExerciseRequestToInput(ex)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		exercises[i] = exInput
	}

	workout, err := h.createWorkoutUC.Execute(ctx, userID, domainworkouts.CreateWorkoutInput{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Intensity:   req.Intensity,
		Duration:    req.Duration,
		ImageURL:    req.ImageURL,
		Exercises:   exercises,
	})
	if err != nil {
		statusCode, errCode, msg := mapDomainErrorToHTTP(err)
		writeError(w, statusCode, errCode, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ApiResponseDTO{Data: mapWorkoutToSummaryDTO(*workout)})
}

// UpdateWorkout godoc
// @Summary Update a workout
// @Description Updates a workout owned by the authenticated user
// @Tags workouts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workout ID (UUID)"
// @Param body body UpdateWorkoutRequest true "Workout data"
// @Success 200 {object} ApiResponseDTO{data=WorkoutSummaryDTO}
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Workout not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/workouts/{id} [put]
func (h *WorkoutsHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.extractUserIDFromJWT(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	workoutIDStr := chi.URLParam(r, "id")
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "workoutId must be a valid UUID")
		return
	}

	var req UpdateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body.")
		return
	}

	var exercises []domainworkouts.WorkoutExerciseInput
	if len(req.Exercises) > 0 {
		exercises = make([]domainworkouts.WorkoutExerciseInput, len(req.Exercises))
		for i, ex := range req.Exercises {
			exInput, err := mapExerciseRequestToInput(ex)
			if err != nil {
				writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
				return
			}
			exercises[i] = exInput
		}
	}

	workout, err := h.updateWorkoutUC.Execute(ctx, userID, workoutID, domainworkouts.UpdateWorkoutInput{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Intensity:   req.Intensity,
		Duration:    req.Duration,
		ImageURL:    req.ImageURL,
		Exercises:   exercises,
	})
	if err != nil {
		statusCode, errCode, msg := mapDomainErrorToHTTP(err)
		writeError(w, statusCode, errCode, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ApiResponseDTO{Data: mapWorkoutToSummaryDTO(*workout)})
}

// DeleteWorkout godoc
// @Summary Delete a workout
// @Description Soft-deletes a workout owned by the authenticated user
// @Tags workouts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workout ID (UUID)"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Workout not found"
// @Failure 409 {object} ErrorResponse "Workout has active sessions"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/workouts/{id} [delete]
func (h *WorkoutsHandler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.extractUserIDFromJWT(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	workoutIDStr := chi.URLParam(r, "id")
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "workoutId must be a valid UUID")
		return
	}

	if err := h.deleteWorkoutUC.Execute(ctx, userID, workoutID); err != nil {
		statusCode, errCode, msg := mapDomainErrorToHTTP(err)
		writeError(w, statusCode, errCode, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutsHandler) extractUserIDFromJWT(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, fmt.Errorf("missing authorization header")
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	return h.jwtManager.ParseToken(tokenStr)
}

func parseIntQueryParam(r *http.Request, key string, defaultValue int) (int, error) {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s", key)
	}
	return value, nil
}
