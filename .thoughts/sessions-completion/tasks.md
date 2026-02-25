# Tasks — Sessions Completion

> **Feature**: RecordSet, FinishSession, AbandonSession  
> **Dependências**: POST /sessions (já implementado)

---

## 📋 Ordem de Implementação

Seguir ordem sequencial (T01 → T10) para minimizar bloqueios.

---

## T01 — Adicionar erros de domínio

### Objetivo
Criar erros específicos para as novas operações.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/errors/errors.go`

### Implementação (passos)
1. Adicionar novos erros:
```go
var (
    // Existentes: ErrNotFound, ErrConflict, ErrActiveSessionExists, ErrWorkoutNotFound
    
    // Novos
    ErrSessionNotActive      = errors.New("session is not active")
    ErrSessionAlreadyClosed  = errors.New("session is already closed")
    ErrSetAlreadyRecorded    = errors.New("set already recorded")
    ErrExerciseNotFound      = errors.New("exercise not found")
)
```

### Critério de aceite
- [ ] Erros compilam sem erro
- [ ] Podem ser usados com `errors.Is()`
- [ ] `make lint` passa

---

## T02 — Criar queries SQLC

### Objetivo
Escrever queries SQL para set_records, sessions (update), exercises.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/repositories/queries/set_records.sql` (novo)
- `internal/kinetria/gateways/repositories/queries/sessions.sql` (atualizar)
- `internal/kinetria/gateways/repositories/queries/exercises.sql` (novo)

### Implementação (passos)

1. **Criar `set_records.sql`**:
```sql
-- name: CreateSetRecord :exec
INSERT INTO set_records (id, session_id, exercise_id, set_number, weight, reps, status, recorded_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindSetRecordBySessionExerciseSet :one
SELECT id, session_id, exercise_id, set_number, weight, reps, status, recorded_at
FROM set_records
WHERE session_id = $1 AND exercise_id = $2 AND set_number = $3;
```

2. **Atualizar `sessions.sql`** (adicionar):
```sql
-- name: FindSessionByID :one
SELECT id, user_id, workout_id, started_at, finished_at, status, notes, created_at, updated_at
FROM sessions
WHERE id = $1;

-- name: UpdateSessionStatus :exec
UPDATE sessions
SET status = $2, finished_at = $3, notes = $4, updated_at = $5
WHERE id = $1;
```

3. **Criar `exercises.sql`**:
```sql
-- name: ExistsExerciseByIDAndWorkoutID :one
SELECT EXISTS(
    SELECT 1 FROM exercises WHERE id = $1 AND workout_id = $2
) AS exists;
```

4. Gerar código:
```bash
make sqlc
```

### Critério de aceite
- [ ] Queries SQL criadas
- [ ] `make sqlc` executa sem erro
- [ ] Código Go gerado compila
- [ ] `make lint` passa

---

## T03 — Criar interfaces de repositório (ports)

### Objetivo
Definir contratos para SetRecordRepository e ExerciseRepository, atualizar SessionRepository.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/ports/repositories.go`

### Implementação (passos)

1. **Adicionar SetRecordRepository**:
```go
type SetRecordRepository interface {
    Create(ctx context.Context, setRecord *entities.SetRecord) error
    FindBySessionExerciseSet(ctx context.Context, sessionID, exerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error)
}
```

2. **Adicionar ExerciseRepository**:
```go
type ExerciseRepository interface {
    ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error)
}
```

3. **Atualizar SessionRepository** (adicionar métodos):
```go
type SessionRepository interface {
    // Existentes: Create, FindActiveByUserID
    FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error)
    UpdateStatus(ctx context.Context, sessionID uuid.UUID, status vos.SessionStatus, finishedAt *time.Time, notes string) error
}
```

### Critério de aceite
- [ ] Interfaces compilam
- [ ] Comentários Godoc
- [ ] `make mocks` gera mocks
- [ ] `make lint` passa

---

## T04 — Implementar Use Case: RecordSet

### Objetivo
Criar use case para registrar séries durante sessão ativa.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/sessions/uc_record_set.go` (novo)

### Implementação (passos)

1. **Criar estrutura**:
```go
type RecordSetInput struct {
    UserID     uuid.UUID
    SessionID  uuid.UUID
    ExerciseID uuid.UUID
    SetNumber  int
    Weight     int     // gramas
    Reps       int
    Status     vos.SetRecordStatus
}

type RecordSetOutput struct {
    SetRecord entities.SetRecord
}

type RecordSetUseCase struct {
    sessionRepo   ports.SessionRepository
    setRecordRepo ports.SetRecordRepository
    exerciseRepo  ports.ExerciseRepository
    auditLogRepo  ports.AuditLogRepository
}
```

