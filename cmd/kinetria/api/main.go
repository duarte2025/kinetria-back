package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
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

			// Database
			repositories.NewDatabasePool,
			repositories.NewSQLDB,

			// JWT
			func(cfg config.Config) *gatewayauth.JWTManager {
				return gatewayauth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
			},

			// Repositories (with interface binding)
			fx.Annotate(
				repositories.NewUserRepository,
				fx.As(new(ports.UserRepository)),
			),
			fx.Annotate(
				repositories.NewRefreshTokenRepository,
				fx.As(new(ports.RefreshTokenRepository)),
			),

			// Use cases
			func(userRepo ports.UserRepository, refreshTokenRepo ports.RefreshTokenRepository, jwtMgr *gatewayauth.JWTManager, cfg config.Config) *domainauth.RegisterUC {
				return domainauth.NewRegisterUC(userRepo, refreshTokenRepo, jwtMgr, cfg.RefreshTokenExpiry)
			},
			func(userRepo ports.UserRepository, refreshTokenRepo ports.RefreshTokenRepository, jwtMgr *gatewayauth.JWTManager, cfg config.Config) *domainauth.LoginUC {
				return domainauth.NewLoginUC(userRepo, refreshTokenRepo, jwtMgr, cfg.RefreshTokenExpiry)
			},
			func(refreshTokenRepo ports.RefreshTokenRepository, jwtMgr *gatewayauth.JWTManager, cfg config.Config) *domainauth.RefreshTokenUC {
				return domainauth.NewRefreshTokenUC(refreshTokenRepo, jwtMgr, cfg.RefreshTokenExpiry)
			},
			domainauth.NewLogoutUC,

			// Validator and HTTP
			validator.New,
			healthhandler.NewHealthHandler,
			httpgateway.NewAuthHandler,
			httpgateway.NewServiceRouter,
			chi.NewRouter,
		),
		fx.Invoke(func(router chi.Router, serviceRouter httpgateway.ServiceRouter) {
			router.Route(serviceRouter.Pattern(), serviceRouter.Router)
		}),
		fx.Invoke(httpgateway.StartHTTPServer),
	).Run()
}
