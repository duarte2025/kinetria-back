package service

import (
	"github.com/go-chi/chi/v5"
)

type ServiceRouter struct {
	handler Handler
}

func NewServiceRouter(handler Handler) ServiceRouter {
	return ServiceRouter{
		handler: handler,
	}
}

func (s ServiceRouter) Pattern() string {
	return "/"
}

func (s ServiceRouter) Router(router chi.Router) {
	router.Get("/health", s.handler.Health)
}
