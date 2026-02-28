# Plan — Workout Exercises N:N Relationship

## 1) Inputs usados
- `.github/instructions/MIGRATION_009_GUIDE.md` (guia oficial da migração)
- `migrations/003_create_exercises.sql` (estrutura AS-IS)
- `migrations/009_refactor_exercises_to_shared_library.sql` (estrutura TO-BE)
- `internal/kinetria/domain/entities/exercise.go` (entity atual)
- `internal/kinetria/domain/entities/set_record.go` (entity atual)
- `internal/kinetria/domain/ports/repositories.go` (ports atuais)
- `internal/kinetria/gateways/repositories/exercise_repository.go` (implementação atual)

## 2) AS-IS (resumo)

### Estrutura atual

**Relacionamento 1:N direto:**
```
workouts (1) ----< exercises (N)
  (user_id)      (workout_id, sets, reps, weight, rest_time)
```

**Tabela `exercises`:**
- Cada exercise pertence a **um único workout** (coluna `workout_id`)
- Configurações de treino (sets, reps, rest_time, weight, order_index) estão **armazenadas no próprio exercise**
- Não há compartilhamento: criar o mesmo exercício em dois workouts cria **2 registros duplicados**

**Entidade `Exercise` (domain):**
```go
type Exercise struct {
    ID           uuid.UUID      // PK
    WorkoutID    uuid.UUID      // FK para workout
    Name         string
    ThumbnailURL string
    Sets         int            // configuração específica
    Reps         string         // configuração específica
    Muscles      []string
    RestTime     int            // configuração específica
    Weight       int            // configuração específica
    OrderIndex   int            // configuração específica
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Tabela `set_records`:**
- Referencia diretamente `exercise_id` (FK para `exercises`)
- Unique constraint: `(session_id, exercise_id, set_number)`

**Port `ExerciseRepository`:**
```go
type ExerciseRepository interface {
    ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error)
}
```

**Queries SQL atuais:**
- `ExistsExerciseByIDAndWorkoutID` (verificação de pertencimento)

### Limitações do AS-IS

1. **Duplicação de dados:** Exercícios repetidos em múltiplos workouts geram N registros idênticos
2. **Impossibilidade de biblioteca compartilhada:** Não há conceito de "exercise template" reutilizável
3. **Manutenção complexa:** Atualizar metadata de um exercise (ex: thumbnail_url, muscles) requer update em todos os registros duplicados
4. **Acoplamento forte:** set_records referencia exercise_id diretamente, o que não diferencia "qual uso do exercise" em caso de reutilização

---

## 3) TO-BE (proposta)

### Estrutura futura

**Relacionamento N:N via tabela de junção:**
```
workouts (1) ----< workout_exercises (N:M) >---- exercises (biblioteca)
                    (carrega configurações)         (compartilhados)
