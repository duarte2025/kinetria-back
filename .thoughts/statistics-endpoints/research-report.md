# 🔎 Research Report — Statistics Endpoints

## 1) Task Summary

### O que é
Implementar 4 endpoints de estatísticas do usuário:
- **GET /api/v1/stats/overview** — Visão geral (total workouts, volume, tempo, streak)
- **GET /api/v1/stats/progression** — Progressão ao longo do tempo (gráfico de volume/força)
- **GET /api/v1/stats/personal-records** — Lista de recordes pessoais por exercício
- **GET /api/v1/stats/frequency** — Heatmap de frequência de treinos (365 dias)

### O que não é (fora de escopo)
- Comparação com outros usuários (ranking)
- Previsões/recomendações baseadas em ML
- Exportação de dados (CSV, PDF)
- Notificações de novos PRs

---

## 2) Clarifying Questions (para o dev)

### Regras de Negócio
1. **Personal Record:** Como definir? Maior peso para mesmo número de reps? Ou maior volume total (peso × reps × sets)?
2. **Streak:** Considerar apenas dias consecutivos? Permitir 1 dia de "folga" (ex: treinar seg-qua-sex conta como streak de 3)?
3. **Progression metric:** Qual métrica principal? Volume total (peso × reps)? Peso máximo? Reps máximas?

### Interface / Contrato
4. **Período padrão:** Se não informar `startDate`/`endDate`, usar qual período? Últimos 30 dias? Últimos 90 dias? Desde o início?
5. **Filtros em progression:** Permitir filtrar por exercício específico? Por muscle group?
6. **Paginação em personal-records:** Limitar a top 10? Ou paginar?

### Performance / NFRs
7. **Volumetria esperada:** Quantas sessions por usuário em média? Quantos set_records por session?
8. **Cache:** Stats de overview podem ser cacheadas? TTL de quanto tempo?
9. **Limite de período:** Limitar range máximo de datas (ex: 2 anos) para evitar queries lentas?

---

## 3) Facts from the Codebase

### Domínio(s) candidato(s)
- `internal/kinetria/domain/statistics/` (novo, a criar)
- `internal/kinetria/domain/dashboard/` (já existe, pode servir de referência)

### Entrypoints (cmd/)
- `cmd/kinetria/api/main.go` — Único entrypoint, usa Fx para DI

### Principais pacotes/símbolos envolvidos

**Entidades existentes:**
```go
// internal/kinetria/domain/entities/session.go
type Session struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    WorkoutID  uuid.UUID
    StartedAt  time.Time
    FinishedAt *time.Time
    Status     vos.SessionStatus // active, completed, abandoned
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// internal/kinetria/domain/entities/set_record.go
type SetRecord struct {
    ID                uuid.UUID
    SessionID         uuid.UUID
    WorkoutExerciseID uuid.UUID
    SetNumber         int
    Reps              int
    Weight            *int // em gramas
    Status            vos.SetRecordStatus // completed, skipped
    CreatedAt         time.Time
}
```

**Ports existentes:**
```go
// internal/kinetria/domain/ports/repositories.go
type SessionRepository interface {
    Create(ctx context.Context, session *entities.Session) error
    GetByID(ctx context.Context, id uuid.UUID) (*entities.Session, error)
    GetCompletedSessionsByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*entities.Session, error)
    // FALTA: Queries agregadas para stats
}

type SetRecordRepository interface {
    Create(ctx context.Context, record *entities.SetRecord) error
    FindBySessionExerciseSet(ctx context.Context, sessionID, workoutExerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error)
    // FALTA: Queries agregadas para PRs e progressão
}
```

**Gateways existentes:**
- `gateways/repositories/session_repository.go` — Implementação com SQLC
- `gateways/repositories/set_record_repository.go` — Implementação com SQLC
- `gateways/repositories/queries/sessions.sql` — Queries SQL tipadas
- `gateways/repositories/queries/set_records.sql` — Queries SQL tipadas
- `gateways/http/handler_dashboard.go` — Exemplo de agregação de dados

**Padrão identificado no Dashboard:**
```go
// internal/kinetria/domain/dashboard/uc_get_dashboard.go
// Agrega dados de múltiplas fontes (sessions, workouts, etc)
// Retorna struct com múltiplos campos calculados
```

---

