package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	_ "github.com/kinetria/kinetria-back/docs"
	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	domaindashboard "github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	domainprofile "github.com/kinetria/kinetria-back/internal/kinetria/domain/profile"
	domainsessions "github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	domainworkouts "github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	httpgateway "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http"
	healthhandler "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http/health"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories"
)

// @title Kinetria API
// @version 1.0
// @description API para gerenciamento de treinos e acompanhamento de progresso
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@kinetria.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

			// Tracer (uses global OTel tracer provider; defaults to noop if not configured)
			func() trace.Tracer {
				return otel.Tracer("kinetria")
			},

			// Database
			repositories.NewDatabasePool,
			repositories.NewSQLDB,
			repositories.NewMigrator,

			// JWT - Provide JWTManager as both concrete type and interface
			func(cfg config.Config) *gatewayauth.JWTManager {
				return gatewayauth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
			},
			// Bind JWTManager to TokenManager interface for use cases
			fx.Annotate(
				func(jwtMgr *gatewayauth.JWTManager) ports.TokenManager {
					return jwtMgr
				},
				fx.As(new(ports.TokenManager)),
			),

			// Repositories (with interface binding)
			fx.Annotate(
				repositories.NewUserRepository,
				fx.As(new(ports.UserRepository)),
			),
			fx.Annotate(
				repositories.NewRefreshTokenRepository,
				fx.As(new(ports.RefreshTokenRepository)),
			),
			fx.Annotate(
				repositories.NewSessionRepository,
				fx.As(new(ports.SessionRepository)),
			),
			fx.Annotate(
				repositories.NewSetRecordRepository,
				fx.As(new(ports.SetRecordRepository)),
			),
			fx.Annotate(
				repositories.NewExerciseRepository,
				fx.As(new(ports.ExerciseRepository)),
			),
			fx.Annotate(
				repositories.NewWorkoutRepository,
				fx.As(new(ports.WorkoutRepository)),
			),
			fx.Annotate(
				repositories.NewAuditLogRepository,
				fx.As(new(ports.AuditLogRepository)),
			),

			// Use cases
			func(userRepo ports.UserRepository, refreshTokenRepo ports.RefreshTokenRepository, tokenMgr ports.TokenManager, cfg config.Config) *domainauth.RegisterUC {
				return domainauth.NewRegisterUC(userRepo, refreshTokenRepo, tokenMgr, cfg.JWTExpiry, cfg.RefreshTokenExpiry)
			},
			func(userRepo ports.UserRepository, refreshTokenRepo ports.RefreshTokenRepository, tokenMgr ports.TokenManager, cfg config.Config) *domainauth.LoginUC {
				return domainauth.NewLoginUC(userRepo, refreshTokenRepo, tokenMgr, cfg.JWTExpiry, cfg.RefreshTokenExpiry)
			},
			func(refreshTokenRepo ports.RefreshTokenRepository, tokenMgr ports.TokenManager, cfg config.Config) *domainauth.RefreshTokenUC {
				return domainauth.NewRefreshTokenUC(refreshTokenRepo, tokenMgr, cfg.JWTExpiry, cfg.RefreshTokenExpiry)
			},
			domainauth.NewLogoutUC,
			domainsessions.NewStartSessionUC,
			domainsessions.NewRecordSetUseCase,
			domainsessions.NewFinishSessionUseCase,
			domainsessions.NewAbandonSessionUseCase,
			domainworkouts.NewListWorkoutsUC,
			domainworkouts.NewGetWorkoutUC,
			domaindashboard.NewGetUserProfileUC,
			domaindashboard.NewGetTodayWorkoutUC,
			domaindashboard.NewGetWeekProgressUC,
			domaindashboard.NewGetWeekStatsUC,
			domainprofile.NewGetProfileUC,
			domainprofile.NewUpdateProfileUC,

			// Validator and HTTP
			validator.New,
			healthhandler.NewHealthHandler,
			httpgateway.NewAuthHandler,
			httpgateway.NewSessionsHandler,
			httpgateway.NewWorkoutsHandler,
			httpgateway.NewDashboardHandler,
			httpgateway.NewProfileHandler,
			httpgateway.NewServiceRouter,
			chi.NewRouter,
		),
		fx.Invoke(func(router *chi.Mux, serviceRouter httpgateway.ServiceRouter) {
			router.Route(serviceRouter.Pattern(), serviceRouter.Router)
		}),
		fx.Invoke(repositories.RunMigrations),
		fx.Invoke(httpgateway.StartHTTPServer),
	).Run()
}
