# Plan — mvp-userflow

## 1) Inputs usados

- `.thoughts/mvp-userflow/api-contract.yaml` — OpenAPI 3.0 com todos os endpoints, schemas e códigos de erro
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — Decisões arquiteturais, modelo de dados, segurança, observabilidade e roadmap
- `.thoughts/mvp-userflow/bff-aggregation-strategy.md` — Estratégia de agregação para o endpoint `/dashboard`
- `.thoughts/foundation-infrastructure/plan.md` — Plano de infraestrutura de fundação (migrations, entidades base, Docker)
- `.thoughts/foundation-infrastructure/tasks.md` — Backlog da feature de fundação (referência para dependências)

---

## 2) AS-IS (resumo)

### Estado Atual do Projeto

O repositório está em estado de **scaffold inicial** (greenfield). A feature `foundation-infrastructure` está planejada mas ainda não implementada.

#### ✅ Já existe (scaffold)
- Estrutura de diretórios hexagonal completa (`domain/`, `gateways/`, `cmd/`)
- Injeção de dependência com Fx (esqueleto no `main.go`)
- Configuração SQLC (`sqlc.yaml`)
- Config via envconfig (`gateways/config/config.go`)
- Erros de domínio genéricos (`domain/errors/errors.go`)
- Makefile com run/build/test/sqlc/mocks
- Linter configurado (golangci-lint)
- Dependências: Go 1.25, Chi v5, Fx, PGX v5, Validator, UUID, Envconfig

#### ❌ Não existe (gaps)
- Nenhuma migration SQL — `migrations/` vazio
- Nenhuma entidade de domínio real — apenas template
- Nenhum VO, port, use case ou handler implementado
- Sem Docker/Docker Compose
- Sem autenticação JWT
- Sem conexão com PostgreSQL
- Módulos `pkg/` (`xuc`, `xhttp`, `xlog`, etc.) não existem
- Nenhum teste implementado

---

## 3) TO-BE (proposta)

### Arquitetura

**Monolito modular** com arquitetura hexagonal, dividido nos seguintes domínios:

```
internal/kinetria/domain/
├── auth/          # Registro, login, refresh, logout
├── dashboard/     # Agregação de dados e estatísticas
├── workouts/      # Consulta de workouts e exercises
└── sessions/      # Tracking de execução de treino
```

### Interface (HTTP REST)

Base path: `/api/v1`

| Método | Path | Auth | Descrição |
|--------|------|------|-----------|
| POST | `/auth/register` | ❌ | Criar usuário |
| POST | `/auth/login` | ❌ | Autenticar e obter tokens |
| POST | `/auth/refresh` | ❌ | Renovar access token |
| POST | `/auth/logout` | ✅ JWT | Revogar refresh token |
| GET | `/dashboard` | ✅ JWT | Dados agregados (user, workout do dia, semana, stats) |
| GET | `/workouts` | ✅ JWT | Listar workouts com paginação |
| GET | `/workouts/{id}` | ✅ JWT | Detalhe do workout com exercises |
| POST | `/sessions` | ✅ JWT | Iniciar sessão de treino |
| POST | `/sessions/{id}/sets` | ✅ JWT | Registrar série (exercício + peso + reps) |
| PATCH | `/sessions/{id}/finish` | ✅ JWT | Finalizar sessão |
| PATCH | `/sessions/{id}/abandon` | ✅ JWT | Abandonar sessão |

**Wrapper de resposta padrão**:
- Sucesso: `{ "data": {...}, "meta": {...} }`
- Erro: `{ "code": "ERROR_CODE", "message": "...", "details": {...} }`

### Contratos (entidades de domínio)

