# Plan — mvp-userflow (Start Workout Session)

## 1) Inputs usados

### Artefatos de Research
- `.thoughts/mvp-userflow/api-contract.yaml` — Contrato OpenAPI completo (endpoints, schemas, validações)
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — Arquitetura AS-IS/TO-BE, modelo de dados, decisões (CRUD + Audit Log), segurança, observabilidade
- `.thoughts/mvp-userflow/bff-aggregation-strategy.md` — Estratégia de agregação (use cases atômicos + agregação no handler HTTP)

### Análise do Repositório
- `internal/kinetria/domain/entities/entities.go` — Template vazio (sem entidades implementadas)
- `internal/kinetria/domain/errors/errors.go` — Erros básicos existem (ErrNotFound, ErrConflict, ErrMalformedParameters, ErrFailedDependency)
- `internal/kinetria/domain/vos/vos.go` — Template vazio (sem VOs implementados)
- `internal/kinetria/domain/ports/repositories.go` — Template vazio (sem ports definidos)
- `internal/kinetria/gateways/config/config.go` — Config básica (APP_NAME, ENVIRONMENT, REQUEST_TIMEOUT)
- `internal/kinetria/gateways/http/handler.go` — Handler vazio (sem rotas implementadas)
- `internal/kinetria/gateways/repositories/repository.go` — Repository vazio (sem implementação)
- `migrations/` — Vazio (apenas .gitkeep)

### Dependências Assumidas
**Do plano foundation-infrastructure** (assumimos já implementado):
- ✅ Migrations SQL criadas (users, workouts, exercises, sessions, set_records, refresh_tokens, audit_log)
- ✅ Entidades de domínio básicas (User, Workout, Exercise, Session, SetRecord, AuditLog)
- ✅ Value Objects (SessionStatus, SetStatus)
- ✅ Autenticação JWT implementada (middleware que injeta userID no contexto)
- ✅ Database pool configurado (PostgreSQL + SQLC)
- ✅ Infraestrutura básica (Docker Compose, health check, config, errors)

**Se essas dependências não existirem, registramos como GAP crítico e ajustamos escopo.**

---

## 2) AS-IS (resumo)

### Estado Atual
- **Repositório**: estado de scaffold inicial (greenfield)
- **Migrations**: vazias (dependência do foundation-infrastructure)
- **Entidades**: não implementadas (apenas templates)
- **Use cases**: nenhum implementado
- **Handlers HTTP**: estrutura existe mas sem rotas
- **Repositories**: estrutura existe mas sem implementação
- **Autenticação**: não implementada (dependência do foundation-infrastructure)

### Gaps Críticos Identificados
Se foundation-infrastructure não foi implementado:
1. ❌ Migrations SQL (sessions, workouts, users, audit_log)
2. ❌ Entidade Session, Workout, User
3. ❌ Middleware JWT (autenticação)
4. ❌ Database pool
5. ❌ Value Objects (SessionStatus)

**Ação**: Implementar foundation-infrastructure **ANTES** desta feature.

---

## 3) TO-BE (proposta)

### 3.1) Interface HTTP

#### Endpoint
```
POST /api/v1/sessions
```

#### Request
```json
{
  "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
}
```

**Headers obrigatórios**:
- `Authorization: Bearer <JWT>` — userID extraído do token

**Validações (handler)**:
- `workoutId`: required, formato UUID válido
- Token JWT válido (middleware)

#### Response Success (201 Created)
```json
{
  "data": {
    "id": "d4e5f6a7-b8c9-0123-defa-234567890123",
    "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "userId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "startedAt": "2026-02-23T14:00:00Z",
    "finishedAt": null,
    "status": "active"
  }
}
```

#### Error Responses

**401 Unauthorized** — Token inválido ou expirado
```json
{
  "code": "UNAUTHORIZED",
  "message": "Invalid or expired access token."
}
```

