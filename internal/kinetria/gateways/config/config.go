package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName     string `envconfig:"APP_NAME" required:"true"`
	Environment string `envconfig:"ENVIRONMENT" required:"true"`

	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`

	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	HTTPPort string `envconfig:"HTTP_PORT" default:"8080"`
}

func ParseConfigFromEnv() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}
