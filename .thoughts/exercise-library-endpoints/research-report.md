# 🔎 Research Report — Exercise Library Endpoints

## 1) Task Summary

### O que é
Implementar 3 endpoints de biblioteca de exercícios:
- **GET /api/v1/exercises** — Listar exercícios com filtros (muscleGroup, equipment, difficulty, search)
- **GET /api/v1/exercises/:id** — Detalhes do exercício + estatísticas do usuário
- **GET /api/v1/exercises/:id/history** — Histórico de execução do exercício pelo usuário

### O que não é (fora de escopo)
- Criação/edição de exercícios pelo usuário (biblioteca é read-only)
- Upload de vídeos/imagens de exercícios (usar URLs mock)
- Exercícios favoritos/salvos
- Recomendações de exercícios baseadas em histórico

---

## 2) Decisions Made

### Persistência
1. **Tabela `exercises` completa?** Não. Criar migration 011 para adicionar: `description`, `instructions`, `tips`, `difficulty`, `equipment`, `video_url`.
2. **Seed de exercícios:** 30-40 exercícios mais comuns. Conteúdo genérico/público. URLs mock por enquanto.
3. **Campo `muscles`:** TEXT[] (já existe). Valores livres, enum pode vir depois.

### Interface / Contrato
4. **Filtros obrigatórios:** `muscleGroup` (essencial), `search` (essencial), `equipment` (opcional), `difficulty` (opcional).
5. **Paginação padrão:** page=1, pageSize=20, max=100 (padrão do projeto).
6. **Search:** Apenas `name` com ILIKE. Full-text search em description fica para v2.
7. **History:** Todas as execuções agrupadas por session. Mostrar todos os sets.

### Regras de Negócio
8. **User stats em GET /exercises/:id:** `lastPerformed`, `bestWeight`, `timesPerformed`, `averageWeight` (últimas 10 execuções).
9. **History ordenação:** Mais recente primeiro (DESC).
10. **History paginação:** Paginar (page=1, pageSize=20). Sem limite fixo, usuário navega todo histórico.

---

## 3) Facts from the Codebase

### Domínio(s) candidato(s)
- `internal/kinetria/domain/exercises/` (novo, a criar)

### Entrypoints (cmd/)
- `cmd/kinetria/api/main.go` — Único entrypoint, usa Fx para DI

### Principais pacotes/símbolos envolvidos

**Entidades existentes:**
```go
// internal/kinetria/domain/entities/exercise.go
type Exercise struct {
    ID           uuid.UUID
    Name         string
    ThumbnailURL *string
    Muscles      []string // ou pq.StringArray
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Entidades relacionadas:**
```go
// internal/kinetria/domain/entities/workout_exercise.go
type WorkoutExercise struct {
    ID         uuid.UUID
    WorkoutID  uuid.UUID
    ExerciseID uuid.UUID
    Sets       int
    Reps       string
    RestTime   int
    Weight     *int
    OrderIndex int
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
    Weight            *int
    Status            vos.SetRecordStatus
    CreatedAt         time.Time
}
```

**Ports existentes:**
```go
// internal/kinetria/domain/ports/repositories.go
type ExerciseRepository interface {
    ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error)
    FindWorkoutExerciseID(ctx context.Context, workoutID, exerciseID uuid.UUID) (*uuid.UUID, error)
    // FALTA: List, GetByID, GetUserStats, GetHistory
}
```

**Migrations existentes:**
- Migration 009: Refatorou exercises para biblioteca compartilhada (N:N com workouts)

```sql
-- Estrutura atual (migration 009)
CREATE TABLE exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    thumbnail_url TEXT,
    muscles TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE workout_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    sets INT NOT NULL,
    reps VARCHAR(50) NOT NULL,
    rest_time INT NOT NULL,
    weight INT,
    order_index INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Campos faltantes na tabela `exercises`:**