```

### Tabelas modificadas/criadas

#### `exercises` (biblioteca compartilhada)
```sql
exercises (
    id UUID PRIMARY KEY,                -- ✅ PK global (não vinculado a workout)
    name VARCHAR(255),
    description VARCHAR(500),           -- ✅ NOVO
    thumbnail_url VARCHAR(500),
    muscles JSONB,                      -- ✅ Mantém (metadata do exercício)
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
    -- ❌ REMOVIDO: workout_id, sets, reps, rest_time, weight, order_index
)
```

**Índices:**
- `idx_exercises_name` (busca por nome)
- `idx_exercises_muscles` GIN (busca por grupo muscular)

#### `workout_exercises` (NOVA — tabela de junção)
```sql
workout_exercises (
    id UUID PRIMARY KEY,                -- ✅ PK próprio (identificador único do vínculo)
    workout_id UUID NOT NULL,           -- FK para workouts
    exercise_id UUID NOT NULL,          -- FK para exercises
    sets INT,                           -- ✅ Configuração específica deste uso
    reps VARCHAR(20),                   -- ✅ Configuração específica deste uso
    rest_time INT,                      -- ✅ Configuração específica deste uso
    weight INT,                         -- ✅ Configuração específica deste uso
    order_index INT,                    -- ✅ Ordem específica deste uso
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    UNIQUE (workout_id, exercise_id)    -- Um exercise só pode aparecer 1x por workout
)
```

**Índices:**
- `idx_workout_exercises_workout_id` (buscar exercises de um workout)
- `idx_workout_exercises_exercise_id` (buscar workouts que usam um exercise)
- `idx_workout_exercises_order` (ordenação: workout_id, order_index)

#### `set_records` (atualizada)
```sql
set_records (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    workout_exercise_id UUID NOT NULL,  -- ✅ Novo FK (referencia workout_exercises)
    set_number INT,
    weight INT,
    reps INT,
    status VARCHAR(20),
    recorded_at TIMESTAMPTZ,
    UNIQUE (session_id, workout_exercise_id, set_number)  -- ✅ Constraint atualizada
    -- ❌ REMOVIDO: exercise_id
)
```

### Entidades (domain layer)

#### `Exercise` (atualizada)
```go
type Exercise struct {
    ID           uuid.UUID
    Name         string
    Description  string         // ✅ NOVO
    ThumbnailURL string
    Muscles      []string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    // ❌ REMOVIDO: WorkoutID, Sets, Reps, RestTime, Weight, OrderIndex
}
```

#### `WorkoutExercise` (NOVA)
```go
type WorkoutExercise struct {
    ID          uuid.UUID
    WorkoutID   uuid.UUID
    ExerciseID  uuid.UUID
    Sets        int
    Reps        string
    RestTime    int
    Weight      int
    OrderIndex  int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Observação:** WorkoutExercise é uma **entidade de domínio completa**, não apenas um DTO de junção. Ela representa "um exercício configurado para um treino específico".

#### `SetRecord` (atualizada)
```go
type SetRecord struct {
    ID                uuid.UUID
    SessionID         uuid.UUID
    WorkoutExerciseID uuid.UUID  // ✅ NOVO
    SetNumber         int
    Weight            int
    Reps              int
    Status            string
    RecordedAt        time.Time
    // ❌ REMOVIDO: ExerciseID
}
```

### Ports (interfaces)

#### `ExerciseRepository` (atualizado)
```go
type ExerciseRepository interface {
    // Operações sobre a biblioteca de exercises
    GetByID(ctx context.Context, id uuid.UUID) (*entities.Exercise, error)
    GetAll(ctx context.Context) ([]entities.Exercise, error)
    Create(ctx context.Context, exercise *entities.Exercise) error
    Update(ctx context.Context, exercise *entities.Exercise) error
    Delete(ctx context.Context, id uuid.UUID) error
    // ❌ REMOVIDO: ExistsByIDAndWorkoutID (essa lógica vai para WorkoutExerciseRepository)
}
```

#### `WorkoutExerciseRepository` (NOVO)
```go
type WorkoutExerciseRepository interface {
    // Operações sobre o vínculo workout <-> exercise
    GetByID(ctx context.Context, id uuid.UUID) (*entities.WorkoutExercise, error)
    GetByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]entities.WorkoutExercise, error)
    ExistsByIDAndWorkoutID(ctx context.Context, id, workoutID uuid.UUID) (bool, error)
    Create(ctx context.Context, we *entities.WorkoutExercise) error
    Update(ctx context.Context, we *entities.WorkoutExercise) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

#### `SetRecordRepository` (atualizado)
```go
type SetRecordRepository interface {
    Create(ctx context.Context, setRecord *entities.SetRecord) error
    FindBySessionExerciseSet(ctx context.Context, sessionID, workoutExerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error)
    // Assinatura alterada: workoutExerciseID no lugar de exerciseID
}
```

### Queries SQL (SQLC)

**Arquivo: `queries/exercises.sql`**
```sql
-- name: GetExerciseByID :one
SELECT id, name, description, thumbnail_url, muscles, created_at, updated_at
FROM exercises
WHERE id = $1;

-- name: GetAllExercises :many
SELECT id, name, description, thumbnail_url, muscles, created_at, updated_at
FROM exercises
ORDER BY name ASC;

-- name: CreateExercise :one
INSERT INTO exercises (name, description, thumbnail_url, muscles)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, thumbnail_url, muscles, created_at, updated_at;

-- name: UpdateExercise :exec
UPDATE exercises
SET name = $2, description = $3, thumbnail_url = $4, muscles = $5, updated_at = NOW()
WHERE id = $1;

-- name: DeleteExercise :exec
DELETE FROM exercises WHERE id = $1;
```

**Arquivo: `queries/workout_exercises.sql` (NOVO)**
```sql
-- name: GetWorkoutExerciseByID :one
SELECT 
    we.id,
    we.workout_id,
    we.exercise_id,
    we.sets,
    we.reps,
    we.rest_time,
    we.weight,
    we.order_index,
    we.created_at,
    we.updated_at,
    e.name,
    e.description,
    e.thumbnail_url,
    e.muscles
FROM workout_exercises we
JOIN exercises e ON we.exercise_id = e.id
WHERE we.id = $1;

-- name: GetWorkoutExercisesByWorkoutID :many
SELECT 
    we.id,
    we.workout_id,
    we.exercise_id,
    we.sets,
    we.reps,
    we.rest_time,
    we.weight,
    we.order_index,
    we.created_at,
    we.updated_at,
    e.name,
    e.description,
    e.thumbnail_url,
    e.muscles
FROM workout_exercises we
JOIN exercises e ON we.exercise_id = e.id
WHERE we.workout_id = $1
ORDER BY we.order_index ASC;

-- name: ExistsWorkoutExerciseByIDAndWorkoutID :one
SELECT EXISTS(
    SELECT 1 FROM workout_exercises
    WHERE id = $1 AND workout_id = $2
) AS exists;

-- name: CreateWorkoutExercise :one
INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, rest_time, weight, order_index)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, workout_id, exercise_id, sets, reps, rest_time, weight, order_index, created_at, updated_at;

