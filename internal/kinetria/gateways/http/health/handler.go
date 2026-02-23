package health

import (
	"encoding/json"
	"net/http"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func NewHealthHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Status:  "healthy",
			Service: cfg.AppName,
			Version: "undefined",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}
