package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/statistics"
)

// StatisticsHandler handles HTTP requests for statistics endpoints.
type StatisticsHandler struct {
	getOverviewUC        *statistics.GetOverviewUC
	getProgressionUC     *statistics.GetProgressionUC
	getPersonalRecordsUC *statistics.GetPersonalRecordsUC
	getFrequencyUC       *statistics.GetFrequencyUC
}

// NewStatisticsHandler creates a new StatisticsHandler.
func NewStatisticsHandler(
	getOverviewUC *statistics.GetOverviewUC,
	getProgressionUC *statistics.GetProgressionUC,
	getPersonalRecordsUC *statistics.GetPersonalRecordsUC,
	getFrequencyUC *statistics.GetFrequencyUC,
) *StatisticsHandler {
	return &StatisticsHandler{
		getOverviewUC:        getOverviewUC,
		getProgressionUC:     getProgressionUC,
		getPersonalRecordsUC: getPersonalRecordsUC,
		getFrequencyUC:       getFrequencyUC,
	}
}

// HandleGetOverview godoc
// @Summary Get statistics overview
// @Description Get aggregated workout statistics for the authenticated user
// @Tags statistics
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "Start date (RFC3339 or YYYY-MM-DD)"
// @Param endDate query string false "End date (RFC3339 or YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse "Overview statistics"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/stats/overview [get]
func (h *StatisticsHandler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	input := statistics.GetOverviewInput{UserID: userID}

	if s := r.URL.Query().Get("startDate"); s != "" {
		t, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid startDate format. Use YYYY-MM-DD or RFC3339.")
			return
		}
		input.StartDate = &t
	}
	if s := r.URL.Query().Get("endDate"); s != "" {
		t, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid endDate format. Use YYYY-MM-DD or RFC3339.")
			return
		}
		input.EndDate = &t
	}

	out, err := h.getOverviewUC.Execute(ctx, input)
	if err != nil {
		if isStatValidationError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve overview statistics.")
		return
	}

	writeSuccess(w, http.StatusOK, mapOverviewToResponse(out))
}

// HandleGetProgression godoc
// @Summary Get workout progression
// @Description Get daily progression data (max weight, total volume) for the authenticated user
// @Tags statistics
// @Produce json
// @Security BearerAuth
// @Param exerciseId query string false "Filter by exercise UUID"
// @Param startDate query string false "Start date (RFC3339 or YYYY-MM-DD)"
// @Param endDate query string false "End date (RFC3339 or YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse "Progression data"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/stats/progression [get]
func (h *StatisticsHandler) HandleGetProgression(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	input := statistics.GetProgressionInput{UserID: userID}

	if s := r.URL.Query().Get("exerciseId"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid exerciseId format.")
			return
		}
		input.ExerciseID = &id
	}

	if s := r.URL.Query().Get("startDate"); s != "" {
		t, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid startDate format. Use YYYY-MM-DD or RFC3339.")
			return
		}
		input.StartDate = &t
	}
	if s := r.URL.Query().Get("endDate"); s != "" {
		t, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid endDate format. Use YYYY-MM-DD or RFC3339.")
			return
		}
		input.EndDate = &t
	}

	out, err := h.getProgressionUC.Execute(ctx, input)
	if err != nil {
		if isStatValidationError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve progression data.")
		return
	}

	writeSuccess(w, http.StatusOK, mapProgressionToResponse(out))
}

// HandleGetPersonalRecords godoc
// @Summary Get personal records
// @Description Get personal records (best performance per muscle group) for the authenticated user
// @Tags statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse "Personal records"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/stats/personal-records [get]
func (h *StatisticsHandler) HandleGetPersonalRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	records, err := h.getPersonalRecordsUC.Execute(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve personal records.")
		return
	}

	writeSuccess(w, http.StatusOK, mapPersonalRecordsToResponse(records))
}

// HandleGetFrequency godoc
// @Summary Get workout frequency
// @Description Get daily workout frequency (heatmap data) for the authenticated user
// @Tags statistics
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "Start date (RFC3339 or YYYY-MM-DD)"
// @Param endDate query string false "End date (RFC3339 or YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse "Frequency data"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/stats/frequency [get]
func (h *StatisticsHandler) HandleGetFrequency(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	input := statistics.GetFrequencyInput{UserID: userID}

	if s := r.URL.Query().Get("startDate"); s != "" {
		t, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid startDate format. Use YYYY-MM-DD or RFC3339.")
			return
		}
		input.StartDate = &t
	}
	if s := r.URL.Query().Get("endDate"); s != "" {
		t, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid endDate format. Use YYYY-MM-DD or RFC3339.")
			return
		}
		input.EndDate = &t
	}

	out, err := h.getFrequencyUC.Execute(ctx, input)
	if err != nil {
		if isStatValidationError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve frequency data.")
		return
	}

	writeSuccess(w, http.StatusOK, mapFrequencyToResponse(out))
}