- `description TEXT`
- `instructions TEXT`
- `tips TEXT`
- `difficulty VARCHAR(50)` (ex: "Iniciante", "Intermediário", "Avançado")
- `equipment VARCHAR(100)` (ex: "Barra", "Halteres", "Peso corporal")
- `video_url TEXT`

---

## 4) Current Flow (AS-IS)

### Fluxo atual de exercises
- Exercises são referenciados em `workout_exercises` (N:N com workouts)
- Não há endpoints públicos para listar/buscar exercises
- Frontend usa dados mockados

### Relacionamento atual
```
exercises (1) ----< workout_exercises >---- (N) workouts
                         ↓
                   set_records (via workout_exercise_id)
```

---

## 5) Change Points (prováveis pontos de alteração)

### 5.1) Migration

**Arquivo a criar:**
- `internal/kinetria/gateways/migrations/011_expand_exercises_table.sql`

```sql
-- Adicionar campos faltantes
ALTER TABLE exercises 
ADD COLUMN description TEXT,
ADD COLUMN instructions TEXT,
ADD COLUMN tips TEXT,
ADD COLUMN difficulty VARCHAR(50),
ADD COLUMN equipment VARCHAR(100),
ADD COLUMN video_url TEXT;

-- Índices para busca e filtros
CREATE INDEX idx_exercises_name ON exercises USING gin(to_tsvector('portuguese', name));
CREATE INDEX idx_exercises_muscles ON exercises USING gin(muscles);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX idx_exercises_equipment ON exercises(equipment);
```

**Arquivo a criar (opcional):**
- `internal/kinetria/gateways/migrations/012_seed_exercises.sql`

```sql
-- Seed com exercícios base
INSERT INTO exercises (name, description, instructions, tips, difficulty, equipment, muscles, thumbnail_url, video_url) VALUES
('Supino Reto', 'Exercício composto para peito', 'Deite no banco, pegue a barra...', 'Mantenha os pés no chão', 'Intermediário', 'Barra', ARRAY['Peito', 'Tríceps', 'Ombro'], 'https://cdn.kinetria.app/exercises/bench-press.jpg', 'https://cdn.kinetria.app/videos/bench-press.mp4'),
('Agachamento Livre', 'Exercício composto para pernas', 'Posicione a barra nas costas...', 'Mantenha o core contraído', 'Intermediário', 'Barra', ARRAY['Quadríceps', 'Glúteos', 'Posterior'], 'https://cdn.kinetria.app/exercises/squat.jpg', 'https://cdn.kinetria.app/videos/squat.mp4'),
-- ... mais 48 exercícios
;
```

---

### 5.2) Domain Layer

**Arquivo a modificar:**
- `internal/kinetria/domain/entities/exercise.go`

Adicionar campos:
```go
type Exercise struct {
    ID           uuid.UUID
    Name         string
    Description  *string
    Instructions *string
    Tips         *string
    Difficulty   *string
    Equipment    *string
    ThumbnailURL *string
    VideoURL     *string
    Muscles      []string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Arquivos a criar:**
- `internal/kinetria/domain/exercises/uc_list_exercises.go`
- `internal/kinetria/domain/exercises/uc_get_exercise.go`
- `internal/kinetria/domain/exercises/uc_get_exercise_history.go`

**Structs auxiliares:**
```go
type ExerciseFilters struct {
    MuscleGroup *string
    Equipment   *string
    Difficulty  *string
    Search      *string
}

type ExerciseWithStats struct {
    Exercise        *entities.Exercise
    LastPerformed   *time.Time
    BestWeight      *int
    TimesPerformed  int
    AverageWeight   *float64
}

type ExerciseHistoryEntry struct {
    SessionID   uuid.UUID
    WorkoutName string
    PerformedAt time.Time
    Sets        []SetDetail
}