**422 Unprocessable Entity** — Validação falhou
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Request body is invalid.",
  "details": {
    "workoutId": "must be a valid UUID"
  }
}
```

**404 Not Found** — Workout não existe ou não pertence ao usuário
```json
{
  "code": "WORKOUT_NOT_FOUND",
  "message": "Workout with id 'b2c3d4e5-f6a7-8901-bcde-f12345678901' was not found."
}
```

**409 Conflict** — Usuário já tem sessão ativa
```json
{
  "code": "ACTIVE_SESSION_EXISTS",
  "message": "User already has an active session. Finish or abandon it before starting a new one."
}
```

---

### 3.2) Contrato de Domínio

#### Use Case: StartSession

**Input**:
```go
type StartSessionInput struct {
    UserID    uuid.UUID // extraído do JWT
    WorkoutID uuid.UUID // do request body
}
```

**Output**:
```go
type StartSessionOutput struct {
    Session entities.Session
}
```

**Regras de Negócio (validações no use case)**:
1. ✅ **WorkoutID obrigatório** — deve ser UUID válido
2. ✅ **Workout ownership** — workout deve existir e pertencer ao userID
3. ✅ **Sessão ativa única** — usuário não pode ter mais de uma sessão ativa
4. ✅ **Registro de auditoria** — toda criação de sessão deve ser registrada no audit_log

**Erros de Domínio**:
- `ErrNotFound` — workout não encontrado ou não pertence ao usuário
- `ErrConflict` — usuário já tem sessão ativa
- `ErrMalformedParameters` — workoutID inválido

---

### 3.3) Persistência

#### Repository Interface (Port)
```go
type SessionRepository interface {
    // Create insere nova sessão no banco
    Create(ctx context.Context, session *entities.Session) error
    
    // FindActiveByUserID retorna sessão ativa do usuário (se existir)
    FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error)
}

type WorkoutRepository interface {
    // FindByID retorna workout por ID
    FindByID(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)
    
    // ExistsByIDAndUserID verifica se workout existe e pertence ao usuário
    ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error)
}

type AuditLogRepository interface {
    // Append registra evento de auditoria (fire-and-forget ou transacional)
    Append(ctx context.Context, entry *entities.AuditLog) error
}
```

#### SQLC Queries
```sql
-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, workout_id, started_at, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: FindActiveSessionByUserID :one
SELECT id, user_id, workout_id, started_at, finished_at, status, notes, created_at, updated_at
FROM sessions
WHERE user_id = $1 AND status = 'active'
LIMIT 1;

-- name: ExistsWorkoutByIDAndUserID :one
SELECT EXISTS(
    SELECT 1 FROM workouts WHERE id = $1 AND user_id = $2
) AS exists;

-- name: AppendAuditLog :exec
INSERT INTO audit_log (id, user_id, entity_type, entity_id, action, action_data, occurred_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
```

**Database Constraint (já deve existir na migration)**:
```sql
-- Garante que usuário só pode ter 1 sessão ativa
CREATE UNIQUE INDEX sessions_active_unique_user 
ON sessions (user_id) 
WHERE status = 'active';
```

---

### 3.4) Fluxo de Execução

```
┌────────────────┐
│ HTTP Request   │ POST /api/v1/sessions
│ Authorization  │ Bearer <JWT>
│ Body: workoutId│
└───────┬────────┘
        │
        ▼
┌────────────────────────────────────────────────┐
│ 1. HTTP Handler (gateways/http)                │
│  - Valida request body (validator)             │
│  - Extrai userID do contexto (JWT middleware)  │
│  - Chama StartSessionUseCase                   │
└───────┬────────────────────────────────────────┘
        │
        ▼
┌────────────────────────────────────────────────┐
│ 2. StartSessionUseCase (domain/sessions)       │
│  ┌──────────────────────────────────────────┐  │
│  │ a) Validar WorkoutID (formato UUID)      │  │
│  │ b) Verificar workout ownership           │  │
│  │    → WorkoutRepo.ExistsByIDAndUserID     │  │
│  │    → Se não existe: return ErrNotFound   │  │
│  │ c) Verificar sessão ativa duplicada      │  │
│  │    → SessionRepo.FindActiveByUserID      │  │
│  │    → Se encontrou: return ErrConflict    │  │
│  │ d) Criar entidade Session                │  │
│  │    → ID: uuid.New()                      │  │
│  │    → UserID: input.UserID                │  │
│  │    → WorkoutID: input.WorkoutID          │  │
│  │    → StartedAt: time.Now()               │  │
│  │    → Status: SessionStatusActive         │  │
│  │    → Notes: ""                           │  │
│  │ e) Persistir sessão                      │  │
│  │    → SessionRepo.Create(session)         │  │
│  │ f) Registrar audit log                   │  │
│  │    → AuditLogRepo.Append(...)            │  │
│  │ g) Retornar StartSessionOutput           │  │
│  └──────────────────────────────────────────┘  │
└───────┬────────────────────────────────────────┘
        │
        ▼
