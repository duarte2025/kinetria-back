# Go Hexagonal Architecture - Code Style & Patterns

Este documento define as convenções de código Go para projetos que seguem arquitetura hexagonal (Ports and Adapters) com injeção de dependência via `go.uber.org/fx`.

## Estrutura de Diretórios

```
project-root/
├── cmd/{service}/           # Pontos de entrada (main.go)
│   ├── api/                 # Servidor HTTP principal
│   ├── read-api/            # API somente leitura (opcional)
│   └── worker/              # Workers assíncronos (opcional)
├── internal/{service}/      # Código fonte do serviço
│   ├── domain/              # Lógica de negócio pura
│   │   ├── {feature}/       # Casos de uso (pastas por feature)
│   │   ├── entities/        # Entidades de domínio
│   │   ├── vos/             # Value Objects
│   │   ├── ports/           # Interfaces (contratos)
│   │   ├── validators/      # Validadores de domínio
│   │   ├── services/        # Serviços de domínio
│   │   └── errors/          # Erros de domínio
│   ├── gateways/            # Adaptadores externos
│   │   ├── http/            # Handlers HTTP
│   │   ├── events/          # Handlers e Publishers de eventos
│   │   ├── repositories/    # Repositórios (banco de dados)
│   │   ├── config/          # Configurações do serviço
│   │   └── {client-name}/   # Clientes HTTP externos
│   ├── extensions/          # Utilitários internos do serviço
│   ├── telemetry/           # Observabilidade customizada
│   └── tests/               # Testes de integração
└── pkg/                     # Pacotes compartilhados
```

## 1. Main (cmd/{service}/api/main.go)

### Estrutura Padrão

```go
package main

import (
    "go.uber.org/fx"
    // imports dos módulos
)

var (
    AppName     = "{service-name}"
    BuildCommit = "undefined"
    BuildTag    = "undefined"
    BuildTime   = "undefined"
)

func main() {
    fx.New(
        // Módulos base (ordem importa)
        xfx.BaseModule(),
        xbuild.Module(AppName, BuildCommit, BuildTime, BuildTag),
        xlog.Module(),
        xtelemetry.Module(),
        xhttp.Module(),
        xhealth.Module(),
        
        fx.Provide(
            // 1. Configuração
            config.ParseConfigFromEnv,
            
            // 2. Clients externos
            fx.Annotate(clientapi.NewClient, fx.As(new(ports.ServiceInterface))),
            
            // 3. Repositórios
            fx.Annotate(repository.NewRepository, fx.As(new(ports.Repository))),
            
            // 4. Producers (Kafka/SNS)
            fx.Annotate(
                publishers.NewProducer,
                fx.As(new(ports.Producer)),
                fx.ParamTags(`name:"xwatermill_kafka"`),
            ),
            
            // 5. Use Cases
            usecase.NewUseCase,
            
            // 6. HTTP Handlers
            validator.New,
            httphandler.NewHandler,
            
            // 7. HTTP Router
            xhttp.AsRouter(httphandler.NewServiceRouter),
            
            // 8. Event Handlers
            eventhandlers.NewHandler,
        ),
        
        // Decorators para registrar handlers de mensagem
        fx.Decorate(eventhandlers.RegisterMessageHandlers),
    ).Run()
}
```

### Regras

- **Ordem de registro**: Config → Clients → Repositories → Producers → Use Cases → Handlers → Routers
- Use `fx.Annotate` com `fx.As(new(ports.Interface))` para bind de interfaces
- Use `fx.ParamTags` para injetar dependências nomeadas
- Use `fx.ResultTags` com `group:` para registrar múltiplas implementações

## 2. Domain Layer

### 2.1 Entities (internal/{service}/domain/entities/)

