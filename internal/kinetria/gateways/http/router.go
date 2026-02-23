package service

import (
	"github.com/go-chi/chi/v5"
)

// ServiceRouter mounts all API routes for the kinetria service.
type ServiceRouter struct {
	authHandler *AuthHandler
}

// NewServiceRouter creates a new ServiceRouter with the provided handlers.
func NewServiceRouter(authHandler *AuthHandler) ServiceRouter {
	return ServiceRouter{authHandler: authHandler}
}

// Pattern returns the base path prefix for all routes.
func (s ServiceRouter) Pattern() string {
	return "/api/v1"
}

// Router registers all routes onto the provided chi.Router.
func (s ServiceRouter) Router(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.authHandler.Register)
		r.Post("/login", s.authHandler.Login)
		r.Post("/refresh", s.authHandler.RefreshToken)
		r.Post("/logout", s.authHandler.Logout)
	})
}
