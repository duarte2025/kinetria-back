# Plan — foundation-infrastructure

## 1) Inputs usados

### Artefatos de Research
- Nenhum artefato de research específico para esta feature
- **Razão**: Esta é a feature inicial de fundação do projeto

### Análise do Repositório
- **Repository**: `/home/runner/work/kinetria-back/kinetria-back`
- **Arquivos analisados**:
  - `NEXT_STEPS.md` - Guia de estrutura inicial e próximos passos
  - `README.md` - Documentação da arquitetura hexagonal
  - `cmd/kinetria/api/main.go` - Entry point com Fx DI (ainda vazio)
  - `internal/kinetria/domain/` - Estrutura de domínio (vazia, apenas templates)
  - `internal/kinetria/gateways/` - Gateways configurados (http, repositories, config, events)
  - `sqlc.yaml` - Configuração do SQLC para geração de código
  - `migrations/` - Diretório vazio (apenas .gitkeep)
  - `Makefile` - Comandos de build, test, sqlc, mocks
  - `.env.example` - Configuração base (APP_NAME, ENVIRONMENT, REQUEST_TIMEOUT)
  - `.thoughts/mvp-userflow/api-contract.yaml` - Contrato da API (referência para entidades)

### Análise de Requisitos da API (mvp-userflow)
Baseado no contrato OpenAPI, identificamos as seguintes entidades necessárias:
- **Users** (usuários com autenticação)
- **Workouts** (planos de treino)
- **Exercises** (exercícios do catálogo)
- **Sessions** (sessões de treino ativas)
- **Set Records** (registros de séries executadas)
- **Refresh Tokens** (tokens para renovação de autenticação)
- **Audit Log** (log de auditoria para rastreabilidade)

---

## 2) AS-IS (resumo)

### Estado Atual do Projeto
O repositório está em estado de **scaffold inicial** (greenfield):

#### ✅ Já Configurado
- Estrutura de diretórios da arquitetura hexagonal completa
- Injeção de dependências com Fx (esqueleto no main.go)
- Configuração do SQLC para geração de queries (`sqlc.yaml`)
- Gateway de configuração com envconfig (`gateways/config/config.go`)
- Erros de domínio básicos (`domain/errors/errors.go`)
- Makefile com comandos úteis (run, build, test, sqlc, mocks)
- Linter configurado (golangci-lint)
- Dependencies: Go 1.25, Chi v5, Fx, PGX v5, Validator, UUID, Envconfig

#### ❌ Não Existe (Gaps)
- **Nenhuma migration SQL** - diretório `migrations/` vazio
- **Nenhuma entidade de domínio** - arquivo template apenas
- **Nenhum Value Object (VO)** - diretório `vos/` vazio
- **Nenhuma interface (port) definida** - diretório `ports/` vazio
- **Nenhum use case implementado** - apenas exemplo comentado
- **Nenhum handler HTTP** - estrutura existe mas sem rotas
- **Nenhum repositório** - estrutura existe mas sem implementação
- **Docker/Docker Compose não existe** - sem containerização
- **Health check não implementado** - comentado no main.go (`xhealth.Module()`)
- **Database pool não configurado** - sem conexão com PostgreSQL
- **Módulos pkg/ não existem** - xfx, xlog, xhttp, xhealth, xuc, etc. estão referenciados mas não criados

#### Observações
- Projeto foi scaffoldado mas **nenhuma feature foi implementada**
- `NEXT_STEPS.md` indica necessidade de criar módulos `pkg/` compartilhados
- Comentários no código indicam "TODO: Adicionar quando disponível"
- Não há testes implementados ainda

---

## 3) TO-BE (proposta)

### Objetivo da Feature: foundation-infrastructure
Criar a **infraestrutura fundamental** do projeto para habilitar desenvolvimento de features:

1. **Migrations SQL** - esquema de banco de dados completo
2. **Docker Compose** - ambiente de desenvolvimento local
3. **Entidades de domínio + VOs** - modelagem core do negócio
4. **Health check endpoint** - `/health` para monitoramento

---

### 3.1) Migrations SQL (PostgreSQL)

#### Estrutura de Arquivos
Seguir padrão de numeração sequencial com timestamp/ordem:
```
migrations/
├── 001_create_users.sql
├── 002_create_workouts.sql
├── 003_create_exercises.sql
├── 004_create_sessions.sql
├── 005_create_set_records.sql
├── 006_create_refresh_tokens.sql
└── 007_create_audit_log.sql
```

