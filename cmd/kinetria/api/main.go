package main

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	httpgateway "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http"
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
		fx.Invoke(httpgateway.StartHTTPServer),
	).Run()
}
