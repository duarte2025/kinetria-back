# Tasks — Exercise Library Endpoints

## T01 — Criar migration para expandir tabela exercises
- **Objetivo:** Adicionar campos faltantes na tabela `exercises`
- **Arquivos:**
  - `internal/kinetria/gateways/migrations/011_expand_exercises_table.sql`
- **Implementação:**
  1. Criar arquivo de migration com `ALTER TABLE exercises ADD COLUMN ...`
  2. Adicionar campos: `description`, `instructions`, `tips`, `difficulty`, `equipment`, `video_url`
  3. Criar índices: GIN em `name` (full-text), GIN em `muscles`, B-tree em `difficulty` e `equipment`
  4. Testar migration localmente
- **Critério de aceite:**
  - Migration aplicada com sucesso ao subir o ambiente
  - Colunas existem na tabela `exercises`
  - Índices criados corretamente

---

## T02 — Criar seed de exercícios
- **Objetivo:** Popular biblioteca com 30-40 exercícios base
- **Arquivos:**
  - `internal/kinetria/gateways/migrations/012_seed_exercises.sql` (ou script Go separado)
- **Implementação:**
  1. Criar INSERT statements para 30-40 exercícios mais comuns
  2. Incluir dados: name, description, instructions, tips, difficulty, equipment, muscles, thumbnail_url, video_url
  3. Usar URLs mock para thumbnails e vídeos
  4. Cobrir principais grupos musculares: Peito, Costas, Pernas, Ombros, Braços, Core
  5. Variar dificuldades: Iniciante, Intermediário, Avançado
  6. Variar equipamentos: Barra, Halteres, Peso corporal, Máquinas
- **Critério de aceite:**
  - Seed aplicado com sucesso
  - 30-40 exercícios inseridos no banco
  - Dados completos e consistentes

---

## T03 — Atualizar entity Exercise com novos campos
- **Objetivo:** Adicionar campos faltantes na entity `Exercise`
- **Arquivos:**
  - `internal/kinetria/domain/entities/exercise.go`
- **Implementação:**
  1. Adicionar campos: `Description`, `Instructions`, `Tips`, `Difficulty`, `Equipment`, `VideoURL`
  2. Todos os novos campos são ponteiros (nullable)
  3. Atualizar construtores/factories se necessário
- **Critério de aceite:**
  - Código compila sem erros
  - Testes existentes de `Exercise` continuam passando

---

## T04 — Criar structs auxiliares no domain
- **Objetivo:** Definir tipos para filtros, stats e histórico
- **Arquivos:**
  - `internal/kinetria/domain/exercises/types.go` (novo)
- **Implementação:**
  1. Criar `ExerciseFilters` com campos opcionais (ponteiros)
  2. Criar `ExerciseWithStats` que combina Exercise + UserStats
  3. Criar `ExerciseUserStats` com lastPerformed, bestWeight, timesPerformed, averageWeight
  4. Criar `ExerciseHistoryEntry` com sessionId, workoutName, performedAt, sets
  5. Criar `SetDetail` com setNumber, reps, weight, status
- **Critério de aceite:**
  - Structs compilam sem erros
  - Tipos bem documentados com comentários

---

## T05 — Adicionar métodos no ExerciseRepository port
- **Objetivo:** Definir contratos para operações de biblioteca
- **Arquivos:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação:**
  1. Adicionar `List(ctx, filters, page, pageSize) ([]*Exercise, int, error)`
  2. Adicionar `GetByID(ctx, exerciseID) (*Exercise, error)`
  3. Adicionar `GetUserStats(ctx, userID, exerciseID) (*ExerciseUserStats, error)`
  4. Adicionar `GetHistory(ctx, userID, exerciseID, page, pageSize) ([]*ExerciseHistoryEntry, int, error)`
- **Critério de aceite:**
  - Interface compila sem erros
  - Implementações existentes quebram (esperado, será corrigido em T06)

---

