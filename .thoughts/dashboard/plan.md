# Plan — Dashboard

## 1) Inputs usados

### Documentação de Research
- `.thoughts/mvp-userflow/api-contract.yaml` — contrato OpenAPI completo (endpoint `/home` renomeado para `/dashboard`)
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — arquitetura CRUD + Audit Log, decisões
- `.thoughts/mvp-userflow/bff-aggregation-strategy.md` — **decisão confirmada**: agregação no handler HTTP com goroutines paralelas

### Estado atual do repositório (AS-IS)
- **Entidades**: `User`, `Workout`, `Session` implementadas em `internal/kinetria/domain/entities/`
- **Ports**: apenas `UserRepository` e `RefreshTokenRepository` existem (faltam: `WorkoutRepository`, `SessionRepository`)
- **HTTP**: `AuthHandler` implementado com helpers `writeSuccess` e `writeError`
- **Auth**: JWT middleware já funcional
- **Migrations**: todas as tabelas criadas (users, workouts, sessions, etc.)

### Dependências externas
- Features **workouts** e **sessions** estão planejadas (`.thoughts/workouts/`, `.thoughts/sessions/`) mas **NÃO implementadas ainda**
- Dashboard precisa que pelo menos os **repositories** e **queries SQLC** dessas features sejam implementados primeiro

---

## 2) AS-IS (resumo)

### ✅ O que existe hoje

**Entidades (domain/entities/)**:
```go
type User struct {
    ID              uuid.UUID
    Email           string
    Name            string
    PasswordHash    string
    ProfileImageURL string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type Workout struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Name        string
    Description string
    Type        string    // "FORÇA", "CARDIO", etc.
    Intensity   string    // "Alta", "Média", "Baixa"
    Duration    int       // minutos estimados
    ImageURL    string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Session struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    WorkoutID  uuid.UUID
    Status     string      // "active", "completed", "abandoned"
    Notes      string
    StartedAt  time.Time
    FinishedAt *time.Time  // null se ainda ativa
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**Repositories existentes**:
- ✅ `UserRepository` (GetByID, GetByEmail, Create)
- ✅ `RefreshTokenRepository`
- ❌ `WorkoutRepository` — **NÃO existe ainda**
- ❌ `SessionRepository` — **NÃO existe ainda**

**HTTP Layer**:
- ✅ `AuthHandler` com helpers `writeSuccess(w, status, data)` e `writeError(w, status, code, message)`
- ✅ JWT middleware funcional (`gateways/auth/jwt_manager.go`)
- ✅ `ServiceRouter` registra rotas em `/api/v1`

**Database**:
- ✅ PostgreSQL com SQLC
- ✅ Migrations todas aplicadas (users, workouts, sessions, set_records, audit_log)

---

## 3) TO-BE (proposta)

### 🎯 Endpoint `/dashboard`

**Path**: `GET /api/v1/dashboard`  
**Auth**: Bearer JWT (obrigatório)  
**Response schema**:

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "string",
      "email": "string",
      "profileImageUrl": "string|null"
    },
    "todayWorkout": {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "type": "string",
      "intensity": "string",
      "duration": int,
      "imageUrl": "string|null"
    } | null,
    "weekProgress": [
      {
        "day": "S|T|Q|Q|S|S|D",  // pt-BR abreviado
        "date": "2026-02-17",
        "status": "completed|missed|future"
      }
      // ... 7 itens
    ],
    "stats": {
      "calories": int,           // total semana
      "totalTimeMinutes": int    // total semana
    }
  }
}
```

---

### 📐 Arquitetura (Agregação no Handler)

Conforme decisão em `bff-aggregation-strategy.md`:

