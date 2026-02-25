.PHONY: help
help: ## Mostra esta mensagem de ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Executa a aplicação
	go run cmd/kinetria/api/main.go

.PHONY: build
build: ## Compila a aplicação
	go build -o bin/kinetria cmd/kinetria/api/main.go

.PHONY: test
test: ## Executa os testes
	go test -v -race ./...

.PHONY: test-coverage
test-coverage: ## Executa os testes com cobertura
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Executa o linter
	golangci-lint run

.PHONY: sqlc
sqlc: ## Gera código a partir das queries SQL
	sqlc generate

.PHONY: mocks
mocks: ## Gera mocks das interfaces
	go generate ./...

.PHONY: tidy
tidy: ## Organiza as dependências
	go mod tidy

.PHONY: deps
deps: ## Instala as dependências
	go mod download

.PHONY: swagger
swagger: ## Gera documentação Swagger/OpenAPI
	swag init -g cmd/kinetria/api/main.go -o docs --parseDependency --parseInternal