2. **Implementar Execute**:
```go
func (uc *RecordSetUseCase) Execute(ctx context.Context, input RecordSetInput) (RecordSetOutput, error) {
    // 1. Validar inputs
    if input.SessionID == uuid.Nil || input.ExerciseID == uuid.Nil {
        return RecordSetOutput{}, errors.ErrMalformedParameters
    }
    if input.SetNumber < 1 || input.Weight < 0 || input.Reps < 0 {
        return RecordSetOutput{}, errors.ErrMalformedParameters
    }
    
    // 2. Buscar session e validar ownership
    session, err := uc.sessionRepo.FindByID(ctx, input.SessionID)
    if err != nil || session == nil {
        return RecordSetOutput{}, errors.ErrNotFound
    }
    if session.UserID != input.UserID {
        return RecordSetOutput{}, errors.ErrNotFound // não vazar que existe
    }
    
    // 3. Validar session ativa
    if session.Status != vos.SessionStatusActive {
        return RecordSetOutput{}, errors.ErrSessionNotActive
    }
    
    // 4. Validar exercise pertence ao workout
    exists, err := uc.exerciseRepo.ExistsByIDAndWorkoutID(ctx, input.ExerciseID, session.WorkoutID)
    if err != nil {
        return RecordSetOutput{}, fmt.Errorf("failed to check exercise: %w", err)
    }
    if !exists {
        return RecordSetOutput{}, errors.ErrExerciseNotFound
    }
    
    // 5. Verificar duplicação
    existing, err := uc.setRecordRepo.FindBySessionExerciseSet(ctx, input.SessionID, input.ExerciseID, input.SetNumber)
    if err != nil && err != sql.ErrNoRows {
        return RecordSetOutput{}, fmt.Errorf("failed to check duplicate: %w", err)
    }
    if existing != nil {
        return RecordSetOutput{}, errors.ErrSetAlreadyRecorded
    }
    
    // 6. Criar SetRecord
    now := time.Now()
    setRecord := entities.SetRecord{
        ID:         uuid.New(),
        SessionID:  input.SessionID,
        ExerciseID: input.ExerciseID,
        SetNumber:  input.SetNumber,
        Weight:     input.Weight,
        Reps:       input.Reps,
        Status:     input.Status,
        RecordedAt: now,
    }
    
    // 7. Persistir
    if err := uc.setRecordRepo.Create(ctx, &setRecord); err != nil {
        return RecordSetOutput{}, fmt.Errorf("failed to create set record: %w", err)
    }
    
    // 8. Audit log
    auditEntry := entities.AuditLog{
        ID:         uuid.New(),
        UserID:     input.UserID,
        EntityType: "set_record",
        EntityID:   setRecord.ID,
        Action:     "created",
        ActionData: json.Marshal(setRecord),
        OccurredAt: now,
    }
    _ = uc.auditLogRepo.Append(ctx, &auditEntry)
    
    return RecordSetOutput{SetRecord: setRecord}, nil
}
```

### Critério de aceite
- [ ] Use case compila
- [ ] Todas as validações implementadas
- [ ] Audit log registrado
- [ ] Comentários Godoc
- [ ] `make lint` passa

---

## T05 — Implementar Use Case: FinishSession

### Objetivo
Criar use case para finalizar sessão ativa.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/sessions/uc_finish_session.go` (novo)

### Implementação (passos)

1. **Criar estrutura**:
```go
type FinishSessionInput struct {
    UserID    uuid.UUID
    SessionID uuid.UUID
    Notes     string // opcional
}

type FinishSessionOutput struct {
    Session entities.Session
}

type FinishSessionUseCase struct {
    sessionRepo  ports.SessionRepository
    auditLogRepo ports.AuditLogRepository
}
```

2. **Implementar Execute** (similar a RecordSet, mas mais simples):
- Buscar session e validar ownership
- Validar status = "active"
- Atualizar: status → "completed", finishedAt → now(), notes
- Audit log com action "completed"

### Critério de aceite
- [ ] Use case compila
- [ ] Validações implementadas
- [ ] Audit log registrado
- [ ] `make lint` passa

---

## T06 — Implementar Use Case: AbandonSession

### Objetivo
Criar use case para abandonar sessão ativa.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/sessions/uc_abandon_session.go` (novo)

### Implementação (passos)

Similar a FinishSession, mas:
- status → "abandoned"
- notes não é atualizado (mantém vazio)
- Audit log com action "abandoned"

### Critério de aceite
- [ ] Use case compila
- [ ] Validações implementadas
- [ ] Audit log registrado
- [ ] `make lint` passa

---