```go
package entities

import (
    "github.com/google/uuid"
    "github.com/guregu/null"
)

// Type aliases para IDs
type (
    EntityID = uuid.UUID
    OwnerID  = uuid.UUID
)

type Entity struct {
    ID            EntityID
    Status        vos.EntityStatus
    Name          string
    OptionalField null.String
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Regras:**
- Use `type aliases` para IDs: `type EntityID = uuid.UUID`
- Use `guregu/null` para campos opcionais: `null.String`, `null.Int`, `null.Bool`
- Use `guregu/null/zero` para strings que devem ser "" quando nulas: `zero.String`
- Referencie VOs para tipos enumerados
- **NUNCA** inclua lógica de negócio nas entities

### 2.2 Value Objects (internal/{service}/domain/vos/)

```go
package vos

type EntityStatus string

const (
    EntityStatusActive   EntityStatus = "active"
    EntityStatusInactive EntityStatus = "inactive"
    EntityStatusBlocked  EntityStatus = "blocked"
)

func (s EntityStatus) String() string {
    return string(s)
}

func (s EntityStatus) IsValid() bool {
    switch s {
    case EntityStatusActive, EntityStatusInactive, EntityStatusBlocked:
        return true
    }
    return false
}
```

**Regras:**
- Use `type X string` para enumerações
- Defina constantes com prefixo do tipo: `EntityStatusActive`
- Implemente `String() string` sempre
- VOs são imutáveis - sem métodos que alterem estado

### 2.3 Ports (internal/{service}/domain/ports/)

```go
package ports

import (
    "context"
    "{module-path}/internal/{service}/domain/entities"
)

//go:generate moq -stub -pkg mocks -out mocks/repositories.go . EntityRepository

type EntityRepository interface {
    FindByID(ctx context.Context, id entities.EntityID) (entities.Entity, error)
    Insert(ctx context.Context, entity entities.Entity) error
    Update(ctx context.Context, entity entities.Entity) error
}

//go:generate moq -stub -pkg mocks -out mocks/services.go . ExternalService

type ExternalService interface {
    GetData(ctx context.Context, input entities.GetDataInput) (entities.DataResult, error)
}
```

**Regras:**
- Separe em arquivos: `repositories.go`, `services.go`
- Use `//go:generate moq` para gerar mocks
- Todas as operações recebem `context.Context` como primeiro parâmetro
- Retorne erros de domínio, não erros de infraestrutura

### 2.4 Errors (internal/{service}/domain/errors/)

```go
package errors

import "errors"

var (
    ErrNotFound            = errors.New("not found")
    ErrConflict            = errors.New("data conflict")
    ErrMalformedParameters = errors.New("malformed parameters")
    ErrFailedDependency    = errors.New("failed dependency")
)

var (
    ErrEntityNotFound = errors.New("entity not found")
    ErrInvalidStatus  = errors.New("invalid entity status")
    ErrBlockedEntity  = errors.New("entity is blocked")
)
```

### 2.5 Use Cases (internal/{service}/domain/{feature}/)

```go
package feature

import (
    "context"
    "{module-path}/internal/{service}/domain/entities"
    "{module-path}/internal/{service}/domain/ports"
    "{module-path}/pkg/xuc"
    "go.opentelemetry.io/otel/trace"
)

type CreateEntityInput struct {
    Name   string
    Status string
}

type CreateEntityOutput struct {
    ID        entities.EntityID
    CreatedAt time.Time
}

type CreateEntityUC struct {
    tracer     trace.Tracer
    repository ports.EntityRepository
    producer   ports.EntityProducer
}

func NewCreateEntityUC(
    tracer trace.Tracer,
    repository ports.EntityRepository,
    producer ports.EntityProducer,
) xuc.UseCase[CreateEntityInput, CreateEntityOutput] {
    return CreateEntityUC{
        tracer:     tracer,
        repository: repository,
        producer:   producer,
    }
}

func (uc CreateEntityUC) Execute(ctx context.Context, input CreateEntityInput) (CreateEntityOutput, error) {
    ctx, span := uc.tracer.Start(ctx, "CreateEntityUC")
    defer span.End()
    
    // 1. Validações de domínio
    // 2. Orquestração de operações
    // 3. Publicação de eventos
    // 4. Retorno do resultado
    
    return CreateEntityOutput{}, nil
}
```

