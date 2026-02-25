# Plan — Sessions Completion (RecordSet, FinishSession, AbandonSession)

## 1) Inputs usados

### Artefatos de Research
- `.thoughts/mvp-userflow/api-contract.yaml` — Contrato OpenAPI (RecordSetRequest, FinishSessionRequest, endpoints)
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — Modelo de dados, audit log obrigatório
- `.thoughts/sessions/plan.md` — Plano de StartSession (já implementado)
- `.thoughts/sessions/execution-report.md` — Implementação existente de StartSession

### Análise do Repositório (AS-IS)
- ✅ `internal/kinetria/domain/entities/session.go` — Session entity implementada
- ✅ `internal/kinetria/domain/entities/set_record.go` — SetRecord entity implementada
- ✅ `internal/kinetria/domain/vos/session_status.go` — SessionStatus (active, completed, abandoned)
- ✅ `internal/kinetria/domain/vos/set_record_status.go` — SetRecordStatus (completed, skipped)
- ✅ `internal/kinetria/domain/sessions/uc_start_session.go` — StartSession use case implementado
- ✅ `internal/kinetria/gateways/http/handler_sessions.go` — Handler de sessions (POST /sessions)
- ✅ `internal/kinetria/gateways/repositories/session_repository.go` — SessionRepository implementado
- ✅ Migrations aplicadas (sessions, set_records, audit_log)

---

## 2) AS-IS (resumo)

### Estado Atual
- ✅ **POST /sessions** implementado e testado
- ✅ Entidades Session e SetRecord existem
- ✅ SessionRepository com Create e FindActiveByUserID
- ✅ Audit log funcionando para criação de sessions
- ✅ JWT middleware autenticando requests
- ❌ **Faltam**: RecordSet, FinishSession, AbandonSession

### Gaps Identificados
1. ❌ SetRecordRepository não existe
2. ❌ Queries SQLC para set_records não existem
3. ❌ Use cases RecordSet, FinishSession, AbandonSession não implementados
4. ❌ Handlers HTTP para os 3 endpoints não implementados
5. ❌ Validações de ownership de session não implementadas

---

## 3) TO-BE (proposta)

### 3.1) POST /sessions/:id/sets (RecordSet)

#### Request
```json
{
  "exerciseId": "c3d4e5f6-a7b8-9012-cdef-123456789012",
  "setNumber": 2,
  "weight": 82.5,
  "reps": 10,
  "status": "completed"
}
```

#### Response (201 Created)
```json
{
  "data": {
    "id": "uuid",
    "sessionId": "uuid",
    "exerciseId": "uuid",
    "setNumber": 2,
    "weight": 82500,
    "reps": 10,
    "status": "completed",
    "recordedAt": "2026-02-24T14:05:30Z"
  }
}
```

#### Validações (Use Case)
1. ✅ Session existe e pertence ao usuário
2. ✅ Session está ativa (status = "active")
3. ✅ Exercise existe e pertence ao workout da session
4. ✅ SetNumber não duplicado (UNIQUE constraint: session_id, exercise_id, set_number)
5. ✅ Weight >= 0, Reps >= 0
6. ✅ Status é "completed" ou "skipped"

#### Erros
- 401 Unauthorized — JWT inválido
- 404 Not Found — Session ou Exercise não encontrado
- 409 Conflict — Set já registrado (duplicação)
- 422 Unprocessable Entity — Validação falhou

---

### 3.2) PATCH /sessions/:id/finish (FinishSession)

#### Request (opcional)
```json
{
  "notes": "Ótimo treino, senti o peitoral muito bem."
}
```