## T06 — Implementar queries SQLC para exercises
- **Objetivo:** Criar queries SQL para operações de biblioteca
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/exercises.sql`
- **Implementação:**
  1. Criar `ListExercises` com filtros opcionais (muscleGroup, equipment, difficulty, search) e paginação
  2. Criar `CountExercises` com mesmos filtros
  3. Criar `GetExerciseByID` simples
  4. Criar `GetExerciseUserStats` com JOINs (exercises → workout_exercises → set_records → sessions)
  5. Criar `GetExerciseHistory` com JOINs (exercises → workout_exercises → set_records → sessions → workouts)
  6. Criar `CountExerciseHistory` para paginação
  7. Rodar `make sqlc` para gerar código Go
- **Critério de aceite:**
  - Queries compilam sem erros no SQLC
  - Código Go gerado em `queries/exercises.sql.go`
  - Queries testadas manualmente no psql

---

## T07 — Implementar métodos no ExerciseRepository
- **Objetivo:** Implementar lógica de acesso a dados
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/exercise_repository.go`
- **Implementação:**
  1. Implementar `List()`:
     - Calcular offset a partir de page e pageSize
     - Chamar `queries.ListExercises()` e `queries.CountExercises()`
     - Mapear rows para entities
     - Retornar lista + total
  2. Implementar `GetByID()`:
     - Chamar `queries.GetExerciseByID()`
     - Mapear row para entity
     - Retornar erro se não encontrado
  3. Implementar `GetUserStats()`:
     - Chamar `queries.GetExerciseUserStats()`
     - Mapear row para struct de stats
     - Retornar stats (null se user nunca executou)
  4. Implementar `GetHistory()`:
     - Calcular offset
     - Chamar `queries.GetExerciseHistory()` e `queries.CountExerciseHistory()`
     - Agrupar sets por sessionID
     - Retornar lista de entries + total
- **Critério de aceite:**
  - Código compila sem erros
  - Métodos retornam dados corretos (testar manualmente)

---

## T08 — Criar use case ListExercisesUC
- **Objetivo:** Implementar lógica de negócio para listar exercícios
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_list_exercises.go` (novo)
- **Implementação:**
  1. Criar struct `ListExercisesUC` com dependência `ExerciseRepository`
  2. Criar struct `ListExercisesInput` com filtros e paginação
  3. Implementar método `Execute(ctx, input) ([]*Exercise, int, error)`
  4. Validar inputs:
     - page >= 1
     - pageSize >= 1 e <= 100
  5. Chamar `exerciseRepo.List()`
  6. Retornar lista + total
- **Critério de aceite:**
  - Use case retorna lista corretamente
  - Validações funcionam (rejeitar inputs inválidos)

---

## T09 — Criar use case GetExerciseUC
- **Objetivo:** Implementar lógica de negócio para obter detalhes de exercício
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_get_exercise.go` (novo)
- **Implementação:**
  1. Criar struct `GetExerciseUC` com dependência `ExerciseRepository`
  2. Implementar método `Execute(ctx, exerciseID, userID *uuid.UUID) (*ExerciseWithStats, error)`
  3. Chamar `exerciseRepo.GetByID()`
  4. Se userID fornecido, chamar `exerciseRepo.GetUserStats()`
  5. Retornar `ExerciseWithStats` (stats null se não autenticado)
  6. Retornar erro de domínio se exercise não encontrado
- **Critério de aceite:**
  - Use case retorna exercise corretamente
  - Inclui stats se userID fornecido
  - Retorna erro apropriado se exercise não existe

---

## T10 — Criar use case GetExerciseHistoryUC
- **Objetivo:** Implementar lógica de negócio para obter histórico
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_get_exercise_history.go` (novo)
- **Implementação:**
  1. Criar struct `GetExerciseHistoryUC` com dependência `ExerciseRepository`
  2. Criar struct `GetExerciseHistoryInput` com exerciseID, userID, page, pageSize
  3. Implementar método `Execute(ctx, input) ([]*ExerciseHistoryEntry, int, error)`
  4. Validar inputs (page >= 1, pageSize <= 100)
  5. Verificar que exercise existe via `exerciseRepo.GetByID()`
  6. Chamar `exerciseRepo.GetHistory()`
  7. Retornar lista de entries + total
- **Critério de aceite:**
  - Use case retorna histórico corretamente
  - Sets agrupados por session
  - Ordenado por mais recente primeiro
  - Retorna erro apropriado se exercise não existe

---

## T11 — Criar DTOs para ExercisesHandler
- **Objetivo:** Definir contratos de request/response
- **Arquivos:**
  - `internal/kinetria/gateways/http/dtos/exercises.go` (novo)
- **Implementação:**
  1. Criar `ExerciseDTO` com todos os campos do exercise
  2. Criar `ExerciseWithStatsDTO` que estende ExerciseDTO + userStats
  3. Criar `UserStatsDTO` com lastPerformed, bestWeight, timesPerformed, averageWeight
  4. Criar `HistoryEntryDTO` com sessionId, workoutName, performedAt, sets
  5. Criar `SetDetailDTO` com setNumber, reps, weight, status
  6. Criar `ListExercisesResponse` com data[] e meta
  7. Criar `ExerciseDetailResponse` com data
  8. Criar `ExerciseHistoryResponse` com data[] e meta
  9. Adicionar tags JSON
- **Critério de aceite:**
  - DTOs compilam sem erros
  - Serialização JSON funciona corretamente

---

## T12 — Criar ExercisesHandler
- **Objetivo:** Implementar handlers HTTP para endpoints de biblioteca
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_exercises.go` (novo)
- **Implementação:**
  1. Criar struct `ExercisesHandler` com dependências dos 3 use cases
  2. Implementar `HandleListExercises()`:
     - Extrair query params (muscleGroup, equipment, difficulty, search, page, pageSize)
     - Validar inputs
     - Chamar `listExercisesUC.Execute()`
     - Mapear entities para DTOs
     - Retornar JSON 200 com paginação
     - Tratar erros (400, 500)
  3. Implementar `HandleGetExercise()`:
     - Extrair exerciseID do path
     - Tentar extrair userID do JWT (opcional, não retornar erro se falhar)
     - Chamar `getExerciseUC.Execute()`
     - Mapear entity para DTO (incluir stats se autenticado)
     - Retornar JSON 200
     - Tratar erros (400, 404, 500)
  4. Implementar `HandleGetExerciseHistory()`:
     - Extrair exerciseID do path
     - Extrair userID do JWT (obrigatório)
     - Extrair query params (page, pageSize)
     - Validar inputs
     - Chamar `getExerciseHistoryUC.Execute()`
     - Mapear entries para DTOs
     - Retornar JSON 200 com paginação
     - Tratar erros (400, 401, 404, 500)