```go
type User struct {
    ID              uuid.UUID
    Name            string
    Email           string       // unique
    PasswordHash    string
    ProfileImageURL string       // default: /assets/avatars/default.png
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type Workout struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Name        string
    Description string           // max 500 chars
    Type        string           // "FORÇA"|"HIPERTROFIA"|"MOBILIDADE"|"CONDICIONAMENTO"
    Intensity   string           // "BAIXA"|"MODERADA"|"ALTA"
    Duration    int              // minutos (calculado a partir dos exercises)
    ImageURL    string           // default baseado no Type
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Exercise struct {
    ID           uuid.UUID
    WorkoutID    uuid.UUID
    Name         string
    ThumbnailURL string          // default: /assets/exercises/generic.png
    Sets         int             // min 1
    Reps         string          // "8-12" ou "10"
    Muscles      []string        // JSONB, min 1
    RestTime     int             // segundos, default 60
    Weight       float64         // kg, 0 para bodyweight
    OrderIndex   int
}

type Session struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    WorkoutID  uuid.UUID
    StartedAt  time.Time
    FinishedAt *time.Time        // null enquanto ativa
    Status     string            // "active"|"completed"|"abandoned"
    Notes      string            // max 1000 chars
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type SetRecord struct {
    ID         uuid.UUID
    SessionID  uuid.UUID
    ExerciseID uuid.UUID
    SetNumber  int              // min 1
    Weight     float64          // kg, 0 para bodyweight
    Reps       int              // min 0 (0 = falha)
    Status     string           // "completed"|"skipped"
    RecordedAt time.Time
}

// UNIQUE constraint: (session_id, exercise_id, set_number)

type RefreshToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Token     string            // hash SHA-256 do token opaco
    ExpiresAt time.Time
    RevokedAt *time.Time        // null = válido
    CreatedAt time.Time
}

type AuditLog struct {
    ID           uuid.UUID
    UserID       uuid.UUID
    EntityType   string         // "session", "set_record", "workout"
    EntityID     uuid.UUID
    Action       string         // "created", "updated", "deleted", "completed", "abandoned"
    ActionData   json.RawMessage
    OccurredAt   time.Time
    IPAddress    string
    UserAgent    string
}
```

### Persistência (tabelas / migrations)

| Migration | Tabela | Índices principais |
|-----------|--------|--------------------|
| 001 | `users` | `UNIQUE(email)` |
| 002 | `workouts` | `(user_id)`, `(user_id, type)` |
| 003 | `exercises` | `(workout_id)`, `(workout_id, order_index)` |
| 004 | `sessions` | `(user_id)`, `(workout_id)`, `(user_id, status)` |
| 005 | `set_records` | `(session_id)`, `UNIQUE(session_id, exercise_id, set_number)` |
| 006 | `refresh_tokens` | `(user_id)`, `UNIQUE(token)` |
| 007 | `audit_log` | `(user_id, occurred_at)`, `(entity_type, entity_id)` |

**Constraints críticos**:
- `sessions`: `UNIQUE(user_id) WHERE status='active'` — impede sessão ativa duplicada
- `set_records`: `UNIQUE(session_id, exercise_id, set_number)` — idempotência em retries
- `refresh_tokens`: `token` é hash SHA-256 do token opaco

### Observabilidade

- **Logs**: JSON estruturado via `xlog` (zerolog), com `user_id`, `session_id`, duração; sem PII/senha/token
- **Tracing**: OpenTelemetry em todos os use cases e queries DB
- **Métricas**: `http_requests_total`, `http_request_duration_seconds`, `active_sessions_total`, `workout_sessions_completed_total`
- **Health check**: `/health` (liveness + readiness com DB ping)

### Autenticação e Segurança

- **Senha**: bcrypt cost 12
- **Access Token**: JWT HS256, TTL 1h, payload: `{ sub: userID, exp, iat }`
- **Refresh Token**: token opaco de 32 bytes, hasheado (SHA-256) no banco, TTL 30 dias
- **Autorização**: middleware JWT em todas as rotas protegidas; todos os use cases validam `userID` na query
- **Rate Limiting**: `/auth/login` 20 req/min por IP, `/auth/register` 10 req/min por IP, `/sessions/:id/sets` 500 req/min por user
- **Validação de input**: `go-playground/validator/v10` em todos os handlers

