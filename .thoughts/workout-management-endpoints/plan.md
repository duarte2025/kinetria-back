# Plan — Workout Management Endpoints

## 1) Inputs usados
- `.thoughts/workout-management-endpoints/research-report.md`

## 2) AS-IS (resumo)

### Estrutura atual
- Tabela `workouts` não possui campo `created_by` (ownership)
- Tabela `workouts` não possui campo `deleted_at` (soft delete)
- **FALTA:** Distinguir workouts template vs customizados
- `WorkoutRepository` possui apenas métodos de leitura (List, GetByID)
- **FALTA:** Métodos Create, Update, Delete
- Endpoints existentes: GET /workouts, GET /workouts/:id
- **FALTA:** POST, PUT, DELETE

### Relacionamento atual
```
users (1) ----< workouts (N)  (FALTA FK created_by)
                  ↓
            workout_exercises (N:N com exercises)
                  ↓
              set_records
```

### Fluxo atual
- GET /workouts retorna todos os workouts (sem filtro de ownership)
- Não há distinção entre templates e workouts customizados
- Não há endpoints para criar/editar/deletar

## 3) TO-BE (proposta)

### Interface HTTP
**Endpoints:**
- `POST /api/v1/workouts` — Criar workout customizado
- `PUT /api/v1/workouts/:id` — Atualizar workout customizado
- `DELETE /api/v1/workouts/:id` — Deletar workout customizado

**Autenticação:** Obrigatória (JWT Bearer) em todos os endpoints

### Contratos

**POST /api/v1/workouts**
```
Request:
{
  "name": "Treino A - Peito e Tríceps",
  "description": "Foco em hipertrofia",
  "type": "HIPERTROFIA",
  "intensity": "ALTA",
  "duration": 60,
  "imageUrl": "https://cdn.example.com/workout.jpg",
  "exercises": [
    {
      "exerciseId": "uuid",
      "sets": 4,
      "reps": "8-12",
      "restTime": 90,
      "weight": 80000,
      "orderIndex": 1
    }
  ]
}

Response 201:
{
  "data": {
    "id": "uuid",
    "name": "Treino A - Peito e Tríceps",
    "description": "Foco em hipertrofia",
    "type": "HIPERTROFIA",
    "intensity": "Alta",
    "duration": 60,
    "imageUrl": "https://cdn.example.com/workout.jpg",
    "exercises": [...]
  }
}

Response 400:
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "at least one exercise is required"
  }
}
```

**PUT /api/v1/workouts/:id**
```
Request:
{
  "name": "Treino A - Atualizado",
  "exercises": [
    {
      "exerciseId": "uuid",
      "sets": 5,
      "reps": "10",
      "restTime": 60,
      "weight": 85000,
      "orderIndex": 1
    }
  ]
}

Response 200:
{
  "data": {
    "id": "uuid",
    "name": "Treino A - Atualizado",
    ...
  }
}

Response 403:
{
  "error": {
    "code": "FORBIDDEN",
    "message": "you can only update your own workouts"
  }
}
```

**DELETE /api/v1/workouts/:id**
```
Response 204: (sem body)

Response 403:
{
  "error": {
    "code": "FORBIDDEN",
    "message": "you can only delete your own workouts"
  }
}

Response 409:
{
  "error": {
    "code": "CONFLICT",
    "message": "cannot delete workout with active sessions"
  }
}
```

### Persistência

**Migration 013:** `013_add_workout_ownership.sql`
```sql
ALTER TABLE workouts 
ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE CASCADE,
ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_workouts_created_by ON workouts(created_by);
CREATE INDEX idx_workouts_deleted_at ON workouts(deleted_at) WHERE deleted_at IS NULL;
```

**Queries SQLC:**
- `CreateWorkout` — INSERT workout
- `UpdateWorkout` — UPDATE workout (validar created_by)
- `SoftDeleteWorkout` — UPDATE deleted_at (validar created_by)
- `HasActiveSessions` — Verificar se tem sessions ativas
- `CreateWorkoutExercise` — INSERT workout_exercise
- `DeleteWorkoutExercises` — DELETE workout_exercises por workout_id

**Transações:**
- Create: workout + workout_exercises (atômico)
- Update: workout + delete old exercises + insert new exercises (atômico)

### Domain Layer

**Entity Workout (adicionar campos):**
```go
CreatedBy *uuid.UUID // NULL = template, NOT NULL = customizado
DeletedAt *time.Time // Soft delete
```

**Structs de input:**
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

**Use Cases:**
- `CreateWorkoutUC` — Valida inputs, cria workout + exercises (transação)
- `UpdateWorkoutUC` — Valida ownership, atualiza workout + exercises (transação)
- `DeleteWorkoutUC` — Valida ownership, verifica sessions ativas, soft delete

### Validações

**Name:**
- Mínimo: 3 caracteres
- Máximo: 255 caracteres
- Não vazio

**Type:**
- Enum válido: FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO

**Intensity:**
- Enum válido: BAIXA, MODERADA, ALTA

**Duration:**
- Mínimo: 1 minuto
- Máximo: 300 minutos

**Exercises:**
- Mínimo: 1 exercise
- Máximo: 20 exercises
- Todos os exerciseID devem existir na biblioteca

