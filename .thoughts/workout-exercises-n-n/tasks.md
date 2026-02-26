# Tasks — Workout Exercises N:N Relationship

> Este backlog deve ser executado na ordem apresentada. Cada task é atômica e testável independentemente.

---

## Grupo 1: Preparação e Setup

### T01 — Validar migration 009 em ambiente local
- **Objetivo:** Garantir que a migration roda sem erros e migra dados corretamente
- **Arquivos/pacotes prováveis:**
  - `migrations/009_refactor_exercises_to_shared_library.sql`
  - Docker Compose (postgres service)
- **Implementação (passos):**
  1. Criar dump de dados mockados na estrutura AS-IS (tabela `exercises` com workout_id)
  2. Subir banco local com `docker compose up postgres`
  3. Restaurar dump no banco local
  4. Executar migration 009: `make migrate-up` ou equivalente
  5. Validar queries de integridade:
     - `SELECT COUNT(*) FROM exercises;` (deve ter exercises deduplicados)
     - `SELECT COUNT(*) FROM workout_exercises;` (deve ter todos os vínculos)
     - `SELECT COUNT(*) FROM set_records WHERE workout_exercise_id IS NULL;` (deve ser 0)
- **Critério de aceite:**
  - Migration executa sem erros
  - Queries de integridade confirmam dados consistentes
  - Rollback manual funciona (restore de backup)

### T02 — Criar backup strategy para produção
- **Objetivo:** Documentar processo de backup/restore para deploy seguro
- **Arquivos/pacotes prováveis:**
  - `docs/deployment/migration-009-rollout.md` (novo)
  - `Makefile` (adicionar targets de backup/restore)
- **Implementação (passos):**
  1. Criar script de backup: `make db-backup`
  2. Criar script de restore: `make db-restore`
  3. Documentar passo a passo do rollout em `docs/deployment/migration-009-rollout.md`:
     - Pre-checks (verificar sessions ativas)
     - Backup
     - Aplicar migration
     - Deploy backend
     - Validação pós-deploy
     - Rollback (se necessário)
- **Critério de aceite:**
  - Scripts de backup/restore funcionando localmente
  - Documentação completa e revisada

---

## Grupo 2: Domain Layer (Entities + Ports)

### T03 — Atualizar entidade Exercise (remover campos de configuração)
- **Objetivo:** Refatorar `Exercise` para representar only metadata da biblioteca
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/entities/exercise.go`
- **Implementação (passos):**
  1. Remover campos: `WorkoutID`, `Sets`, `Reps`, `RestTime`, `Weight`, `OrderIndex`
  2. Adicionar campo: `Description string`
  3. Manter campos: `ID`, `Name`, `ThumbnailURL`, `Muscles`, `CreatedAt`, `UpdatedAt`
  4. Atualizar comentários Godoc
- **Critério de aceite:**
  - `Exercise` compila sem erros
  - Struct reflete modelo TO-BE (apenas metadata)
  - Comentários Godoc atualizados

### T04 — Criar entidade WorkoutExercise (nova)
- **Objetivo:** Criar entidade de domínio para representar vínculo workout <-> exercise
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/entities/workout_exercise.go` (novo)
- **Implementação (passos):**
  1. Criar struct `WorkoutExercise`:
     ```go
     type WorkoutExerciseID = uuid.UUID

     type WorkoutExercise struct {
         ID          WorkoutExerciseID
         WorkoutID   WorkoutID
         ExerciseID  ExerciseID
         Sets        int
         Reps        string
         RestTime    int
         Weight      int
         OrderIndex  int
         CreatedAt   time.Time
         UpdatedAt   time.Time
     }
     ```
  2. Adicionar comentários Godoc explicando o relacionamento N:N
- **Critério de aceite:**
  - `WorkoutExercise` compila sem erros
  - Todos os campos necessários presentes
  - Comentários Godoc claros