## 4) Current Flow (AS-IS)

### Fluxo do Dashboard (referência)
1. **HTTP Request** → Chi router (`router.go`)
2. **Auth Middleware** → Valida JWT, extrai userID
3. **Handler** (`handler_dashboard.go`) → Extrai userID
4. **Use Case** (`uc_get_dashboard.go`) → Agrega dados:
   - Busca sessions completadas (últimos 30 dias)
   - Busca workout ativo
   - Calcula total de workouts, tempo médio, etc
5. **Repositories** → Executam queries agregadas via SQLC
6. **Response** → Handler mapeia para DTO, retorna JSON

### Queries agregadas existentes
- `GetCompletedSessionsByUserAndDateRange` — Retorna lista de sessions
- Dashboard calcula agregações em memória (Go)

### Índices existentes
- Migration 008: `CREATE INDEX idx_sessions_dashboard ON sessions(user_id, started_at DESC, status);`

---

## 5) Change Points (prováveis pontos de alteração)

### 5.1) Domain Layer

**Arquivos a criar:**
- `internal/kinetria/domain/statistics/uc_get_overview.go`
- `internal/kinetria/domain/statistics/uc_get_progression.go`
- `internal/kinetria/domain/statistics/uc_get_personal_records.go`
- `internal/kinetria/domain/statistics/uc_get_frequency.go`

**Structs de retorno (exemplos):**
```go
type OverviewStats struct {
    TotalWorkouts    int
    TotalSets        int
    TotalReps        int
    TotalVolume      int64 // em gramas
    TotalTime        int   // em minutos
    CurrentStreak    int   // dias consecutivos
    LongestStreak    int
    AveragePerWeek   float64
}

type ProgressionData struct {
    ExerciseID   uuid.UUID
    ExerciseName string
    DataPoints   []ProgressionPoint
}

type ProgressionPoint struct {
    Date   time.Time
    Value  float64 // volume, peso máximo, ou outra métrica
    Change float64 // % de mudança em relação ao anterior
}

type PersonalRecord struct {
    ExerciseID   uuid.UUID
    ExerciseName string
    Weight       int       // em gramas
    Reps         int
    Volume       int64     // peso × reps
    AchievedAt   time.Time
    PreviousBest *int      // peso anterior (para mostrar melhoria)
}

type FrequencyData struct {
    Date  time.Time
    Count int // número de workouts nesse dia
}
```

---

### 5.2) Ports

**Arquivo a modificar:**
- `internal/kinetria/domain/ports/repositories.go`

Adicionar métodos agregados:
```go
type SessionRepository interface {
    // ... métodos existentes
    
    // Stats
    GetStatsByUserAndPeriod(ctx context.Context, userID uuid.UUID, start, end time.Time) (*SessionStats, error)
    GetFrequencyByUserAndPeriod(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]FrequencyData, error)
}

type SetRecordRepository interface {
    // ... métodos existentes
    
    // Personal Records
    GetPersonalRecordsByUser(ctx context.Context, userID uuid.UUID) ([]PersonalRecord, error)
    
    // Progression
    GetProgressionByUserAndExercise(ctx context.Context, userID uuid.UUID, exerciseID *uuid.UUID, start, end time.Time) ([]ProgressionPoint, error)
}

type SessionStats struct {
    TotalWorkouts int
    TotalTime     int // minutos
}
```

---

### 5.3) Repository Layer

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/queries/sessions.sql`

Adicionar queries:
```sql
-- name: GetStatsByUserAndPeriod :one
SELECT 
    COUNT(*) as total_workouts,
    COALESCE(SUM(EXTRACT(EPOCH FROM (finished_at - started_at)) / 60), 0)::int as total_time_minutes
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND started_at >= $2
  AND started_at <= $3;

-- name: GetFrequencyByUserAndPeriod :many
SELECT 
    DATE(started_at) as date,
    COUNT(*) as count
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND started_at >= $2
  AND started_at <= $3
GROUP BY DATE(started_at)
ORDER BY date;
```

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/queries/set_records.sql`

