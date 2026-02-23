package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	httphandler "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http"
)

var (
	AppName     = "kinetria"
	BuildCommit = "undefined"
	BuildTag    = "undefined"
	BuildTime   = "undefined"
)

func newLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func newDBPool(lc fx.Lifecycle, cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

func newRouter(handler httphandler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	sr := httphandler.NewServiceRouter(handler)
	r.Route(sr.Pattern(), sr.Router)

	return r
}

func startHTTPServer(lc fx.Lifecycle, cfg config.Config, router *chi.Mux, logger *zap.Logger) {
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}
			logger.Info("HTTP server started", zap.String("addr", server.Addr))
			go server.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

func main() {
	fx.New(
		fx.Provide(
			config.ParseConfigFromEnv,
			newLogger,
			newDBPool,
			validator.New,
			httphandler.NewHandler,
			newRouter,
		),
		fx.Invoke(startHTTPServer),
	).Run()
}