type SetDetail struct {
    SetNumber int
    Reps      int
    Weight    *int
    Status    string
}
```

---

### 5.3) Ports

**Arquivo a modificar:**
- `internal/kinetria/domain/ports/repositories.go`

Adicionar métodos:
```go
type ExerciseRepository interface {
    // Existentes
    ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error)
    FindWorkoutExerciseID(ctx context.Context, workoutID, exerciseID uuid.UUID) (*uuid.UUID, error)
    
    // Novos
    List(ctx context.Context, filters ExerciseFilters, page, pageSize int) ([]*entities.Exercise, int, error)
    GetByID(ctx context.Context, exerciseID uuid.UUID) (*entities.Exercise, error)
    GetUserStats(ctx context.Context, userID, exerciseID uuid.UUID) (*ExerciseUserStats, error)
    GetHistory(ctx context.Context, userID, exerciseID uuid.UUID, page, pageSize int) ([]*ExerciseHistoryEntry, int, error)
}

type ExerciseUserStats struct {
    LastPerformed  *time.Time
    BestWeight     *int
    TimesPerformed int
    AverageWeight  *float64
}
```

---

### 5.4) Repository Layer

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/queries/exercises.sql`

Adicionar queries:
```sql
-- name: ListExercises :many
SELECT * FROM exercises
WHERE 
    ($1::text IS NULL OR $1 = ANY(muscles))
    AND ($2::text IS NULL OR equipment = $2)
    AND ($3::text IS NULL OR difficulty = $3)
    AND ($4::text IS NULL OR name ILIKE '%' || $4 || '%')
ORDER BY name
LIMIT $5 OFFSET $6;

-- name: CountExercises :one
SELECT COUNT(*) FROM exercises
WHERE 
    ($1::text IS NULL OR $1 = ANY(muscles))
    AND ($2::text IS NULL OR equipment = $2)
    AND ($3::text IS NULL OR difficulty = $3)
    AND ($4::text IS NULL OR name ILIKE '%' || $4 || '%');

-- name: GetExerciseByID :one
SELECT * FROM exercises WHERE id = $1;

-- name: GetExerciseUserStats :one
SELECT 
    MAX(s.started_at) as last_performed,
    MAX(sr.weight) as best_weight,
    COUNT(DISTINCT s.id) as times_performed,
    AVG(sr.weight)::int as average_weight
FROM exercises e
LEFT JOIN workout_exercises we ON e.id = we.exercise_id
LEFT JOIN set_records sr ON we.id = sr.workout_exercise_id
LEFT JOIN sessions s ON sr.session_id = s.id
WHERE e.id = $1
  AND s.user_id = $2
  AND s.status = 'completed'
  AND sr.status = 'completed';

-- name: GetExerciseHistory :many
SELECT 
    s.id as session_id,
    w.name as workout_name,
    s.started_at as performed_at,
    sr.set_number,
    sr.reps,
    sr.weight,
    sr.status
FROM exercises e
JOIN workout_exercises we ON e.id = we.exercise_id
JOIN set_records sr ON we.id = sr.workout_exercise_id
JOIN sessions s ON sr.session_id = s.id
JOIN workouts w ON s.workout_id = w.id
WHERE e.id = $1
  AND s.user_id = $2
  AND s.status = 'completed'
ORDER BY s.started_at DESC, sr.set_number
LIMIT $3 OFFSET $4;

-- name: CountExerciseHistory :one
SELECT COUNT(DISTINCT s.id)
FROM exercises e
JOIN workout_exercises we ON e.id = we.exercise_id
JOIN set_records sr ON we.id = sr.workout_exercise_id
JOIN sessions s ON sr.session_id = s.id
WHERE e.id = $1
  AND s.user_id = $2
  AND s.status = 'completed';
```

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/exercise_repository.go`

Implementar métodos:
```go
func (r *exerciseRepository) List(ctx context.Context, filters ports.ExerciseFilters, page, pageSize int) ([]*entities.Exercise, int, error) {
    offset := (page - 1) * pageSize
    
    rows, err := r.queries.ListExercises(ctx, queries.ListExercisesParams{
        MuscleGroup: filters.MuscleGroup,
        Equipment:   filters.Equipment,
        Difficulty:  filters.Difficulty,
        Search:      filters.Search,
        Limit:       int32(pageSize),
        Offset:      int32(offset),
    })
    if err != nil {
        return nil, 0, err
    }
    
    total, err := r.queries.CountExercises(ctx, queries.CountExercisesParams{
        MuscleGroup: filters.MuscleGroup,
        Equipment:   filters.Equipment,
        Difficulty:  filters.Difficulty,
        Search:      filters.Search,
    })
    if err != nil {
        return nil, 0, err
    }
    
    exercises := make([]*entities.Exercise, len(rows))
    for i, row := range rows {
        exercises[i] = mapToExerciseEntity(row)
    }
    
    return exercises, int(total), nil
}

