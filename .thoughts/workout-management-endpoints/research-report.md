# 🔎 Research Report — Workout Management Endpoints

## 1) Task Summary

### O que é
Implementar 3 endpoints de gerenciamento de workouts customizados:
- **POST /api/v1/workouts** — Criar workout customizado pelo usuário
- **PUT /api/v1/workouts/:id** — Atualizar workout customizado
- **DELETE /api/v1/workouts/:id** — Deletar workout customizado

### O que não é (fora de escopo)
- Duplicação de workouts (clone)
- Compartilhamento de workouts entre usuários
- Versionamento de workouts
- Importação/exportação de workouts (JSON, CSV)

---

## 2) Clarifying Questions (para o dev)

### Regras de Negócio
1. **Ownership model:** Usar `created_by UUID` (FK para users) ou `is_custom BOOLEAN`? `created_by` é mais flexível para futuras features (compartilhamento).
2. **Workout template vs customizado:** Workouts sem `created_by` são templates (read-only)? Ou todos os workouts pertencem a um usuário?
3. **Deleção de workout:** Soft delete (adicionar campo `deleted_at`) ou hard delete? O que acontece com sessions ativas que referenciam o workout deletado?
4. **Atualização de workout:** Permitir atualizar workout que já tem sessions completadas? Ou bloquear/versionar?

### Interface / Contrato
5. **POST /workouts:** Criar workout vazio (sem exercises) e adicionar depois? Ou exigir pelo menos 1 exercise?
6. **PUT /workouts:** Atualização completa (substituir todos os exercises) ou parcial (merge)?
7. **Validações:** Limites de sets (min/max)? Limites de exercises por workout (max 20)?

### Persistência
8. **Transação:** Criar/atualizar workout + workout_exercises deve ser atômico (transação)?
9. **Order index:** Validar que não há duplicatas? Reordenar automaticamente (1, 2, 3...) ou aceitar gaps?
10. **Cascade delete:** Se deletar workout, deletar workout_exercises automaticamente (ON DELETE CASCADE já existe)?

---

## 3) Facts from the Codebase

### Domínio(s) candidato(s)
- `internal/kinetria/domain/workouts/` (já existe, expandir)

### Entrypoints (cmd/)
- `cmd/kinetria/api/main.go` — Único entrypoint, usa Fx para DI

### Principais pacotes/símbolos envolvidos

**Entidades existentes:**
```go
// internal/kinetria/domain/entities/workout.go
type Workout struct {
    ID          uuid.UUID
    Name        string
    Description *string
    Type        vos.WorkoutType      // FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO
    Intensity   vos.WorkoutIntensity // BAIXA, MODERADA, ALTA
    Duration    int                  // minutos estimados
    ImageURL    *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// internal/kinetria/domain/entities/workout_exercise.go
type WorkoutExercise struct {
    ID         uuid.UUID
    WorkoutID  uuid.UUID
    ExerciseID uuid.UUID
    Sets       int
    Reps       string // "8-12" ou "10"
    RestTime   int    // segundos
    Weight     *int   // gramas (opcional)
    OrderIndex int
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**Ports existentes:**
```go
// internal/kinetria/domain/ports/repositories.go
type WorkoutRepository interface {
    ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entities.Workout, int, error)
    GetByID(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)
    ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error)
    GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error)
    // FALTA: Create, Update, Delete
}
```

**Gateways existentes:**
- `gateways/repositories/workout_repository.go` — Implementação com SQLC
- `gateways/repositories/queries/workouts.sql` — Queries SQL tipadas
- `gateways/http/handler_workouts.go` — Já tem GET /workouts e GET /workouts/:id

**Migrations existentes:**
- Migration 002: Criou tabela `workouts`
- Migration 009: Refatorou para N:N com `workout_exercises`

```sql
-- Estrutura atual (migration 009)
CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    intensity VARCHAR(50) NOT NULL,
    duration INT NOT NULL,
    image_url TEXT,
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
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(workout_id, order_index)
);
```

**Campo faltante:**
- `created_by UUID REFERENCES users(id)` (para ownership)

---

## 4) Current Flow (AS-IS)

### Fluxo atual de workouts
1. **GET /workouts** → Lista workouts (sem filtro de ownership, retorna todos)
2. **GET /workouts/:id** → Detalhes do workout + exercises
3. Não há endpoints para criar/atualizar/deletar

### Relacionamento atual
```
users (1) ----< workouts (N)  (FALTA FK created_by)
                  ↓
            workout_exercises (N:N com exercises)
                  ↓
              set_records (via workout_exercise_id)