```
┌──────────────────────────────────────────────┐
│  DashboardHandler (gateways/http/)           │
│  ┌────────────────────────────────────┐      │
│  │  GET /api/v1/dashboard             │      │
│  │  ┌──────────────────────────┐      │      │
│  │  │ 1. Extract userID (JWT)  │      │      │
│  │  │ 2. Parallel calls:       │      │      │
│  │  │    - GetUserProfileUC    │      │      │
│  │  │    - GetTodayWorkoutUC   │      │      │
│  │  │    - GetWeekProgressUC   │      │      │
│  │  │    - GetWeekStatsUC      │      │      │
│  │  │ 3. Aggregate into DTO    │      │      │
│  │  │ 4. writeSuccess()        │      │      │
│  │  └──────────────────────────┘      │      │
│  └────────────────────────────────────┘      │
└──────────────────────────────────────────────┘
           ↓        ↓        ↓        ↓
    ┌──────────┬──────────┬──────────┬──────────┐
    │ GetUser  │ GetToday │ GetWeek  │ GetWeek  │  ← Use Cases (domain)
    │ ProfileUC│ WorkoutUC│ Progress │  StatsUC │
    │          │          │    UC    │          │
    └──────────┴──────────┴──────────┴──────────┘
           ↓        ↓        ↓        ↓
    ┌──────────┬──────────┬──────────┬──────────┐
    │   User   │ Workout  │ Session  │ Session  │  ← Repositories
    │   Repo   │   Repo   │   Repo   │   Repo   │
    └──────────┴──────────┴──────────┴──────────┘
```

**Decisão confirmada**: 
- ✅ Use cases **atômicos** no domain (reutilizáveis por múltiplos clientes)
- ✅ Agregação **paralela** no handler HTTP (melhor performance)
- ✅ DTOs específicos de cliente apenas no handler

---

### 🧩 Use Cases necessários

#### 1. `GetUserProfileUC`
**Input**: `{ UserID uuid.UUID }`  
**Output**: `{ ID, Name, Email, ProfileImageURL }`  
**Repository**: `UserRepository.GetByID()`

#### 2. `GetTodayWorkoutUC`
**Input**: `{ UserID uuid.UUID }`  
**Output**: `Workout | nil`  
**Lógica**:
- Sem sistema de agendamento no MVP → retornar o **primeiro workout ativo do usuário**
- Se não houver workouts → retornar `nil`
- Alternativa (decisão futura): retornar workout da sessão ativa, se houver

**Repository**: `WorkoutRepository.GetFirstByUserID(userID)`

**Query SQLC**:
```sql
-- name: GetFirstWorkoutByUserID :one
SELECT id, user_id, name, description, type, intensity, duration, image_url, created_at, updated_at
FROM workouts
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT 1;
```

#### 3. `GetWeekProgressUC`
**Input**: `{ UserID uuid.UUID }`  
**Output**: `[]DayProgress` (7 itens, últimos 7 dias incluindo hoje)

**DayProgress**:
```go
type DayProgress struct {
    Day    string  // "S", "T", "Q", "Q", "S", "S", "D"
    Date   string  // "2026-02-17"
    Status string  // "completed", "missed", "future"
}
```

**Lógica**:
1. Calcular últimos 7 dias (hoje - 6 dias até hoje)
2. Para cada dia:
   - Se dia > hoje → `"future"`
   - Se dia ≤ hoje e existe sessão `completed` naquele dia → `"completed"`
   - Se dia ≤ hoje e NÃO existe sessão → `"missed"`

**Repository**: `SessionRepository.GetCompletedSessionsByUserAndDateRange(userID, startDate, endDate)`

**Query SQLC**:
```sql
-- name: GetCompletedSessionsByDateRange :many
SELECT id, user_id, workout_id, status, started_at, finished_at
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND DATE(started_at) BETWEEN $2 AND $3
ORDER BY started_at DESC;
```

#### 4. `GetWeekStatsUC`
**Input**: `{ UserID uuid.UUID }`  
**Output**: `{ Calories int, TotalTimeMinutes int }`

**Lógica**:
1. Buscar todas as sessões `completed` dos últimos 7 dias
2. Somar `duration` de cada sessão (calculado: `finished_at - started_at`)
3. **Calorias estimadas** (sem sensor no MVP): `totalTimeMinutes * 7 kcal/min` (média ACSM para exercício moderado)

**Repository**: `SessionRepository.GetCompletedSessionsByUserAndDateRange()` (mesma query do weekProgress)

**Alternativa (decisão)**: usar `workouts.duration` como fallback se `finished_at` for `null` (não deve acontecer para `status=completed`, mas boa prática)

---

### 🗄️ Repositories necessários

#### `WorkoutRepository` (novo)
```go
type WorkoutRepository interface {
    GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error)
    // Outros métodos serão adicionados pela feature workouts
}
```