┌────────────────────────────────────────────────┐
│ 3. HTTP Response                               │
│  - Success: 201 Created + Session DTO          │
│  - Error: 401/404/409/422 + ApiError           │
└────────────────────────────────────────────────┘
```

---

### 3.5) Auditoria (Audit Log)

**Pattern obrigatório**: toda mutação de Session deve ser registrada.

```go
// Após criação bem-sucedida
auditEntry := &entities.AuditLog{
    ID:         uuid.New(),
    UserID:     session.UserID,
    EntityType: "session",
    EntityID:   session.ID,
    Action:     "created",
    ActionData: json.Marshal(map[string]interface{}{
        "workoutId": session.WorkoutID,
        "startedAt": session.StartedAt,
    }),
    OccurredAt: time.Now(),
    IPAddress:  extractIPFromContext(ctx), // do request
    UserAgent:  extractUserAgentFromContext(ctx),
}
uc.auditRepo.Append(ctx, auditEntry) // fire-and-forget ou dentro da tx
```

**Nota**: Se auditoria for crítica, deve estar dentro da mesma transação. Se for fire-and-forget, pode ser async.

---

### 3.6) Observabilidade

#### Logs Estruturados (zerolog)
```go
log.Info().
    Str("method", "POST").
    Str("path", "/sessions").
    Str("user_id", userID.String()).
    Str("workout_id", workoutID.String()).
    Str("session_id", session.ID.String()).
    Int("status", 201).
    Dur("duration_ms", duration).
    Msg("session_created")
```

**Regras**:
- ✅ Log de sucesso (INFO)
- ✅ Log de erro de validação (WARN)
- ✅ Log de erro de domínio (ERROR)
- ❌ Nunca logar tokens, passwords, PII sensível

#### Métricas (Prometheus)
```go
// Counters
sessions_started_total{user_id}
sessions_start_errors_total{error_type}

