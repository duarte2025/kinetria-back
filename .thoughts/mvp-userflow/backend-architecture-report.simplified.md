# 🧱 Backend Architecture Report — MVP Kinetria (Simplified)

## 1) Scope

### Problema/objetivo
MVP de plataforma de treinos com:
- **Tracking de treinos**: registro de séries, peso e reps
- **Histórico e progressão**: comparação de performance ao longo do tempo
- **Auditoria completa**: rastreabilidade de todas as ações do usuário

### Domínio/app
**Kinetria Backend Platform**: serviço REST monolítico modular em Go que gerencia o domínio completo de treinos e usuários.

### Interfaces
- **HTTP REST** (`/api/v1`): 11 endpoints públicos (auth, dashboard, workouts, sessions)
- **Autenticação**: JWT Bearer
- **Persistência**: PostgreSQL via SQLC
- **Clientes**: apps mobile (iOS/Android), web app (futuro), integrações externas (futuro)

---

## 2) AS-IS (resumo)

- **Estrutura**: scaffolding hexagonal (domain/gateways/cmd) preparado para múltiplos clientes
- **Estado**: migrations vazias, config básica de Fx e SQLC
- **Contratos**: OpenAPI 3.0 documentado (API REST pública), sem implementação
- **Arquitetura**: preparada para evolução (GraphQL, gRPC ou novos clients podem ser adicionados)

---

## 3) TO-BE (proposta)

### ✅ DECISÃO ARQUITETURAL: CRUD + Audit Log

**Abordagem simplificada para MVP**:
- 🟢 **CRUD tradicional** para todas as entidades (User, Workout, Exercise, Session, SetRecord)
- 🟢 **Audit Log obrigatório** para rastreabilidade completa e compliance
- 🟢 **API RESTful agnóstica**: serve múltiplos tipos de clientes (mobile, web, integrações)
- 🟢 **Complexidade reduzida**: sem event sourcing, snapshots ou read models
- 🟢 **Tempo de implementação**: 50% mais rápido que Event Sourcing

**Justificativa**:
- Plataforma backend deve ser **client-agnostic** (suportar mobile, web, APIs externas)
- Time pequeno (1-2 devs) com padrão CRUD conhecido
- MVP precisa ser entregue rápido (< 8 semanas)
- Audit log bem estruturado fornece rastreabilidade suficiente
- API REST genérica facilita onboarding de novos clientes
- Possibilidade de migração futura para ES se necessário

---

### Service boundaries (monolito modular)

```
internal/kinetria/domain/
├── auth/          # Registro, login, refresh, logout
├── dashboard/     # Agregação de dados e estatísticas do usuário
├── workouts/      # CRUD de workouts e exercises
└── sessions/      # Tracking de execução de treino (histórico e progressão)
```

**Nota sobre a feature "dashboard"** (anteriormente "home"):
- Fornece dados agregados e estatísticas independentes de cliente
- Clientes podem usar os dados do jeito que quiserem (cards, gráficos, listas)
- Não é acoplado a layouts específicos de telas mobile/web

---

### Modelo de dados (CRUD)

**Princípios de design**:
- ✅ Ponteiros **apenas** quando null tem significado semântico (`FinishedAt`, `RevokedAt`)
- ✅ Valores default para todos os outros campos (strings vazias, 0, arrays vazios)
- ✅ Enums obrigatórios validados no use case
- ✅ Assets default configurados (avatares, imagens de workout)

```go
// Entidades core
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
    Duration    int              // minutos (calculado)
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
    Weight     int              // min 0 use grams
    Reps       int              // min 0 (0 = falha)
    Status     string           // "completed"|"skipped"
    RecordedAt time.Time
}

// UNIQUE constraint: (session_id, exercise_id, set_number)

type RefreshToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Token     string            // hash do token
    ExpiresAt time.Time
    RevokedAt *time.Time        // null = válido
    CreatedAt time.Time
}
```

---

### 📋 Audit Log (rastreabilidade completa)