func (r *exerciseRepository) GetUserStats(ctx context.Context, userID, exerciseID uuid.UUID) (*ports.ExerciseUserStats, error) {
    stats, err := r.queries.GetExerciseUserStats(ctx, queries.GetExerciseUserStatsParams{
        ExerciseID: exerciseID,
        UserID:     userID,
    })
    if err != nil {
        return nil, err
    }
    
    return &ports.ExerciseUserStats{
        LastPerformed:  stats.LastPerformed,
        BestWeight:     stats.BestWeight,
        TimesPerformed: int(stats.TimesPerformed),
        AverageWeight:  stats.AverageWeight,
    }, nil
}
```

---

### 5.5) Use Cases

**Arquivo a criar:**
- `internal/kinetria/domain/exercises/uc_list_exercises.go`

Lógica:
1. Recebe filtros + paginação
2. Valida inputs (page >= 1, pageSize <= 100)
3. Chama `exerciseRepo.List()`
4. Retorna lista + total

**Arquivo a criar:**
- `internal/kinetria/domain/exercises/uc_get_exercise.go`

Lógica:
1. Recebe exerciseID + userID (opcional, para stats)
2. Chama `exerciseRepo.GetByID()`
3. Se userID fornecido, chama `exerciseRepo.GetUserStats()`
4. Retorna `ExerciseWithStats`

**Arquivo a criar:**
- `internal/kinetria/domain/exercises/uc_get_exercise_history.go`

Lógica:
1. Recebe userID + exerciseID + paginação
2. Valida que exercise existe
3. Chama `exerciseRepo.GetHistory()`
4. Agrupa sets por session
5. Retorna lista de `ExerciseHistoryEntry`

---

### 5.6) HTTP Layer

**Arquivo a criar:**
- `internal/kinetria/gateways/http/handler_exercises.go`

Estrutura:
```go
type ExercisesHandler struct {
    listExercisesUC        *exercises.ListExercisesUC
    getExerciseUC          *exercises.GetExerciseUC
    getExerciseHistoryUC   *exercises.GetExerciseHistoryUC
}

// DTOs
type ListExercisesResponse struct {
    Data []ExerciseDTO `json:"data"`
    Meta PaginationMeta `json:"meta"`
}

type ExerciseDTO struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Description  *string  `json:"description"`
    Instructions *string  `json:"instructions"`
    Tips         *string  `json:"tips"`
    Difficulty   *string  `json:"difficulty"`
    Equipment    *string  `json:"equipment"`
    ThumbnailURL *string  `json:"thumbnailUrl"`
    VideoURL     *string  `json:"videoUrl"`
    Muscles      []string `json:"muscles"`
}

type ExerciseDetailResponse struct {
    Data ExerciseWithStatsDTO `json:"data"`
}

type ExerciseWithStatsDTO struct {
    ExerciseDTO
    UserStats *UserStatsDTO `json:"userStats,omitempty"`
}

type UserStatsDTO struct {
    LastPerformed  *string  `json:"lastPerformed"`  // ISO 8601
    BestWeight     *int     `json:"bestWeight"`     // em gramas
    TimesPerformed int      `json:"timesPerformed"`
    AverageWeight  *float64 `json:"averageWeight"`  // em gramas
}