// Histogram
session_start_duration_seconds
```

---

## 4) Decisões e Assunções

### Decisões Arquiteturais
1. ✅ **Use case atômico** — `StartSessionUseCase` faz apenas iniciar sessão (não registra sets)
2. ✅ **Agregação no handler** — handler pode agregar múltiplos use cases se necessário (conforme bff-aggregation-strategy.md)
3. ✅ **Validação em camadas**:
   - Handler: validação de sintaxe (formato UUID, required fields)
   - Use case: validação de negócio (ownership, sessão duplicada)
4. ✅ **Audit log obrigatório** — decisão do backend-architecture-report.simplified.md
5. ✅ **SessionStatus como enum** — value object para segurança de tipo

### Assunções
1. ⚙️ **Autenticação JWT** — userID sempre disponível no contexto (middleware já implementado)
2. ⚙️ **Database pool configurado** — SQLC queries podem ser executadas
3. ⚙️ **Migrations aplicadas** — constraint UNIQUE de sessão ativa já existe
4. ⚙️ **Workouts seed data existe** — usuário tem workouts para iniciar sessões
5. ⚙️ **Config de timeout** — REQUEST_TIMEOUT já configurado (do config básico)

### Gaps Registrados
Se foundation-infrastructure não foi implementado:
- ❌ Criar migrations SQL
- ❌ Criar entidades de domínio
- ❌ Configurar database pool
- ❌ Implementar middleware JWT

---

## 5) Riscos / Edge Cases

### Riscos Principais

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Sessão ativa duplicada** (race condition) | Alta | Alto | Database constraint `UNIQUE(user_id) WHERE status='active'` + validação no use case |
| **Workout não pertence ao usuário** | Média | Alto | Validação obrigatória no use case (`ExistsByIDAndUserID`) |
| **Client retry cria duplicação** | Média | Médio | Idempotência: se erro for conflict (409), client pode consultar sessão ativa |
| **Audit log falha mas sessão é criada** | Baixa | Médio | Decidir: transacional (rollback se falhar) vs fire-and-forget (melhor performance) |
| **JWT expirado durante request** | Média | Baixo | Middleware rejeita com 401, client faz refresh |

### Edge Cases a Cobrir

1. **Workout deletado entre validação e criação**:
   - FK constraint no DB vai falhar → 500 Internal Server Error
   - Mitigação: foreign key constraint garante integridade

2. **UserID inválido no JWT**:
   - Middleware deve validar JWT antes de chegar no handler
   - Se falhar validação: 401 Unauthorized

3. **Concorrência**: dois requests simultâneos tentam criar sessão:
   - Database constraint UNIQUE vai rejeitar o segundo
   - Use case retorna ErrConflict → 409 Conflict

4. **Workout existe mas não pertence ao usuário**:
   - Validação `ExistsByIDAndUserID` retorna false
   - Use case retorna ErrNotFound → 404 Not Found (não vazar que workout existe)

---

## 6) Rollout / Compatibilidade

### Estratégia de Deploy

**Fase 1: Preparação**
- ✅ Criar migrations (se não existem)
- ✅ Aplicar migrations em dev environment
- ✅ Verificar constraints criados (`sessions_active_unique_user`)

**Fase 2: Implementação**
- ✅ Implementar entidades, VOs, use case, repository, handler
- ✅ Escrever testes (unitários + integração)
- ✅ Code review + ajustes

**Fase 3: Validação**
- ✅ Testes de integração contra DB real (Docker Compose)
- ✅ Teste de concorrência (criar 2 sessões simultâneas → 1 deve falhar)
- ✅ Teste de ownership (tentar iniciar sessão de workout de outro usuário)
- ✅ `make lint` sem warnings

**Fase 4: Deploy**
- ✅ Merge to main
- ✅ Deploy em staging
- ✅ Smoke test: `POST /sessions` retorna 201
- ✅ Health check: `/health` responde OK
- ✅ Monitorar métricas: `sessions_started_total`, `session_start_errors_total`

### Rollback Plan
Se houver problemas após deploy:
1. Reverter código (git revert)
2. Migrations **não** devem ser revertidas (apenas expand/contract)
3. Se constraint UNIQUE causar problemas, pode ser removido em migration posterior

### Compatibilidade
- ✅ **Backward compatible**: endpoint novo, não afeta endpoints existentes
- ✅ **Database**: constraint UNIQUE é seguro (apenas rejeita duplicação)
- ✅ **API contract**: segue OpenAPI spec (sem breaking changes)

---

## 7) Próximos Passos (Pós-Implementação)

Após implementar StartSession, features relacionadas:
1. **RecordSet** — `POST /sessions/{id}/sets` (registrar séries)
2. **FinishSession** — `PATCH /sessions/{id}/finish` (finalizar sessão)
3. **AbandonSession** — `PATCH /sessions/{id}/abandon` (abandonar sessão)
4. **GetSession** — `GET /sessions/{id}` (consultar sessão e sets registrados)
5. **Dashboard** — `GET /dashboard` (agregação: sessão ativa + stats)

---

## 8) Checklist de Implementação

Antes de considerar feature completa:
- [ ] Entidades criadas: Session, Workout, User, AuditLog
- [ ] Value Objects: SessionStatus (active, completed, abandoned)
- [ ] Ports: SessionRepository, WorkoutRepository, AuditLogRepository
- [ ] SQLC queries: CreateSession, FindActiveSessionByUserID, ExistsWorkoutByIDAndUserID, AppendAuditLog
- [ ] Use Case: StartSessionUseCase implementado com todas validações
- [ ] Handler HTTP: POST /sessions com validação de input
- [ ] DTOs: StartSessionRequestDTO, SessionResponseDTO
- [ ] Error mapping: domain errors → HTTP status codes
- [ ] Audit log registrado em toda criação
- [ ] Logs estruturados (zerolog)
- [ ] Testes unitários: use case (table-driven, mocks)
- [ ] Testes de integração: handler + DB real
- [ ] Testes de edge cases: duplicação, ownership, concorrência
- [ ] Documentação: comentários Godoc em funções exportadas
- [ ] `make lint` passa sem warnings
- [ ] `make test` passa com cobertura > 70%
- [ ] Validação contra OpenAPI spec (formato, status codes)

---

**Documento gerado em**: 2026-02-23  
**Feature**: mvp-userflow (Start Workout Session)  
**Status**: ✅ Pronto para implementação  
**Dependências**: foundation-infrastructure (migrations, auth, entities básicas)