## T07 — Criar testes unitários dos Use Cases

### Objetivo
Testar os 3 use cases com table-driven tests.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/sessions/uc_record_set_test.go` (novo)
- `internal/kinetria/domain/sessions/uc_finish_session_test.go` (novo)
- `internal/kinetria/domain/sessions/uc_abandon_session_test.go` (novo)

### Implementação (passos)

Para cada use case, criar testes cobrindo:
- **RecordSet**: sucesso, session não encontrada, session não ativa, exercise não encontrado, set duplicado, ownership
- **FinishSession**: sucesso com/sem notes, session não encontrada, session já fechada, ownership
- **AbandonSession**: sucesso, session não encontrada, session já fechada, ownership

### Critério de aceite
- [ ] Testes cobrem happy + sad paths
- [ ] Usam inline mocks
- [ ] `make test` passa
- [ ] Cobertura > 80%

---

## T08 — Implementar Handlers HTTP

### Objetivo
Criar handlers para os 3 endpoints.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/http/handler_sessions.go` (atualizar)
- `internal/kinetria/gateways/http/dto.go` (atualizar)

### Implementação (passos)

1. **Adicionar DTOs**:
```go
type RecordSetRequestDTO struct {
    ExerciseID uuid.UUID `json:"exerciseId" validate:"required,uuid"`
    SetNumber  int       `json:"setNumber" validate:"required,min=1"`
    Weight     float64   `json:"weight" validate:"required,min=0"`
    Reps       int       `json:"reps" validate:"required,min=0"`
    Status     string    `json:"status" validate:"required,oneof=completed skipped"`
}

type FinishSessionRequestDTO struct {
    Notes string `json:"notes"`
}

type SetRecordResponseDTO struct {
    ID         uuid.UUID `json:"id"`
    SessionID  uuid.UUID `json:"sessionId"`
    ExerciseID uuid.UUID `json:"exerciseId"`
    SetNumber  int       `json:"setNumber"`
    Weight     int       `json:"weight"`
    Reps       int       `json:"reps"`
    Status     string    `json:"status"`
    RecordedAt time.Time `json:"recordedAt"`
}
```

2. **Implementar handlers**:
- `RecordSet(w http.ResponseWriter, r *http.Request)`
- `FinishSession(w http.ResponseWriter, r *http.Request)`
- `AbandonSession(w http.ResponseWriter, r *http.Request)`

3. **Registrar rotas** (em `router.go`):
```go
r.With(AuthMiddleware).Post("/sessions/{sessionId}/sets", handler.RecordSet)
r.With(AuthMiddleware).Patch("/sessions/{sessionId}/finish", handler.FinishSession)
r.With(AuthMiddleware).Patch("/sessions/{sessionId}/abandon", handler.AbandonSession)
```

### Critério de aceite
- [ ] Handlers compilam
- [ ] DTOs com validação
- [ ] Error mapping correto (404, 409, 422)
- [ ] Rotas registradas
- [ ] `make lint` passa

---

## T09 — Implementar Repository Adapters

### Objetivo
Implementar SetRecordRepository e ExerciseRepository, atualizar SessionRepository.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/repositories/set_record_repository.go` (novo)
- `internal/kinetria/gateways/repositories/exercise_repository.go` (novo)
- `internal/kinetria/gateways/repositories/session_repository.go` (atualizar)

### Implementação (passos)

Seguir padrão existente (usar `*sql.DB`, mapear entities ↔ SQLC params).

### Critério de aceite
- [ ] Repositories compilam
- [ ] Seguem padrão existente
- [ ] `make lint` passa

---

## T10 — Wiring com Fx DI

### Objetivo
Registrar use cases, repositories e handlers no `main.go`.

### Arquivos/pacotes prováveis
- `cmd/kinetria/api/main.go`

### Implementação (passos)

```go
// Repositories
fx.Provide(fx.Annotate(repositories.NewSetRecordRepository, fx.As(new(ports.SetRecordRepository)))),
fx.Provide(fx.Annotate(repositories.NewExerciseRepository, fx.As(new(ports.ExerciseRepository)))),

// Use Cases
fx.Provide(domainsessions.NewRecordSetUC),
fx.Provide(domainsessions.NewFinishSessionUC),
fx.Provide(domainsessions.NewAbandonSessionUC),

// Handler já existe, apenas atualizar construtor para receber novos use cases
```

### Critério de aceite
- [ ] Aplicação compila
- [ ] `make build` passa
- [ ] Aplicação inicia sem erro

---

## Resumo

**Total de tasks**: 10  
**Estimativa**: 2-3 dias (1 dev experiente)  
**Ordem**: T01 → T10 (sequencial)