**Regras:**
- Implemente `xuc.UseCase[TInput, TOutput]` interface
- Use nomes descritivos: `{Verbo}{Entidade}UC`
- Inicie span de trace no início: `ctx, span := uc.tracer.Start(ctx, "UCName")`
- Receba dependências via construtor (injeção)
- Retorne sempre `(Output, error)`
- **NUNCA** acesse infraestrutura diretamente - use ports

### 2.6 Validators (internal/{service}/domain/validators/)

```go
package validators

import (
    "context"
    "{module-path}/internal/{service}/domain/entities"
)

type EntityValidator interface {
    GetID() entities.ValidatorID
    Validate(ctx context.Context, input *entities.ValidateInput) error
}

type StatusValidator struct{}

func NewStatusValidator() StatusValidator {
    return StatusValidator{}
}

func (v StatusValidator) GetID() entities.ValidatorID {
    return entities.ValidatorID("status_validator")
}

func (v StatusValidator) Validate(ctx context.Context, input *entities.ValidateInput) error {
    if input.Status != vos.EntityStatusActive {
        return domerrors.ErrInvalidStatus
    }
    return nil
}
```

**Regras:**
- Cada validador implementa uma regra específica
- Use `GetID()` para identificação do validador
- Retorne erro de domínio em caso de falha
- Registre múltiplos validadores com `fx.ResultTags` no main

## 3. Gateways Layer

### 3.1 Config (internal/{service}/gateways/config/)

```go
package config

import (
    "fmt"
    "time"
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    AppName     string `envconfig:"APP_NAME" required:"true"`
    Environment string `envconfig:"ENVIRONMENT" required:"true"`
    
    ExternalService ExternalServiceConfig
    KafkaConfig     KafkaConfig
    
    RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`
}

type ExternalServiceConfig struct {
    URL     string        `envconfig:"EXTERNAL_SERVICE_URL" required:"true"`
    Timeout time.Duration `envconfig:"EXTERNAL_SERVICE_TIMEOUT" default:"500ms"`
}

func ParseConfigFromEnv() (Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return Config{}, fmt.Errorf("failed to parse config: %w", err)
    }
    return cfg, nil
}
```

### 3.2 HTTP Handlers (internal/{service}/gateways/http/)

#### Router

```go
package service

import (
    "github.com/go-chi/chi/v5"
    "{module-path}/pkg/xhttp/rest"
)

type ServiceRouter struct {
    handler Handler
}

func NewServiceRouter(handler Handler) ServiceRouter {
    return ServiceRouter{handler: handler}
}

func (s ServiceRouter) Pattern() string {
    return "/service/v1/{service}"
}

func (s ServiceRouter) Router(router chi.Router) {
    router.Post("/entities", rest.Handle(s.handler.Create))
    router.Get("/entities/{id}", rest.Handle(s.handler.GetByID))
}
```

#### Handler

```go
package service

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-playground/validator/v10"
    "{module-path}/internal/{service}/domain/{feature}"
    "{module-path}/pkg/xhttp/rest"
    "{module-path}/pkg/xlog"
    "{module-path}/pkg/xuc"
)

type Handler struct {
    validate *validator.Validate
    uc       xuc.UseCase[feature.Input, feature.Output]
}

func NewHandler(
    validate *validator.Validate,
    uc xuc.UseCase[feature.Input, feature.Output],
) Handler {
    return Handler{validate: validate, uc: uc}
}