- **Critério de aceite:**
  - Handlers compilam sem erros
  - Retornam status codes corretos
  - Mapeiam erros de domínio para HTTP corretamente
  - Autenticação opcional funciona em GET /exercises/:id

---

## T13 — Criar helper para autenticação opcional
- **Objetivo:** Extrair userID do JWT sem exigir autenticação
- **Arquivos:**
  - `internal/kinetria/gateways/http/middleware_auth.go`
- **Implementação:**
  1. Criar função `tryExtractUserIDFromJWT(r *http.Request) *uuid.UUID`
  2. Tentar extrair token do header Authorization
  3. Tentar validar e parsear JWT
  4. Se sucesso, retornar userID
  5. Se falhar, retornar nil (não retornar erro)
- **Critério de aceite:**
  - Função retorna userID se JWT válido
  - Função retorna nil se JWT ausente ou inválido
  - Não retorna erro (silent fail)

---

## T14 — Registrar rotas de exercises no router
- **Objetivo:** Adicionar endpoints de biblioteca no Chi router
- **Arquivos:**
  - `internal/kinetria/gateways/http/router.go`
- **Implementação:**
  1. Adicionar rotas públicas (ou autenticação opcional):
     - `GET /api/v1/exercises` → `exercisesHandler.HandleListExercises`
     - `GET /api/v1/exercises/{id}` → `exercisesHandler.HandleGetExercise`
  2. Adicionar rota protegida:
     - `GET /api/v1/exercises/{id}/history` → `exercisesHandler.HandleGetExerciseHistory`
  3. Garantir que middleware de autenticação está aplicado apenas em /history
- **Critério de aceite:**
  - Rotas registradas corretamente
  - Middleware de autenticação aplicado apenas em /history
  - Endpoints acessíveis via HTTP

---

## T15 — Registrar dependências no Fx (main.go)
- **Objetivo:** Configurar injeção de dependências
- **Arquivos:**
  - `cmd/kinetria/api/main.go`
- **Implementação:**
  1. Adicionar `exercises.NewListExercisesUC` no `fx.Provide`
  2. Adicionar `exercises.NewGetExerciseUC` no `fx.Provide`
  3. Adicionar `exercises.NewGetExerciseHistoryUC` no `fx.Provide`
  4. Adicionar `http.NewExercisesHandler` no `fx.Provide`
  5. Garantir que handler é injetado no router
- **Critério de aceite:**
  - Aplicação inicia sem erros
  - Dependências resolvidas corretamente pelo Fx

---

## T16 — Criar testes unitários para ListExercisesUC
- **Objetivo:** Testar lógica de negócio de listar exercícios
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_list_exercises_test.go` (novo)
- **Implementação:**
  1. Criar mock de `ExerciseRepository`
  2. Testar cenários (table-driven):
     - Happy path: retorna lista paginada
     - Filtrar por muscleGroup
     - Filtrar por equipment
     - Filtrar por difficulty
     - Buscar por search
     - Combinar múltiplos filtros
     - Paginação inválida (page < 1, pageSize > 100)
     - Biblioteca vazia (retorna array vazio)
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case

---

## T17 — Criar testes unitários para GetExerciseUC
- **Objetivo:** Testar lógica de negócio de obter detalhes
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_get_exercise_test.go` (novo)
- **Implementação:**
  1. Criar mock de `ExerciseRepository`
  2. Testar cenários:
     - Happy path sem autenticação: retorna exercise sem stats
     - Happy path com autenticação: retorna exercise com stats
     - User nunca executou: stats null/zero
     - Exercise não encontrado: retorna erro
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case