#### `SessionRepository` (novo)
```go
type SessionRepository interface {
    GetCompletedSessionsByUserAndDateRange(
        ctx context.Context, 
        userID uuid.UUID, 
        startDate time.Time, 
        endDate time.Time,
    ) ([]entities.Session, error)
    // Outros métodos serão adicionados pela feature sessions
}
```

---

### 📊 Queries SQLC necessárias

#### `queries/workouts.sql` (novo)
```sql
-- name: GetFirstWorkoutByUserID :one
SELECT id, user_id, name, description, type, intensity, duration, image_url, created_at, updated_at
FROM workouts
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT 1;
```

#### `queries/sessions.sql` (novo)
```sql
-- name: GetCompletedSessionsByDateRange :many
SELECT id, user_id, workout_id, status, started_at, finished_at, notes, created_at, updated_at
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND DATE(started_at) BETWEEN $2 AND $3
ORDER BY started_at DESC;
```

---

### 🔀 Agregação Paralela (Handler)

```go
// gateways/http/handler_dashboard.go
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := extractUserIDFromContext(ctx) // middleware JWT injeta userID

    // Estrutura para coletar resultados
    type result struct {
        user         *GetUserProfileOutput
        todayWorkout *GetTodayWorkoutOutput
        weekProgress *GetWeekProgressOutput
        weekStats    *GetWeekStatsOutput
        err          error
    }

    // Canal para sincronizar goroutines
    ch := make(chan result, 4)

    // 1. GetUserProfile
    go func() {
        out, err := h.getUserProfileUC.Execute(ctx, GetUserProfileInput{UserID: userID})
        ch <- result{user: &out, err: err}
    }()

    // 2. GetTodayWorkout
    go func() {
        out, err := h.getTodayWorkoutUC.Execute(ctx, GetTodayWorkoutInput{UserID: userID})
        ch <- result{todayWorkout: &out, err: err}
    }()

    // 3. GetWeekProgress
    go func() {
        out, err := h.getWeekProgressUC.Execute(ctx, GetWeekProgressInput{UserID: userID})
        ch <- result{weekProgress: &out, err: err}
    }()

    // 4. GetWeekStats
    go func() {
        out, err := h.getWeekStatsUC.Execute(ctx, GetWeekStatsInput{UserID: userID})
        ch <- result{weekStats: &out, err: err}
    }()

    // Coletar resultados
    var res result
    for i := 0; i < 4; i++ {
        r := <-ch
        if r.err != nil {
            writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load dashboard data")
            return
        }
        // Merge
        if r.user != nil { res.user = r.user }
        if r.todayWorkout != nil { res.todayWorkout = r.todayWorkout }
        if r.weekProgress != nil { res.weekProgress = r.weekProgress }
        if r.weekStats != nil { res.weekStats = r.weekStats }
    }

    // Montar DTO
    response := DashboardResponseDTO{
        User: UserProfileDTO{
            ID:              res.user.ID.String(),
            Name:            res.user.Name,
            Email:           res.user.Email,
            ProfileImageURL: res.user.ProfileImageURL,
        },
        TodayWorkout: mapTodayWorkoutToDTO(res.todayWorkout),
        WeekProgress: mapWeekProgressToDTO(res.weekProgress),
        Stats: UserStatsDTO{
            Calories:          res.weekStats.Calories,
            TotalTimeMinutes:  res.weekStats.TotalTimeMinutes,
        },
    }

    writeSuccess(w, http.StatusOK, response)
}
```

---

## 4) Decisões e Assunções

### ✅ Decisões confirmadas

| # | Decisão | Justificativa |
|---|---------|---------------|
| 1 | **Agregação no handler HTTP** (não no domain) | Seguir `bff-aggregation-strategy.md`: domain agnóstico a clientes, use cases reutilizáveis |
| 2 | **Agregação paralela** com goroutines | Reduzir latência total (4 queries paralelas vs sequenciais) |
| 3 | **"Today's workout" = primeiro workout do usuário** | Sem agendamento no MVP; alternativa: usar workout da sessão ativa se houver |
| 4 | **Calorias estimadas** (7 kcal/min) | Sem sensor/wearable no MVP; valor baseado em ACSM guidelines (exercício moderado) |
| 5 | **WeekProgress: últimos 7 dias** (hoje inclusive) | Incluir "hoje" permite mostrar progresso do dia atual |
| 6 | **Status "future"** para dias > hoje | Evitar mostrar "missed" para dias que ainda não aconteceram |
| 7 | **Usar `DATE(started_at)`** na query de sessões | Sessão iniciada às 23:55 e terminada às 00:10 = mesmo dia (pela data de início) |