#### Esquema Proposto

##### **001_create_users.sql**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    profile_image_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Decisões**:
- `id` como UUID para distribuição e segurança
- `email` único com índice para lookup rápido
- `password_hash` armazenado com bcrypt (não plaintext)
- `profile_image_url` nullable (usuário pode não ter foto de perfil)
- `created_at/updated_at` para auditoria temporal

---

##### **002_create_workouts.sql**
```sql
CREATE TYPE workout_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status workout_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workouts_user_id ON workouts(user_id);
CREATE INDEX idx_workouts_status ON workouts(status);
CREATE INDEX idx_workouts_user_status ON workouts(user_id, status);
```

**Decisões**:
- ENUM para status (segurança de tipo)
- Foreign key para `users` com CASCADE delete (user deletado = workouts deletados)
- Índices compostos para queries comuns (filtrar workouts por user + status)

---

##### **003_create_exercises.sql**
```sql
CREATE TYPE exercise_category AS ENUM ('strength', 'cardio', 'flexibility', 'balance');
CREATE TYPE muscle_group AS ENUM ('chest', 'back', 'legs', 'shoulders', 'arms', 'core', 'full_body');

CREATE TABLE exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category exercise_category NOT NULL,
    primary_muscle_group muscle_group NOT NULL,
    equipment_required VARCHAR(255),
    difficulty_level INT NOT NULL CHECK (difficulty_level BETWEEN 1 AND 5),
    video_url VARCHAR(500),
    thumbnail_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_muscle_group ON exercises(primary_muscle_group);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty_level);
```

**Decisões**:
- Catálogo global de exercícios (sem user_id, compartilhado)
- ENUMs para categorias e grupos musculares (validação no DB)
- `difficulty_level` com constraint (1-5)
- URLs para assets de vídeo/imagem
- Índices para filtros comuns (categoria, músculo, dificuldade)

---

##### **004_create_sessions.sql**
```sql
CREATE TYPE session_status AS ENUM ('in_progress', 'completed', 'cancelled');

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT,
    status session_status NOT NULL DEFAULT 'in_progress',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_started_at ON sessions(started_at DESC);
CREATE INDEX idx_sessions_user_status ON sessions(user_id, status);
```

**Decisões**:
- `workout_id` com RESTRICT (evita deletar workout com sessões ativas)
- `completed_at` nullable (só preenchido quando finalizado)
- Índice em `started_at DESC` para listar sessões recentes
- ENUM para status da sessão

---

##### **005_create_set_records.sql**
```sql
CREATE TABLE set_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    set_number INT NOT NULL CHECK (set_number > 0),
    reps INT CHECK (reps >= 0),
    weight_kg DECIMAL(6,2) CHECK (weight_kg >= 0),
    duration_seconds INT CHECK (duration_seconds >= 0),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_set_records_session_id ON set_records(session_id);
CREATE INDEX idx_set_records_exercise_id ON set_records(exercise_id);
CREATE INDEX idx_set_records_session_exercise ON set_records(session_id, exercise_id);
```

**Decisões**:
- Registros de séries com CASCADE delete (session deletada = sets deletados)
- `reps`, `weight_kg`, `duration_seconds` todos nullable (depende do tipo de exercício)
- `set_number` para ordenação (1ª série, 2ª série, etc.)
- Constraints de validação (valores >= 0)
- DECIMAL para peso (precisão de 2 casas)

---

##### **006_create_refresh_tokens.sql**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_user_revoked ON refresh_tokens(user_id, revoked);
```

**Decisões**:
- `token_hash` único (lookup rápido, segurança - não armazenar token plaintext)
- `revoked` para invalidação manual (logout, security breach)
- `expires_at` com índice para cleanup de tokens expirados
- Índice composto para buscar tokens válidos por usuário

---

##### **007_create_audit_log.sql**
```sql
CREATE TYPE audit_action AS ENUM (
    'user_created', 'user_updated', 'user_deleted',
    'workout_created', 'workout_updated', 'workout_deleted',
    'session_started', 'session_completed', 'session_cancelled',
    'set_recorded', 'set_updated', 'set_deleted',
    'login', 'logout', 'token_refreshed'
);

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action audit_action NOT NULL,
    entity_type VARCHAR(100),
    entity_id UUID,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_metadata ON audit_log USING GIN(metadata);