### T05 — Atualizar entidade SetRecord (referenciar workout_exercise_id)
- **Objetivo:** Alterar `SetRecord` para referenciar `WorkoutExerciseID` no lugar de `ExerciseID`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/entities/set_record.go`
- **Implementação (passos):**
  1. Remover campo: `ExerciseID`
  2. Adicionar campo: `WorkoutExerciseID WorkoutExerciseID`
  3. Atualizar comentários Godoc
- **Critério de aceite:**
  - `SetRecord` compila sem erros
  - Campo `WorkoutExerciseID` presente
  - Comentários Godoc atualizados

### T06 — Atualizar port ExerciseRepository
- **Objetivo:** Refatorar interface para operações sobre a biblioteca de exercises
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação (passos):**
  1. Remover método: `ExistsByIDAndWorkoutID`
  2. Adicionar métodos:
     ```go
     type ExerciseRepository interface {
         GetByID(ctx context.Context, id uuid.UUID) (*entities.Exercise, error)
         GetAll(ctx context.Context) ([]entities.Exercise, error)
         Create(ctx context.Context, exercise *entities.Exercise) error
         Update(ctx context.Context, exercise *entities.Exercise) error
         Delete(ctx context.Context, id uuid.UUID) error
     }
     ```
  3. Atualizar comentários Godoc
- **Critério de aceite:**
  - Interface compila sem erros
  - Métodos refletem operações sobre biblioteca compartilhada
  - Comentários Godoc atualizados

### T07 — Criar port WorkoutExerciseRepository (novo)
- **Objetivo:** Criar interface para operações sobre vínculos workout <-> exercise
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação (passos):**
  1. Adicionar interface:
     ```go
     type WorkoutExerciseRepository interface {
         GetByID(ctx context.Context, id uuid.UUID) (*entities.WorkoutExercise, error)
         GetByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]entities.WorkoutExercise, error)
         ExistsByIDAndWorkoutID(ctx context.Context, id, workoutID uuid.UUID) (bool, error)
         Create(ctx context.Context, we *entities.WorkoutExercise) error
         Update(ctx context.Context, we *entities.WorkoutExercise) error
         Delete(ctx context.Context, id uuid.UUID) error
     }
     ```
  2. Adicionar comentários Godoc explicando responsabilidade
- **Critério de aceite:**
  - Interface compila sem erros
  - Métodos cobrem CRUD + validações
  - Comentários Godoc claros

### T08 — Atualizar port SetRecordRepository
- **Objetivo:** Alterar assinatura de métodos para usar `workoutExerciseID`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação (passos):**
  1. Atualizar método `FindBySessionExerciseSet`:
     ```go
     FindBySessionExerciseSet(ctx context.Context, sessionID, workoutExerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error)
     ```
  2. Atualizar comentários Godoc
- **Critério de aceite:**
  - Interface compila sem erros
  - Assinatura reflete uso de `workoutExerciseID`
  - Comentários Godoc atualizados

---

## Grupo 3: Queries SQL (SQLC)

### T09 — Criar queries SQL para exercises (biblioteca)
- **Objetivo:** Implementar queries CRUD para tabela `exercises`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/queries/exercises.sql`
- **Implementação (passos):**
  1. Criar arquivo se não existir
  2. Implementar queries:
     - `GetExerciseByID` (SELECT por id)
     - `GetAllExercises` (SELECT all, ORDER BY name)
     - `CreateExercise` (INSERT ... RETURNING)
     - `UpdateExercise` (UPDATE ... SET name, description, thumbnail_url, muscles)
     - `DeleteExercise` (DELETE WHERE id = ?)
  3. Usar comentários SQLC padrão: `-- name: QueryName :type`
- **Critério de aceite:**
  - Queries SQL válidas (testar syntax com psql)
  - Comentários SQLC corretos
  - Cobrem CRUD completo de exercises