// --- Helpers ---

// parseDate parses a date string in YYYY-MM-DD or RFC3339 format.
func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	return time.Parse(time.RFC3339, s)
}

// isStatValidationError reports whether the error is a validation/business error.
func isStatValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "must be") ||
		strings.Contains(msg, "must not") ||
		strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "exceed")
}

// --- Response mappers ---

type overviewResponse struct {
	StartDate        string  `json:"startDate"`
	EndDate          string  `json:"endDate"`
	TotalWorkouts    int     `json:"totalWorkouts"`
	AveragePerWeek   float64 `json:"averagePerWeek"`
	TotalTimeMinutes int     `json:"totalTimeMinutes"`
	TotalSets        int     `json:"totalSets"`
	TotalReps        int     `json:"totalReps"`
	TotalVolume      int64   `json:"totalVolume"`
	CurrentStreak    int     `json:"currentStreak"`
	LongestStreak    int     `json:"longestStreak"`
}

func mapOverviewToResponse(out *statistics.OverviewStats) overviewResponse {
	return overviewResponse{
		StartDate:        out.StartDate.Format("2006-01-02"),
		EndDate:          out.EndDate.Format("2006-01-02"),
		TotalWorkouts:    out.TotalWorkouts,
		AveragePerWeek:   out.AveragePerWeek,
		TotalTimeMinutes: out.TotalTimeMinutes,
		TotalSets:        out.TotalSets,
		TotalReps:        out.TotalReps,
		TotalVolume:      out.TotalVolume,
		CurrentStreak:    out.CurrentStreak,
		LongestStreak:    out.LongestStreak,
	}
}

type progressionPointResponse struct {
	Date        string  `json:"date"`
	MaxWeight   int64   `json:"maxWeight"`
	TotalVolume int64   `json:"totalVolume"`
	Change      float64 `json:"change"`
}

type progressionResponse struct {
	ExerciseID string                     `json:"exerciseId,omitempty"`
	Points     []progressionPointResponse `json:"points"`
}

func mapProgressionToResponse(out *statistics.ProgressionData) progressionResponse {
	points := make([]progressionPointResponse, 0, len(out.Points))
	for _, p := range out.Points {
		points = append(points, progressionPointResponse{
			Date:        p.Date.Format("2006-01-02"),
			MaxWeight:   p.MaxWeight,
			TotalVolume: p.TotalVolume,
			Change:      p.Change,
		})
	}
	resp := progressionResponse{Points: points}
	if out.ExerciseID != nil {
		resp.ExerciseID = out.ExerciseID.String()
	}
	return resp
}

type personalRecordResponse struct {
	ExerciseID   string `json:"exerciseId"`
	ExerciseName string `json:"exerciseName"`
	Weight       int    `json:"weight"`
	Reps         int    `json:"reps"`
	Volume       int64  `json:"volume"`
	AchievedAt   string `json:"achievedAt"`
}

type personalRecordsResponse struct {
	Records []personalRecordResponse `json:"records"`
}

func mapPersonalRecordsToResponse(records []statistics.PersonalRecord) personalRecordsResponse {
	dtos := make([]personalRecordResponse, 0, len(records))
	for _, pr := range records {
		dtos = append(dtos, personalRecordResponse{
			ExerciseID:   pr.ExerciseID.String(),
			ExerciseName: pr.ExerciseName,
			Weight:       pr.Weight,
			Reps:         pr.Reps,
			Volume:       pr.Volume,
			AchievedAt:   pr.AchievedAt.Format("2006-01-02"),
		})
	}
	return personalRecordsResponse{Records: dtos}
}

type frequencyDataResponse struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type frequencyResponse struct {
	Data []frequencyDataResponse `json:"data"`
}

func mapFrequencyToResponse(data []statistics.FrequencyData) frequencyResponse {
	dtos := make([]frequencyDataResponse, 0, len(data))
	for _, d := range data {
		dtos = append(dtos, frequencyDataResponse{
			Date:  d.Date.Format("2006-01-02"),
			Count: d.Count,
		})
	}
	return frequencyResponse{Data: dtos}
}
