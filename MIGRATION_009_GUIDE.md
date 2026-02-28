# Guia de Migração: Exercises N:N (Migration 009)

## Resumo das Mudanças

A migration 009 refatora o relacionamento entre `workouts` e `exercises` de 1:N para N:N através de uma tabela de junção `workout_exercises`.

## Estrutura Anterior
```
workouts (1) ----< exercises (N)
  (user_id)      (workout_id, sets, reps, weight, rest_time)
```

## Nova Estrutura
```
workouts (1) ----< workout_exercises >---- exercises (compartilhados)
                    (carrega configurações)   (biblioteca)
```

## Impactos na Base de Dados

### Tabelas Modificadas

#### `exercises` (antes: 1:N com workouts)
**Antes:**
```sql
exercises (
    id UUID PRIMARY KEY,
    workout_id UUID NOT NULL,           -- ❌ REMOVIDO
    name VARCHAR(255),
    thumbnail_url VARCHAR(500),
    sets INT,
    reps VARCHAR(20),
    muscles JSONB,
    rest_time INT,
    weight INT,
    order_index INT,
    created_at, updated_at
)
```

**Depois:**
```sql
exercises (
    id UUID PRIMARY KEY,                -- ✅ Cada exercise é único na biblioteca
    name VARCHAR(255),
    description VARCHAR(500),           -- ✅ NOVO
    thumbnail_url VARCHAR(500),
    muscles JSONB,                      -- ✅ Mantém músculos
    created_at, updated_at
    -- ❌ Removido: sets, reps, rest_time, weight, order_index
    -- ❌ Removido: workout_id
)
```

#### `workout_exercises` (NOVA)
```sql
workout_exercises (                     -- ✅ NOVA TABELA
    id UUID PRIMARY KEY,
    workout_id UUID NOT NULL,           -- FK para workouts
    exercise_id UUID NOT NULL,          -- FK para exercises
    sets INT,                           -- ✅ Específico de cada uso
    reps VARCHAR(20),                   -- ✅ Específico de cada uso
    rest_time INT,                      -- ✅ Específico de cada uso
    weight INT,                         -- ✅ Específico de cada uso
    order_index INT,                    -- ✅ Específico de cada uso
    created_at, updated_at,
    UNIQUE (workout_id, exercise_id)
)
```

#### `set_records` (Adiciona coluna)
**Antes:**
```sql
set_records (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    exercise_id UUID NOT NULL,          -- ❌ REMOVIDO
    set_number INT,
    weight INT,
    reps INT,
    status VARCHAR(20),
    recorded_at TIMESTAMPTZ
)
```

**Depois:**
```sql
set_records (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    workout_exercise_id UUID NOT NULL,  -- ✅ NOVO (referencia workout_exercises)
    set_number INT,
    weight INT,
    reps INT,
    status VARCHAR(20),
    recorded_at TIMESTAMPTZ
)
```

## Impactos no Código Go

### 1. Entidades (Domain Layer)

#### Entity: `Exercise`
```go
// ANTES
type Exercise struct {
    ID           uuid.UUID      `db:"id"`
    WorkoutID    uuid.UUID      `db:"workout_id"`        // ❌ REMOVER
    Name         string         `db:"name"`
    ThumbnailURL string         `db:"thumbnail_url"`
    Sets         int            `db:"sets"`              // ❌ REMOVER
    Reps         string         `db:"reps"`              // ❌ REMOVER
    Muscles      pq.StringArray `db:"muscles"`
    RestTime     int            `db:"rest_time"`         // ❌ REMOVER
    Weight       int            `db:"weight"`            // ❌ REMOVER
    OrderIndex   int            `db:"order_index"`       // ❌ REMOVER
    ...
}

// DEPOIS
type Exercise struct {
    ID           uuid.UUID      `db:"id"`
    Name         string         `db:"name"`
    Description  string         `db:"description"`       // ✅ NOVO
    ThumbnailURL string         `db:"thumbnail_url"`
    Muscles      pq.StringArray `db:"muscles"`
    CreatedAt    time.Time      `db:"created_at"`
    UpdatedAt    time.Time      `db:"updated_at"`
}
```

#### Entity: `WorkoutExercise` (NOVO)
```go
// ✅ NOVA ENTIDADE
type WorkoutExercise struct {
    ID           uuid.UUID `db:"id"`
    WorkoutID    uuid.UUID `db:"workout_id"`
    ExerciseID   uuid.UUID `db:"exercise_id"`
    Sets         int       `db:"sets"`
    Reps         string    `db:"reps"`
    RestTime     int       `db:"rest_time"`
    Weight       int       `db:"weight"`
    OrderIndex   int       `db:"order_index"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}
```

#### Entity: `SetRecord`
```go
// ANTES
type SetRecord struct {
    ID          uuid.UUID `db:"id"`
    SessionID   uuid.UUID `db:"session_id"`
    ExerciseID  uuid.UUID `db:"exercise_id"`           // ❌ REMOVER
    SetNumber   int       `db:"set_number"`
    Weight      int       `db:"weight"`
    Reps        int       `db:"reps"`
    Status      string    `db:"status"`
    RecordedAt  time.Time `db:"recorded_at"`
}