```

---

## 5) Change Points (prováveis pontos de alteração)

### 5.1) Migration

**Arquivo a criar:**
- `internal/kinetria/gateways/migrations/013_add_workout_ownership.sql`

```sql
-- Adicionar coluna created_by (nullable para workouts template)
ALTER TABLE workouts 
ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE CASCADE;

-- Índice para buscar workouts por usuário
CREATE INDEX idx_workouts_created_by ON workouts(created_by);

-- Opcional: soft delete
ALTER TABLE workouts 
ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_workouts_deleted_at ON workouts(deleted_at) WHERE deleted_at IS NULL;
```

**Nota:** Se usar soft delete, queries devem filtrar `deleted_at IS NULL`.

---

### 5.2) Domain Layer

**Arquivo a modificar:**
- `internal/kinetria/domain/entities/workout.go`

Adicionar campo `CreatedBy`:
```go
type Workout struct {
    ID          uuid.UUID
    Name        string
    Description *string
    Type        vos.WorkoutType
    Intensity   vos.WorkoutIntensity
    Duration    int
    ImageURL    *string
    CreatedBy   *uuid.UUID // NULL = template, NOT NULL = customizado
    DeletedAt   *time.Time // Opcional, se usar soft delete
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Arquivos a criar:**
- `internal/kinetria/domain/workouts/uc_create_workout.go`
- `internal/kinetria/domain/workouts/uc_update_workout.go`
- `internal/kinetria/domain/workouts/uc_delete_workout.go`

**Structs auxiliares:**
```go
type CreateWorkoutInput struct {
    Name        string
    Description *string
    Type        vos.WorkoutType
    Intensity   vos.WorkoutIntensity
    Duration    int
    ImageURL    *string
    Exercises   []WorkoutExerciseInput
}

type WorkoutExerciseInput struct {
    ExerciseID uuid.UUID
    Sets       int
    Reps       string
    RestTime   int
    Weight     *int
    OrderIndex int
}

type UpdateWorkoutInput struct {
    Name        *string
    Description *string
    Type        *vos.WorkoutType
    Intensity   *vos.WorkoutIntensity
    Duration    *int
    ImageURL    *string
    Exercises   []WorkoutExerciseInput // Substituir todos
}
```

---

### 5.3) Ports

**Arquivo a modificar:**
- `internal/kinetria/domain/ports/repositories.go`

Adicionar métodos:
```go
type WorkoutRepository interface {
    // Existentes
    ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entities.Workout, int, error)
    GetByID(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)
    ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error)
    GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error)
    
    // Novos
    Create(ctx context.Context, workout *entities.Workout, exercises []*entities.WorkoutExercise) error
    Update(ctx context.Context, workout *entities.Workout, exercises []*entities.WorkoutExercise) error
    Delete(ctx context.Context, workoutID, userID uuid.UUID) error // Soft ou hard delete
    HasActiveSessions(ctx context.Context, workoutID uuid.UUID) (bool, error) // Validar antes de deletar
}
```

---

### 5.4) Repository Layer

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/queries/workouts.sql`

Adicionar queries:
```sql
-- name: CreateWorkout :one
INSERT INTO workouts (name, description, type, intensity, duration, image_url, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateWorkout :exec
UPDATE workouts
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    type = COALESCE($4, type),
    intensity = COALESCE($5, intensity),
    duration = COALESCE($6, duration),
    image_url = COALESCE($7, image_url),
    updated_at = NOW()
WHERE id = $1 AND created_by = $8;

-- name: SoftDeleteWorkout :exec
UPDATE workouts
SET deleted_at = NOW()
WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL;

-- name: HardDeleteWorkout :exec
DELETE FROM workouts
WHERE id = $1 AND created_by = $2;

-- name: HasActiveSessions :one
SELECT EXISTS(
    SELECT 1 FROM sessions
    WHERE workout_id = $1 AND status = 'active'
);

-- name: CreateWorkoutExercise :exec
INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, rest_time, weight, order_index)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteWorkoutExercises :exec
DELETE FROM workout_exercises WHERE workout_id = $1;
```

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/workout_repository.go`

Implementar métodos:
```go
func (r *workoutRepository) Create(ctx context.Context, workout *entities.Workout, exercises []*entities.WorkoutExercise) error {
    // Iniciar transação
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    qtx := r.queries.WithTx(tx)
    
    // Criar workout
    createdWorkout, err := qtx.CreateWorkout(ctx, queries.CreateWorkoutParams{
        Name:        workout.Name,
        Description: workout.Description,
        Type:        string(workout.Type),
        Intensity:   string(workout.Intensity),
        Duration:    int32(workout.Duration),
        ImageUrl:    workout.ImageURL,
        CreatedBy:   workout.CreatedBy,
    })
    if err != nil {
        return err
    }
    
    workout.ID = createdWorkout.ID
    
    // Criar workout_exercises
    for _, ex := range exercises {
        err = qtx.CreateWorkoutExercise(ctx, queries.CreateWorkoutExerciseParams{
            WorkoutID:  workout.ID,
            ExerciseID: ex.ExerciseID,
            Sets:       int32(ex.Sets),
            Reps:       ex.Reps,
            RestTime:   int32(ex.RestTime),
            Weight:     ex.Weight,
            OrderIndex: int32(ex.OrderIndex),
        })
        if err != nil {
            return err
        }
    }
    
    return tx.Commit(ctx)
}

func (r *workoutRepository) Update(ctx context.Context, workout *entities.Workout, exercises []*entities.WorkoutExercise) error {
    // Iniciar transação
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    qtx := r.queries.WithTx(tx)
    
    // Atualizar workout
    err = qtx.UpdateWorkout(ctx, queries.UpdateWorkoutParams{
        ID:          workout.ID,
        Name:        &workout.Name,
        Description: workout.Description,
        Type:        (*string)(&workout.Type),
        Intensity:   (*string)(&workout.Intensity),
        Duration:    &workout.Duration,
        ImageUrl:    workout.ImageURL,
        CreatedBy:   workout.CreatedBy,
    })
    if err != nil {
        return err
    }
    
    // Deletar workout_exercises antigos
    err = qtx.DeleteWorkoutExercises(ctx, workout.ID)
    if err != nil {
        return err
    }
    
    // Criar novos workout_exercises
    for _, ex := range exercises {
        err = qtx.CreateWorkoutExercise(ctx, queries.CreateWorkoutExerciseParams{
            WorkoutID:  workout.ID,
            ExerciseID: ex.ExerciseID,
            Sets:       int32(ex.Sets),
            Reps:       ex.Reps,
            RestTime:   int32(ex.RestTime),
            Weight:     ex.Weight,
            OrderIndex: int32(ex.OrderIndex),
        })
        if err != nil {
            return err
        }
    }
    
    return tx.Commit(ctx)
}

func (r *workoutRepository) Delete(ctx context.Context, workoutID, userID uuid.UUID) error {
    // Verificar se tem sessions ativas
    hasActive, err := r.queries.HasActiveSessions(ctx, workoutID)
    if err != nil {
        return err
    }
    if hasActive {
        return errors.New("cannot delete workout with active sessions")
    }
    
    // Soft delete (ou hard delete)
    return r.queries.SoftDeleteWorkout(ctx, queries.SoftDeleteWorkoutParams{
        ID:        workoutID,
        CreatedBy: &userID,
    })
}
```

---

### 5.5) Use Cases

**Arquivo a criar:**
- `internal/kinetria/domain/workouts/uc_create_workout.go`

Lógica:
1. Recebe userID + `CreateWorkoutInput`
2. Valida inputs:
   - Name não vazio
   - Duration > 0
   - Sets > 0, RestTime >= 0
   - Exercises não vazio (pelo menos 1)
   - Order index sem duplicatas
   - Exercises existem na biblioteca
3. Cria entity `Workout` com `CreatedBy = userID`
4. Cria entities `WorkoutExercise`
5. Chama `workoutRepo.Create()` (transação)
6. Retorna workout criado

**Arquivo a criar:**
- `internal/kinetria/domain/workouts/uc_update_workout.go`

Lógica:
1. Recebe userID + workoutID + `UpdateWorkoutInput`
2. Valida ownership (`workout.CreatedBy == userID`)
3. Valida inputs (similar ao create)
4. Busca workout atual
5. Atualiza campos modificados
6. Chama `workoutRepo.Update()` (transação)
7. Retorna workout atualizado

**Arquivo a criar:**
- `internal/kinetria/domain/workouts/uc_delete_workout.go`

Lógica:
1. Recebe userID + workoutID
2. Valida ownership
3. Verifica se tem sessions ativas (via repository)
4. Chama `workoutRepo.Delete()` (soft ou hard delete)
5. Retorna sucesso

---

### 5.6) HTTP Layer

**Arquivo a modificar:**
- `internal/kinetria/gateways/http/handler_workouts.go`

Adicionar handlers:
```go
type WorkoutsHandler struct {
    listWorkoutsUC   *workouts.ListWorkoutsUC
    getWorkoutUC     *workouts.GetWorkoutUC
    createWorkoutUC  *workouts.CreateWorkoutUC  // NOVO
    updateWorkoutUC  *workouts.UpdateWorkoutUC  // NOVO
    deleteWorkoutUC  *workouts.DeleteWorkoutUC  // NOVO
}

// DTOs
type CreateWorkoutRequest struct {
    Name        string                      `json:"name"`
    Description *string                     `json:"description"`
    Type        string                      `json:"type"`        // "FORÇA", "HIPERTROFIA", etc
    Intensity   string                      `json:"intensity"`   // "BAIXA", "MODERADA", "ALTA"
    Duration    int                         `json:"duration"`    // minutos
    ImageURL    *string                     `json:"imageUrl"`
    Exercises   []CreateWorkoutExerciseDTO  `json:"exercises"`
}

type CreateWorkoutExerciseDTO struct {
    ExerciseID string  `json:"exerciseId"`
    Sets       int     `json:"sets"`
    Reps       string  `json:"reps"`
    RestTime   int     `json:"restTime"`
    Weight     *int    `json:"weight"`
    OrderIndex int     `json:"orderIndex"`
}

type UpdateWorkoutRequest struct {
    Name        *string                     `json:"name"`
    Description *string                     `json:"description"`
    Type        *string                     `json:"type"`
    Intensity   *string                     `json:"intensity"`
    Duration    *int                        `json:"duration"`
    ImageURL    *string                     `json:"imageUrl"`
    Exercises   []CreateWorkoutExerciseDTO  `json:"exercises"` // Substituir todos
}

type WorkoutResponse struct {
    Data WorkoutDTO `json:"data"`
}
```

**Handlers:**
- `POST /api/v1/workouts` → `HandleCreateWorkout()`
- `PUT /api/v1/workouts/:id` → `HandleUpdateWorkout()`
- `DELETE /api/v1/workouts/:id` → `HandleDeleteWorkout()`

**Validações no handler:**
- Name: 3-255 caracteres
- Type: enum válido (FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO)
- Intensity: enum válido (BAIXA, MODERADA, ALTA)
- Duration: 1-300 minutos
- Sets: 1-10
- RestTime: 0-600 segundos
- OrderIndex: sem duplicatas, sequencial (1, 2, 3...)
- Exercises: 1-20 exercícios

---

### 5.7) Router

**Arquivo a modificar:**
- `internal/kinetria/gateways/http/router.go`

Adicionar rotas protegidas:
```go
r.Route("/api/v1/workouts", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    
    r.Get("/", workoutsHandler.HandleListWorkouts)
    r.Post("/", workoutsHandler.HandleCreateWorkout)      // NOVO
    r.Get("/{id}", workoutsHandler.HandleGetWorkout)
    r.Put("/{id}", workoutsHandler.HandleUpdateWorkout)   // NOVO
    r.Delete("/{id}", workoutsHandler.HandleDeleteWorkout) // NOVO
})
```

---

### 5.8) Dependency Injection

**Arquivo a modificar:**
- `cmd/kinetria/api/main.go`

Registrar novos use cases:
```go
fx.Provide(
    // Use cases existentes
    workouts.NewListWorkoutsUC,
    workouts.NewGetWorkoutUC,
    
    // Novos use cases
    workouts.NewCreateWorkoutUC,
    workouts.NewUpdateWorkoutUC,
    workouts.NewDeleteWorkoutUC,
    
    // Handler (já existe, apenas injetar novos UCs)
    fx.Annotate(
        http.NewWorkoutsHandler,
        fx.As(new(http.WorkoutsHandler)),
    ),
),
```

---

### 5.9) Modificar GET /workouts

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/queries/workouts.sql`

Atualizar query `ListWorkoutsByUserID` para filtrar por ownership:
```sql
-- name: ListWorkoutsByUserID :many
SELECT * FROM workouts
WHERE (created_by = $1 OR created_by IS NULL) -- Templates + customizados do usuário
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

---

## 6) Risks / Edge Cases

### Ownership
- **Workout template (created_by = NULL):** Não pode ser editado/deletado por usuários
- **Validação:** Sempre verificar `workout.CreatedBy == userID` antes de update/delete
- **Mitigação:** Retornar 403 Forbidden se tentar editar workout de outro usuário

### Transações
- **Create/Update:** Deve ser atômico (workout + workout_exercises)
- **Rollback:** Se falhar ao criar workout_exercises, reverter criação do workout
- **Mitigação:** Usar transações SQL (BEGIN/COMMIT/ROLLBACK)

### Deleção
- **Sessions ativas:** Não permitir deletar workout com sessions ativas
- **Sessions completadas:** Permitir deletar (soft delete mantém referência)
- **Cascade:** `workout_exercises` são deletados automaticamente (ON DELETE CASCADE)
- **Mitigação:** Validar `HasActiveSessions()` antes de deletar

### Validações
- **Exercises não existem:** Validar que todos os `exerciseID` existem na biblioteca
- **Order index duplicado:** Validar que não há duplicatas (ou reordenar automaticamente)
- **Sets/RestTime negativos:** Validar valores positivos
- **Workout vazio:** Exigir pelo menos 1 exercise
- **Mitigação:** Validações no use case, retornar 400 com mensagem clara

### Soft Delete
- **Queries:** Sempre filtrar `deleted_at IS NULL`
- **Restauração:** Não implementar por enquanto (fora de escopo)
- **Cleanup:** Considerar job para deletar permanentemente após X dias (fora de escopo)

### Performance
- **Transação longa:** Create/Update com muitos exercises pode ser lento
- **Mitigação:** Limitar a 20 exercises por workout
- **Índices:** `idx_workouts_created_by` para buscar workouts do usuário

---

## 7) Suggested Implementation Strategy (alto nível, sem código)

### Etapa 1: Migration (30min)
1. Criar migration `013_add_workout_ownership.sql`
2. Adicionar coluna `created_by` (nullable)
3. Adicionar coluna `deleted_at` (opcional, se usar soft delete)
4. Adicionar índices

### Etapa 2: Domain (30min)
1. Atualizar `entities.Workout` com `CreatedBy` e `DeletedAt`
2. Criar structs `CreateWorkoutInput`, `UpdateWorkoutInput`, `WorkoutExerciseInput`

### Etapa 3: Repository (2h)
1. Adicionar métodos em `ports.WorkoutRepository`
2. Criar queries em `queries/workouts.sql`:
   - `CreateWorkout`
   - `UpdateWorkout`
   - `SoftDeleteWorkout` (ou `HardDeleteWorkout`)
   - `HasActiveSessions`
   - `CreateWorkoutExercise`
   - `DeleteWorkoutExercises`
3. Rodar `make sqlc`
4. Implementar métodos em `workout_repository.go` com transações

### Etapa 4: Use Cases (2-3h)
1. Criar `uc_create_workout.go`:
   - Valida inputs
   - Valida que exercises existem
   - Cria workout + workout_exercises (transação)
2. Criar `uc_update_workout.go`:
   - Valida ownership
   - Valida inputs
   - Atualiza workout + recria workout_exercises (transação)
3. Criar `uc_delete_workout.go`:
   - Valida ownership
   - Verifica sessions ativas
   - Deleta workout (soft ou hard)

### Etapa 5: HTTP Handler (1-2h)
1. Atualizar `handler_workouts.go` com novos handlers
2. Criar DTOs (`CreateWorkoutRequest`, `UpdateWorkoutRequest`)
3. Implementar validações de input
4. Mapear DTOs para entities

### Etapa 6: Routing e DI (15min)
1. Registrar rotas em `router.go`
2. Registrar use cases em `main.go` (Fx)

### Etapa 7: Atualizar GET /workouts (30min)
1. Modificar query `ListWorkoutsByUserID` para filtrar por ownership
2. Testar que retorna templates + workouts customizados do usuário

### Etapa 8: Testes (2-3h)
1. Unit tests para use cases (mock repository)
2. Integration tests para endpoints (DB real)
3. Edge cases: ownership, sessions ativas, validações

---

## 8) Handoff Notes to Plan

### Assunções feitas
- Usar `created_by UUID` (FK para users) para ownership
- Workouts com `created_by = NULL` são templates (read-only)
- Usar soft delete (`deleted_at`) para preservar histórico
- Não permitir deletar workout com sessions ativas
- Update substitui todos os exercises (não é merge)
- Exigir pelo menos 1 exercise ao criar workout

### Dependências
- **Decisão de negócio:**
  - Soft delete ou hard delete?
  - Permitir atualizar workout com sessions completadas?
  - Limites de exercises por workout (max 20?)
- **Decisão técnica:**
  - Validar order_index sem duplicatas ou reordenar automaticamente?

### Recomendações para Plano de Testes

**Unit tests:**
- `CreateWorkoutUC`: cria workout + exercises, valida inputs, valida que exercises existem
- `UpdateWorkoutUC`: atualiza workout, valida ownership, valida inputs
- `DeleteWorkoutUC`: deleta workout, valida ownership, bloqueia se tem sessions ativas

**Integration tests:**
- `POST /workouts`: retorna 201 com workout criado, valida inputs inválidos (400)
- `PUT /workouts/:id`: retorna 200 com workout atualizado, valida ownership (403)
- `DELETE /workouts/:id`: retorna 204, valida ownership (403), bloqueia se tem sessions ativas (409)

**Edge cases:**
- Criar workout sem exercises (400)
- Criar workout com exerciseID inválido (400)
- Atualizar workout de outro usuário (403)
- Deletar workout com session ativa (409)
- Order index duplicado (400)
- Transação falha (rollback correto)

**Performance tests:**
- Criar workout com 20 exercises (tempo de resposta)
- Atualizar workout com muitos exercises (tempo de transação)

### Próximos passos
1. Responder perguntas da seção 2
2. Criar plano detalhado com tasks granulares
3. Implementar migration + domain
4. Implementar repository com transações
5. Implementar use cases + handlers
6. Testes