### T10 — Criar queries SQL para workout_exercises (nova tabela)
- **Objetivo:** Implementar queries CRUD e JOIN para tabela `workout_exercises`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/queries/workout_exercises.sql` (novo)
- **Implementação (passos):**
  1. Criar arquivo `workout_exercises.sql`
  2. Implementar queries:
     - `GetWorkoutExerciseByID` (SELECT we.* + e.* com JOIN)
     - `GetWorkoutExercisesByWorkoutID` (SELECT we.* + e.* com JOIN, ORDER BY we.order_index)
     - `ExistsWorkoutExerciseByIDAndWorkoutID` (SELECT EXISTS)
     - `CreateWorkoutExercise` (INSERT ... RETURNING)
     - `UpdateWorkoutExercise` (UPDATE sets, reps, rest_time, weight, order_index)
     - `DeleteWorkoutExercise` (DELETE WHERE id = ?)
  3. Usar JOINs para retornar metadata de `exercises` junto com configurações
- **Critério de aceite:**
  - Queries SQL válidas (testar syntax com psql)
  - JOINs retornam colunas de `exercises` + `workout_exercises`
  - Comentários SQLC corretos

### T11 — Atualizar queries SQL para set_records
- **Objetivo:** Alterar queries para usar `workout_exercise_id` no lugar de `exercise_id`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/queries/set_records.sql`
- **Implementação (passos):**
  1. Atualizar query `CreateSetRecord`:
     - Coluna: `workout_exercise_id` (no lugar de `exercise_id`)
  2. Atualizar query `FindSetRecordBySessionExerciseSet`:
     - WHERE: `workout_exercise_id = ?` (no lugar de `exercise_id = ?`)
  3. Adicionar query (se não existir): `GetSetRecordsBySessionID`
     - SELECT * FROM set_records WHERE session_id = ? ORDER BY recorded_at
- **Critério de aceite:**
  - Queries SQL válidas
  - Referenciam `workout_exercise_id` corretamente
  - Comentários SQLC atualizados