-- name: UpdateWorkoutExercise :exec
UPDATE workout_exercises
SET sets = $2, reps = $3, rest_time = $4, weight = $5, order_index = $6, updated_at = NOW()
WHERE id = $1;

-- name: DeleteWorkoutExercise :exec
DELETE FROM workout_exercises WHERE id = $1;
```

**Arquivo: `queries/set_records.sql` (ATUALIZADO)**
```sql
-- name: CreateSetRecord :exec
INSERT INTO set_records (session_id, workout_exercise_id, set_number, weight, reps, status)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FindSetRecordBySessionExerciseSet :one
SELECT id, session_id, workout_exercise_id, set_number, weight, reps, status, recorded_at
FROM set_records
WHERE session_id = $1 AND workout_exercise_id = $2 AND set_number = $3;
```

### Contratos HTTP (API)

#### GET /api/v1/workouts/{id}

**Antes:**
```json
{
  "data": {
    "id": "uuid",
    "name": "Treino A",
    "exercises": [
      {
        "id": "exercise-id",           // ID do exercises
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

**Depois:**
```json
{
  "data": {
    "id": "uuid",
    "name": "Treino A",
    "exercises": [
      {
        "id": "workout-exercise-id",    // ✅ ID de workout_exercises
        "exerciseId": "exercise-id",     // ✅ Novo campo (ID na biblioteca)
        "name": "Supino",
        "description": "",               // ✅ Novo campo
        "sets": 4,
        "reps": "8-12",
        "weight": 80000,
        "restTime": 90,
        "muscles": ["Peito"],
        "thumbnailUrl": "..."
      }
    ]
  }
}
```

**Impactos:**
- Frontend agora recebe `exerciseId` (para referenciar exercise da biblioteca)
- Campo `id` passa a ser o ID de `workout_exercises` (necessário para set_records)

#### POST /api/v1/sessions/{sessionId}/set-records

**Antes:**
```json
{
  "exerciseId": "uuid",    // ID do exercises
  "setNumber": 1,
  "weight": 80000,
  "reps": 10,
  "status": "completed"
}
```

**Depois:**
```json
{
  "workoutExerciseId": "uuid",  // ✅ ID de workout_exercises
  "setNumber": 1,
  "weight": 80000,
  "reps": 10,
  "status": "completed"
}
```

**Validação no backend:**
```go
// Verificar se workout_exercise pertence ao workout da session
session := getSession(sessionID)
exists := workoutExerciseRepo.ExistsByIDAndWorkoutID(ctx, workoutExerciseID, session.WorkoutID)
if !exists {
    return ErrWorkoutExerciseNotBelongsToWorkout
}
```

### Observabilidade

**Logs relevantes:**
- `exercise.created` (novo exercise na biblioteca)
- `workout_exercise.linked` (exercise vinculado a workout)
- `workout_exercise.updated` (configuração de workout_exercise alterada)
- `set_record.created` (com `workout_exercise_id`)

**Métricas:**
- `kinetria_exercises_library_total` (gauge: total de exercises na biblioteca)
- `kinetria_workout_exercises_per_workout` (histogram: distribuição de exercises por workout)
- `kinetria_set_records_created_total` (counter, label: `workout_exercise_id`)

**Tracing:**
- Span: `ExerciseRepository.GetByID`
- Span: `WorkoutExerciseRepository.GetByWorkoutID` (com JOIN)
- Span: `SetRecordRepository.Create` (incluir `workout_exercise_id` em atributos)

---

## 4) Decisões e Assunções

### Decisões de design

1. **WorkoutExercise é uma entidade de domínio completa**
   - Não é apenas um DTO de junção
   - Possui seu próprio ID (não composto)
   - Carrega tanto dados de configuração (`sets`, `reps`) quanto metadata do exercise (`name`, `muscles`)

2. **Exercise.Description é opcional**
   - Pode ser vazio inicialmente
   - Permite enriquecer a biblioteca no futuro sem quebra

3. **Unique constraint (workout_id, exercise_id)**
   - Um exercise só pode aparecer 1x por workout
   - Se quiser usar o mesmo exercise com configurações diferentes, duplicar não faz sentido no modelo N:N
   - Alternativa: ajustar a lógica de negócio para permitir múltiplos usos

4. **Migration 009 é destrutiva (com migração de dados)**
   - A migration DROP + RENAME a tabela `exercises`
   - Dados são migrados via JOIN por `name` e `thumbnail_url` (não preserva IDs originais)
   - **Rollback complexo**: necessário backup antes de aplicar

5. **SetRecord referencia WorkoutExercise diretamente**
   - Isso permite saber **qual configuração** foi usada na série
   - Se o workout for editado, set_records históricos não são afetados

### Assunções

1. **Não há set_records órfãos no AS-IS**
   - Assumimos que `exercise_id` sempre referencia um exercise válido
   - A migration falha se houver FK violada

2. **Nome + thumbnail_url identificam exercises únicos na migração**
   - A migration assume que exercises com mesmo `name` e `thumbnail_url` são "equivalentes"
   - Se houver exercises idênticos mas com configurações diferentes, apenas um será criado na biblioteca

3. **Frontend será atualizado junto com o backend**
   - Não há período de convivência com API antiga
   - Se necessário rollout gradual, será preciso versionamento de API (v1 vs v2)

4. **Não há concorrência na edição de workout_exercises durante sessão ativa**
   - Se um workout for editado enquanto uma session está ativa, set_records continuam válidos (referencia workout_exercise_id imutável)
   - Mas futuros sets usarão a nova configuração (se order_index/exerciseID mudar)

---

## 5) Riscos / Edge Cases

### Riscos da migração

| Risco | Severidade | Mitigação |
|-------|-----------|-----------|
| **Migration falha no meio** (órfãos em set_records) | `blocker` | Testar migration em ambiente de staging com dump de produção. Criar backup antes de aplicar. |
| **Perda de IDs originais** (exercises) | `high` | IDs novos são gerados. Impacto: se frontend/app mobile cachear IDs de exercises, cache ficará inválido. Solução: forçar refresh de workouts no app após deploy. |
| **Duplicação de exercises na biblioteca** | `medium` | Migration usa `ON CONFLICT DO NOTHING`. Se houver exercises com mesmo nome mas thumbnails diferentes, ambos serão criados. Avaliar deduplicação manual pós-migração. |
| **Constraint UNIQUE (workout_id, exercise_id) bloqueia usos múltiplos** | `medium` | Se um usuário quiser fazer "Supino 4x8" e "Supino 3x12" no mesmo workout, não será possível. Decisão: aceitar limitação ou remover constraint e adicionar lógica de ordering. |
| **Rollback complexo** | `high` | Migration é destrutiva (DROP TABLE). Rollback requer restore de backup. Não há migration DOWN automática. |

### Edge cases

1. **Session ativa durante deploy da migration**
   - Se uma session está ativa (status=`active`) e a migration roda, set_records poderão ficar inconsistentes
   - **Mitigação:** Deploy em janela de baixo tráfego. Validar que não há sessions ativas antes de rodar migration.

2. **Workout editado durante session ativa**
   - Se um workout_exercise for deletado enquanto uma session o referencia, set_records ficarão órfãos? 
   - **Resposta:** Não. `workout_exercises` tem `ON DELETE RESTRICT` na FK com `set_records`. Delete falhará se houver set_records vinculados.

3. **Usuario cria exercise duplicado na biblioteca**
   - Se houver UI para criar exercises, usuário pode criar "Supino" 2x (com mesma thumbnail)
   - **Mitigação:** Implementar busca/autocomplete ao criar workout. Sugerir exercises existentes antes de criar novo.

4. **SetRecord com workout_exercise_id inválido**
   - API deve validar que `workoutExerciseID` pertence ao workout da session antes de criar set_record
   - Caso contrário, usuário poderia registrar série de exercise de outro workout

5. **Concorrência: dois usuários editam mesmo workout**
   - Se dois usuários editarem configurações de workout_exercises ao mesmo tempo, last-write-wins
   - **Mitigação:** Adicionar `updated_at` e validação otimista se necessário (fora do escopo desta migration)

---

## 6) Rollout / Compatibilidade

### Estratégia de rollout

**Opção 1: Big Bang (recomendada para MVP)**
1. Deploy migration 009 em staging e validar
2. Criar backup completo de produção
3. Deploy migration 009 em produção (janela de manutenção)
4. Deploy backend com novo código
5. Deploy frontend com novos contratos
6. Validar fluxo end-to-end (criar workout, iniciar session, registrar serie)

**Vantagens:**
- Simples, sem período de convivência
- Não requer versionamento de API

**Desvantagens:**
- Requer janela de manutenção
- Rollback complexo (restore de backup)

**Opção 2: Expand/Contract (se necessário rollout sem downtime)**
1. Phase 1 (Expand): Deploy migration 009 + código compatível com ambas estruturas
   - Backend lê de `workout_exercises` mas mantém lógica de fallback
   - Frontend ainda consome contratos antigos (se possível)
2. Phase 2 (Contract): Remover código/tabelas antigas após validação
3. Phase 3 (Cleanup): Remover coluna `exercise_id` de set_records

**Complexidade:** Alta. Requer código dual-mode e testes extensivos. Não recomendado para MVP.

### Retrocompatibilidade

**API:** Não compatível. Contratos HTTP mudam:
- `GET /workouts/{id}` retorna `exerciseId` + novos campos
- `POST /sessions/{id}/set-records` requer `workoutExerciseId` no lugar de `exerciseId`

**Database:** Migration é destrutiva. Rollback requer restore de backup.

### Checklist de deploy

- [ ] Testar migration 009 em ambiente local com dados reais
- [ ] Testar migration 009 em staging com dump de produção
- [ ] Criar backup completo de produção (DB + arquivos estáticos)
- [ ] Validar que não há sessions ativas (`SELECT * FROM sessions WHERE status='active'`)
- [ ] Deploy migration 009 em produção
- [ ] Deploy backend com novo código
- [ ] Regenerar SQLC (`make sqlc`)
- [ ] Rodar testes end-to-end em produção (happy path)
- [ ] Deploy frontend (se houver mudanças nos contratos)
- [ ] Monitorar logs/métricas por 24h
- [ ] Validar com usuários beta
- [ ] Documentar em changelog/release notes

### Validação pós-deploy

```sql
-- Verificar que exercises foram criados
SELECT COUNT(*) FROM exercises;  -- Deve ser > 0

-- Verificar que workout_exercises foram criados
SELECT COUNT(*) FROM workout_exercises;  -- Deve ser = COUNT antigo de exercises

-- Verificar que set_records foram migrados
SELECT COUNT(*) FROM set_records WHERE workout_exercise_id IS NOT NULL;  -- 100%

-- Verificar integridade referencial
SELECT sr.id 
FROM set_records sr
LEFT JOIN workout_exercises we ON sr.workout_exercise_id = we.id
WHERE we.id IS NULL;  -- Deve retornar 0 linhas (sem órfãos)
```

---

## 7) Referências

- Guia oficial: `MIGRATION_009_GUIDE.md`
- Migration script: `migrations/009_refactor_exercises_to_shared_library.sql`
- Padrão arquitetural: `.github/instructions/global.instructions.md` (Hexagonal + DI via Fx)
- Documentação SQLC: https://docs.sqlc.dev/
