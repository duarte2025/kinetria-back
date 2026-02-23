package repositories_test

import (
	"context"
	"os"
	"testing"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories"
)

func TestNewDatabasePool_Success(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION_TEST=1 to run")
	}

	cfg := config.Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     5432,
		DBUser:     getEnvOrDefault("DB_USER", "kinetria"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "kinetria_dev_pass"),
		DBName:     getEnvOrDefault("DB_NAME", "kinetria"),
		DBSSLMode:  "disable",
	}

	pool, err := repositories.NewDatabasePool(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		t.Errorf("expected ping to succeed, got %v", err)
	}
}

func TestNewDatabasePool_Failure_WrongPort(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: set INTEGRATION_TEST=1 to run")
	}

	cfg := config.Config{
		DBHost:     "localhost",
		DBPort:     9999,
		DBUser:     "kinetria",
		DBPassword: "kinetria_dev_pass",
		DBName:     "kinetria",
		DBSSLMode:  "disable",
	}

	_, err := repositories.NewDatabasePool(cfg)
	if err == nil {
		t.Fatal("expected error for wrong port, got nil")
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