### T12 — Regenerar código SQLC
- **Objetivo:** Gerar structs e métodos Go a partir das queries SQL
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/queries/*.go` (gerados)
- **Implementação (passos):**
  1. Executar `make sqlc` (ou `sqlc generate`)
  2. Validar que arquivos `.go` foram regenerados
  3. Verificar que não há erros de compilação
  4. Revisar structs geradas (ex: `WorkoutExerciseRow`)
- **Critério de aceite:**
  - Comando `make sqlc` executa sem erros
  - Arquivos `.go` gerados dentro de `queries/`
  - Projeto compila: `go build ./...`

---

## Grupo 4: Gateways (Repository Implementations)

### T13 — Implementar ExerciseRepository (novo CRUD)
- **Objetivo:** Implementar port `ExerciseRepository` usando queries SQLC
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/exercise_repository.go`
- **Implementação (passos):**
  1. Atualizar struct `ExerciseRepository` (já existe)
  2. Implementar métodos:
     - `GetByID` → chama `q.GetExerciseByID`
     - `GetAll` → chama `q.GetAllExercises`
     - `Create` → chama `q.CreateExercise`
     - `Update` → chama `q.UpdateExercise`
     - `Delete` → chama `q.DeleteExercise`
  3. Mapear structs SQLC para entidades de domínio (adapter layer)
  4. Adicionar tratamento de erros (sql.ErrNoRows → domain.ErrNotFound)
- **Critério de aceite:**
  - Todos os métodos do port implementados
  - Mapeamento SQLC → domain correto
  - Compila sem erros

### T14 — Criar WorkoutExerciseRepository (nova implementação)
- **Objetivo:** Implementar port `WorkoutExerciseRepository` usando queries SQLC
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/workout_exercise_repository.go` (novo)
- **Implementação (passos):**
  1. Criar struct `WorkoutExerciseRepository`:
     ```go
     type WorkoutExerciseRepository struct {
         db *sql.DB
         q  *queries.Queries
     }
     ```
  2. Implementar métodos (seguir padrão de outros repositories):
     - `GetByID`
     - `GetByWorkoutID`
     - `ExistsByIDAndWorkoutID`
     - `Create`
     - `Update`
     - `Delete`
  3. Mapear structs SQLC (com JOIN) para entidades de domínio
  4. Adicionar tratamento de erros
  5. Adicionar comentários Godoc
- **Critério de aceite:**
  - Todos os métodos do port implementados
  - Mapeamento SQLC → domain correto (incluindo JOIN)
  - Compila sem erros
  - Comentários Godoc presentes

### T15 — Atualizar SetRecordRepository (usar workout_exercise_id)
- **Objetivo:** Alterar implementação para usar `workout_exercise_id`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/set_record_repository.go`
- **Implementação (passos):**
  1. Atualizar método `Create`:
     - Passar `setRecord.WorkoutExerciseID` para query SQLC (no lugar de `ExerciseID`)
  2. Atualizar método `FindBySessionExerciseSet`:
     - Passar `workoutExerciseID` para query SQLC
  3. Atualizar mapeamento de structs SQLC → domain
- **Critério de aceite:**
  - Métodos usam `WorkoutExerciseID` corretamente
  - Compila sem erros
  - Mapeamento atualizado

---

## Grupo 5: Dependency Injection (Fx Module)

### T16 — Registrar WorkoutExerciseRepository no módulo Fx
- **Objetivo:** Disponibilizar `WorkoutExerciseRepository` via DI
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/module.go` (ou equivalente)
- **Implementação (passos):**
  1. Adicionar no módulo Fx:
     ```go
     fx.Provide(
         fx.Annotate(
             NewWorkoutExerciseRepository,
             fx.As(new(ports.WorkoutExerciseRepository)),
         ),
     )
     ```
  2. Verificar que outros repositories seguem o mesmo padrão
- **Critério de aceite:**
  - `WorkoutExerciseRepository` disponível via DI
  - App inicia sem erros de DI

### T17 — Atualizar use cases para injetar WorkoutExerciseRepository
- **Objetivo:** Adicionar dependência de `WorkoutExerciseRepository` nos use cases relevantes
- **Arquivos/pacotes prováveis:**
  - Use cases que criam/validam set_records (ex: `internal/kinetria/domain/sessions/uc_record_set.go`)
- **Implementação (passos):**
  1. Identificar use cases que validam `exerciseID` + `workoutID`
  2. Adicionar campo `workoutExerciseRepo ports.WorkoutExerciseRepository` nas structs de use case
  3. Atualizar construtores para injetar dependência
  4. Atualizar lógica de validação:
     - Antes: `exerciseRepo.ExistsByIDAndWorkoutID(exerciseID, workoutID)`
     - Depois: `workoutExerciseRepo.ExistsByIDAndWorkoutID(workoutExerciseID, workoutID)`
- **Critério de aceite:**
  - Use cases compilam sem erros
  - Dependências injetadas via Fx
  - Lógica de validação atualizada

---

## Grupo 6: API Layer (Handlers + DTOs)

### T18 — Atualizar DTOs de response (GET /workouts/{id})
- **Objetivo:** Adicionar campos `exerciseId` e metadata de exercise no response
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/http/swagger_models.go` (ou DTOs)
- **Implementação (passos):**
  1. Atualizar struct de response (ex: `WorkoutExerciseDTO`):
     ```go
     type WorkoutExerciseDTO struct {
         ID           string   `json:"id"`           // workout_exercise_id
         ExerciseID   string   `json:"exerciseId"`   // exercise_id (biblioteca)
         Name         string   `json:"name"`
         Description  string   `json:"description"`
         Sets         int      `json:"sets"`
         Reps         string   `json:"reps"`
         RestTime     int      `json:"restTime"`
         Weight       int      `json:"weight"`
         Muscles      []string `json:"muscles"`
         ThumbnailURL string   `json:"thumbnailUrl"`
         OrderIndex   int      `json:"orderIndex,omitempty"`
     }
     ```
  2. Atualizar comentários Swagger
- **Critério de aceite:**
  - DTOs refletem contrato TO-BE
  - Comentários Swagger atualizados
  - Compila sem erros

### T19 — Atualizar handler GET /workouts/{id}
- **Objetivo:** Retornar exercises via `WorkoutExerciseRepository` (com JOIN)
- **Arquivos/pacotes prováveis:**
  - Handler de workout (ex: `internal/kinetria/gateways/http/handlers/workout_handler.go`)
- **Implementação (passos):**
  1. Injetar `WorkoutExerciseRepository` no handler
  2. Buscar workout: `workoutRepo.GetByID(workoutID)`
  3. Buscar exercises do workout: `workoutExerciseRepo.GetByWorkoutID(workoutID)`
  4. Mapear `[]WorkoutExercise` → `[]WorkoutExerciseDTO`
  5. Retornar response com exercises
- **Critério de aceite:**
  - Endpoint retorna 200 OK com exercises
  - Response segue contrato TO-BE (com `exerciseId`)
  - Compila sem erros

### T20 — Atualizar DTOs de request (POST /sessions/{id}/set-records)
- **Objetivo:** Alterar campo de `exerciseId` para `workoutExerciseId`
- **Arquivos/pacotes prováveis:**
  - DTOs de request (ex: `internal/kinetria/gateways/http/swagger_models.go`)
- **Implementação (passos):**
  1. Atualizar struct de request:
     ```go
     type RecordSetRequest struct {
         WorkoutExerciseID string `json:"workoutExerciseId" binding:"required,uuid"`
         SetNumber         int    `json:"setNumber" binding:"required,min=1"`
         Weight            int    `json:"weight" binding:"required,min=0"`
         Reps              int    `json:"reps" binding:"required,min=0"`
         Status            string `json:"status" binding:"required,oneof=completed skipped"`
     }
     ```
  2. Atualizar comentários Swagger
- **Critério de aceite:**
  - DTO reflete contrato TO-BE
  - Validações corretas (uuid, required)
  - Comentários Swagger atualizados

### T21 — Atualizar handler POST /sessions/{id}/set-records
- **Objetivo:** Validar `workoutExerciseId` e criar set_record
- **Arquivos/pacotes prováveis:**
  - Handler de sessions (ex: `internal/kinetria/gateways/http/handlers/session_handler.go`)
- **Implementação (passos):**
  1. Parsear `workoutExerciseId` do request
  2. Buscar session: `sessionRepo.FindByID(sessionID)`
  3. Validar que workout_exercise pertence ao workout da session:
     ```go
     exists := workoutExerciseRepo.ExistsByIDAndWorkoutID(workoutExerciseID, session.WorkoutID)
     if !exists {
         return 400, "workout_exercise não pertence ao workout desta sessão"
     }
     ```
  4. Criar set_record com `workoutExerciseID`
  5. Retornar 201 Created
- **Critério de aceite:**
  - Endpoint valida `workoutExerciseId` corretamente
  - Retorna 400 se workout_exercise não pertence ao workout
  - Retorna 201 Created se sucesso
  - Compila sem erros

---

## Grupo 7: Testes (Unit + Integration)

### T22 — Atualizar testes unitários de ExerciseRepository
- **Objetivo:** Criar/atualizar testes para novo CRUD de exercises
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/exercise_repository_test.go`
- **Implementação (passos):**
  1. Criar table-driven tests para:
     - `GetByID` (happy path + not found)
     - `GetAll` (empty + multiple exercises)
     - `Create` (success + constraint violation)
     - `Update` (success + not found)
     - `Delete` (success + FK constraint violation)
  2. Usar testcontainers ou mock DB
  3. Seguir padrão AAA (Arrange, Act, Assert)
- **Critério de aceite:**
  - Todos os métodos testados (coverage > 80%)
  - Table-driven tests
  - Testes passam: `go test ./internal/kinetria/gateways/repositories/...`

### T23 — Criar testes unitários de WorkoutExerciseRepository
- **Objetivo:** Testar novo repository (CRUD + JOIN)
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/workout_exercise_repository_test.go` (novo)
- **Implementação (passos):**
  1. Criar table-driven tests para:
     - `GetByID` (happy path + not found)
     - `GetByWorkoutID` (empty + multiple + ordering por order_index)
     - `ExistsByIDAndWorkoutID` (true + false)
     - `Create` (success + unique constraint violation)
     - `Update` (success + not found)
     - `Delete` (success + FK constraint violation)
  2. Validar que JOIN retorna metadata de exercises
  3. Usar testcontainers ou mock DB
- **Critério de aceite:**
  - Todos os métodos testados (coverage > 80%)
  - Table-driven tests
  - Testes de JOIN validam metadata de exercises
  - Testes passam

### T24 — Atualizar testes unitários de SetRecordRepository
- **Objetivo:** Atualizar testes para usar `workout_exercise_id`
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/repositories/set_record_repository_test.go`
- **Implementação (passos):**
  1. Atualizar testes existentes:
     - Substituir `exerciseID` por `workoutExerciseID` nos fixtures
  2. Adicionar teste: unique constraint (session + workout_exercise + set_number)
  3. Validar que FK constraint funciona (workout_exercise_id inválido)
- **Critério de aceite:**
  - Testes passam com novo schema
  - Coverage mantido/aumentado
  - Table-driven tests

### T25 — Criar testes de integração para fluxo completo
- **Objetivo:** Testar fluxo end-to-end (criar workout → session → set_record)
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/tests/integration/workout_exercise_flow_test.go` (novo)
- **Implementação (passos):**
  1. Setup: criar exercise na biblioteca
  2. Criar workout + vincular exercise (workout_exercise)
  3. Iniciar session do workout
  4. Criar set_record com workout_exercise_id
  5. Validar que set_record foi criado corretamente
  6. Validar query de set_records por session
  7. Usar testcontainers com Postgres real
- **Critério de aceite:**
  - Teste end-to-end passa
  - Valida integridade referencial (FKs)
  - Usa banco real (testcontainers)

### T26 — Criar testes de validação de API (contrato)
- **Objetivo:** Validar que API retorna contratos corretos (TO-BE)
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/gateways/http/handlers/workout_handler_test.go`
  - `internal/kinetria/gateways/http/handlers/session_handler_test.go`
- **Implementação (passos):**
  1. Testar GET /workouts/{id}:
     - Validar que response contém `exerciseId` + `id` (workout_exercise_id)
     - Validar que metadata de exercise está presente
  2. Testar POST /sessions/{id}/set-records:
     - Validar que request com `workoutExerciseId` válido retorna 201
     - Validar que request com `workoutExerciseId` inválido retorna 400
     - Validar mensagem de erro
  3. Usar mocks de repositories
- **Critério de aceite:**
  - Testes de contrato passam
  - Validam estrutura JSON correta
  - Table-driven tests

---

## Grupo 8: Documentação

### T27 — Atualizar documentação Swagger (OpenAPI)
- **Objetivo:** Atualizar specs da API com novos contratos
- **Arquivos/pacotes prováveis:**
  - `docs/swagger.yaml` (ou `docs/swagger.json`)
  - `internal/kinetria/gateways/http/handlers/*.go` (comentários Swagger)
- **Implementação (passos):**
  1. Atualizar endpoint GET /workouts/{id}:
     - Response: adicionar `exerciseId`, `description` no schema de exercise
     - Exemplo de response atualizado
  2. Atualizar endpoint POST /sessions/{id}/set-records:
     - Request: campo `workoutExerciseId` (no lugar de `exerciseId`)
     - Response de erro 400: documentar "workout_exercise não pertence ao workout"
  3. Adicionar schemas:
     - `Exercise` (biblioteca)
     - `WorkoutExercise` (vínculo)
  4. Regenerar docs: `make swagger` (se houver)
- **Critério de aceite:**
  - Swagger reflete contratos TO-BE
  - Exemplos de request/response atualizados
  - Schemas de error documentados
  - `make swagger` executa sem erros

### T28 — Documentar migration 009 no README
- **Objetivo:** Adicionar seção sobre relacionamento N:N no README do projeto
- **Arquivos/pacotes prováveis:**
  - `README.md`
  - `docs/architecture/exercises-model.md` (novo, opcional)
- **Implementação (passos):**
  1. Adicionar seção no README:
     - "Modelo de Exercises (N:N)"
     - Explicar biblioteca compartilhada
     - Explicar tabela de junção `workout_exercises`
     - Link para MIGRATION_009_GUIDE.md
  2. (Opcional) Criar diagrama ER:
     - workouts ← workout_exercises → exercises
     - sessions → workout_exercises (via set_records)
  3. Adicionar exemplo de uso da API
- **Critério de aceite:**
  - README atualizado com seção de exercises
  - Documentação clara e objetiva
  - Links para guias relevantes

### T29 — Adicionar comentários Godoc nas interfaces e entidades
- **Objetivo:** Garantir que todas as novas entidades/interfaces têm comentários claros
- **Arquivos/pacotes prováveis:**
  - `internal/kinetria/domain/entities/workout_exercise.go`
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação (passos):**
  1. Revisar entidade `WorkoutExercise`:
     - Adicionar comentário explicando que representa vínculo N:N
  2. Revisar interface `WorkoutExerciseRepository`:
     - Adicionar comentário explicando responsabilidade
  3. Atualizar comentários de `Exercise` e `SetRecord`
  4. Verificar que comentários seguem padrão Godoc (iniciam com nome do tipo)
- **Critério de aceite:**
  - Todas as entidades/interfaces exportadas têm comentários Godoc
  - Comentários são claros e objetivos
  - `golangci-lint` não reporta warnings de missing comments

---

## Grupo 9: Deploy e Validação

### T30 — Executar migration 009 em ambiente de staging
- **Objetivo:** Validar migration em ambiente próximo a produção
- **Arquivos/pacotes prováveis:**
  - `migrations/009_refactor_exercises_to_shared_library.sql`
- **Implementação (passos):**
  1. Criar dump de produção (ou dados realistas)
  2. Restaurar dump em staging
  3. Executar migration: `make migrate-up` (ou equivalente)
  4. Validar queries de integridade (ver T01)
  5. Executar testes de integração em staging
  6. Validar com cliente HTTP (Postman/Insomnia):
     - GET /workouts/{id}
     - POST /sessions/{id}/set-records
- **Critério de aceite:**
  - Migration executa sem erros em staging
  - Queries de integridade OK
  - API responde com contratos corretos
  - Testes de integração passam

### T31 — Deploy de backend em staging
- **Objetivo:** Deploy do código novo em staging e validação end-to-end
- **Arquivos/pacotes prováveis:**
  - `Makefile` (target de deploy)
  - `Dockerfile`
- **Implementação (passos):**
  1. Build da aplicação: `make build`
  2. Deploy em staging (Docker Compose, Kubernetes, etc.)
  3. Validar logs de inicialização (Fx DI)
  4. Executar smoke tests:
     - Health check endpoint
     - GET /workouts/{id}
     - POST /sessions/{id}/set-records
  5. Monitorar métricas (se houver)
- **Critério de aceite:**
  - App inicia sem erros
  - Smoke tests passam
  - Logs indicam sucesso de DI
  - API responde corretamente

### T32 — Validar rollback strategy em staging
- **Objetivo:** Testar processo de rollback caso algo dê errado
- **Arquivos/pacotes prováveis:**
  - Scripts de backup/restore (T02)
- **Implementação (passos):**
  1. Criar snapshot do banco pós-migration
  2. Simular problema (ex: deploy errado)
  3. Executar rollback: `make db-restore`
  4. Validar que banco voltou ao estado AS-IS
  5. Validar que app antiga funciona
- **Critério de aceite:**
  - Rollback restaura banco corretamente
  - App antiga funciona após rollback
  - Processo documentado (T02)

### T33 — Preparar checklist de produção
- **Objetivo:** Criar checklist final de deploy em produção
- **Arquivos/pacotes prováveis:**
  - `docs/deployment/migration-009-rollout.md`
- **Implementação (passos):**
  1. Consolidar checklist (ver seção 6 do plan.md):
     - Pre-checks (sessions ativas, backup)
     - Deploy (migration + backend)
     - Validação pós-deploy (queries, API, logs)
     - Rollback (se necessário)
  2. Adicionar contatos de suporte
  3. Adicionar SLAs (tempo de rollback, janela de manutenção)
- **Critério de aceite:**
  - Checklist completo e revisado
  - Aprovado por tech lead
  - Documentado em `docs/deployment/`

### T34 — Executar deploy em produção
- **Objetivo:** Aplicar migration e deploy de backend em produção
- **Arquivos/pacotes prováveis:**
  - `migrations/009_refactor_exercises_to_shared_library.sql`
  - `Makefile`, `Dockerfile`
- **Implementação (passos):**
  1. Seguir checklist (T33)
  2. Criar backup de produção
  3. Validar que não há sessions ativas
  4. Executar migration em produção
  5. Deploy de backend
  6. Validar queries de integridade
  7. Executar smoke tests em produção
  8. Monitorar logs/métricas por 1h
  9. Notificar equipe de sucesso
- **Critério de aceite:**
  - Migration executada sem erros
  - App em produção funcionando
  - Smoke tests passam
  - Sem alarmes críticos em 1h

---

## Grupo 10: Pós-Deploy e Monitoramento

### T35 — Monitorar logs e métricas pós-deploy (D+1)
- **Objetivo:** Acompanhar comportamento da aplicação no primeiro dia
- **Arquivos/pacotes prováveis:**
  - N/A (observabilidade externa)
- **Implementação (passos):**
  1. Monitorar logs:
     - Erros de FK violations
     - Erros de validação (workoutExerciseId inválido)
     - Latência de queries com JOIN
  2. Monitorar métricas:
     - Taxa de erro de POST /sessions/{id}/set-records
     - Latência P99 de GET /workouts/{id}
     - Total de exercises na biblioteca (gauge)
  3. Validar com usuários beta:
     - Criar workout
     - Iniciar session
     - Registrar série
- **Critério de aceite:**
  - Sem erros críticos em 24h
  - Latência dentro do SLA
  - Feedback positivo de usuários beta

### T36 — Criar análise de performance das queries com JOIN
- **Objetivo:** Validar que queries com JOIN não degradam performance
- **Arquivos/pacotes prováveis:**
  - N/A (análise manual)
- **Implementação (passos):**
  1. Executar EXPLAIN ANALYZE nas queries principais:
     - `GetWorkoutExercisesByWorkoutID` (com JOIN)
     - `GetSetRecordsBySessionID` (se houver JOIN)
  2. Validar que índices estão sendo usados:
     - `idx_workout_exercises_workout_id`
     - `idx_workout_exercises_order`
  3. Comparar latência AS-IS vs TO-BE (se possível)
  4. Documentar findings
- **Critério de aceite:**
  - Queries usam índices corretamente
  - Latência dentro do esperado (< 50ms P99)
  - Documentação criada (em docs/ ou wiki)

### T37 — Revisar e deduplicar exercises na biblioteca (se necessário)
- **Objetivo:** Limpar possíveis duplicatas criadas pela migration
- **Arquivos/pacotes prováveis:**
  - Scripts SQL ad-hoc
- **Implementação (passos):**
  1. Executar query para identificar duplicatas:
     ```sql
     SELECT name, COUNT(*) FROM exercises GROUP BY name HAVING COUNT(*) > 1;
     ```
  2. Revisar manualmente exercícios duplicados
  3. Se necessário, consolidar:
     - Atualizar `workout_exercises` para referenciar exercise único
     - Deletar exercise duplicado
  4. Documentar processo de deduplicação
- **Critério de aceite:**
  - Biblioteca de exercises sem duplicatas desnecessárias
  - Processo documentado (caso precise repetir)

---

## Resumo de Entregas

| Grupo | Entregas | Criticidade |
|-------|---------|-------------|
| 1. Preparação | Migration validada, backup strategy | `blocker` |
| 2. Domain Layer | Entities + Ports atualizados | `blocker` |
| 3. Queries SQL | Queries SQLC + regeneração | `blocker` |
| 4. Gateways | Repositories implementados | `blocker` |
| 5. DI | Fx module atualizado | `blocker` |
| 6. API Layer | Handlers + DTOs atualizados | `blocker` |
| 7. Testes | Unit + Integration tests | `high` (mas bloquear merge sem) |
| 8. Documentação | Swagger + README + Godoc | `high` |
| 9. Deploy | Staging + Produção | `blocker` |
| 10. Pós-Deploy | Monitoramento + Performance | `medium` (mas crítico para estabilidade) |

**Total de tarefas:** 37

**Estimativa de esforço:** 
- Desenvolvimento: ~5-7 dias (1 dev full-time)
- Testes: ~2-3 dias
- Documentação: ~1 dia
- Deploy + Validação: ~1-2 dias

**Total:** ~10-13 dias (sprint de 2-3 semanas)
