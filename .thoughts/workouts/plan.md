# Plan — lista-workouts

## 1) Inputs usados

### Artefatos de research
- `.thoughts/mvp-userflow/api-contract.yaml` — Contrato OpenAPI com endpoint `GET /workouts`
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — Arquitetura CRUD + Audit Log, modelo de dados, decisões
- `.thoughts/mvp-userflow/bff-aggregation-strategy.md` — Estratégia de agregação no handler (não no domain)

### Artefatos de código
- `.github/instructions/global.instructions.md` — Padrões de código e arquitetura hexagonal com fx
- Estrutura de diretórios existente: `internal/kinetria/domain/`, `internal/kinetria/gateways/`

### Dependências conhecidas
- **Foundation-infrastructure** (`.thoughts/foundation-infrastructure/`): migrations, entidades de domínio, Docker Compose
- **Feature AUTH**: autenticação JWT Bearer (middleware para extração de `userID`)

---

## 2) AS-IS (resumo)

### Estado atual do repositório
- ✅ Estrutura de diretórios hexagonal preparada
- ✅ Configuração fx e SQLC configurada (`sqlc.yaml`)
- ✅ Makefile com targets de build/lint/test
- ❌ **Sem migrations** SQL implementadas (workouts, exercises)
- ❌ **Sem entidades** de domínio implementadas (Workout, Exercise)
- ❌ **Sem ports** (interfaces) definidos
- ❌ **Sem use cases** implementados
- ❌ **Sem handlers** HTTP implementados
- ❌ **Sem repositórios** implementados

### Estado das dependências
- **Foundation-infrastructure**: planejada em `.thoughts/foundation-infrastructure/` (migrations, entidades, Docker) — **precisa ser implementada antes**
- **AUTH**: não verificada, assumimos que middleware de autenticação estará disponível (extrai `userID` do JWT)

### Gaps identificados
1. **Entidades de domínio** (Workout, Exercise) não existem → dependência de foundation-infrastructure
2. **Migrations** não criadas → dependência de foundation-infrastructure
3. **Middleware de autenticação** não verificado → assumimos disponível ou será criado em paralelo
4. **Seed data** de workouts → fora do escopo desta feature (será feature separada)

---

## 3) TO-BE (proposta)

### Escopo da feature "lista-workouts"

Implementar **apenas** o endpoint `GET /workouts` conforme contrato OpenAPI:

#### Interface HTTP

**Endpoint**: `GET /api/v1/workouts`

**Autenticação**: Bearer JWT (obrigatório)

**Query Parameters**:
- `page` (integer, default: 1, min: 1)
- `pageSize` (integer, default: 20, min: 1, max: 100)

**Response Success (200)**:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Treino de Peito e Tríceps",
      "description": "Foco em hipertrofia...",
      "type": "FORÇA",
      "intensity": "Alta",
      "duration": 45,
      "imageUrl": "https://cdn.kinetria.app/workouts/chest.jpg"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 57,
    "totalPages": 3
  }
}
```

**Response Error (401 Unauthorized)**:
```json
{
  "code": "UNAUTHORIZED",
  "message": "Invalid or expired access token."
}
```

**Response Error (422 Validation Error)**:
```json
{
  "code": "VALIDATION_ERROR",
  "message": "pageSize must be between 1 and 100"
}
```

---

### Camadas afetadas

#### 1) Domain Layer (`internal/kinetria/domain/workouts/`)

**Ports (interfaces)**:
```go
// ports/workout_repository.go
type WorkoutRepository interface {
    ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error)
}
```

**Use Case**:
```go
// uc_list_workouts.go
type ListWorkoutsInput struct {
    UserID   uuid.UUID
    Page     int
    PageSize int
}

type ListWorkoutsOutput struct {
    Workouts   []entities.Workout
    Total      int
    Page       int
    PageSize   int
    TotalPages int
}

type ListWorkoutsUC struct {
    repo ports.WorkoutRepository
}

