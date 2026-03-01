# Tasks — Workout Management Endpoints

## T01 — Criar migration para adicionar ownership e soft delete
- **Objetivo:** Adicionar colunas `created_by` e `deleted_at` na tabela `workouts`
- **Arquivos:**
  - `internal/kinetria/gateways/migrations/013_add_workout_ownership.sql`
- **Implementação:**
  1. Criar arquivo de migration com `ALTER TABLE workouts ADD COLUMN ...`
  2. Adicionar `created_by UUID REFERENCES users(id) ON DELETE CASCADE` (nullable)
  3. Adicionar `deleted_at TIMESTAMP` (nullable)
  4. Criar índice `idx_workouts_created_by` em (created_by)
  5. Criar índice `idx_workouts_deleted_at` em (deleted_at) WHERE deleted_at IS NULL
  6. Testar migration localmente
- **Critério de aceite:**
  - Migration aplicada com sucesso ao subir o ambiente
  - Colunas existem na tabela `workouts`
  - Índices criados corretamente
  - Workouts existentes ficam com `created_by = NULL` (templates)

---

## T02 — Atualizar entity Workout com novos campos
- **Objetivo:** Adicionar campos `CreatedBy` e `DeletedAt` na entity
- **Arquivos:**
  - `internal/kinetria/domain/entities/workout.go`
- **Implementação:**
  1. Adicionar campo `CreatedBy *uuid.UUID` (nullable)
  2. Adicionar campo `DeletedAt *time.Time` (nullable)
  3. Atualizar construtores/factories se necessário
- **Critério de aceite:**
  - Código compila sem erros
  - Testes existentes de `Workout` continuam passando

---

## T03 — Criar structs de input no domain
- **Objetivo:** Definir tipos para criar/atualizar workouts
- **Arquivos:**
  - `internal/kinetria/domain/workouts/types.go` (novo ou existente)
- **Implementação:**
  1. Criar `CreateWorkoutInput` com todos os campos obrigatórios
  2. Criar `UpdateWorkoutInput` com campos opcionais (ponteiros)
  3. Criar `WorkoutExerciseInput` com exerciseId, sets, reps, restTime, weight, orderIndex
  4. Adicionar comentários de documentação
- **Critério de aceite:**
  - Structs compilam sem erros
  - Tipos bem documentados

---

## T04 — Adicionar métodos no WorkoutRepository port
- **Objetivo:** Definir contratos para operações de CRUD
- **Arquivos:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação:**
  1. Adicionar `Create(ctx, workout, exercises) error`
  2. Adicionar `Update(ctx, workout, exercises) error`
  3. Adicionar `Delete(ctx, workoutID, userID) error`
  4. Adicionar `HasActiveSessions(ctx, workoutID) (bool, error)`
- **Critério de aceite:**
  - Interface compila sem erros
  - Implementações existentes quebram (esperado, será corrigido em T05-T06)

---

## T05 — Implementar queries SQLC para workouts
- **Objetivo:** Criar queries SQL para CRUD
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/workouts.sql`
- **Implementação:**
  1. Criar `CreateWorkout` — INSERT workout
  2. Criar `UpdateWorkout` — UPDATE workout (validar created_by)
  3. Criar `SoftDeleteWorkout` — UPDATE deleted_at (validar created_by)
  4. Criar `HasActiveSessions` — SELECT EXISTS com status='active'
  5. Criar `CreateWorkoutExercise` — INSERT workout_exercise
  6. Criar `DeleteWorkoutExercises` — DELETE workout_exercises por workout_id
  7. Atualizar `ListWorkoutsByUserID` para filtrar por ownership e deleted_at
  8. Rodar `make sqlc` para gerar código Go
- **Critério de aceite:**
  - Queries compilam sem erros no SQLC
  - Código Go gerado em `queries/workouts.sql.go`
  - Queries testadas manualmente no psql

---

## T06 — Implementar métodos no WorkoutRepository com transações
- **Objetivo:** Implementar lógica de acesso a dados com transações
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/workout_repository.go`
- **Implementação:**
  1. Implementar `Create()`:
     - Iniciar transação (BEGIN)
     - Chamar `queries.CreateWorkout()`
     - Loop: chamar `queries.CreateWorkoutExercise()` para cada exercise
     - Commit se sucesso, Rollback se erro
  2. Implementar `Update()`:
     - Iniciar transação
     - Chamar `queries.UpdateWorkout()`
     - Chamar `queries.DeleteWorkoutExercises()`
     - Loop: chamar `queries.CreateWorkoutExercise()` para cada exercise
     - Commit/Rollback
  3. Implementar `Delete()`:
     - Chamar `queries.HasActiveSessions()`
     - Se tem sessions ativas, retornar erro
     - Chamar `queries.SoftDeleteWorkout()`
  4. Implementar `HasActiveSessions()`:
     - Chamar `queries.HasActiveSessions()`