```go
// Tabela append-only para auditoria
type AuditLog struct {
    ID           uuid.UUID
    UserID       uuid.UUID      // indexed
    EntityType   string         // "session", "set_record", "workout"
    EntityID     uuid.UUID      // ID da entidade afetada
    Action       string         // "created", "updated", "deleted", "completed"
    ActionData   json.RawMessage // estado antes/depois ou payload da ação
    OccurredAt   time.Time      // indexed
    IPAddress    string
    UserAgent    string
}

// Indices: (user_id, occurred_at), (entity_type, entity_id)
```

**Uso do Audit Log**:
- ✅ Registrar toda mutação de Session e SetRecord
- ✅ Analytics: "quantas séries por dia/semana?"
- ✅ Debugging: "o que aconteceu com a sessão X?"
- ✅ Compliance: rastreabilidade completa

**Pattern de uso**:
```go
// Dentro do use case, após mutação bem-sucedida
func (uc *RecordSetUseCase) Execute(...) error {
    // 1. Validar + persistir SetRecord
    setRecord := &SetRecord{...}
    if err := uc.repo.CreateSet(ctx, setRecord); err != nil {
        return err
    }
    
    // 2. Registrar no audit log
    auditEntry := &AuditLog{
        UserID:     userID,
        EntityType: "set_record",
        EntityID:   setRecord.ID,
        Action:     "created",
        ActionData: json.Marshal(setRecord),
        OccurredAt: time.Now(),
    }
    uc.auditRepo.Append(ctx, auditEntry) // fire-and-forget ou tx
    
    return nil
}
```

---

### Endpoints principais

| Método | Path | Auth | Descrição |
|--------|------|------|-----------|
| POST | `/auth/register` | ❌ | Criar usuário |
| POST | `/auth/login` | ❌ | Autenticar |
| POST | `/auth/refresh` | ❌ | Renovar token |
| POST | `/auth/logout` | ✅ | Revogar token |
| GET | `/dashboard` | ✅ | Dados agregados (stats, recent, active session) |
| GET | `/workouts` | ✅ | Listar workouts |
| GET | `/workouts/:id` | ✅ | Detalhes do workout |
| POST | `/sessions` | ✅ | Iniciar sessão |
| POST | `/sessions/:id/sets` | ✅ | Registrar série |
| PATCH | `/sessions/:id/finish` | ✅ | Finalizar sessão |
| PATCH | `/sessions/:id/abandon` | ✅ | Abandonar sessão |

**Wrapper padrão**:
- Success: `{ "data": {...}, "meta": {...} }`
- Error: `{ "code": "ERROR_CODE", "message": "...", "details": {...} }`

---

### Integrações

**MVP (síncrono)**:
- Clientes (mobile/web) → HTTP REST API → Use Cases → PostgreSQL
- Autenticação via JWT Bearer (stateless, agnóstico de client)

**Pós-MVP**:
- **Eventos assíncronos**: Kafka/SNS para `session.completed`, `workout.created` (analytics, notificações)
- **APIs de terceiros**: integração com wearables (Apple Health, Google Fit)
- **GraphQL** (opcional): para clientes web que precisam de queries flexíveis
- **gRPC** (opcional): para integrações server-to-server de alto desempenho

---

## 4) Segurança & Governança

### Autenticação
- **Senha**: bcrypt cost 12
- **Access Token**: JWT, 1h de validade, HS256
- **Refresh Token**: 30 dias, hasheado no DB

### Autorização
- Validação obrigatória de `userID` em todos os use cases
- Queries sempre filtram por `user_id`

### Validação de Input
```go
type RecordSetRequest struct {
    ExerciseID uuid.UUID `json:"exerciseId" validate:"required,uuid"`
    SetNumber  int       `json:"setNumber" validate:"required,min=1,max=20"`
    Weight     float64   `json:"weight" validate:"required,min=0,max=500"`
    Reps       int       `json:"reps" validate:"required,min=0,max=100"`
    Status     string    `json:"status" validate:"required,oneof=completed skipped"`
}
```