// DEPOIS
type SetRecord struct {
    ID                 uuid.UUID `db:"id"`
    SessionID          uuid.UUID `db:"session_id"`
    WorkoutExerciseID  uuid.UUID `db:"workout_exercise_id"` // ✅ NOVO
    SetNumber          int       `db:"set_number"`
    Weight             int       `db:"weight"`
    Reps               int       `db:"reps"`
    Status             string    `db:"status"`
    RecordedAt         time.Time `db:"recorded_at"`
}
```

### 2. Ports (Interfaces)

#### Repository Ports

**ExerciseRepository**
```go
// ANTES
type ExerciseRepository interface {
    GetByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]Exercise, error)
    Save(ctx context.Context, exercise *Exercise) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// DEPOIS - Mudanças
type ExerciseRepository interface {
    // ✅ Operações em exercises (biblioteca compartilhada)
    GetByID(ctx context.Context, id uuid.UUID) (*Exercise, error)
    GetAll(ctx context.Context) ([]Exercise, error)
    Create(ctx context.Context, exercise *Exercise) error
    Update(ctx context.Context, exercise *Exercise) error
    // ... removido: GetByWorkoutID
}

// ✅ NOVO + WorkoutExerciseRepository
type WorkoutExerciseRepository interface {
    GetByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]WorkoutExercise, error)
    GetByID(ctx context.Context, id uuid.UUID) (*WorkoutExercise, error)
    Save(ctx context.Context, we *WorkoutExercise) error
    Update(ctx context.Context, we *WorkoutExercise) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// ✅ NOVO + SetRecordRepository
type SetRecordRepository interface {
    GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]SetRecord, error)
    Save(ctx context.Context, record *SetRecord) error
    // exercise_id não é mais necessário no filtro
}
```

### 3. Use Cases

#### Estrutura de Resposta na API

**GET /api/v1/workouts/{id}**
```json
// ANTES - exercises vinculados ao workout
{
  "data": {
    "id": "...",
    "name": "Treino A",
    "exercises": [
      {
        "id": "ex-1",              // ID do exercises (direto)
        "name": "Supino",
        "sets": 4,
        "reps": "8-12",
        "weight": 80000,
        "restTime": 90,
        "muscles": ["Peito"]
      }
    ]
  }
}

// DEPOIS - exercises via workout_exercises (incluir ambos IDs)
{
  "data": {
    "id": "...",
    "name": "Treino A",
    "exercises": [
      {
        "id": "we-1",                      // ✅ ID de workout_exercises
        "exerciseId": "ex-1",               // ✅ Novo campo
        "name": "Supino",
        "sets": 4,
        "reps": "8-12",
        "weight": 80000,
        "restTime": 90,
        "muscles": ["Peito"]
      }
    ]
  }
}
```

### 4. Queries SQL (SQLC)

**Criar/atualizar as queries:**

```sql
-- ✅ NOVO: Buscar exercises de um workout com configurações
-- queries/workout_exercises.sql
SELECT 
    we.id,
    we.workout_id,
    we.exercise_id,
    we.sets,
    we.reps,
    we.rest_time,
    we.weight,
    we.order_index,
    e.name,
    e.description,
    e.thumbnail_url,
    e.muscles
FROM workout_exercises we
JOIN exercises e ON we.exercise_id = e.id
WHERE we.workout_id = $1
ORDER BY we.order_index ASC;

-- ✅ NOVO: Vincular exercise a workout
INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, rest_time, weight, order_index)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- ✅ ATUALIZADO: set_records usa workout_exercise_id
SELECT 
    sr.id,
    sr.session_id,
    sr.workout_exercise_id,      -- ✅ Novo campo
    sr.set_number,
    sr.weight,
    sr.reps,
    sr.status,
    sr.recorded_at
FROM set_records sr
WHERE sr.session_id = $1
ORDER BY sr.set_number ASC;
```

### 5. Checklist de Migração do Código

- [ ] Atualizar entidade `Exercise` (remover campos de configuração)
- [ ] Criar entidade `WorkoutExercise`
- [ ] Atualizar entidade `SetRecord`
- [ ] Refatorar `ExerciseRepository`
- [ ] Criar `WorkoutExerciseRepository`
- [ ] Atualizar `SetRecordRepository`
- [ ] Regenerar queries com `make sqlc`
- [ ] Atualizar DTOs de resposta da API
- [ ] Atualizar handlers de GET /workouts/{id}
- [ ] Atualizar handlers de sessions/set_records
- [ ] Atualizar testes (mocks e table-driven tests)
- [ ] Testar fluxo completo (criar workout com exercises → registrar série)

## Rollout

1. **Deploy migration 009** (sem código novo)
2. **Deploy código com novo schema** (compatible com ambas estruturas temporariamente)
3. **Remover código legado** em release posterior

## Referências

- Arquivo: [migrations/009_refactor_exercises_to_shared_library.sql](../migrations/009_refactor_exercises_to_shared_library.sql)
- Padrão: Hexagonal Architecture com Ports & Adapters
- Template: [global.instructions.md](.github/instructions/global.instructions.md)