type ExerciseHistoryResponse struct {
    Data []HistoryEntryDTO `json:"data"`
    Meta PaginationMeta    `json:"meta"`
}

type HistoryEntryDTO struct {
    SessionID   string      `json:"sessionId"`
    WorkoutName string      `json:"workoutName"`
    PerformedAt string      `json:"performedAt"` // ISO 8601
    Sets        []SetDetail `json:"sets"`
}
```

**Handlers:**
- `GET /api/v1/exercises?muscleGroup=&equipment=&difficulty=&search=&page=&pageSize=` → `HandleListExercises()`
- `GET /api/v1/exercises/:id` → `HandleGetExercise()` (inclui user stats se autenticado)
- `GET /api/v1/exercises/:id/history?page=&pageSize=` → `HandleGetExerciseHistory()` (requer autenticação)

---

### 5.7) Router

**Arquivo a modificar:**
- `internal/kinetria/gateways/http/router.go`

Adicionar rotas:
```go
r.Route("/api/v1/exercises", func(r chi.Router) {
    // Públicas (ou autenticadas opcionalmente)
    r.Get("/", exercisesHandler.HandleListExercises)
    r.Get("/{id}", exercisesHandler.HandleGetExercise)
    
    // Requer autenticação
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        r.Get("/{id}/history", exercisesHandler.HandleGetExerciseHistory)
    })
})
```

---

### 5.8) Dependency Injection

**Arquivo a modificar:**
- `cmd/kinetria/api/main.go`

Registrar use cases e handler:
```go
fx.Provide(
    // Use cases
    exercises.NewListExercisesUC,
    exercises.NewGetExerciseUC,
    exercises.NewGetExerciseHistoryUC,
    
    // Handler
    fx.Annotate(
        http.NewExercisesHandler,
        fx.As(new(http.ExercisesHandler)),
    ),
),
```

---

## 6) Risks / Edge Cases

### Seed de Exercícios
- **Conteúdo:** Quem vai criar descrições, instruções, dicas para 50+ exercícios?
- **Imagens/vídeos:** URLs mock ou reais? Onde hospedar?
- **Mitigação:** Começar com 20-30 exercícios mais comuns, expandir depois

### Performance
- **Busca full-text:** Índice GIN em `name` pode ser lento com muitos registros
- **Mitigação:** Limitar resultados (max 100), usar paginação
- **User stats query:** JOIN de 4 tabelas pode ser lento
- **Mitigação:** Índices compostos, cache de stats (se necessário)

### Validações
- **ExerciseID inválido:** Retornar 404
- **Filtros inválidos:** Validar valores de difficulty, equipment (enum?)
- **Paginação:** Validar page >= 1, pageSize <= 100

### Dados vazios
- **Biblioteca vazia:** Se não rodar seed, endpoints retornam arrays vazios
- **User sem histórico:** Stats retornam null/zero
- **Mitigação:** Retornar estrutura válida, não erro

### Autenticação opcional
- **GET /exercises e GET /exercises/:id:** Podem ser públicos ou autenticados
- **Se autenticado:** Incluir user stats
- **Se não autenticado:** Retornar apenas dados do exercício
- **Mitigação:** Verificar JWT no handler, mas não exigir (middleware opcional)

---

## 7) Suggested Implementation Strategy (alto nível, sem código)

### Etapa 1: Migration e Seed (1-2h)
1. Criar migration `011_expand_exercises_table.sql`
2. Criar migration `012_seed_exercises.sql` (ou script Go separado)
3. Decidir: quantos exercícios? quais dados?
4. Rodar migrations

### Etapa 2: Domain e Entities (30min)
1. Atualizar `entities.Exercise` com novos campos
2. Criar structs auxiliares (`ExerciseFilters`, `ExerciseWithStats`, etc)

### Etapa 3: Repository (2h)
1. Adicionar métodos em `ports.ExerciseRepository`
2. Criar queries em `queries/exercises.sql`:
   - `ListExercises` (com filtros)
   - `CountExercises`
   - `GetExerciseByID`
   - `GetExerciseUserStats`
   - `GetExerciseHistory`
   - `CountExerciseHistory`
3. Rodar `make sqlc`
4. Implementar métodos em `exercise_repository.go`

### Etapa 4: Use Cases (1-2h)
1. Criar `uc_list_exercises.go`:
   - Valida filtros e paginação
   - Chama repository
2. Criar `uc_get_exercise.go`:
   - Busca exercise
   - Se userID fornecido, busca stats
3. Criar `uc_get_exercise_history.go`:
   - Busca histórico
   - Agrupa sets por session

### Etapa 5: HTTP Handler (1-2h)
1. Criar `handler_exercises.go` com DTOs
2. Implementar `HandleListExercises()`:
   - Extrai query params
   - Valida inputs
   - Chama use case
   - Retorna JSON com paginação
3. Implementar `HandleGetExercise()`:
   - Extrai exerciseID
   - Verifica se há JWT (opcional)
   - Chama use case
   - Retorna JSON
4. Implementar `HandleGetExerciseHistory()`:
   - Extrai userID do JWT
   - Extrai exerciseID e paginação
   - Chama use case
   - Retorna JSON

### Etapa 6: Routing e DI (15min)
1. Registrar rotas em `router.go`
2. Registrar use cases e handler em `main.go` (Fx)

### Etapa 7: Testes (2h)
1. Unit tests para use cases (mock repository)
2. Integration tests para endpoints (DB real com seed)
3. Edge cases: filtros inválidos, exercise não encontrado, user sem histórico

---

## 8) Handoff Notes to Plan

### Assunções feitas
- Biblioteca de exercícios é read-only (usuários não criam exercícios)
- GET /exercises e GET /exercises/:id são públicos (ou autenticação opcional)
- GET /exercises/:id/history requer autenticação
- Seed inicial com 30-40 exercícios mais comuns
- URLs de imagens/vídeos são mock por enquanto
- Campo `muscles` como TEXT[] com valores livres
- Search apenas em `name` (ILIKE), não em description
- History mostra todas execuções (não apenas best sets), paginado

### Dependências
- **Decisões implementadas:**
  - Migration 011: adicionar campos faltantes em `exercises`
  - Seed com 30-40 exercícios (conteúdo genérico)
  - Filtros: muscleGroup e search (essenciais), equipment e difficulty (opcionais)
  - User stats: lastPerformed, bestWeight, timesPerformed, averageWeight
  - History: todas execuções, paginado, ordenado por mais recente
- **Decisão técnica:**
  - URLs de imagens/vídeos mock (CDN/S3 fica para depois)
  - Campo `muscles` como TEXT[] (valores livres)
  - Search apenas em name (ILIKE)

### Recomendações para Plano de Testes

**Unit tests:**
- `ListExercisesUC`: filtra corretamente, pagina corretamente
- `GetExerciseUC`: retorna exercise + stats (se autenticado)
- `GetExerciseHistoryUC`: agrupa sets por session, pagina corretamente

**Integration tests:**
- `GET /exercises`: retorna 200 com lista paginada, filtra por muscleGroup/equipment/difficulty/search
- `GET /exercises/:id`: retorna 200 com detalhes, inclui stats se autenticado
- `GET /exercises/:id/history`: retorna 200 com histórico paginado, requer autenticação

**Edge cases:**
- Biblioteca vazia (sem seed)
- Filtros inválidos
- ExerciseID não encontrado (404)
- User sem histórico (stats null/zero)
- Paginação inválida (page < 1, pageSize > 100)

### Próximos passos
1. Responder perguntas da seção 2
2. Criar conteúdo para seed (ou decidir usar mock)
3. Criar plano detalhado com tasks granulares
4. Implementar migration + seed
5. Implementar repository + use cases + handlers
6. Testes
