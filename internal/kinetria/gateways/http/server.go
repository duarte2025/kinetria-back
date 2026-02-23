package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
)

func StartHTTPServer(lc fx.Lifecycle, cfg config.Config, router chi.Router, healthHandler http.HandlerFunc) {
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
