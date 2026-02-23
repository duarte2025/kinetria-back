package service

import (
	"github.com/go-chi/chi/v5"
)

type ServiceRouter struct {
	// handler Handler
}

func NewServiceRouter( /* handler Handler */ ) ServiceRouter {
	return ServiceRouter{
		// handler: handler,
	}
}

func (s ServiceRouter) Pattern() string {
	return "/api/v1"
}

func (s ServiceRouter) Router(router chi.Router) {
	// Adicione suas rotas aqui
	// router.Post("/example", rest.Handle(s.handler.Create))
}