**Sets:**
- Mínimo: 1
- Máximo: 10

**RestTime:**
- Mínimo: 0 segundos
- Máximo: 600 segundos (10 minutos)

**OrderIndex:**
- Sem duplicatas
- Sequencial (gaps permitidos: 1, 2, 5 é válido)

### Regras de Negócio

**Ownership:**
- `created_by = NULL` → Template (read-only, não pode ser editado/deletado)
- `created_by = userID` → Customizado (pode ser editado/deletado pelo dono)
- Validar ownership antes de update/delete (retornar 403 se não for dono)

**Deleção:**
- Soft delete (atualizar `deleted_at`)
- Bloquear se houver sessions ativas (retornar 409)
- Permitir deletar se houver apenas sessions completadas/abandonadas
- Cascade: `workout_exercises` são deletados automaticamente (ON DELETE CASCADE)

**Atualização:**
- PUT = atualização completa (substituir todos os exercises)
- Permitir atualizar workout com sessions completadas (sem versionamento)
- Validar ownership antes de atualizar

**GET /workouts:**
- Retornar templates (created_by = NULL) + workouts customizados do usuário (created_by = userID)
- Filtrar `deleted_at IS NULL`

### Observabilidade
- Logs de erro em caso de falha na transação
- Retornar erros de domínio apropriados (validation, forbidden, conflict, internal)

## 4) Decisões e Assunções

1. **Ownership model:** `created_by UUID` (nullable). NULL = template, NOT NULL = customizado.
2. **Soft delete:** Usar `deleted_at` para preservar histórico.
3. **Bloquear deleção:** Se houver sessions ativas (status = 'active').
4. **PUT = replace:** Substituir todos os exercises (não é merge).
5. **Exigir 1+ exercise:** Workout vazio não é válido.
6. **Validações:** Sets 1-10, RestTime 0-600s, Exercises 1-20, Order index sem duplicatas (gaps ok).
7. **Transações:** Create/Update devem ser atômicos (workout + workout_exercises).
8. **Permitir atualizar com sessions completadas:** Sem versionamento na v1.

## 5) Riscos / Edge Cases

### Ownership
- **Risco:** Usuário tenta editar/deletar workout de outro usuário
- **Mitigação:** Validar `workout.CreatedBy == userID`, retornar 403 Forbidden

### Transações
- **Risco:** Falha ao criar workout_exercises após criar workout
- **Mitigação:** Usar transações SQL (BEGIN/COMMIT/ROLLBACK)
- **Risco:** Transação longa com muitos exercises
- **Mitigação:** Limitar a 20 exercises por workout

### Deleção
- **Risco:** Deletar workout com sessions ativas quebra integridade
- **Mitigação:** Validar `HasActiveSessions()` antes de deletar, retornar 409 Conflict
- **Risco:** Soft delete não limpa dados antigos
- **Mitigação:** Considerar job de cleanup (fora de escopo v1)

### Validações
- **Risco:** ExerciseID não existe na biblioteca
- **Mitigação:** Validar que todos os exerciseID existem antes de criar/atualizar
- **Risco:** Order index duplicado
- **Mitigação:** Validar no use case, retornar 400
- **Risco:** Workout vazio (sem exercises)
- **Mitigação:** Validar no use case, retornar 400

### Soft Delete
- **Risco:** Queries esquecem de filtrar `deleted_at IS NULL`
- **Mitigação:** Sempre incluir filtro em queries de leitura
- **Risco:** Restauração de workouts deletados
- **Mitigação:** Não implementar por enquanto (fora de escopo)

### Performance
- **Risco:** Criar/atualizar workout com 20 exercises pode ser lento
- **Mitigação:** Índices, limitar a 20 exercises
- **Risco:** Transação longa bloqueia outras operações
- **Mitigação:** Manter transação curta, apenas operações essenciais

## 6) Rollout / Compatibilidade

### Fase 1: Migration
1. Aplicar migration `013_add_workout_ownership.sql`
2. Adicionar colunas `created_by` e `deleted_at`
3. Workouts existentes ficam com `created_by = NULL` (templates)

### Fase 2: Repository
1. Adicionar métodos Create, Update, Delete no port
2. Implementar queries SQLC
3. Implementar métodos com transações

### Fase 3: Use Cases
1. Criar use cases para Create, Update, Delete
2. Implementar validações
3. Implementar lógica de ownership

### Fase 4: HTTP Layer
1. Criar handlers para POST, PUT, DELETE
2. Criar DTOs
3. Registrar rotas

### Fase 5: Atualizar GET /workouts
1. Modificar query para filtrar por ownership
2. Filtrar `deleted_at IS NULL`

### Fase 6: Testes
1. Unit tests para use cases
2. Integration tests para endpoints
3. Edge cases (ownership, sessions ativas, transações)

### Compatibilidade
- ✅ Migration é aditiva (não quebra código existente)
- ✅ Workouts existentes ficam como templates (created_by = NULL)
- ✅ GET /workouts continua funcionando (retorna templates + customizados)
- ✅ Soft delete preserva histórico (sessions antigas continuam válidas)