func (uc *ListWorkoutsUC) Execute(ctx context.Context, input ListWorkoutsInput) (ListWorkoutsOutput, error)
```

**Validações**:
- `Page` >= 1 (default 1 se não informado)
- `PageSize` >= 1 e <= 100 (default 20 se não informado)
- `UserID` deve ser válido (não zero UUID)
- Cálculo de `offset = (page - 1) * pageSize`
- Cálculo de `totalPages = ceil(total / pageSize)`

---

#### 2) Gateway Layer (`internal/kinetria/gateways/`)

**Repository (SQLC)**:
```go
// gateways/repositories/workout_repository.go
type WorkoutRepository struct {
    queries *sqlc.Queries
}

func (r *WorkoutRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error) {
    // SQLC queries: ListWorkoutsByUserID + CountWorkoutsByUserID
}
```

**HTTP Handler**:
```go
// gateways/http/handler_workouts.go
type WorkoutsHandler struct {
    listWorkoutsUC *workouts.ListWorkoutsUC
    validator      *validator.Validate
}

func (h *WorkoutsHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
    // 1. Extrair userID do context (JWT middleware)
    // 2. Parse query params (page, pageSize)
    // 3. Validar params
    // 4. Chamar use case
    // 5. Mapear para DTO (WorkoutSummary)
    // 6. Responder com ApiResponse + PaginationMeta
}
```

**DTOs**:
```go
// gateways/http/dto_workouts.go
type WorkoutSummaryDTO struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Description *string `json:"description"` // nullable
    Type        *string `json:"type"`        // nullable
    Intensity   *string `json:"intensity"`   // nullable
    Duration    int     `json:"duration"`
    ImageURL    *string `json:"imageUrl"`    // nullable
}

type PaginationMetaDTO struct {
    Page       int `json:"page"`
    PageSize   int `json:"pageSize"`
    Total      int `json:"total"`
    TotalPages int `json:"totalPages"`
}

type ApiResponseDTO struct {
    Data interface{}        `json:"data"`
    Meta *PaginationMetaDTO `json:"meta,omitempty"`
}
```

**Router**:
```go
// gateways/http/router.go
func NewServiceRouter(handler *WorkoutsHandler, authMiddleware *AuthMiddleware) http.Handler {
    r := chi.NewRouter()
    
    r.Route("/api/v1", func(r chi.Router) {
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware.Require) // Extrai userID do JWT
            
            r.Get("/workouts", handler.ListWorkouts)
        })
    })
    
    return r
}
```

---

#### 3) Persistência (SQLC Queries)

**Query: ListWorkoutsByUserID**
```sql
-- name: ListWorkoutsByUserID :many
SELECT 
    id, 
    user_id, 
    name, 
    description, 
    type, 
    intensity, 
    duration, 
    image_url, 
    created_at, 
    updated_at
FROM workouts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

**Query: CountWorkoutsByUserID**
```sql
-- name: CountWorkoutsByUserID :one
SELECT COUNT(*) 
FROM workouts
WHERE user_id = $1;
```

**Nota**: As migrations da tabela `workouts` devem estar criadas pela feature foundation-infrastructure.

---

### Contratos de dados

#### Entidade de domínio (Workout)
```go
// domain/entities/workout.go
type Workout struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Name        string
    Description string     // pode ser vazio
    Type        string     // enum: "FORÇA"|"HIPERTROFIA"|"MOBILIDADE"|"CONDICIONAMENTO"
    Intensity   string     // enum: "BAIXA"|"MODERADA"|"ALTA"
    Duration    int        // minutos
    ImageURL    string     // default baseado no Type
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Observação**: Como `WorkoutSummary` não inclui a lista de exercises, **não precisamos** carregar exercises nesta query (performance otimizada).

---

### Observabilidade

**Logs estruturados** (zerolog):
```go
log.Info().
    Str("user_id", userID.String()).
    Int("page", page).
    Int("page_size", pageSize).
    Int("total", total).
    Dur("duration_ms", duration).
    Msg("list_workouts_success")
