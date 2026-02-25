package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
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

// WorkoutsHandler handles HTTP requests for workouts endpoints.
type WorkoutsHandler struct {
	listWorkoutsUC *domainworkouts.ListWorkoutsUC
	getWorkoutUC   *domainworkouts.GetWorkoutUC
	jwtManager     *gatewayauth.JWTManager
}

// NewWorkoutsHandler creates a new WorkoutsHandler.
func NewWorkoutsHandler(
	listWorkoutsUC *domainworkouts.ListWorkoutsUC,
	getWorkoutUC *domainworkouts.GetWorkoutUC,
	jwtManager *gatewayauth.JWTManager,
) *WorkoutsHandler {
	return &WorkoutsHandler{
		listWorkoutsUC: listWorkoutsUC,
		getWorkoutUC:   getWorkoutUC,
		jwtManager:     jwtManager,
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