### Rate Limiting
- `/auth/register`: 10 req/min por IP
- `/auth/login`: 20 req/min por IP
- Endpoints autenticados: 100 req/min por `user_id`
- `/sessions/:id/sets`: 500 req/min (permite registrar sets rapidamente)

### Secrets
- `JWT_SECRET` (256 bits via secrets manager)
- `DATABASE_URL` via env vars
- Nunca commitar secrets

---

## 5) Riscos e Trade-offs

### Riscos principais

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| SetRecord duplicado (client retry) | Alta | Médio | UNIQUE constraint `(session_id, exercise_id, set_number)` |
| Sessão ativa duplicada | Média | Alto | Validação no use case + constraint `UNIQUE(user_id)` WHERE `status='active'` |
| JWT secret vazado | Baixa | Crítico | Rotação automática, TTL curto (1h) |
| N+1 query no /dashboard | Média | Médio | SQLC query com JOINs |
| Lock de tabela em migration | Média | Alto | Migrations expand/contract |

### Trade-offs

**✅ CRUD + Audit Log vs Event Sourcing**
- ✅ **Vantagens**: simplicidade, tempo de implementação 50% menor, padrão conhecido
- ⚠️ **Desvantagens**: sem replay de eventos, auditoria menos granular que ES
- ⚙️ **Mitigação**: audit log estruturado permite analytics e debugging

**✅ Monolito modular vs Microserviços**
- ✅ **Vantagens**: deploy simples, transações ACID, debug fácil
- ⚠️ **Desvantagens**: acoplamento entre features
- ⚙️ **Revisitar**: quando > 10k usuários ativos

**✅ JWT stateless vs Session-based**
- ✅ **Vantagens**: zero latência de lookup
- ⚠️ **Desvantagens**: impossível revogar antes de expirar
- ⚙️ **Mitigação**: TTL curto (1h), refresh token revogável

---

## 6) Observabilidade

### Logs (estruturados, JSON via zerolog)
```go
log.Info().
    Str("method", "POST").
    Str("path", "/sessions/123/sets").
    Str("user_id", userID).
    Int("status", 201).
    Dur("duration_ms", duration).
    Msg("http_request")
```

**Regras**:
- ✅ Log de todas as requests HTTP
- ✅ Log de erros de domínio
- ❌ Nunca logar senhas, tokens, PII

### Métricas (Prometheus)
- `http_requests_total{method, path, status}`
- `http_request_duration_seconds{method, path}`
- `db_query_duration_seconds{query}`
- `active_sessions_total` (gauge)
- `workout_sessions_completed_total` (counter)

### Tracing (OpenTelemetry, opcional MVP)
- Trace de requests HTTP
- Span de queries DB

---

## 7) Dependências

### Infra
- **PostgreSQL 15+**: obrigatório desde dia 1
- **Redis** (opcional MVP): rate limiting (pode usar in-memory)
- **Prometheus + Grafana**: métricas (Docker Compose)

