package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName     string `envconfig:"APP_NAME" required:"true"`
	Environment string `envconfig:"ENVIRONMENT" required:"true"`

	// Database
	DBHost     string `envconfig:"DB_HOST" required:"true"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBName     string `envconfig:"DB_NAME" required:"true"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"require"`

	// HTTP Server
	HTTPPort int `envconfig:"HTTP_PORT" default:"8080"`

	// JWT
	JWTSecret          string        `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiry          time.Duration `envconfig:"JWT_EXPIRY" default:"1h"`
	RefreshTokenExpiry time.Duration `envconfig:"REFRESH_TOKEN_EXPIRY" default:"720h"`
}

func ParseConfigFromEnv() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}
