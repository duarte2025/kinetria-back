package health

import (
	"encoding/json"
	"net/http"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
)

type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"kinetria"`
	Version string `json:"version" example:"undefined"`
}

// HealthCheck godoc
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func NewHealthHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Status:  "healthy",
			Service: cfg.AppName,
			Version: "undefined",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}