func (h Handler) Create(r *http.Request) rest.Response {
    ctx := r.Context()
    
    xlog.DebugContext(ctx, "handler called")
    
    var req RequestBody
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return rest.BadRequest(err, rest.NewErrorPayload(srn.BadRequest, "invalid body"))
    }
    
    if err := h.validate.Struct(req); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            return rest.BadRequestValidator(req, validationErrors)
        }
        return rest.BadRequest(err, rest.NewErrorPayload(srn.BadRequest, "invalid body"))
    }
    
    output, err := h.uc.Execute(ctx, mapToInput(req))
    if err != nil {
        return handleError(err)
    }
    
    return rest.Created(mapToResponse(output))
}
```

**Regras:**
- Use `rest.Handle()` para wrap do handler
- Use `validator/v10` para validação de input
- Retorne `rest.Response` (Created, OK, BadRequest, etc.)
- Separe Request/Response DTOs no pacote `scheme`
- Mapeie DTOs para Input/Output do Use Case

### 3.3 HTTP Clients (internal/{service}/gateways/{client-name}/)

```go
package clientname

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    
    "{module-path}/internal/{service}/gateways/config"
    "{module-path}/pkg/xapiclient"
)

type Client struct {
    client xapiclient.APIClient
}

func NewClient(config config.Config, httpclient *http.Client) (Client, error) {
    baseURL, err := url.Parse(config.ExternalService.URL)
    if err != nil {
        return Client{}, fmt.Errorf("failed to parse URL: %w", err)
    }
    
    return Client{
        client: xapiclient.NewAPIClient(httpclient, baseURL),
    }, nil
}

func (c Client) GetData(ctx context.Context, input entities.GetDataInput) (entities.DataResult, error) {
    const route = "api/v1/data"
    
    headers := map[string]string{
        "Content-Type": "application/json",
        "Accept":       "application/json",
    }
    
    body, statusCode, err := c.client.DoRequest(ctx, http.MethodGet, route, nil, headers)
    if err != nil {
        if statusCode == http.StatusNotFound {
            return entities.DataResult{}, domerrors.ErrNotFound
        }
        return entities.DataResult{}, fmt.Errorf("request failed: %w", err)
    }
    
    var response DataResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return entities.DataResult{}, fmt.Errorf("unmarshal failed: %w", err)
    }
    
    return mapToEntity(response), nil
}
```

**Regras:**
- Use `xapiclient.APIClient` como client base
- Receba `config.Config` e `*http.Client` via construtor
- Converta erros HTTP para erros de domínio
- Implemente a interface definida em `ports`

### 3.4 Repositories (internal/{service}/gateways/repositories/)

#### Repository Base

```go
package repository

import (
    "{module-path}/internal/{service}/gateways/repositories/queries"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    db *pgxpool.Pool
    q  *queries.Queries
}

func NewRepository(db *pgxpool.Pool) Repository {
    return Repository{
        db: db,
        q:  queries.New(),
    }
}
```

#### Implementação

```go
package repository

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "{module-path}/internal/{service}/domain/entities"
    domerr "{module-path}/internal/{service}/domain/errors"
)

func (r Repository) FindByID(ctx context.Context, id entities.EntityID) (entities.Entity, error) {
    row, err := r.q.GetEntityByID(ctx, r.db, pgtype.UUID{Bytes: id, Valid: true})
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return entities.Entity{}, domerr.ErrEntityNotFound
        }
        return entities.Entity{}, fmt.Errorf("error finding entity: %w", err)
    }
    
    return mapToEntity(row), nil
}
```

#### SQLC Queries

```sql
-- name: GetEntityByID :one
SELECT id, name, status, created_at, updated_at
FROM entities
WHERE id = @id;

-- name: InsertEntity :exec
INSERT INTO entities (id, name, status, created_at, updated_at)
VALUES (@id, @name, @status, @created_at, @updated_at);
```

**Regras:**
- Use SQLC para geração de código type-safe
- Converta `pgx.ErrNoRows` para erro de domínio
- Use `pgtype.UUID`, `pgtype.Text` para conversão de tipos
- Mapeie entre `queries.Model` e `entities.Entity`
- **IMPORTANTE**: Verifique sempre se as colunas utilizadas nas condições `WHERE` das queries possuem índices:
  ```sql
  CREATE INDEX IF NOT EXISTS idx_table_column ON table_name (column_name);
  CREATE INDEX IF NOT EXISTS idx_table_columns ON table_name (column1, column2);
  ```

### 3.5 Event Publishers (internal/{service}/gateways/events/publishers/)

```go
package publishers

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/ThreeDotsLabs/watermill/message"
    "github.com/google/uuid"
)

