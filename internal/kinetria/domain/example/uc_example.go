package example

import (
	"context"
)

// Este é um exemplo de use case seguindo o padrão xuc.UseCase[TInput, TOutput]
// Remova este arquivo quando criar seus próprios use cases

type ExampleInput struct {
	Name string
}

type ExampleOutput struct {
	ID      string
	Message string
}

type ExampleUC struct {
	// Injete suas dependências aqui
	// tracer     trace.Tracer
	// repository ports.ExampleRepository
}

func NewExampleUC( /* dependências */ ) ExampleUC {
	return ExampleUC{
		// Inicialize as dependências
	}
}

func (uc ExampleUC) Execute(ctx context.Context, input ExampleInput) (ExampleOutput, error) {
	// ctx, span := uc.tracer.Start(ctx, "ExampleUC")
	// defer span.End()

	// 1. Validações de domínio
	// 2. Orquestração de operações
	// 3. Publicação de eventos
	// 4. Retorno do resultado

	return ExampleOutput{
		ID:      "example-id",
		Message: "Hello, " + input.Name,
	}, nil
}