- **Critério de aceite:**
  - Código compila sem erros
  - Transações funcionam corretamente (testar rollback)
  - Métodos retornam dados corretos

---

## T07 — Criar use case CreateWorkoutUC
- **Objetivo:** Implementar lógica de negócio para criar workout
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_create_workout.go` (novo)
- **Implementação:**
  1. Criar struct `CreateWorkoutUC` com dependências (WorkoutRepository, ExerciseRepository)
  2. Implementar método `Execute(ctx, userID, input) (*Workout, error)`
  3. Validar inputs:
     - Name: 3-255 caracteres
     - Type: enum válido
     - Intensity: enum válido
     - Duration: 1-300 minutos
     - Exercises: 1-20 exercícios
     - Sets: 1-10
     - RestTime: 0-600 segundos
     - OrderIndex: sem duplicatas
  4. Validar que todos os exerciseID existem (chamar ExerciseRepository)
  5. Criar entity `Workout` com `CreatedBy = userID`
  6. Criar entities `WorkoutExercise`
  7. Chamar `workoutRepo.Create()` (transação)
  8. Retornar workout criado
- **Critério de aceite:**
  - Use case cria workout corretamente
  - Validações funcionam (rejeitar inputs inválidos)
  - Transação funciona (rollback se falhar)

---

## T08 — Criar use case UpdateWorkoutUC
- **Objetivo:** Implementar lógica de negócio para atualizar workout
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_update_workout.go` (novo)
- **Implementação:**
  1. Criar struct `UpdateWorkoutUC` com dependências
  2. Implementar método `Execute(ctx, userID, workoutID, input) (*Workout, error)`
  3. Buscar workout atual via `workoutRepo.GetByID()`
  4. Validar ownership (`workout.CreatedBy == userID`)
  5. Validar que não é template (`workout.CreatedBy != NULL`)
  6. Validar inputs (similar ao create)
  7. Validar que todos os exerciseID existem
  8. Atualizar campos modificados
  9. Chamar `workoutRepo.Update()` (transação)
  10. Retornar workout atualizado
- **Critério de aceite:**
  - Use case atualiza workout corretamente
  - Validações de ownership funcionam (retornar erro se não for dono)
  - Validações de inputs funcionam
  - Transação funciona

---

## T09 — Criar use case DeleteWorkoutUC
- **Objetivo:** Implementar lógica de negócio para deletar workout
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_delete_workout.go` (novo)
- **Implementação:**
  1. Criar struct `DeleteWorkoutUC` com dependência WorkoutRepository
  2. Implementar método `Execute(ctx, userID, workoutID) error`
  3. Buscar workout atual via `workoutRepo.GetByID()`
  4. Validar ownership (`workout.CreatedBy == userID`)
  5. Validar que não é template (`workout.CreatedBy != NULL`)
  6. Chamar `workoutRepo.HasActiveSessions()`
  7. Se tem sessions ativas, retornar erro (conflict)
  8. Chamar `workoutRepo.Delete()` (soft delete)
  9. Retornar sucesso
- **Critério de aceite:**
  - Use case deleta workout corretamente
  - Validações de ownership funcionam
  - Bloqueia deleção se tem sessions ativas
  - Soft delete funciona (deleted_at preenchido)

---

## T10 — Criar DTOs para WorkoutsHandler
- **Objetivo:** Definir contratos de request/response
- **Arquivos:**
  - `internal/kinetria/gateways/http/dtos/workouts.go` (novo ou existente)
- **Implementação:**
  1. Criar `CreateWorkoutRequest` com todos os campos obrigatórios
  2. Criar `CreateWorkoutExerciseDTO` com exerciseId, sets, reps, restTime, weight, orderIndex
  3. Criar `UpdateWorkoutRequest` com campos opcionais (ponteiros)
  4. Criar `WorkoutResponse` com data
  5. Adicionar tags JSON e validação
- **Critério de aceite:**
  - DTOs compilam sem erros
  - Serialização JSON funciona corretamente

---

## T11 — Atualizar WorkoutsHandler com novos handlers
- **Objetivo:** Implementar handlers HTTP para POST, PUT, DELETE
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_workouts.go`
- **Implementação:**
  1. Adicionar dependências dos novos use cases no struct `WorkoutsHandler`
  2. Implementar `HandleCreateWorkout()`:
     - Extrair userID do JWT
     - Decodificar request body
     - Validar inputs
     - Chamar `createWorkoutUC.Execute()`
     - Mapear entity para DTO
     - Retornar JSON 201
     - Tratar erros (400, 401, 500)
  3. Implementar `HandleUpdateWorkout()`:
     - Extrair userID do JWT
     - Extrair workoutID do path
     - Decodificar request body
     - Validar inputs
     - Chamar `updateWorkoutUC.Execute()`
     - Mapear entity para DTO
     - Retornar JSON 200
     - Tratar erros (400, 401, 403, 404, 500)
  4. Implementar `HandleDeleteWorkout()`:
     - Extrair userID do JWT
     - Extrair workoutID do path
     - Chamar `deleteWorkoutUC.Execute()`
     - Retornar 204 No Content
     - Tratar erros (401, 403, 404, 409, 500)
