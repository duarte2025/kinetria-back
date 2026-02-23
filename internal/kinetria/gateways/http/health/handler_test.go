package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/http/health"
)

func TestHealthHandler_ReturnsOK(t *testing.T) {
	cfg := config.Config{AppName: "test-service"}
	handler := health.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	cfg := config.Config{AppName: "test-service"}
	handler := health.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestHealthHandler_ResponseBody(t *testing.T) {
	cfg := config.Config{AppName: "test-service"}
	handler := health.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp health.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp.Status)
	}
	if resp.Service != "test-service" {
		t.Errorf("expected service 'test-service', got '%s'", resp.Service)
	}
	if resp.Version == "" {
		t.Error("expected version to be non-empty")
	}
}