```

**Decisões**:
- `user_id` com SET NULL (preservar audit mesmo se user deletado)
- ENUM para ações auditáveis (tipagem segura)
- `metadata` JSONB para dados flexíveis (contexto adicional)
- INET para IP (validação de tipo)
- GIN index no JSONB para queries em metadata
- Apenas `created_at` (audit log é imutável)

---

### 3.2) Docker Compose (Desenvolvimento Local)

#### Arquivo: `docker-compose.yml` (raiz do projeto)

```yaml
version: '3.9'

services:
  postgres:
    image: postgres:16-alpine
    container_name: kinetria-postgres
    environment:
      POSTGRES_DB: kinetria
      POSTGRES_USER: kinetria
      POSTGRES_PASSWORD: kinetria_dev_pass
      POSTGRES_HOST_AUTH_METHOD: trust
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U kinetria"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - kinetria-network

  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: kinetria-app
    environment:
      APP_NAME: kinetria
      ENVIRONMENT: development
      REQUEST_TIMEOUT: 10s
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: kinetria
      DB_PASSWORD: kinetria_dev_pass
      DB_NAME: kinetria
      DB_SSL_MODE: disable
      HTTP_PORT: 8080
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - kinetria-network
    volumes:
      - .:/app
    command: ["go", "run", "cmd/kinetria/api/main.go"]

volumes:
  postgres_data:

networks:
  kinetria-network:
    driver: bridge
```

#### Arquivo: `Dockerfile` (raiz do projeto)

```dockerfile
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/kinetria /app/kinetria

EXPOSE 8080

CMD ["/app/kinetria"]
```

#### Arquivo: `.dockerignore`

```
.git
.github
.kiro
.thoughts
bin/
coverage.out
coverage.html
.env
.env.local
*.log
```

**Decisões Docker**:
- PostgreSQL 16 Alpine (leve e moderna)
- Multi-stage build (builder + runtime pequeno)
- Health check no Postgres (app só sobe quando DB estiver pronto)
- Volumes para persistência de dados
- Network isolada para comunicação interna
- Migrations aplicadas automaticamente no init do Postgres
- Environment variables segregadas (dev vs prod)

---

### 3.3) Entidades de Domínio + VOs + Constants

#### Estrutura de Arquivos no Domain

```
internal/kinetria/domain/
├── entities/
│   ├── entities.go          # Arquivo existente (atualizar)
│   ├── user.go              # User entity
│   ├── workout.go           # Workout entity
│   ├── exercise.go          # Exercise entity
│   ├── session.go           # Session entity
│   ├── set_record.go        # SetRecord entity
│   ├── refresh_token.go     # RefreshToken entity
│   └── audit_log.go         # AuditLog entity
├── vos/
│   ├── workout_status.go    # WorkoutStatus enum + validation
│   ├── session_status.go    # SessionStatus enum + validation
│   ├── exercise_category.go # ExerciseCategory enum + validation
│   ├── muscle_group.go      # MuscleGroup enum + validation
│   └── audit_action.go      # AuditAction enum + validation
└── constants/
    ├── defaults.go          # Asset defaults (thumbnails, vídeos)
    └── validation.go        # Validation rules (min/max length, ranges)
```

#### Exemplo de Implementação

##### `entities/user.go`
```go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type UserID = uuid.UUID