- **Critério de aceite:**
  - Handlers compilam sem erros
  - Retornam status codes corretos
  - Mapeiam erros de domínio para HTTP corretamente

---

## T12 — Registrar rotas de workouts no router
- **Objetivo:** Adicionar endpoints de CRUD no Chi router
- **Arquivos:**
  - `internal/kinetria/gateways/http/router.go`
- **Implementação:**
  1. Adicionar rotas protegidas:
     - `POST /api/v1/workouts` → `workoutsHandler.HandleCreateWorkout`
     - `PUT /api/v1/workouts/{id}` → `workoutsHandler.HandleUpdateWorkout`
     - `DELETE /api/v1/workouts/{id}` → `workoutsHandler.HandleDeleteWorkout`
  2. Garantir que middleware de autenticação está aplicado
- **Critério de aceite:**
  - Rotas registradas corretamente
  - Middleware de autenticação aplicado
  - Endpoints acessíveis via HTTP

---

## T13 — Registrar dependências no Fx (main.go)
- **Objetivo:** Configurar injeção de dependências
- **Arquivos:**
  - `cmd/kinetria/api/main.go`
- **Implementação:**
  1. Adicionar `workouts.NewCreateWorkoutUC` no `fx.Provide`
  2. Adicionar `workouts.NewUpdateWorkoutUC` no `fx.Provide`
  3. Adicionar `workouts.NewDeleteWorkoutUC` no `fx.Provide`
  4. Garantir que handler é injetado no router com novos use cases
- **Critério de aceite:**
  - Aplicação inicia sem erros
  - Dependências resolvidas corretamente pelo Fx

---

## T14 — Atualizar query ListWorkoutsByUserID
- **Objetivo:** Filtrar workouts por ownership e deleted_at
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/workouts.sql`
- **Implementação:**
  1. Modificar query `ListWorkoutsByUserID` para:
     - Filtrar `(created_by = $1 OR created_by IS NULL)` (templates + customizados do usuário)
     - Filtrar `deleted_at IS NULL` (não deletados)
  2. Rodar `make sqlc`
- **Critério de aceite:**
  - Query retorna templates + workouts customizados do usuário
  - Query não retorna workouts deletados
  - Query não retorna workouts de outros usuários

---

## T15 — Criar testes unitários para CreateWorkoutUC
- **Objetivo:** Testar lógica de negócio de criar workout
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_create_workout_test.go` (novo)
- **Implementação:**
  1. Criar mocks de repositories
  2. Testar cenários (table-driven):
     - Happy path: cria workout + exercises corretamente
     - Workout sem exercises (retorna erro)
     - ExerciseID inválido (retorna erro)
     - Name muito curto (retorna erro)
     - Type inválido (retorna erro)
     - Sets fora do range (retorna erro)
     - RestTime negativo (retorna erro)
     - OrderIndex duplicado (retorna erro)
     - Mais de 20 exercises (retorna erro)
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case
  - Todos os cenários BDD cobertos

---

## T16 — Criar testes unitários para UpdateWorkoutUC
- **Objetivo:** Testar lógica de negócio de atualizar workout
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_update_workout_test.go` (novo)
- **Implementação:**
  1. Criar mocks de repositories
  2. Testar cenários:
     - Happy path: atualiza workout corretamente
     - Atualizar workout de outro usuário (retorna erro 403)
     - Atualizar workout template (retorna erro 403)
     - Workout não encontrado (retorna erro 404)
     - Validações inválidas (similar ao create)
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case
  - Validações de ownership testadas

---

## T17 — Criar testes unitários para DeleteWorkoutUC
- **Objetivo:** Testar lógica de negócio de deletar workout
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_delete_workout_test.go` (novo)
- **Implementação:**
  1. Criar mocks de repositories
  2. Testar cenários:
     - Happy path: deleta workout corretamente
     - Deletar workout de outro usuário (retorna erro 403)
     - Deletar workout template (retorna erro 403)
     - Deletar workout com session ativa (retorna erro 409)
     - Deletar workout com sessions completadas (sucesso)
     - Workout não encontrado (retorna erro 404)
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case
  - Validações de ownership e sessions ativas testadas

---