### Bibliotecas Go
- `go.uber.org/fx` — DI
- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/sqlc-dev/sqlc` — SQL codegen
- `github.com/golang-jwt/jwt/v5` — JWT
- `golang.org/x/crypto/bcrypt` — hashing
- `github.com/go-playground/validator/v10` — validação
- `github.com/rs/zerolog` — logging
- `github.com/prometheus/client_golang` — métricas

---

## 8) Recomendações para Plan

### Tasks prioritárias (ordem de implementação)

**🔴 Sprint 1 (2 semanas) — Fundação**
1. ✅ Criar migrations SQL (users, workouts, exercises, sessions, set_records, refresh_tokens, **audit_log**)
2. ✅ Docker Compose (PostgreSQL + app)
3. ✅ Entidades de domínio + constants (enums, validações, asset defaults)
4. ✅ Feature AUTH completa (register, login, refresh, logout) + testes
5. ✅ Feature WORKOUTS básica (list, get) + seed data

**🟡 Sprint 2 (2-3 semanas) — Core do Produto**
6. ✅ Feature SESSIONS completa:
   - Use cases: StartSession, RecordSet, FinishSession, AbandonSession
   - Validações: sessão única ativa, set number sequencial, ownership
   - **Audit log** em todas as mutações
   - Testes de cenários complexos (duplicação, concorrência)
7. ✅ Feature DASHBOARD (agregação de dados e estatísticas)
   - Endpoint genérico que serve dados agnósticos de cliente
   - Queries otimizadas com aggregations
8. ✅ Observabilidade básica (logs, métricas HTTP/DB)

**🟢 Sprint 3 (1 semana) — Qualidade**
9. ✅ Rate limiting
10. ✅ Testes de integração (cobertura > 70%)
11. ✅ Validação de compatibilidade com OpenAPI
12. ✅ Documentação Swagger UI

---

## 9) Checklist de Implementação

### Funcionalidades

- [x] POST /auth/register
- [x] POST /auth/login
- [x] POST /auth/refresh
- [x] POST /auth/logout
- [ ] GET /workouts
- [ ] GET /workouts/:id
- [ ] POST /sessions
- [ ] POST /sessions/:id/sets
- [ ] PATCH /sessions/:id/finish
- [ ] PATCH /sessions/:id/abandon
- [ ] GET /dashboard

### Infraestrutura

- [x] Migration: users
- [x] Migration: workouts
- [x] Migration: exercises
- [x] Migration: sessions
- [x] Migration: set_records
- [x] Migration: refresh_tokens
- [x] Migration: audit_log
- [x] Docker Compose (PostgreSQL)
- [ ] Seed data (workouts)
- [ ] Rate limiting
- [x] JWT middleware
- [ ] Audit log em mutações
- [ ] Logs estruturados (zerolog)
- [ ] Métricas Prometheus
- [x] Healthcheck /health
- [ ] Testes cobertura > 70%

### Produção

- [x] Migrations aplicadas com sucesso
- [ ] Índices criados: `user_id`, `workout_id`, `session_id`, `(user_id, occurred_at)` em audit_log
- [ ] Constraints: UNIQUE, FK, CHECK configurados
- [ ] JWT_SECRET via secrets manager
- [ ] Rate limiting habilitado
- [ ] Logs sem PII
- [ ] Métricas exportadas para Prometheus
- [x] Healthcheck `/health` respondendo
- [ ] Testes cobertura > 70%
- [ ] Teste de carga: 100 req/s por 1 min sem erro
- [ ] Rollback plan documentado

---

**Documento gerado em**: 2026-02-23  
**Versão**: 3.2 (revisado)  
**Status**: ✅ Decisão tomada — pronto para implementação  
**Próxima revisão**: após Sprint 1

**Changelog v3.2 (revisão)**:
- 🔧 Corrigido: contagem de endpoints (11, não 13)
- 🔧 Corrigido: referência "home" → "dashboard" na seção 1
- 🔧 Consolidado: checklist de produção dentro da seção 9
- 📐 Estrutura: documento mais limpo e consistente

**Changelog v3.1 (platform-centric)**:
- 🔄 **Mudança de escopo**: BFF mobile-only → Backend Platform multi-client
- 🔄 Feature "home" renomeada para "dashboard" (client-agnostic)
- ➕ Adicionado: suporte explícito para múltiplos clientes (mobile, web, integrações)
- ➕ Adicionado: visão de evolução (GraphQL, gRPC, APIs externas)
- 📐 **Arquitetura**: API RESTful genérica, não acoplada a layouts de telas

**Changelog v3.0 (simplified)**:
- ✅ **Decisão final**: CRUD + Audit Log (Event Sourcing removido)
- ➖ Removido: 700+ linhas de modelagem ES (eventos, aggregates, snapshots, read models)
- ➖ Removido: seções de trade-off ES vs CRUD (decisão já tomada)
- ➖ Removido: detalhamento excessivo de riscos de migração
- ➕ Adicionado: modelagem de Audit Log obrigatório
- ➕ Adicionado: pattern de uso do audit log nos use cases
- 📉 **Redução de tamanho**: 1222 → ~400 linhas (67% menor)
- 🎯 **Foco**: decisões essenciais, modelo de dados claro, próximos passos acionáveis