type User struct {
    ID           UserID
    Email        string
    Name         string
    PasswordHash string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

##### `vos/workout_status.go`
```go
package vos

import (
    "fmt"
    "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

type WorkoutStatus string

const (
    WorkoutStatusDraft     WorkoutStatus = "draft"
    WorkoutStatusPublished WorkoutStatus = "published"
    WorkoutStatusArchived  WorkoutStatus = "archived"
)

func (s WorkoutStatus) Validate() error {
    switch s {
    case WorkoutStatusDraft, WorkoutStatusPublished, WorkoutStatusArchived:
        return nil
    default:
        return fmt.Errorf("%w: invalid workout status '%s'", errors.ErrMalformedParameters, s)
    }
}

func (s WorkoutStatus) String() string {
    return string(s)
}
```

##### `constants/defaults.go`
```go
package constants

const (
    DefaultExerciseThumbnail = "/assets/defaults/exercise-placeholder.png"
    DefaultExerciseVideo     = ""
    DefaultDifficultyLevel   = 3
)
```

##### `constants/validation.go`
```go
package constants

const (
    MinNameLength         = 1
    MaxNameLength         = 255
    MinDescriptionLength  = 0
    MaxDescriptionLength  = 5000
    MinDifficultyLevel    = 1
    MaxDifficultyLevel    = 5
    MinSetNumber          = 1
    MaxSetNumber          = 100
)
```

**Decisões Domain**:
- Type aliases para IDs (`UserID = uuid.UUID`) - type safety
- VOs com método `Validate()` para business rules
- Constants centralizados para manutenção fácil
- Separação clara: entities (dados), vos (regras), constants (valores fixos)
- Sem dependências externas no domain (puro Go)

---

### 3.4) Health Check Endpoint `/health`

#### Implementação

##### **Criar módulo pkg/xhealth** (simplificado)

Dado que módulos `pkg/` não existem, criar versão **simplificada inline** ou **minimal module**:

**Opção A: Inline no gateway/http** (pragmático)
```go
// internal/kinetria/gateways/http/health/handler.go
package health

import (
    "encoding/json"
    "net/http"
)

type HealthResponse struct {
    Status  string `json:"status"`
    Service string `json:"service"`
    Version string `json:"version"`
}

func NewHealthHandler(appName, version string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        resp := HealthResponse{
            Status:  "healthy",
            Service: appName,
            Version: version,
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(resp)
    }
}
```

**Opção B: Criar pkg/xhealth minimal** (escalável)
```go
// pkg/xhealth/module.go
package xhealth

import (
    "go.uber.org/fx"
    "github.com/go-chi/chi/v5"
)

type Config struct {
    AppName string
    Version string
}

func Module() fx.Option {
    return fx.Module("health",
        fx.Provide(NewHealthHandler),
        fx.Invoke(RegisterHealthRoute),
    )
}

func NewHealthHandler(cfg Config) *HealthHandler {
    return &HealthHandler{appName: cfg.AppName, version: cfg.Version}
}

func RegisterHealthRoute(router chi.Router, handler *HealthHandler) {
    router.Get("/health", handler.Handle)
}
```

**Decisão**: Começar com **Opção A (inline)** para MVP, migrar para **Opção B** quando escalar.

##### **Registrar rota no main.go**

```go
// cmd/kinetria/api/main.go
func main() {
    fx.New(
        fx.Provide(
            config.ParseConfigFromEnv,
            // Database pool provider (TODO)
            // Repositories providers (TODO)
            // Use cases providers (TODO)
            health.NewHealthHandler,
        ),
        fx.Invoke(
            registerRoutes,
        ),
    ).Run()
}

func registerRoutes(router chi.Router, healthHandler http.HandlerFunc) {
    router.Get("/health", healthHandler)
}
```

**Contrato do Endpoint**:
```
GET /health

Response 200 OK:
{
  "status": "healthy",
  "service": "kinetria",
  "version": "undefined"
}
```

**Futuras melhorias**:
- Checar conexão com banco (`db.Ping()`)
- Checar dependências externas (se houver)
- Incluir uptime, memória, etc.

---

### 3.5) Atualização do Config

Adicionar variáveis de ambiente para banco de dados:

```go
// internal/kinetria/gateways/config/config.go
type Config struct {
    AppName     string `envconfig:"APP_NAME" required:"true"`
    Environment string `envconfig:"ENVIRONMENT" required:"true"`
    
    RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`
    
    // Database
    DBHost     string `envconfig:"DB_HOST" required:"true"`
    DBPort     int    `envconfig:"DB_PORT" default:"5432"`
    DBUser     string `envconfig:"DB_USER" required:"true"`
    DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
    DBName     string `envconfig:"DB_NAME" required:"true"`
    DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"require"`
    
    // HTTP Server
    HTTPPort int `envconfig:"HTTP_PORT" default:"8080"`
}
```

Atualizar `.env.example`:
```env
APP_NAME=kinetria
ENVIRONMENT=development
REQUEST_TIMEOUT=5s

DB_HOST=localhost
DB_PORT=5432
DB_USER=kinetria
DB_PASSWORD=kinetria_dev_pass
DB_NAME=kinetria
DB_SSL_MODE=disable

HTTP_PORT=8080
```

---

### 3.6) Database Connection Pool (Provider)

Criar provider para pgxpool:

```go
// internal/kinetria/gateways/repositories/pool.go
package repositories

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
)

func NewDatabasePool(cfg config.Config) (*pgxpool.Pool, error) {
    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s",
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBHost,
        cfg.DBPort,
        cfg.DBName,
        cfg.DBSSLMode,
    )
    
    pool, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to create database pool: %w", err)
    }
    
    if err := pool.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return pool, nil
}
```

Registrar no main.go:
```go
fx.Provide(
    config.ParseConfigFromEnv,
    repositories.NewDatabasePool,
    // ... outros providers
)
```

---

## 4) Decisões e Assunções

### Decisões de Arquitetura

1. **Migrations como SQL puro**
   - Não usar ORM (Gorm, Ent, etc.) - seguir decisão do SQLC
   - Migrations manuais com numeração sequencial
   - Aplicação via `docker-entrypoint-initdb.d` no Postgres (dev) ou ferramenta de migration (prod)

2. **UUIDs como Primary Keys**
   - Melhor distribuição em sistemas distribuídos
   - Segurança (não sequencial/predizível)
   - Compatibilidade com padrões modernos

3. **ENUMs no PostgreSQL**
   - Validação no nível de banco
   - Performance (storage menor que VARCHAR)
   - Type safety no schema

4. **Índices Strategy**
   - Todos os FKs têm índices
   - Colunas em WHERE/JOIN têm índices
   - Índices compostos para queries comuns
   - GIN para JSONB (audit_log.metadata)

5. **Soft Delete vs Hard Delete**
   - **Hard delete** para MVP (simplicidade)
   - Audit log preserva histórico (user deletado = user_id SET NULL no audit)
   - Considerar soft delete em features específicas se necessário

6. **Health Check Inline**
   - Começar simples sem pkg/xhealth completo
   - Migrar para módulo completo quando necessário
   - MVP: apenas status 200 + JSON básico

7. **Docker Compose para Dev**
   - Prod usará Kubernetes/ECS (futuro)
   - Dev: Docker Compose é suficiente
   - Migrations aplicadas automaticamente no init

### Assunções

1. **PostgreSQL 16** é a versão target (moderna, estável)
2. **Go 1.25** é compatível com todas as libs usadas
3. **Não há necessidade de múltiplos bancos** (single DB por enquanto)
4. **Sem autenticação no health endpoint** (público)
5. **Migration tool será decidido depois** (goose, migrate, manual, etc.)
6. **Módulos pkg/ serão criados incrementalmente** (não bloqueante para MVP)
7. **Audit log é append-only** (sem updates/deletes)
8. **Passwords serão hasheados com bcrypt** (feature de auth definirá detalhes)

---

## 5) Riscos / Edge Cases

### Riscos Identificados

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Migrations falharem no Docker init** | Média | Alto | Testar migrations manualmente antes, validar syntax SQL, adicionar rollback scripts |
| **Connection pool esgotar** | Baixa | Médio | Configurar limites adequados no pgxpool, monitorar conexões ativas |
| **ENUMs não sincronizados entre DB e Go** | Alta | Médio | Criar testes de validação, documentar processo de adição de novos valores |
| **Índices insuficientes** | Média | Médio | Monitorar slow queries, usar EXPLAIN ANALYZE, adicionar índices sob demanda |
| **Audit log crescer muito** | Alta | Baixo | Implementar particionamento por data (futuro), cleanup de logs antigos |
| **Health check não verificar DB** | Baixa | Baixo | Aceitável para MVP, adicionar db.Ping() em iteração futura |
| **Docker volumes com permissões incorretas** | Baixa | Baixo | Documentar permissões, usar user mapping se necessário |

### Edge Cases

1. **User deletado com sessões ativas**
   - Comportamento: Cascade delete (sessions deletadas)
   - Alternativa: Considerar soft delete para users críticos

2. **Workout deletado com sessões referenciadas**
   - Comportamento: RESTRICT (bloqueia delete)
   - Solução: Workflow de "archive" ao invés de delete

3. **Token refresh após user deletado**
   - Comportamento: FK cascade (refresh_tokens deletados)
   - Logout forçado (esperado)

4. **Audit log com user_id NULL**
   - Comportamento: ON DELETE SET NULL (preserva audit)
   - Query por user_id = NULL retorna audits de usuários deletados

5. **Migrations já aplicadas no re-run**
   - Docker Compose: `/docker-entrypoint-initdb.d` só roda em DB vazio
   - Solução: Para dev, recriar volume (`docker-compose down -v`)

6. **Concorrência em set_records**
   - Cenário: 2 requests simultâneos criando sets
   - Mitigação: Transações no repository, idempotency keys (futuro)

---

## 6) Rollout / Compatibilidade

### Estratégia de Rollout

Esta feature é **fundacional**, portanto:

#### Fase 1: Infraestrutura Local (Dev)
- ✅ Criar migrations SQL
- ✅ Criar Docker Compose
- ✅ Testar aplicação das migrations
- ✅ Validar conexão app ↔ Postgres
- ✅ Validar health check endpoint

#### Fase 2: Domain Layer
- ✅ Criar entidades de domínio
- ✅ Criar Value Objects com validação
- ✅ Criar constants
- ✅ Atualizar config com variáveis de DB

#### Fase 3: Integração
- ✅ Registrar database pool no Fx
- ✅ Testar end-to-end (Docker up → health check → logs)
- ✅ Documentar setup no README

#### Fase 4: Preparação para Features
- ✅ Migrations aplicadas e validadas
- ✅ Estrutura pronta para criar repositories (SQLC queries)
- ✅ Estrutura pronta para criar use cases
- ✅ Próxima feature pode começar desenvolvimento

### Compatibilidade

#### Backward Compatibility
- **N/A** - Primeira versão, sem compatibilidade retroativa necessária

#### Forward Compatibility
- **Migrations**: Usar migrations versionadas permite rollback
- **ENUMs**: Adicionar novos valores é compatível, remover quebra
- **Schema changes**: Futuras alterações devem usar migrations aditivas (ADD COLUMN, não DROP/RENAME)

### Deployment

#### Desenvolvimento
```bash
# Subir ambiente
docker-compose up -d

# Verificar health
curl http://localhost:8080/health

# Logs
docker-compose logs -f app
```

#### CI/CD (Futuro)
- Pipeline rodará migrations via ferramenta (goose, migrate, etc.)
- Testes de integração validarão schema
- Health check será usado pelo load balancer

---

## 7) Dependências Externas

### Dependências Necessárias

#### Go Packages (já instalados)
- ✅ `github.com/jackc/pgx/v5` - PostgreSQL driver
- ✅ `github.com/google/uuid` - UUID generation
- ✅ `github.com/kelseyhightower/envconfig` - Config parsing
- ✅ `go.uber.org/fx` - Dependency injection

#### Ferramentas de Desenvolvimento
- ✅ Docker & Docker Compose (para ambiente local)
- ✅ SQLC (para geração de queries - já configurado)
- ⚠️ Migration tool (goose, golang-migrate, ou manual) - **decisão pendente**

#### Infraestrutura
- ✅ PostgreSQL 16 (via Docker)

### Novos Pacotes Necessários

**Nenhum** - todas as dependências já estão no `go.mod`.

---

## 8) Próximos Passos (Após Esta Feature)

Após completar `foundation-infrastructure`, o projeto estará pronto para:

1. **Feature: Authentication & Authorization**
   - Implementar use cases de login, registro, refresh token
   - Criar repositories usando SQLC queries
   - Handlers HTTP com JWT

2. **Feature: Workouts Management**
   - CRUD de workouts
   - Associação workout ↔ exercises
   - Listagem/filtros

3. **Feature: Session Tracking**
   - Start/stop session
   - Record sets
   - Histórico de treinos

4. **Feature: Exercise Catalog**
   - Listar exercícios
   - Filtros por categoria/músculo
   - Seed inicial de exercícios

---

## 9) Critérios de Aceite (Feature Completa)

Esta feature está **completa** quando:

### Infraestrutura
- [ ] `docker-compose.yml` criado e funcional
- [ ] `Dockerfile` criado (multi-stage build)
- [ ] `.dockerignore` criado
- [ ] Comando `docker-compose up` sobe app + postgres sem erros
- [ ] PostgreSQL aceita conexões na porta 5432
- [ ] App conecta com sucesso ao Postgres

### Migrations
- [ ] 7 arquivos de migration criados (`001_*.sql` a `007_*.sql`)
- [ ] Migrations aplicadas com sucesso no Postgres
- [ ] Todas as tabelas existem (`\dt` no psql mostra 7 tabelas)
- [ ] Todos os índices foram criados corretamente
- [ ] ENUMs foram criados (workout_status, session_status, etc.)

### Domain Layer
- [ ] 7 arquivos de entidades criados (`user.go`, `workout.go`, etc.)
- [ ] 5 arquivos de VOs criados (`workout_status.go`, etc.)
- [ ] VOs possuem método `Validate()` implementado
- [ ] Arquivo `constants/defaults.go` criado
- [ ] Arquivo `constants/validation.go` criado

### Configuration
- [ ] `config.Config` atualizado com variáveis de DB
- [ ] `.env.example` atualizado com todas as variáveis necessárias
- [ ] Database pool provider criado (`repositories/pool.go`)
- [ ] Pool registrado no Fx (main.go)

### Health Check
- [ ] Handler de `/health` implementado
- [ ] Rota registrada no Chi router
- [ ] Endpoint responde 200 com JSON válido
- [ ] Response contém `status`, `service`, `version`

### Testes
- [ ] `docker-compose up` → `curl /health` retorna 200
- [ ] Logs do app não mostram erros de conexão
- [ ] `docker-compose down -v` e `up` novamente funciona

### Documentação
- [ ] README.md atualizado com instruções Docker
- [ ] Migrations documentadas (comentários no SQL)
- [ ] `.env.example` possui todas as variáveis explicadas

---

## 10) Checklist de Implementação

### Arquivos a Criar (26 arquivos novos)

#### Docker (3 arquivos)
- [ ] `docker-compose.yml`
- [ ] `Dockerfile`
- [ ] `.dockerignore`

#### Migrations (7 arquivos)
- [ ] `migrations/001_create_users.sql`
- [ ] `migrations/002_create_workouts.sql`
- [ ] `migrations/003_create_exercises.sql`
- [ ] `migrations/004_create_sessions.sql`
- [ ] `migrations/005_create_set_records.sql`
- [ ] `migrations/006_create_refresh_tokens.sql`
- [ ] `migrations/007_create_audit_log.sql`

#### Entities (7 arquivos)
- [ ] `internal/kinetria/domain/entities/user.go`
- [ ] `internal/kinetria/domain/entities/workout.go`
- [ ] `internal/kinetria/domain/entities/exercise.go`
- [ ] `internal/kinetria/domain/entities/session.go`
- [ ] `internal/kinetria/domain/entities/set_record.go`
- [ ] `internal/kinetria/domain/entities/refresh_token.go`
- [ ] `internal/kinetria/domain/entities/audit_log.go`

#### Value Objects (5 arquivos)
- [ ] `internal/kinetria/domain/vos/workout_status.go`
- [ ] `internal/kinetria/domain/vos/session_status.go`
- [ ] `internal/kinetria/domain/vos/exercise_category.go`
- [ ] `internal/kinetria/domain/vos/muscle_group.go`
- [ ] `internal/kinetria/domain/vos/audit_action.go`

#### Constants (2 arquivos)
- [ ] `internal/kinetria/domain/constants/defaults.go`
- [ ] `internal/kinetria/domain/constants/validation.go`

#### Gateways (2 arquivos)
- [ ] `internal/kinetria/gateways/repositories/pool.go`
- [ ] `internal/kinetria/gateways/http/health/handler.go`

### Arquivos a Atualizar (4 arquivos)
- [ ] `internal/kinetria/domain/entities/entities.go` (remover comentários de exemplo)
- [ ] `internal/kinetria/gateways/config/config.go` (adicionar DB vars)
- [ ] `.env.example` (adicionar DB vars)
- [ ] `cmd/kinetria/api/main.go` (registrar providers e rotas)
- [ ] `README.md` (adicionar instruções Docker)

---

**Total de mudanças: 26 arquivos novos + 5 arquivos atualizados = 31 arquivos**

---

## Conclusão

Esta feature estabelece a **fundação completa** do projeto Kinetria Backend:
- ✅ Banco de dados estruturado e versionado
- ✅ Ambiente de desenvolvimento containerizado
- ✅ Modelagem de domínio completa
- ✅ Health check para monitoramento
- ✅ Configuração extensível

Após a implementação, o projeto estará pronto para desenvolvimento ágil de features de negócio (auth, workouts, sessions, etc.) seguindo a arquitetura hexagonal estabelecida.