Adicionar queries:
```sql
-- name: GetPersonalRecordsByUser :many
WITH ranked_sets AS (
    SELECT 
        we.exercise_id,
        e.name as exercise_name,
        sr.weight,
        sr.reps,
        (sr.weight * sr.reps) as volume,
        sr.created_at,
        ROW_NUMBER() OVER (
            PARTITION BY we.exercise_id 
            ORDER BY sr.weight DESC, sr.reps DESC, sr.created_at DESC
        ) as rank
    FROM set_records sr
    JOIN sessions s ON sr.session_id = s.id
    JOIN workout_exercises we ON sr.workout_exercise_id = we.id
    JOIN exercises e ON we.exercise_id = e.id
    WHERE s.user_id = $1
      AND s.status = 'completed'
      AND sr.status = 'completed'
      AND sr.weight IS NOT NULL
)
SELECT 
    exercise_id,
    exercise_name,
    weight,
    reps,
    volume,
    created_at as achieved_at
FROM ranked_sets
WHERE rank = 1
ORDER BY volume DESC;

-- name: GetProgressionByUserAndExercise :many
SELECT 
    DATE(s.started_at) as date,
    MAX(sr.weight) as max_weight,
    SUM(sr.weight * sr.reps) as total_volume
FROM set_records sr
JOIN sessions s ON sr.session_id = s.id
JOIN workout_exercises we ON sr.workout_exercise_id = we.id
WHERE s.user_id = $1
  AND s.status = 'completed'
  AND sr.status = 'completed'
  AND sr.weight IS NOT NULL
  AND s.started_at >= $2
  AND s.started_at <= $3
  AND ($4::uuid IS NULL OR we.exercise_id = $4)
GROUP BY DATE(s.started_at)
ORDER BY date;
```

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/session_repository.go`

Implementar métodos:
```go
func (r *sessionRepository) GetStatsByUserAndPeriod(ctx context.Context, userID uuid.UUID, start, end time.Time) (*ports.SessionStats, error) {
    result, err := r.queries.GetStatsByUserAndPeriod(ctx, queries.GetStatsByUserAndPeriodParams{
        UserID: userID,
        StartedAt: start,
        StartedAt_2: end,
    })
    if err != nil {
        return nil, err
    }
    
    return &ports.SessionStats{
        TotalWorkouts: int(result.TotalWorkouts),
        TotalTime:     int(result.TotalTimeMinutes),
    }, nil
}
```

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/set_record_repository.go`

Implementar métodos similares.

---

### 5.4) Use Cases

**Arquivo a criar:**
- `internal/kinetria/domain/statistics/uc_get_overview.go`

Lógica:
1. Recebe userID + período (opcional)
2. Chama `sessionRepo.GetStatsByUserAndPeriod()`
3. Chama `setRecordRepo.GetTotalSetsRepsVolume()` (nova query)
4. Calcula streak (lógica em Go):
   - Busca sessions dos últimos 365 dias
   - Ordena por data
   - Conta dias consecutivos
5. Retorna `OverviewStats`

**Arquivo a criar:**
- `internal/kinetria/domain/statistics/uc_get_progression.go`

Lógica:
1. Recebe userID + período + exerciseID (opcional)
2. Chama `setRecordRepo.GetProgressionByUserAndExercise()`
3. Calcula % de mudança entre pontos
4. Retorna `ProgressionData`

**Arquivo a criar:**
- `internal/kinetria/domain/statistics/uc_get_personal_records.go`

Lógica:
1. Recebe userID
2. Chama `setRecordRepo.GetPersonalRecordsByUser()`
3. Retorna lista de `PersonalRecord`

**Arquivo a criar:**
- `internal/kinetria/domain/statistics/uc_get_frequency.go`

Lógica:
1. Recebe userID + período (últimos 365 dias)
2. Chama `sessionRepo.GetFrequencyByUserAndPeriod()`
3. Preenche dias sem treino com count=0
4. Retorna array de 365 `FrequencyData`

---

### 5.5) HTTP Layer

**Arquivo a criar:**
- `internal/kinetria/gateways/http/handler_statistics.go`