#### Response (200 OK)
```json
{
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "workoutId": "uuid",
    "startedAt": "2026-02-24T14:00:00Z",
    "finishedAt": "2026-02-24T15:30:00Z",
    "status": "completed",
    "notes": "Ótimo treino...",
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

#### Validações (Use Case)
1. ✅ Session existe e pertence ao usuário
2. ✅ Session está ativa (status = "active")
3. ✅ Atualiza: status → "completed", finishedAt → now(), notes (se fornecido)

#### Erros
- 401 Unauthorized — JWT inválido
- 404 Not Found — Session não encontrada
- 409 Conflict — Session já finalizada/abandonada

---

### 3.3) PATCH /sessions/:id/abandon (AbandonSession)

#### Request
Sem body.

#### Response (200 OK)
```json
{
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "workoutId": "uuid",
    "startedAt": "2026-02-24T14:00:00Z",
    "finishedAt": "2026-02-24T14:15:00Z",
    "status": "abandoned",
    "notes": "",
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

#### Validações (Use Case)
1. ✅ Session existe e pertence ao usuário
2. ✅ Session está ativa (status = "active")
3. ✅ Atualiza: status → "abandoned", finishedAt → now()

#### Erros
- 401 Unauthorized — JWT inválido
- 404 Not Found — Session não encontrada
- 409 Conflict — Session já finalizada/abandonada

---

### 3.4) Persistência

#### SQLC Queries Necessárias

**set_records.sql**:
```sql
-- name: CreateSetRecord :exec
INSERT INTO set_records (id, session_id, exercise_id, set_number, weight, reps, status, recorded_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindSetRecordBySessionExerciseSet :one
SELECT id, session_id, exercise_id, set_number, weight, reps, status, recorded_at
FROM set_records
WHERE session_id = $1 AND exercise_id = $2 AND set_number = $3;
```

**sessions.sql** (adicionar):
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

**exercises.sql** (novo):
```sql
-- name: ExistsExerciseByIDAndWorkoutID :one
SELECT EXISTS(
    SELECT 1 FROM exercises WHERE id = $1 AND workout_id = $2
) AS exists;
```

#### Repository Interfaces (Ports)

```go
type SetRecordRepository interface {
    Create(ctx context.Context, setRecord *entities.SetRecord) error
    FindBySessionExerciseSet(ctx context.Context, sessionID, exerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error)
}

type SessionRepository interface {
    // Já existem: Create, FindActiveByUserID
    FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error)
    UpdateStatus(ctx context.Context, sessionID uuid.UUID, status vos.SessionStatus, finishedAt *time.Time, notes string) error
}

type ExerciseRepository interface {
    ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error)
}
```

---

### 3.5) Audit Log

**Pattern**: registrar todas as mutações.

**RecordSet**:
```go
auditEntry := &entities.AuditLog{
    UserID:     userID,
    EntityType: "set_record",
    EntityID:   setRecord.ID,
    Action:     "created",
    ActionData: json.Marshal(setRecord),
    OccurredAt: time.Now(),
}
```

**FinishSession**:
```go
auditEntry := &entities.AuditLog{
    UserID:     session.UserID,
    EntityType: "session",
    EntityID:   session.ID,
    Action:     "completed",
    ActionData: json.Marshal(map[string]interface{}{
        "finishedAt": session.FinishedAt,
        "notes": session.Notes,
    }),
    OccurredAt: time.Now(),
}
```

**AbandonSession**:
```go
auditEntry := &entities.AuditLog{
    UserID:     session.UserID,
    EntityType: "session",
    EntityID:   session.ID,
    Action:     "abandoned",
    ActionData: json.Marshal(map[string]interface{}{
        "finishedAt": session.FinishedAt,
    }),
    OccurredAt: time.Now(),
}
```

---

## 4) Decisões e Assunções

### Decisões Arquiteturais
1. ✅ **Use cases atômicos** — cada operação é um use case separado
2. ✅ **Ownership validation** — sempre validar que session pertence ao userID
3. ✅ **Status validation** — apenas sessions "active" podem receber sets/finish/abandon
4. ✅ **Weight em gramas** — armazenar como int (82.5kg → 82500g) para evitar float precision issues
5. ✅ **UNIQUE constraint** — (session_id, exercise_id, set_number) previne duplicação
6. ✅ **Audit log obrigatório** — todas as mutações registradas

### Assunções
1. ⚙️ **Migrations aplicadas** — tabelas sessions, set_records, exercises existem
2. ⚙️ **JWT middleware** — userID sempre disponível no contexto
3. ⚙️ **Exercises seed data** — workouts têm exercises associados
4. ⚙️ **Session ativa existe** — usuário já iniciou sessão via POST /sessions

---

## 5) Riscos / Edge Cases

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Set duplicado (client retry)** | Alta | Médio | UNIQUE constraint + validação no use case |
| **Finish/Abandon session já fechada** | Média | Baixo | Validação de status no use case → 409 Conflict |
| **Exercise não pertence ao workout** | Média | Alto | Validação obrigatória (ExistsByIDAndWorkoutID) |
| **Concorrência (2 sets simultâneos)** | Baixa | Médio | UNIQUE constraint garante atomicidade |
| **Weight negativo** | Baixa | Baixo | Validação no handler (validator) |

### Edge Cases
1. **Set com reps = 0**: válido (falha na execução)
2. **Set com weight = 0**: válido (bodyweight exercises)
3. **Finish sem notes**: válido (notes é opcional)
4. **Abandon sem sets registrados**: válido (usuário desistiu antes de começar)

---

## 6) Rollout / Compatibilidade

### Estratégia de Deploy
1. ✅ Implementar use cases + repositories
2. ✅ Implementar handlers HTTP
3. ✅ Testes unitários (use cases)
4. ✅ Testes de integração (handlers + DB)
5. ✅ Deploy em staging
6. ✅ Smoke tests (criar session → registrar set → finalizar)

### Compatibilidade
- ✅ **Backward compatible**: endpoints novos, não afetam POST /sessions
- ✅ **Database**: UNIQUE constraint já existe na migration
- ✅ **API contract**: segue OpenAPI spec

---

## 7) Próximos Passos (Pós-Implementação)

Após implementar RecordSet, FinishSession, AbandonSession:
1. **GetSession** — `GET /sessions/:id` (consultar sessão + sets registrados)
2. **ListSessions** — `GET /sessions?userId=...` (histórico de treinos)
3. **Dashboard** — `GET /dashboard` (agregação: sessão ativa + stats)
4. **Analytics** — queries no audit_log para métricas de progressão

---

**Documento gerado em**: 2026-02-24  
**Feature**: Sessions Completion (RecordSet, FinishSession, AbandonSession)  
**Status**: ✅ Pronto para implementação  
**Dependências**: POST /sessions (já implementado)