---

## T18 — Criar testes unitários para GetExerciseHistoryUC
- **Objetivo:** Testar lógica de negócio de obter histórico
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_get_exercise_history_test.go` (novo)
- **Implementação:**
  1. Criar mock de `ExerciseRepository`
  2. Testar cenários:
     - Happy path: retorna histórico paginado
     - Sets agrupados por session
     - Ordenação por mais recente
     - User sem histórico: retorna array vazio
     - Exercise não encontrado: retorna erro
     - Paginação inválida
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case

---

## T19 — Criar testes de integração para GET /exercises
- **Objetivo:** Testar endpoint de listagem com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_exercises_integration_test.go` (novo)
- **Implementação:**
  1. Setup: aplicar seed de exercícios
  2. Testar cenários (table-driven):
     - GET /exercises: retorna 200 com lista paginada
     - GET /exercises?muscleGroup=Peito: filtra corretamente
     - GET /exercises?equipment=Barra: filtra corretamente
     - GET /exercises?difficulty=Intermediário: filtra corretamente
     - GET /exercises?search=supino: busca corretamente
     - GET /exercises?page=2&pageSize=10: pagina corretamente
     - GET /exercises com biblioteca vazia: retorna array vazio
     - GET /exercises?page=0: retorna 400
     - GET /exercises?pageSize=200: retorna 400
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Cenários BDD cobertos

---

## T20 — Criar testes de integração para GET /exercises/:id
- **Objetivo:** Testar endpoint de detalhes com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_exercises_integration_test.go`
- **Implementação:**
  1. Setup: criar exercise, user, sessions, set_records
  2. Testar cenários:
     - GET /exercises/:id sem JWT: retorna 200 sem stats
     - GET /exercises/:id com JWT: retorna 200 com stats
     - GET /exercises/:id user nunca executou: stats null/zero
     - GET /exercises/:id exercise não encontrado: retorna 404
     - GET /exercises/invalid-id: retorna 400
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Cenários BDD cobertos
  - Stats calculados corretamente

---

## T21 — Criar testes de integração para GET /exercises/:id/history
- **Objetivo:** Testar endpoint de histórico com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_exercises_integration_test.go`
- **Implementação:**
  1. Setup: criar exercise, user, múltiplas sessions com set_records
  2. Testar cenários:
     - GET /exercises/:id/history com JWT: retorna 200 com histórico
     - Sets agrupados por session
     - Ordenação por mais recente
     - Paginação funciona
     - User sem histórico: retorna array vazio
     - Exercise não encontrado: retorna 404
     - Sem JWT: retorna 401
     - JWT inválido: retorna 401
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Todos os cenários BDD cobertos
  - Histórico agrupado e ordenado corretamente

---

## T22 — Documentar endpoints no README
- **Objetivo:** Atualizar documentação da API
- **Arquivos:**
  - `README.md`
- **Implementação:**
  1. Adicionar seção "Exercise Library" na tabela de endpoints
  2. Adicionar exemplos de request/response para os 3 endpoints
  3. Documentar filtros disponíveis
  4. Documentar estrutura de user stats e histórico
  5. Documentar validações e erros possíveis
  6. Adicionar exemplos de uso com curl
- **Critério de aceite:**
  - Documentação clara e completa
  - Exemplos funcionam corretamente
  - Alinhada com comportamento implementado

---

## T23 — Adicionar comentários Godoc nos use cases
- **Objetivo:** Documentar código para geração de docs
- **Arquivos:**
  - `internal/kinetria/domain/exercises/uc_list_exercises.go`
  - `internal/kinetria/domain/exercises/uc_get_exercise.go`
  - `internal/kinetria/domain/exercises/uc_get_exercise_history.go`
- **Implementação:**
  1. Adicionar comentário de pacote em `doc.go`
  2. Adicionar comentários Godoc em structs e métodos públicos
  3. Documentar parâmetros e retornos
  4. Documentar erros possíveis
- **Critério de aceite:**
  - Comentários seguem padrão Godoc
  - `go doc` exibe documentação corretamente

---

## T24 — Atualizar documentação Swagger
- **Objetivo:** Gerar documentação interativa da API
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_exercises.go`
- **Implementação:**
  1. Adicionar annotations Swagger nos handlers:
     - `@Summary`, `@Description`, `@Tags`
     - `@Accept`, `@Produce`
     - `@Param` (query params, path params)
     - `@Success`, `@Failure`
     - `@Security` (JWT apenas em /history)
  2. Rodar `make swagger` para regenerar docs
  3. Testar endpoints no Swagger UI
- **Critério de aceite:**
  - Swagger UI exibe endpoints de biblioteca
  - Exemplos de request/response corretos
  - Filtros documentados
  - Autenticação JWT funciona no Swagger UI para /history