### Padrão de agregação — Dashboard

Conforme decisão em `bff-aggregation-strategy.md`:
- **Agregação no Handler HTTP** (não no domain)
- Chamadas paralelas via goroutines para user, workouts, sessions, stats
- Use cases permanecem atômicos e reutilizáveis por qualquer cliente

```
Handler /dashboard
  ├── go GetUserUC.Execute(ctx)
  ├── go ListWorkoutsUC.Execute(ctx) [scheduled today]
  ├── go GetWeekProgressUC.Execute(ctx) [últimos 7 dias]
  └── go GetWeekStatsUC.Execute(ctx) [calorias, tempo total]
```

---

## 4) Decisões e Assunções

| # | Decisão | Justificativa |
|---|---------|---------------|
| D1 | CRUD + Audit Log (sem Event Sourcing) | Simplicidade para MVP; time pequeno; auditabilidade suficiente |
| D2 | Monolito modular (sem microserviços) | Deploy simples, transações ACID, debugging fácil para MVP |
| D3 | JWT stateless (access) + refresh token revogável | TTL curto mitiga impossibilidade de revogação; refresh permite logout real |
| D4 | Agregação de dashboard no handler HTTP | Domain permanece agnóstico ao cliente; reutilizável por web/GraphQL/gRPC |
| D5 | SQLC para type-safe queries | Zero overhead de ORM; código gerado verificado em tempo de compilação |
| D6 | Sem seed de workouts hardcodado | Workouts são por usuário (`user_id`); seed pode ser gerado via migration específica ou API futura |
| D7 | Weight em float64 (kg) | Permite precisão decimal (ex.: 82.5 kg) |
| D8 | Reps como string na entidade Exercise | Suporta ranges "8-12"; SetRecord.Reps é int (valor real executado) |

**Assunções**:
- Não há integração com serviços externos no MVP (sem wearables, sem notificações push)
- Workouts podem ser criados pelo próprio usuário (CRUD futuro) ou via seed de dados inicial
- `/dashboard` retorna dados da **semana corrente** para weekProgress e stats
- `todayWorkout` é o próximo workout não concluído agendado (ou null — lógica de agendamento fora do MVP v1)

---

## 5) Riscos / Edge Cases

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| SetRecord duplicado (client retry) | Alta | Médio | `UNIQUE(session_id, exercise_id, set_number)` → 409 idempotente |
| Sessão ativa duplicada | Média | Alto | `UNIQUE(user_id) WHERE status='active'` + validação no use case |
| JWT secret vazado | Baixa | Crítico | TTL 1h + rotação via secrets manager; refresh token revogável |
| N+1 no `/dashboard` | Média | Médio | Queries com JOINs + chamadas paralelas no handler |
| Lock de tabela em migrations | Média | Alto | Migrations expand/contract; índices com `CONCURRENTLY` onde possível |
| Race condition no finish/abandon simultâneo | Baixa | Médio | `UPDATE ... WHERE status='active'` é atômico; retorna 409 se já fechada |
| Overflow de `set_number` | Baixa | Baixo | Validação `max=20` no input; limite razoável para qualquer treino |

---

## 6) Rollout / Compatibilidade

- **MVP (v1)**: todos os endpoints novos — sem breaking changes
- **Pós-MVP**: adicionar criação de workouts via API (`POST /workouts`) sem alterar contratos existentes
- **Expansão de cliente**: handlers BFF adicionais (web, GraphQL) podem ser criados sem alterar domain
- **Eventos assíncronos (futuro)**: publicar `session.completed` e `workout.created` no Kafka/SNS sem alterar contratos HTTP
- **Migrations**: sempre additive (expand) antes de contractive (contract); nunca dropar colunas sem ciclo de deprecação
