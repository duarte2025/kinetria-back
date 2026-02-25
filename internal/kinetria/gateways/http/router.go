package service

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// ServiceRouter mounts all API routes for the kinetria service.
type ServiceRouter struct {
	authHandler      *AuthHandler
	sessionsHandler  *SessionsHandler
	workoutsHandler  *WorkoutsHandler
	dashboardHandler *DashboardHandler
	jwtManager       *gatewayauth.JWTManager
}

// NewServiceRouter creates a new ServiceRouter with the provided handlers.
func NewServiceRouter(
	authHandler *AuthHandler,
	sessionsHandler *SessionsHandler,
	workoutsHandler *WorkoutsHandler,
	dashboardHandler *DashboardHandler,
	jwtManager *gatewayauth.JWTManager,
) ServiceRouter {
	return ServiceRouter{
		authHandler:      authHandler,
		sessionsHandler:  sessionsHandler,
		workoutsHandler:  workoutsHandler,
		dashboardHandler: dashboardHandler,
		jwtManager:       jwtManager,
	}
}

// Pattern returns the base path prefix for all routes.
func (s ServiceRouter) Pattern() string {
	return "/api/v1"
}

// Router registers all routes onto the provided chi.Router.
func (s ServiceRouter) Router(router chi.Router) {
	// Swagger documentation
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.authHandler.Register)
		r.Post("/login", s.authHandler.Login)
		r.Post("/refresh", s.authHandler.RefreshToken)
		r.Post("/logout", s.authHandler.Logout)
	})

	// Protected routes
	router.With(AuthMiddleware(s.jwtManager)).Post("/sessions", s.sessionsHandler.StartSession)
	router.With(AuthMiddleware(s.jwtManager)).Post("/sessions/{sessionId}/sets", s.sessionsHandler.RecordSet)
	router.With(AuthMiddleware(s.jwtManager)).Patch("/sessions/{sessionId}/finish", s.sessionsHandler.FinishSession)
	router.With(AuthMiddleware(s.jwtManager)).Patch("/sessions/{sessionId}/abandon", s.sessionsHandler.AbandonSession)

	// Workouts (authenticated)
	router.With(AuthMiddleware(s.jwtManager)).Get("/workouts", s.workoutsHandler.ListWorkouts)

	// Dashboard (authenticated)
	router.With(AuthMiddleware(s.jwtManager)).Get("/dashboard", s.dashboardHandler.GetDashboard)
}