const EntityCreatedTopic = "{service}_fct_entity_created_0"

type EntityCreatedProducer struct {
    publisher message.Publisher
}

func NewEntityCreatedProducer(producer message.Publisher) EntityCreatedProducer {
    return EntityCreatedProducer{publisher: producer}
}

func (p EntityCreatedProducer) PublishEntityCreated(ctx context.Context, entity entities.Entity) error {
    msg := entityCreatedMessage{
        ID:        entity.ID.String(),
        Name:      entity.Name,
        Status:    string(entity.Status),
        CreatedAt: entity.CreatedAt.Format(time.RFC3339),
    }
    
    payload, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("could not serialize entity: %w", err)
    }
    
    wmsg := message.NewMessage(uuid.NewString(), payload)
    wmsg.SetContext(ctx)
    
    return p.publisher.Publish(EntityCreatedTopic, wmsg)
}
```

**Regras:**
- Use `message.Publisher` do Watermill
- Nome do tópico: `{service}_fct_{evento}_0`
- Serialize para JSON
- Use `message.NewMessage(uuid, payload)`
- Sempre `wmsg.SetContext(ctx)`

### 3.6 Event Handlers (internal/{service}/gateways/events/handlers/)

```go
package handlers

import (
    "encoding/json"
    
    "github.com/ThreeDotsLabs/watermill/message"
    "{module-path}/pkg/xlog"
    "{module-path}/pkg/xuc"
)

type EntityEventHandler struct {
    uc xuc.UseCase[feature.Input, feature.Output]
}

func NewEntityEventHandler(uc xuc.UseCase[feature.Input, feature.Output]) EntityEventHandler {
    return EntityEventHandler{uc: uc}
}

