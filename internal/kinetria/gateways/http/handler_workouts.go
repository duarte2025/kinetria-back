package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	domainworkouts "github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
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

// WorkoutsHandler handles HTTP requests for workouts endpoints.
type WorkoutsHandler struct {
	listWorkoutsUC *domainworkouts.ListWorkoutsUC
	jwtManager     *gatewayauth.JWTManager
}

// NewWorkoutsHandler creates a new WorkoutsHandler.
func NewWorkoutsHandler(listWorkoutsUC *domainworkouts.ListWorkoutsUC, jwtManager *gatewayauth.JWTManager) *WorkoutsHandler {
	return &WorkoutsHandler{
		listWorkoutsUC: listWorkoutsUC,
		jwtManager:     jwtManager,
	}
}

// ListWorkouts handles GET /api/v1/workouts
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