## T18 — Criar testes de integração para POST /workouts
- **Objetivo:** Testar endpoint de criar workout com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_workouts_integration_test.go` (novo ou existente)
- **Implementação:**
  1. Setup: criar user, exercises no DB, gerar JWT
  2. Testar cenários (table-driven):
     - POST /workouts com dados válidos: retorna 201 com workout criado
     - Workout sem exercises: retorna 400
     - ExerciseID inválido: retorna 400
     - Validações inválidas: retorna 400
     - Sem JWT: retorna 401
  3. Verificar que workout foi criado no DB com created_by correto
  4. Verificar que workout_exercises foram criados
  5. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Cenários BDD cobertos
  - Transação funciona (rollback se falhar)

---

## T19 — Criar testes de integração para PUT /workouts
- **Objetivo:** Testar endpoint de atualizar workout com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_workouts_integration_test.go`
- **Implementação:**
  1. Setup: criar user, workout, exercises no DB
  2. Testar cenários:
     - PUT /workouts/:id com dados válidos: retorna 200 com workout atualizado
     - Atualizar workout de outro usuário: retorna 403
     - Atualizar workout template: retorna 403
     - Workout não encontrado: retorna 404
     - Validações inválidas: retorna 400
     - Sem JWT: retorna 401
  3. Verificar que workout foi atualizado no DB
  4. Verificar que workout_exercises antigos foram deletados e novos criados
  5. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Cenários BDD cobertos
  - Transação funciona

---

## T20 — Criar testes de integração para DELETE /workouts
- **Objetivo:** Testar endpoint de deletar workout com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_workouts_integration_test.go`
- **Implementação:**
  1. Setup: criar user, workout, sessions no DB
  2. Testar cenários:
     - DELETE /workouts/:id sem sessions ativas: retorna 204
     - Deletar workout de outro usuário: retorna 403
     - Deletar workout template: retorna 403
     - Deletar workout com session ativa: retorna 409
     - Deletar workout com sessions completadas: retorna 204
     - Workout não encontrado: retorna 404
     - Sem JWT: retorna 401
  3. Verificar que workout foi soft deleted (deleted_at preenchido)
  4. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Todos os cenários BDD cobertos
  - Soft delete funciona corretamente

---

## T21 — Criar testes de integração para GET /workouts (ownership)
- **Objetivo:** Testar que GET /workouts filtra por ownership
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_workouts_integration_test.go`
- **Implementação:**
  1. Setup: criar múltiplos users, workouts (templates e customizados)
  2. Testar cenários:
     - GET /workouts retorna templates + workouts customizados do usuário
     - GET /workouts não retorna workouts de outros usuários
     - GET /workouts não retorna workouts deletados
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Filtro de ownership funciona corretamente

---

## T22 — Documentar endpoints no README
- **Objetivo:** Atualizar documentação da API
- **Arquivos:**
  - `README.md`
- **Implementação:**
  1. Adicionar seção "Workout Management" na tabela de endpoints
  2. Adicionar exemplos de request/response para POST, PUT, DELETE
  3. Documentar validações e regras de negócio (ownership, soft delete, sessions ativas)
  4. Documentar erros possíveis (400, 403, 404, 409)
  5. Adicionar exemplos de uso com curl
- **Critério de aceite:**
  - Documentação clara e completa
  - Exemplos funcionam corretamente
  - Alinhada com comportamento implementado

---

## T23 — Adicionar comentários Godoc nos use cases
- **Objetivo:** Documentar código para geração de docs
- **Arquivos:**
  - `internal/kinetria/domain/workouts/uc_create_workout.go`
  - `internal/kinetria/domain/workouts/uc_update_workout.go`
  - `internal/kinetria/domain/workouts/uc_delete_workout.go`
- **Implementação:**
  1. Adicionar comentário de pacote em `doc.go`
  2. Adicionar comentários Godoc em structs e métodos públicos
  3. Documentar parâmetros e retornos
  4. Documentar erros possíveis
  5. Documentar regras de negócio (ownership, validações)
- **Critério de aceite:**
  - Comentários seguem padrão Godoc
  - `go doc` exibe documentação corretamente

---

## T24 — Atualizar documentação Swagger
- **Objetivo:** Gerar documentação interativa da API
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_workouts.go`
- **Implementação:**
  1. Adicionar annotations Swagger nos novos handlers:
     - `@Summary`, `@Description`, `@Tags`
     - `@Accept`, `@Produce`
     - `@Param` (body, path params)
     - `@Success`, `@Failure`
     - `@Security` (JWT)
  2. Rodar `make swagger` para regenerar docs
  3. Testar endpoints no Swagger UI
- **Critério de aceite:**
  - Swagger UI exibe endpoints de workout management
  - Exemplos de request/response corretos
  - Validações documentadas
  - Autenticação JWT funciona no Swagger UI
