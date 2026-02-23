package service

import (
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	validate *validator.Validate
}

func NewHandler(validate *validator.Validate) Handler {
	return Handler{
		validate: validate,
	}
}

// Adicione seus handlers HTTP aqui