### 🤔 Assunções

| # | Assunção | Risco se falso |
|---|----------|----------------|
| 1 | Sessões `completed` sempre têm `finished_at != null` | Se null, cálculo de duration vai falhar → usar `workouts.duration` como fallback |
| 2 | Usuário sempre tem pelo menos 1 workout | Se não, `todayWorkout = null` (spec permite) |
| 3 | Middleware JWT injeta `userID` no context | Se não houver, retornar 401 Unauthorized |
| 4 | Queries SQLC retornam slice vazio (não erro) quando não há resultados | Se retornar erro, tratar como caso válido (usuário sem dados) |

### 🔄 Alternativas consideradas

| Decisão | Alternativa descartada | Por que descartou |
|---------|------------------------|-------------------|
| "Today's workout" = primeiro workout | Retornar workout da sessão ativa | Mais complexo (precisa verificar se há sessão ativa); deixar para feature sessions |
| Cálculo de calorias: estimativa fixa | Usar METs específicos por `workout.type` | Mais complexo; estimativa genérica suficiente para MVP |
| WeekProgress: 7 dias fixos | Semana corrente (seg-dom) | Menos intuitivo para usuário; preferir "últimos 7 dias" |

---

## 5) Riscos / Edge Cases

### ⚠️ Riscos de implementação

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| **N+1 query** em weekProgress | Performance ruim se consultar sessions dia a dia | ✅ Usar query única com `BETWEEN startDate AND endDate` |
| **Sessão ativa no momento** | `finished_at = null` → cálculo de duration falha | ✅ Filtrar apenas `status = 'completed'` ou usar `workouts.duration` como fallback |
| **Timezone mismatch** | Servidor e cliente em timezones diferentes → "hoje" diferente | ✅ Usar `DATE(started_at)` no servidor e documentar que datas são UTC |
| **Usuário sem dados** | Primeira vez no app → arrays vazios ou nil | ✅ Tratar como caso válido: `todayWorkout = null`, `weekProgress` vazio, `stats` zerados |
| **Error em 1 de 4 goroutines** | Agregação paralela falha totalmente | ✅ Se qualquer goroutine falhar, retornar 500 (fail-fast) |
| **Goroutine leak** | Se ctx cancelar, goroutines continuam rodando | ✅ Passar `ctx` para todos os use cases (eles cancelam queries automaticamente) |

### 🧪 Edge cases a testar

| Caso | Comportamento esperado |
|------|------------------------|
| Usuário sem workouts | `todayWorkout = null` |
| Usuário sem sessões na semana | `weekProgress = 7 dias com "missed"/"future"`, `stats = {0, 0}` |
| Usuário com sessão ativa (não completed) | Não conta para weekProgress nem stats |
| Usuário com sessão `abandoned` | Não conta para weekProgress nem stats |
| Dia = hoje e há sessão completed | `status = "completed"` |
| Dia = hoje e NÃO há sessão | `status = "missed"` |
| Dia > hoje | `status = "future"` |
| Sessão com `finished_at = null` mas `status = completed` | **Bug** → logar warning e usar `workouts.duration` como fallback |

---

## 6) Rollout / Compatibilidade

### 📦 Dependências (ordem de implementação)

```
1. ✅ Auth (JWT middleware) → já implementado
2. ⏳ Workouts: implementar repository + queries SQLC mínimas
3. ⏳ Sessions: implementar repository + queries SQLC mínimas
4. ⏳ Dashboard: implementar use cases + handler + agregação
```

**Bloqueios**:
- Dashboard **bloqueia** se `WorkoutRepository` e `SessionRepository` não existirem
- Se features `workouts` e `sessions` forem implementadas primeiro (completas), dashboard só precisa **reusar** os repositories

**Estratégia**:
- Opção A (recomendado): implementar dashboard **após** workouts e sessions completas
- Opção B: implementar **stubs** dos repositories com queries mínimas para dashboard funcionar primeiro

### 🔄 Evolução futura

