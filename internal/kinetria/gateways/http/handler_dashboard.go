package service

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
)

type DashboardHandler struct {
	getUserProfileUC  *dashboard.GetUserProfileUC
	getTodayWorkoutUC *dashboard.GetTodayWorkoutUC
	getWeekProgressUC *dashboard.GetWeekProgressUC
	getWeekStatsUC    *dashboard.GetWeekStatsUC
}

func NewDashboardHandler(
	getUserProfileUC *dashboard.GetUserProfileUC,
	getTodayWorkoutUC *dashboard.GetTodayWorkoutUC,
	getWeekProgressUC *dashboard.GetWeekProgressUC,
	getWeekStatsUC *dashboard.GetWeekStatsUC,
) *DashboardHandler {
	return &DashboardHandler{
		getUserProfileUC:  getUserProfileUC,
		getTodayWorkoutUC: getTodayWorkoutUC,
		getWeekProgressUC: getWeekProgressUC,
		getWeekStatsUC:    getWeekStatsUC,
	}
}

// GetDashboard handles GET /api/v1/dashboard
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extrair userID do JWT (middleware injeta no context)
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	// Estrutura para coletar resultados das goroutines
	type result struct {
		user         *dashboard.GetUserProfileOutput
		todayWorkout *dashboard.GetTodayWorkoutOutput
		weekProgress *dashboard.GetWeekProgressOutput
		weekStats    *dashboard.GetWeekStatsOutput
		err          error
		source       string // para debug
	}

	ch := make(chan result, 4)

	// Executar use cases em paralelo
	go func() {
		out, err := h.getUserProfileUC.Execute(ctx, dashboard.GetUserProfileInput{UserID: userID})
		ch <- result{user: out, err: err, source: "user"}
	}()

	go func() {
		out, err := h.getTodayWorkoutUC.Execute(ctx, dashboard.GetTodayWorkoutInput{UserID: userID})
		ch <- result{todayWorkout: out, err: err, source: "todayWorkout"}
	}()

	go func() {
		out, err := h.getWeekProgressUC.Execute(ctx, dashboard.GetWeekProgressInput{UserID: userID})
		ch <- result{weekProgress: out, err: err, source: "weekProgress"}
	}()

	go func() {
		out, err := h.getWeekStatsUC.Execute(ctx, dashboard.GetWeekStatsInput{UserID: userID})
		ch <- result{weekStats: out, err: err, source: "weekStats"}
	}()

	// Coletar resultados
	var res result
	for i := 0; i < 4; i++ {
		r := <-ch
		if r.err != nil {
			// Fail-fast: se qualquer use case falhar, retornar erro
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load dashboard data")
			return
		}

		// Merge resultados
		if r.user != nil {
			res.user = r.user
		}
		if r.todayWorkout != nil {
			res.todayWorkout = r.todayWorkout
		}
		if r.weekProgress != nil {
			res.weekProgress = r.weekProgress
		}
		if r.weekStats != nil {
			res.weekStats = r.weekStats
		}
	}

	// Montar DTO de resposta
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":              res.user.ID.String(),
			"name":            res.user.Name,
			"email":           res.user.Email,
			"profileImageUrl": res.user.ProfileImageURL,
		},
		"todayWorkout": nil, // default null
		"weekProgress": mapWeekProgressToDTO(res.weekProgress.Days),
		"stats": map[string]interface{}{
			"calories":         res.weekStats.Calories,
			"totalTimeMinutes": res.weekStats.TotalTimeMinutes,
		},
	}

	// TodayWorkout pode ser null
	if res.todayWorkout.Workout != nil {
		w := res.todayWorkout.Workout
		response["todayWorkout"] = map[string]interface{}{
			"id":          w.ID.String(),
			"name":        w.Name,
			"description": w.Description,
			"type":        w.Type,
			"intensity":   w.Intensity,
			"duration":    w.Duration,
			"imageUrl":    w.ImageURL,
		}
	}

	writeSuccess(w, http.StatusOK, response)
}

// Helper: mapear weekProgress para DTO
func mapWeekProgressToDTO(days []dashboard.DayProgress) []map[string]interface{} {
	result := make([]map[string]interface{}, len(days))
	for i, d := range days {
		result[i] = map[string]interface{}{
			"day":    d.Day,
			"date":   d.Date,
			"status": d.Status,
		}
	}
	return result
}