func (h EntityEventHandler) Handle(msg *message.Message) error {
    ctx := msg.Context()
    
    xlog.DebugContext(ctx, "handling message")
    
    var payload EventPayload
    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        xlog.ErrorContext(ctx, "error unmarshalling", xlog.Error(err))
        return err
    }
    
    if _, err := h.uc.Execute(ctx, mapToInput(payload)); err != nil {
        xlog.ErrorContext(ctx, "error executing use case", xlog.Error(err))
        return err
    }
    
    msg.Ack()
    return nil
}
```

**Regras:**
- Use `message.Message` do Watermill
- Extraia contexto: `ctx := msg.Context()`
- Chame `msg.Ack()` após processamento bem-sucedido
- Retorne erro para retry automático

## 4. Convenções de Nomenclatura

### Arquivos
- Use snake_case: `entity_repository.go`, `card_status.go`
- Sufixo `_test.go` para testes
- Prefixo `uc_` para use cases: `uc_authorize.go`

### Tipos
- Use PascalCase: `AuthorizationRepository`, `CreateEntityUC`
- Sufixo `UC` para use cases
- Sufixo `Handler` para handlers HTTP/Event
- Sufixo `Producer` para publishers
- Sufixo `Client` para API clients

### Variáveis
- Use camelCase: `entityRepository`, `kafkaProducer`
- Prefixo `err` para variáveis de erro: `errNotFound`

### Constantes
- Use PascalCase para exportadas: `EntityStatusActive`
- Use UPPER_SNAKE_CASE em envconfig: `APP_NAME`

## 5. Testes

```go
func TestUseCase_Execute(t *testing.T) {
    // Arrange
    repoMock := &mocks.EntityRepositoryMock{
        FindByIDFunc: func(ctx context.Context, id uuid.UUID) (entities.Entity, error) {
            return entities.Entity{ID: id, Name: "test"}, nil
        },
    }
    
    uc := feature.NewUseCase(repoMock)
    
    // Act
    result, err := uc.Execute(context.Background(), input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

**Regras:**
- Use `testify` (require, assert)
- Use mocks gerados pelo `moq`
- Padrão AAA: Arrange, Act, Assert
- Teste edge cases e erros

## 6. Diretrizes Go

### Versão
- Use a mesma versão do projeto (verificar `go.mod`)

### Go moderno
- Use generics quando simplificar e tipar melhor
- Use `context` para cancelamento/timeouts
- Trate erros com wrapping (`%w`) e erros sentinela
- Entenda impactos de GC/memória e alocações

### Concorrência
- Goroutines com lifecycle claro, sem leaks
- Padrões com canais: worker pool, fan-in/fan-out, pipelines
- `select` com cancelamento e operações não-bloqueantes
- `sync` (mutex, waitgroup) e atomics quando apropriado
- Evite races e respeite o memory model

### Performance
- Otimize apenas com medição (pprof/benchmarks)
- Pooling/caching com parcimônia
- Atenção a DB/network: pooling, timeouts, prepared statements

### Arquitetura
- Prefira composição, interfaces pequenas
- Ports/gateways (hexagonal/clean)
- Integre com padrões do repo: Fx (DI), Chi (HTTP), pgx/sqlc (DB), Watermill (eventos), OpenTelemetry

### Testes
- Table-driven tests como default
- Integração com testcontainers quando já existir padrão
- Benchmarks apenas quando fizer parte do critério de aceite

### Segurança
- Não logar secrets/PII
- Validação/sanitização e prevenção de injeção
- TLS/crypto apenas com libs padrão ou já adotadas

## 7. Git

### Commits
- Formato: `<type>(<scope>): <subject>`
- Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- Scope: área/módulo alterado (ex: `auth`, `api`, `database`)
- Subject: descrição breve (máximo 50 caracteres)

Exemplos:
- `feat(auth): add new endpoint for user authentication`
- `fix(api): correct date handling in response`
- `docs(project): update README with new setup instructions`

### Branches
- Formato: `<type>/<scope>/<subject>`
- Exemplo: `feat/auth/add-user-authentication-endpoint`

### Pull Requests
- Título: `<type>(<scope>): <subject>`
- Descrição: contexto, tipo de mudança, issue relacionada, checklist

## 8. Observabilidade

### Logging

```go
xlog.InfoContext(ctx, "operation completed",
    xlog.String("entity_id", entity.ID.String()),
    xlog.Int("count", count),
)

xlog.ErrorContext(ctx, "operation failed",
    xlog.Error(err),
    xlog.String("entity_id", entity.ID.String()),
)
```

### Tracing

```go
func (uc UseCase) Execute(ctx context.Context, input Input) (Output, error) {
    ctx, span := uc.tracer.Start(ctx, "UseCaseName")
    defer span.End()
    
    span.SetAttributes(attribute.String("entity_id", id.String()))
    
    return output, nil
}
```

## Resumo das Responsabilidades

| Camada | Responsabilidade |
|--------|------------------|
| `cmd/` | Bootstrap da aplicação com Fx |
| `domain/entities` | Estruturas de dados do domínio |
| `domain/vos` | Value Objects e enumerações |
| `domain/ports` | Interfaces/contratos |
| `domain/errors` | Erros de domínio |
| `domain/{feature}` | Use Cases (lógica de negócio) |
| `domain/validators` | Regras de validação isoladas |
| `gateways/config` | Parsing de configuração |
| `gateways/http` | Handlers REST |
| `gateways/events` | Handlers e Publishers de eventos |
| `gateways/repositories` | Acesso a banco de dados |
| `gateways/{client}` | Clientes HTTP externos |
| `extensions` | Utilitários do domínio |