Estrutura:
```go
type StatisticsHandler struct {
    getOverviewUC        *statistics.GetOverviewUC
    getProgressionUC     *statistics.GetProgressionUC
    getPersonalRecordsUC *statistics.GetPersonalRecordsUC
    getFrequencyUC       *statistics.GetFrequencyUC
}

// DTOs
type OverviewResponse struct {
    TotalWorkouts  int     `json:"totalWorkouts"`
    TotalSets      int     `json:"totalSets"`
    TotalReps      int     `json:"totalReps"`
    TotalVolume    int64   `json:"totalVolume"` // em gramas
    TotalTime      int     `json:"totalTime"`   // em minutos
    CurrentStreak  int     `json:"currentStreak"`
    LongestStreak  int     `json:"longestStreak"`
    AveragePerWeek float64 `json:"averagePerWeek"`
}

type ProgressionResponse struct {
    ExerciseID   string              `json:"exerciseId"`
    ExerciseName string              `json:"exerciseName"`
    DataPoints   []ProgressionPoint  `json:"dataPoints"`
}

type PersonalRecordsResponse struct {
    Records []PersonalRecordDTO `json:"records"`
}

type FrequencyResponse struct {
    Data []FrequencyDataDTO `json:"data"`
}
```

**Handlers:**
- `GET /api/v1/stats/overview?startDate=&endDate=` → `HandleGetOverview()`
- `GET /api/v1/stats/progression?startDate=&endDate=&exerciseId=` → `HandleGetProgression()`
- `GET /api/v1/stats/personal-records` → `HandleGetPersonalRecords()`
- `GET /api/v1/stats/frequency?startDate=&endDate=` → `HandleGetFrequency()`

---

### 5.6) Router

**Arquivo a modificar:**
- `internal/kinetria/gateways/http/router.go`

Adicionar rotas protegidas:
```go
r.Route("/api/v1/stats", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    
    r.Get("/overview", statsHandler.HandleGetOverview)
    r.Get("/progression", statsHandler.HandleGetProgression)
    r.Get("/personal-records", statsHandler.HandleGetPersonalRecords)
    r.Get("/frequency", statsHandler.HandleGetFrequency)
})
```

---

### 5.7) Dependency Injection

**Arquivo a modificar:**
- `cmd/kinetria/api/main.go`

Registrar use cases e handler:
```go
fx.Provide(
    // Use cases
    statistics.NewGetOverviewUC,
    statistics.NewGetProgressionUC,
    statistics.NewGetPersonalRecordsUC,
    statistics.NewGetFrequencyUC,
    
    // Handler
    fx.Annotate(
        http.NewStatisticsHandler,
        fx.As(new(http.StatisticsHandler)),
    ),
),
```

---

### 5.8) Otimizações (se necessário)

**Arquivo a criar (opcional):**
- `internal/kinetria/gateways/migrations/014_add_stats_indexes.sql`

```sql
-- Índice para queries de set_records por user
CREATE INDEX idx_set_records_user_stats 
ON set_records(session_id, workout_exercise_id, weight DESC, reps DESC);

-- Índice para queries de progression
CREATE INDEX idx_sessions_user_date 
ON sessions(user_id, started_at) 
WHERE status = 'completed';
```

---

## 6) Risks / Edge Cases

### Performance
- **Personal Records query:** JOIN de 3 tabelas + window function pode ser lento com muitos dados
- **Mitigação:** Índices compostos, limitar a top 50 PRs
- **Progression query:** Agregação por dia pode gerar muitos registros
- **Mitigação:** Limitar período máximo (ex: 2 anos), paginar se necessário
- **Frequency (365 dias):** Preencher dias vazios em Go pode ser custoso
- **Mitigação:** Fazer em SQL (generate_series) ou cache

### Cálculo de Streak
- **Lógica complexa:** Dias consecutivos vs permitir 1 dia de folga
- **Timezone:** Considerar timezone do usuário ou UTC?
- **Mitigação:** Definir regra clara, documentar

### Personal Record
- **Empates:** Mesmo peso, mesmas reps em datas diferentes
- **Critério de desempate:** Mais recente? Maior volume total da sessão?
- **Mitigação:** Usar `ORDER BY weight DESC, reps DESC, created_at DESC` (mais recente ganha)

### Dados vazios
- **Usuário novo:** Sem sessions, stats retornam zeros
- **Período sem treinos:** Progression retorna array vazio
- **Mitigação:** Retornar estrutura válida com valores zero, não erro

### Validações
- **Período inválido:** `startDate > endDate`
- **Período muito longo:** > 2 anos
- **ExerciseID inválido:** Não existe ou não pertence ao usuário
- **Mitigação:** Validar no handler, retornar 400

---

## 7) Suggested Implementation Strategy (alto nível, sem código)