| Mudança futura | Impacto no dashboard |
|----------------|----------------------|
| Adicionar agendamento de workouts | Alterar `GetTodayWorkoutUC` para retornar workout agendado para hoje |
| Integrar wearable (calorias reais) | Substituir estimativa por dados do sensor na tabela `sessions` |
| Adicionar GraphQL | Reusar os mesmos use cases (apenas criar resolver GraphQL) |
| Dashboard diferente para coach vs atleta | Criar `handler_dashboard_coach.go` separado (mesmos use cases) |
| Cache de weekProgress/stats | Adicionar Redis entre handler e use cases (não altera domain) |

### 🧩 Compatibilidade com outros módulos

| Módulo | Compatibilidade |
|--------|-----------------|
| Auth | ✅ Reutiliza JWT middleware existente |
| Workouts | ✅ Consome `WorkoutRepository` (será implementado por workouts) |
| Sessions | ✅ Consome `SessionRepository` (será implementado por sessions) |
| Audit Log | ⚠️ Dashboard é read-only → não gera audit log (decisão: logar apenas writes) |

---

## 7) Observabilidade

### 📊 Métricas

| Métrica | Descrição |
|---------|-----------|
| `dashboard.load_duration_ms` | Latência total do endpoint (agregação paralela) |
| `dashboard.user_profile.duration_ms` | Tempo do use case GetUserProfile |
| `dashboard.today_workout.duration_ms` | Tempo do use case GetTodayWorkout |
| `dashboard.week_progress.duration_ms` | Tempo do use case GetWeekProgress |
| `dashboard.week_stats.duration_ms` | Tempo do use case GetWeekStats |

### 🔍 Tracing

- **Span principal**: `GET /dashboard` (handler)
- **Spans filhos**: cada use case (propagam `ctx` automaticamente)
- **Atributos**:
  - `user.id`
  - `dashboard.today_workout.found` (bool)
  - `dashboard.week_progress.days_completed` (int)

### 📝 Logs estruturados

```go
// Exemplo de log no handler
log.Info().
    Str("user_id", userID.String()).
    Bool("today_workout_found", res.todayWorkout != nil).
    Int("week_days_completed", countCompletedDays(res.weekProgress)).
    Msg("Dashboard loaded successfully")
```

---

## 8) Testes

### 🧪 Estratégia de testes

| Nível | O que testar |
|-------|--------------|
| **Use cases (unitário)** | Testar cada use case isolado com mocks de repositories |
| **Handler (integração)** | Testar agregação paralela com mocks dos use cases |
| **Queries SQLC (integração)** | Testar queries com banco real (testcontainers) |
| **E2E** | Testar endpoint completo com banco real + JWT válido |

### ✅ Critérios de aceite

- [ ] Endpoint retorna 200 com dados válidos para usuário autenticado
- [ ] Endpoint retorna 401 se JWT inválido/ausente
- [ ] `todayWorkout = null` se usuário não tem workouts
- [ ] `weekProgress` tem exatamente 7 itens (últimos 7 dias)
- [ ] `weekProgress[today].status = "completed"` se há sessão completed hoje
- [ ] `weekProgress[today].status = "missed"` se NÃO há sessão hoje
- [ ] `weekProgress[future].status = "future"` para dias > hoje
- [ ] `stats.calories` calculado corretamente (total duration * 7)
- [ ] `stats.totalTimeMinutes` soma duration de todas as sessões da semana
- [ ] Agregação paralela executa em < 500ms (4 queries)
- [ ] Logs estruturados registram cada chamada com user_id

---

## 9) Próximos passos

### ✅ Backlog gerado

Ver `tasks.md` para lista completa de tarefas executáveis.

### 📋 Checklist de validação

Antes de implementar:
- [ ] Validar schema `DashboardData` no contrato OpenAPI
- [ ] Confirmar se `workouts.duration` é sempre > 0 (ou permitir 0?)
- [ ] Confirmar timezone padrão do servidor (UTC recomendado)
- [ ] Confirmar se `writeSuccess` e `writeError` são suficientes (ou criar DTOs específicos?)
- [ ] Revisar estimativa de calorias (7 kcal/min) com stakeholders

### 🚀 Go-live

- Implementar tasks em ordem (T01 → T12)
- Rodar testes após cada task
- Deploy em staging → validar manualmente → deploy produção
- Monitorar métricas de latência e taxa de erro nos primeiros dias
