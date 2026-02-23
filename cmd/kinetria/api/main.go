package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	healthhandler "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http/health"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories"
)

var (
	AppName     = "kinetria"
	BuildCommit = "undefined"
	BuildTag    = "undefined"
	BuildTime   = "undefined"
)

func main() {
	fx.New(
		fx.Provide(
			config.ParseConfigFromEnv,
			repositories.NewDatabasePool,
			healthhandler.NewHealthHandler,
			chi.NewRouter,
		),
		fx.Invoke(startHTTPServer),
	).Run()
}

func startHTTPServer(lc fx.Lifecycle, cfg config.Config, router chi.Router, healthHandler http.HandlerFunc) {
	router.Get("/health", healthHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
