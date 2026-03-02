package tests

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	domaindashboard "github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
	domainexercises "github.com/kinetria/kinetria-back/internal/kinetria/domain/exercises"
	domainprofile "github.com/kinetria/kinetria-back/internal/kinetria/domain/profile"
	domainsessions "github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	domainstatistics "github.com/kinetria/kinetria-back/internal/kinetria/domain/statistics"
	domainworkouts "github.com/kinetria/kinetria-back/internal/kinetria/domain/workouts"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	service "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel"
)

type TestServer struct {
	Router     chi.Router
	DB         *sql.DB
	JWTManager *gatewayauth.JWTManager
	Container  *postgres.PostgresContainer
	BaseURL    string
	HTTPServer *httptest.Server
}

func SetupTestServer(t *testing.T) *TestServer {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("kinetria_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	migrator := repositories.NewMigrator(db)
	err = migrator.Run(ctx)
	require.NoError(t, err)

	cfg := config.Config{
		JWTSecret: "test-secret-key",
		JWTExpiry: 15 * time.Minute,
	}

	jwtManager := gatewayauth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
	validate := validator.New()
	tracer := otel.Tracer("kinetria-test")

	userRepo := repositories.NewUserRepository(db)
	workoutRepo := repositories.NewWorkoutRepository(db)
	exerciseRepo := repositories.NewExerciseRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	setRecordRepo := repositories.NewSetRecordRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)

	registerUC := domainauth.NewRegisterUC(userRepo, refreshTokenRepo, jwtManager, cfg.JWTExpiry, 7*24*time.Hour)
	loginUC := domainauth.NewLoginUC(userRepo, refreshTokenRepo, jwtManager, cfg.JWTExpiry, 7*24*time.Hour)
	refreshTokenUC := domainauth.NewRefreshTokenUC(refreshTokenRepo, jwtManager, cfg.JWTExpiry, 7*24*time.Hour)
	logoutUC := domainauth.NewLogoutUC(refreshTokenRepo)

	startSessionUC := domainsessions.NewStartSessionUC(sessionRepo, workoutRepo, auditLogRepo)
	recordSetUC := domainsessions.NewRecordSetUseCase(sessionRepo, setRecordRepo, exerciseRepo, auditLogRepo)
	finishSessionUC := domainsessions.NewFinishSessionUseCase(sessionRepo, auditLogRepo)
	abandonSessionUC := domainsessions.NewAbandonSessionUseCase(sessionRepo, auditLogRepo)

	listWorkoutsUC := domainworkouts.NewListWorkoutsUC(workoutRepo)
	getWorkoutUC := domainworkouts.NewGetWorkoutUC(workoutRepo)
	createWorkoutUC := domainworkouts.NewCreateWorkoutUC(workoutRepo, exerciseRepo)
	updateWorkoutUC := domainworkouts.NewUpdateWorkoutUC(workoutRepo, exerciseRepo)
	deleteWorkoutUC := domainworkouts.NewDeleteWorkoutUC(workoutRepo)

	getUserProfileUC := domaindashboard.NewGetUserProfileUC(tracer, userRepo)
	getTodayWorkoutUC := domaindashboard.NewGetTodayWorkoutUC(tracer, workoutRepo)
	getWeekProgressUC := domaindashboard.NewGetWeekProgressUC(tracer, sessionRepo)
	getWeekStatsUC := domaindashboard.NewGetWeekStatsUC(tracer, sessionRepo)

	getProfileUC := domainprofile.NewGetProfileUC(tracer, userRepo)
	updateProfileUC := domainprofile.NewUpdateProfileUC(tracer, userRepo)

	listExercisesUC := domainexercises.NewListExercisesUC(exerciseRepo)
	getExerciseUC := domainexercises.NewGetExerciseUC(exerciseRepo)
	getExerciseHistoryUC := domainexercises.NewGetExerciseHistoryUC(exerciseRepo)

	getOverviewUC := domainstatistics.NewGetOverviewUC(sessionRepo, setRecordRepo)
	getProgressionUC := domainstatistics.NewGetProgressionUC(setRecordRepo)
	getPersonalRecordsUC := domainstatistics.NewGetPersonalRecordsUC(setRecordRepo)
	getFrequencyUC := domainstatistics.NewGetFrequencyUC(sessionRepo)

	authHandler := service.NewAuthHandler(registerUC, loginUC, refreshTokenUC, logoutUC, jwtManager, validate)
	sessionsHandler := service.NewSessionsHandler(startSessionUC, recordSetUC, finishSessionUC, abandonSessionUC)
	workoutsHandler := service.NewWorkoutsHandler(listWorkoutsUC, getWorkoutUC, createWorkoutUC, updateWorkoutUC, deleteWorkoutUC, jwtManager)
	dashboardHandler := service.NewDashboardHandler(getUserProfileUC, getTodayWorkoutUC, getWeekProgressUC, getWeekStatsUC)
	profileHandler := service.NewProfileHandler(getProfileUC, updateProfileUC)
	exercisesHandler := service.NewExercisesHandler(listExercisesUC, getExerciseUC, getExerciseHistoryUC, jwtManager)
	statisticsHandler := service.NewStatisticsHandler(getOverviewUC, getProgressionUC, getPersonalRecordsUC, getFrequencyUC)

	router := chi.NewRouter()
	serviceRouter := service.NewServiceRouter(authHandler, sessionsHandler, workoutsHandler, dashboardHandler, profileHandler, exercisesHandler, statisticsHandler, jwtManager)
	router.Route(serviceRouter.Pattern(), serviceRouter.Router)

	httpServer := httptest.NewServer(router)

	return &TestServer{
		Router:     router,
		DB:         db,
		JWTManager: jwtManager,
		Container:  pgContainer,
		BaseURL:    httpServer.URL,
		HTTPServer: httpServer,
	}
}

func (ts *TestServer) Cleanup(t *testing.T) {
	ctx := context.Background()

	if ts.HTTPServer != nil {
		ts.HTTPServer.Close()
	}

	if ts.DB != nil {
		ts.DB.Close()
	}

	if ts.Container != nil {
		err := ts.Container.Terminate(ctx)
		require.NoError(t, err)
	}
}

func (ts *TestServer) URL(path string) string {
	return fmt.Sprintf("%s/api/v1%s", ts.BaseURL, path)
}

func (ts *TestServer) AuthHeader(token string) http.Header {
	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	headers.Set("Content-Type", "application/json")
	return headers
}