```

**Métricas** (Prometheus):
- `http_requests_total{method="GET", path="/workouts", status="200|401|422"}`
- `http_request_duration_seconds{method="GET", path="/workouts"}`
- `db_query_duration_seconds{query="ListWorkoutsByUserID"}`

**Erros a logar**:
- Falha ao parsear query params → 422
- Usuário não autenticado → 401
- Erro ao consultar DB → 500
- Validação de input falhou → 422

---

## 4) Decisões e Assunções

### Decisões arquiteturais

1. **✅ Agregação no handler, não no domain**
   - Justificativa: seguir estratégia definida em `bff-aggregation-strategy.md`
   - Use case retorna entidade de domínio (`Workout`)
   - Handler mapeia para DTO (`WorkoutSummaryDTO`)

2. **✅ WorkoutSummary não inclui exercises**
   - Justificativa: otimização de performance (menos dados trafegados)
   - Detalhes do workout (com exercises) serão obtidos em `GET /workouts/:id` (feature separada)

3. **✅ Paginação obrigatória**
   - Justificativa: prevenir queries sem limite
   - Default: `page=1, pageSize=20`
   - Máximo: `pageSize=100`

4. **✅ Ordem: created_at DESC**
   - Justificativa: workouts mais recentes aparecem primeiro
   - Possível evolução: permitir ordenação customizada via query param

5. **✅ Validação de ownership**
   - Justificativa: segurança — usuário só vê seus próprios workouts
   - Implementado via `WHERE user_id = $1` na query SQL

6. **✅ Retorno de total mesmo sem dados**
   - Justificativa: frontend precisa saber se existem workouts
   - `{ "data": [], "meta": { "total": 0, "page": 1, ... } }`

### Assunções

1. **Middleware de autenticação disponível**
   - Extrai `userID` do JWT e injeta no `context.Context`
   - Handler acessa via `userID := ctx.Value("userID").(uuid.UUID)`
   - Se não houver: retornar 401

2. **Foundation-infrastructure implementada**
   - Migrations da tabela `workouts` criadas
   - Entidade `Workout` de domínio implementada
   - Docker Compose com PostgreSQL rodando

3. **Seed data separado**
   - Esta feature não cria workouts de exemplo
   - Assumimos que haverá uma task separada de seed ou feature de criação de workouts

4. **Campos nullable no DTO**
   - `description`, `type`, `intensity`, `imageUrl` podem ser null no JSON
   - No domínio são strings vazias (go convention)
   - Mapeamento handler → DTO converte string vazia para `nil` pointer

5. **Rate limiting será implementado em feature separada**
   - Backend-architecture-report define: 100 req/min por `user_id` para endpoints autenticados
   - Esta feature não implementa rate limiting (será middleware global)
   - Assumimos que rate limiting será adicionado antes de produção

---

## 5) Riscos / Edge Cases

### Riscos de implementação

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| Tabela `workouts` não existe | Alta | Blocker | Verificar dependência foundation-infrastructure antes de começar |
| Middleware auth não disponível | Média | Blocker | Criar stub de middleware ou coordenar com feature AUTH |
| N+1 query ao expandir para exercises | Baixa | Médio | Não aplicável (WorkoutSummary não tem exercises) |
| Query sem LIMIT | Baixa | Alto | Validação obrigatória de pageSize no use case |
| Usuário sem workouts | Alta | Baixo | Retornar array vazio + total=0 (comportamento válido) |
| Paginação além do total | Média | Baixo | Permitir (retorna array vazio, page válido) |

### Edge cases a cobrir nos testes

1. **Usuário sem workouts**
   - Input: `userID` válido, sem workouts no DB
   - Output: `{ "data": [], "meta": { "total": 0, "page": 1, ... } }`

2. **Página além do total**
   - Input: `page=10` mas total de 5 workouts (1 página)
   - Output: `{ "data": [], "meta": { "page": 10, "total": 5, ... } }`

3. **PageSize maior que total**
   - Input: `pageSize=100` mas total de 3 workouts
   - Output: `{ "data": [3 workouts], "meta": { "pageSize": 100, "total": 3, "totalPages": 1 } }`

4. **Campos opcionais vazios**
   - Workout sem `description`, `type`, `intensity` no DB
   - Output: DTOs com campos `null`

5. **Parâmetros inválidos**
   - `page=0` → erro 422
   - `page=-1` → erro 422
   - `pageSize=0` → erro 422
   - `pageSize=101` → erro 422
   - `page=abc` → erro 422

6. **Usuário não autenticado**
   - Request sem header `Authorization`
   - Output: 401 Unauthorized

7. **Token JWT inválido/expirado**
   - Request com token malformado
   - Output: 401 Unauthorized

---

## 6) Rollout / Compatibilidade

### Estratégia de rollout

1. **Pré-requisitos (bloquantes)**:
   - ✅ Foundation-infrastructure implementada (migrations, entidades, Docker)
   - ✅ Feature AUTH com middleware de autenticação disponível
   - ✅ Índice `(user_id, created_at)` criado na tabela `workouts` (verificar migration)

2. **Ordem de implementação** (ver `tasks.md`):
   - T01: Criar port WorkoutRepository
   - T02: Implementar queries SQLC
   - T03: Implementar WorkoutRepository
   - T04: Implementar ListWorkoutsUC
   - T05: Implementar DTOs
   - T06: Implementar WorkoutsHandler
   - T07: Registrar rota no router
   - T08: Implementar testes unitários
   - T09: Implementar testes de integração
   - T10: Documentar API

3. **Validação de integração**:
   - Teste manual via `curl` ou Postman
   - Teste com usuário autenticado (token JWT válido)
   - Validar paginação com diferentes valores de `page` e `pageSize`
   - Validar response schema contra OpenAPI spec

### Compatibilidade

#### Backward compatibility
- **N/A**: feature nova, sem código existente

#### Forward compatibility
- ✅ DTO `WorkoutSummaryDTO` é subset de `WorkoutDTO` (futuro)
- ✅ Endpoint `/workouts` pode evoluir com query params adicionais:
  - `sort` (ex: `sort=name`, `sort=duration`)
  - `filter` (ex: `filter[type]=FORÇA`)
  - `search` (ex: `search=peito`)

#### Breaking changes (futuro)
- ⚠️ Se mudar estrutura de `PaginationMeta` → versionar API (`/api/v2`)
- ⚠️ Se adicionar campos obrigatórios no DTO → breaking change

### Database migrations
- **N/A**: não cria novas tabelas (usa tabela `workouts` da foundation-infrastructure)
- Se houver necessidade de índice adicional:
  ```sql
  CREATE INDEX idx_workouts_user_id_created_at ON workouts(user_id, created_at DESC);
  ```

---

## 7) Checklist de "Definition of Done"

Antes de considerar a feature completa:

- [ ] Todas as tasks em `tasks.md` implementadas
- [ ] Testes unitários implementados e passando (cobertura > 80% do código da feature)
- [ ] Testes de integração implementados e passando
- [ ] Cenários BDD em `test-scenarios.feature` cobertos
- [ ] Código segue `.github/instructions/global.instructions.md`
- [ ] `make lint` sem warnings
- [ ] `make test` passando
- [ ] Logs estruturados implementados (sem PII)
- [ ] Documentação da API atualizada (README ou docs/)
- [ ] Response schema validado contra OpenAPI spec
- [ ] Teste manual com token JWT válido executado
- [ ] Edge cases testados (usuário sem workouts, paginação, validações)
- [ ] PR reviewed e aprovado

---

## 8) Próximos Passos

Após implementação desta feature:

1. **Feature GET /workouts/:id** (detalhes do workout com exercises)
   - Usa mesma base (repository, entidades)
   - Adiciona join com tabela `exercises`
   - Retorna `WorkoutDTO` completo (com array de exercises)

2. **Feature seed-data** (popular workouts de exemplo)
   - Script de seed para popular DB com workouts padrão
   - Útil para testes e demo

3. **Feature dashboard** (agregação de dados)
   - Usa `ListWorkoutsUC` para obter workouts recentes
   - Agrega com outros dados (user, stats, sessions)

---

**Documento gerado em**: 2026-02-23  
**Versão**: 1.0  
**Status**: ✅ Pronto para implementação  
**Próximo artefato**: `test-scenarios.feature`