### Etapa 1: Queries SQL (2h)
1. Criar queries em `sessions.sql`:
   - `GetStatsByUserAndPeriod` (count, tempo total)
   - `GetFrequencyByUserAndPeriod` (group by date)
2. Criar queries em `set_records.sql`:
   - `GetPersonalRecordsByUser` (window function)
   - `GetProgressionByUserAndExercise` (agregação por dia)
   - `GetTotalSetsRepsVolume` (para overview)
3. Rodar `make sqlc` para gerar código
4. Testar queries manualmente no psql

### Etapa 2: Repository (1h)
1. Adicionar métodos em `ports.SessionRepository` e `ports.SetRecordRepository`
2. Implementar métodos em `session_repository.go` e `set_record_repository.go`
3. Mapear resultados SQLC para structs de domínio

### Etapa 3: Use Cases (2-3h)
1. Criar `uc_get_overview.go`:
   - Agregar dados de sessions e set_records
   - Calcular streak (lógica em Go)
2. Criar `uc_get_progression.go`:
   - Buscar dados de progressão
   - Calcular % de mudança
3. Criar `uc_get_personal_records.go`:
   - Buscar PRs do repository
4. Criar `uc_get_frequency.go`:
   - Buscar frequência
   - Preencher dias vazios (0-365)

### Etapa 4: HTTP Handler (1-2h)
1. Criar `handler_statistics.go` com DTOs
2. Implementar 4 handlers:
   - Extrair query params (startDate, endDate, exerciseId)
   - Validar inputs
   - Chamar use case
   - Mapear para DTO
   - Retornar JSON

### Etapa 5: Routing e DI (15min)
1. Registrar rotas em `router.go`
2. Registrar use cases e handler em `main.go` (Fx)

### Etapa 6: Testes (2-3h)
1. Unit tests para use cases (mock repositories)
2. Integration tests para endpoints (DB real com dados de teste)
3. Testes de performance (simular 1000+ sessions)

### Etapa 7 (Opcional): Otimizações (1-2h)
1. Adicionar índices se queries forem lentas
2. Implementar cache em memória para overview (TTL 5min)
3. Limitar período máximo em queries

---

## 8) Handoff Notes to Plan

### Assunções feitas
- Personal Record = maior peso para mesmo exercício (desempate: mais reps, depois mais recente)
- Streak = dias consecutivos (sem permitir folga)
- Período padrão = últimos 30 dias (se não informar startDate/endDate)
- Frequency = últimos 365 dias, preencher dias vazios com count=0
- Progression metric = volume total (peso × reps)

### Dependências
- **Decisão de negócio:**
  - Definição exata de Personal Record (peso vs volume)
  - Regra de streak (consecutivo vs permitir folga)
  - Métrica principal de progressão (volume, peso máximo, reps)
- **Performance:**
  - Volumetria esperada (quantas sessions por usuário?)
  - Necessidade de cache (overview pode ser cacheado?)
- **Validações:**
  - Limite de período máximo (2 anos?)

### Recomendações para Plano de Testes

**Unit tests:**
- `GetOverviewUC`: calcula stats corretamente, calcula streak
- `GetProgressionUC`: calcula % de mudança, filtra por exercício
- `GetPersonalRecordsUC`: retorna PRs ordenados por volume
- `GetFrequencyUC`: preenche dias vazios

**Integration tests:**
- `GET /stats/overview`: retorna 200 com stats corretos
- `GET /stats/progression`: retorna 200 com datapoints, filtra por exercício
- `GET /stats/personal-records`: retorna 200 com PRs ordenados
- `GET /stats/frequency`: retorna 200 com 365 dias

**Edge cases:**
- Usuário sem sessions (retorna zeros)
- Período sem treinos (retorna arrays vazios)
- Período inválido (startDate > endDate, retorna 400)
- ExerciseID inválido (retorna 400)
- Empate em PR (desempate correto)

**Performance tests:**
- Simular 1000+ sessions, 10000+ set_records
- Medir tempo de resposta de cada endpoint
- Verificar se índices estão sendo usados (EXPLAIN ANALYZE)

### Próximos passos
1. Responder perguntas da seção 2
2. Criar plano detalhado com tasks granulares
3. Implementar queries SQL + repository
4. Implementar use cases
5. Implementar handlers
6. Testes + otimizações
