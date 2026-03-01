# Tasks — Statistics Endpoints

## T01 — Criar query GetStatsByUserAndPeriod
- **Objetivo:** Agregar workouts e tempo total por período
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/sessions.sql`
- **Implementação:**
  1. Criar query `GetStatsByUserAndPeriod` que retorna:
     - `total_workouts` (COUNT)
     - `total_time_minutes` (SUM de diferença entre finished_at e started_at)
  2. Filtrar por user_id, status='completed', período (started_at)
  3. Rodar `make sqlc` para gerar código
- **Critério de aceite:**
  - Query compila sem erros no SQLC
  - Código Go gerado em `queries/sessions.sql.go`
  - Query testada manualmente no psql

---

## T02 — Criar query GetFrequencyByUserAndPeriod
- **Objetivo:** Agrupar workouts por dia para heatmap
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/sessions.sql`
- **Implementação:**
  1. Criar query `GetFrequencyByUserAndPeriod` que retorna:
     - `date` (DATE(started_at))
     - `count` (COUNT)
  2. Filtrar por user_id, status='completed', período
  3. GROUP BY DATE(started_at)
  4. ORDER BY date
  5. Rodar `make sqlc`
- **Critério de aceite:**
  - Query compila sem erros
  - Retorna apenas dias com treinos (dias vazios preenchidos em Go)

---

## T03 — Criar query GetSessionsForStreak
- **Objetivo:** Buscar sessions dos últimos 365 dias para cálculo de streak
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/sessions.sql`
- **Implementação:**
  1. Criar query `GetSessionsForStreak` que retorna:
     - `started_at` (apenas data, sem hora)
  2. Filtrar por user_id, status='completed', últimos 365 dias
  3. GROUP BY DATE(started_at) para evitar duplicatas
  4. ORDER BY date DESC
  5. Rodar `make sqlc`
- **Critério de aceite:**
  - Query compila sem erros
  - Retorna lista de datas únicas

---

## T04 — Criar query GetTotalSetsRepsVolume
- **Objetivo:** Agregar sets, reps e volume por período
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/set_records.sql`
- **Implementação:**
  1. Criar query `GetTotalSetsRepsVolume` que retorna:
     - `total_sets` (COUNT)
     - `total_reps` (SUM(reps))
     - `total_volume` (SUM(weight * reps))
  2. JOIN com sessions para filtrar por user_id e período
  3. Filtrar por session.status='completed', set_record.status='completed'
  4. Rodar `make sqlc`
- **Critério de aceite:**
  - Query compila sem erros
  - Agregações corretas (testar manualmente)

---

## T05 — Criar query GetPersonalRecordsByUser
- **Objetivo:** Buscar PRs por grupo muscular (window function)
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/set_records.sql`
- **Implementação:**
  1. Criar CTE `exercise_frequency` que calcula:
     - Frequência de uso de cada exercício
     - Melhor peso, reps, data
  2. Criar CTE `ranked_by_muscle` com window function:
     - ROW_NUMBER() OVER (PARTITION BY primary_muscle ORDER BY times_used DESC, best_weight DESC)
  3. Selecionar apenas rank_in_muscle = 1
  4. ORDER BY best_weight DESC
  5. LIMIT 15
  6. Rodar `make sqlc`
- **Critério de aceite:**
  - Query compila sem erros
  - Retorna apenas 1 exercício por grupo muscular
  - Limitado a top 15

---

## T06 — Criar query GetProgressionByUserAndExercise
- **Objetivo:** Agregar volume e peso máximo por dia
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/set_records.sql`
- **Implementação:**
  1. Criar query `GetProgressionByUserAndExercise` que retorna:
     - `date` (DATE(session.started_at))
     - `max_weight` (MAX(weight))
     - `total_volume` (SUM(weight * reps))
  2. JOIN com sessions, workout_exercises, exercises
  3. Filtrar por user_id, período, exerciseId (opcional)
  4. GROUP BY DATE(started_at)
  5. ORDER BY date
  6. Rodar `make sqlc`
- **Critério de aceite:**
  - Query compila sem erros
  - Filtra por exercício quando fornecido
  - Retorna agregação por dia

---

## T07 — Adicionar métodos no SessionRepository port
- **Objetivo:** Definir contratos para queries de stats
- **Arquivos:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação:**
  1. Adicionar `GetStatsByUserAndPeriod(ctx, userID, start, end) (*SessionStats, error)`
  2. Adicionar `GetFrequencyByUserAndPeriod(ctx, userID, start, end) ([]FrequencyData, error)`
  3. Adicionar `GetSessionsForStreak(ctx, userID) ([]time.Time, error)`
  4. Criar struct `SessionStats` com TotalWorkouts, TotalTime
- **Critério de aceite:**
  - Interface compila sem erros
  - Implementações existentes quebram (esperado, será corrigido em T09)

---

## T08 — Adicionar métodos no SetRecordRepository port
- **Objetivo:** Definir contratos para queries de PRs e progressão
- **Arquivos:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação:**
  1. Adicionar `GetTotalSetsRepsVolume(ctx, userID, start, end) (*SetRecordStats, error)`
  2. Adicionar `GetPersonalRecordsByUser(ctx, userID) ([]PersonalRecord, error)`
  3. Adicionar `GetProgressionByUserAndExercise(ctx, userID, exerciseID, start, end) ([]ProgressionPoint, error)`
  4. Criar structs auxiliares
- **Critério de aceite:**
  - Interface compila sem erros

---

## T09 — Implementar métodos no SessionRepository
- **Objetivo:** Implementar lógica de acesso a dados de sessions
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/session_repository.go`
- **Implementação:**
  1. Implementar `GetStatsByUserAndPeriod()`:
     - Chamar `queries.GetStatsByUserAndPeriod()`
     - Mapear resultado para `SessionStats`
  2. Implementar `GetFrequencyByUserAndPeriod()`:
     - Chamar `queries.GetFrequencyByUserAndPeriod()`
     - Mapear rows para slice de `FrequencyData`
  3. Implementar `GetSessionsForStreak()`:
     - Chamar `queries.GetSessionsForStreak()`
     - Mapear rows para slice de `time.Time`
- **Critério de aceite:**
  - Código compila sem erros
  - Métodos retornam dados corretos (testar manualmente)

---

## T10 — Implementar métodos no SetRecordRepository
- **Objetivo:** Implementar lógica de acesso a dados de set_records
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/set_record_repository.go`
- **Implementação:**
  1. Implementar `GetTotalSetsRepsVolume()`:
     - Chamar `queries.GetTotalSetsRepsVolume()`
     - Mapear resultado para `SetRecordStats`
  2. Implementar `GetPersonalRecordsByUser()`:
     - Chamar `queries.GetPersonalRecordsByUser()`
     - Mapear rows para slice de `PersonalRecord`
  3. Implementar `GetProgressionByUserAndExercise()`:
     - Chamar `queries.GetProgressionByUserAndExercise()`
     - Mapear rows para slice de `ProgressionPoint`
- **Critério de aceite:**
  - Código compila sem erros
  - Métodos retornam dados corretos

---

## T11 — Criar structs de domínio para statistics
- **Objetivo:** Definir tipos de retorno dos use cases
- **Arquivos:**
  - `internal/kinetria/domain/statistics/types.go` (novo)
- **Implementação:**
  1. Criar `OverviewStats` com todos os campos de overview
  2. Criar `ProgressionData` com exerciseId, exerciseName, dataPoints
  3. Criar `ProgressionPoint` com date, value, change
  4. Criar `PersonalRecord` com todos os campos de PR
  5. Criar `FrequencyData` com date, count
- **Critério de aceite:**
  - Structs compilam sem erros
  - Tipos bem documentados com comentários

---

## T12 — Criar use case GetOverviewUC
- **Objetivo:** Implementar lógica de negócio para overview
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_overview.go` (novo)
- **Implementação:**
  1. Criar struct `GetOverviewUC` com dependências dos repositories
  2. Criar struct `GetOverviewInput` com userID, startDate, endDate
  3. Implementar método `Execute(ctx, input) (*OverviewStats, error)`
  4. Validar período (startDate <= endDate, máximo 2 anos)
  5. Aplicar defaults (últimos 30 dias se não informado)
  6. Chamar `sessionRepo.GetStatsByUserAndPeriod()`
  7. Chamar `setRecordRepo.GetTotalSetsRepsVolume()`
  8. Chamar `sessionRepo.GetSessionsForStreak()`
  9. Calcular currentStreak e longestStreak (lógica em Go)
  10. Calcular averagePerWeek
  11. Retornar `OverviewStats`
- **Critério de aceite:**
  - Use case retorna stats corretamente
  - Streak calculado corretamente
  - Validações funcionam

---

## T13 — Criar use case GetProgressionUC
- **Objetivo:** Implementar lógica de negócio para progressão
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_progression.go` (novo)
- **Implementação:**
  1. Criar struct `GetProgressionUC` com dependência `SetRecordRepository`
  2. Criar struct `GetProgressionInput` com userID, exerciseID, startDate, endDate
  3. Implementar método `Execute(ctx, input) (*ProgressionData, error)`
  4. Validar período
  5. Aplicar defaults
  6. Chamar `setRecordRepo.GetProgressionByUserAndExercise()`
  7. Calcular % de mudança entre pontos consecutivos
  8. Retornar `ProgressionData`
- **Critério de aceite:**
  - Use case retorna progressão corretamente
  - % de mudança calculado corretamente
  - Filtra por exercício quando fornecido

---

## T14 — Criar use case GetPersonalRecordsUC
- **Objetivo:** Implementar lógica de negócio para PRs
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_personal_records.go` (novo)
- **Implementação:**
  1. Criar struct `GetPersonalRecordsUC` com dependência `SetRecordRepository`
  2. Implementar método `Execute(ctx, userID) ([]PersonalRecord, error)`
  3. Chamar `setRecordRepo.GetPersonalRecordsByUser()`
  4. Retornar lista de PRs
- **Critério de aceite:**
  - Use case retorna PRs corretamente
  - Ordenado por peso (maior primeiro)
  - Limitado a top 15

---

## T15 — Criar use case GetFrequencyUC
- **Objetivo:** Implementar lógica de negócio para frequência
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_frequency.go` (novo)
- **Implementação:**
  1. Criar struct `GetFrequencyUC` com dependência `SessionRepository`
  2. Criar struct `GetFrequencyInput` com userID, startDate, endDate
  3. Implementar método `Execute(ctx, input) ([]FrequencyData, error)`
  4. Validar período
  5. Aplicar defaults (últimos 365 dias)
  6. Chamar `sessionRepo.GetFrequencyByUserAndPeriod()`
  7. Preencher dias vazios com count=0 (lógica em Go)
  8. Retornar array de 365 `FrequencyData`
- **Critério de aceite:**
  - Use case retorna frequência corretamente
  - Dias vazios preenchidos com count=0
  - Array completo (365 dias ou período especificado)

---

## T16 — Criar DTOs para StatisticsHandler
- **Objetivo:** Definir contratos de request/response
- **Arquivos:**
  - `internal/kinetria/gateways/http/dtos/statistics.go` (novo)
- **Implementação:**
  1. Criar `OverviewResponse` com todos os campos de overview
  2. Criar `ProgressionResponse` com exerciseId, exerciseName, dataPoints
  3. Criar `ProgressionPointDTO` com date, value, change
  4. Criar `PersonalRecordsResponse` com records[]
  5. Criar `PersonalRecordDTO` com todos os campos de PR
  6. Criar `FrequencyResponse` com data[]
  7. Criar `FrequencyDataDTO` com date, count
  8. Adicionar tags JSON
- **Critério de aceite:**
  - DTOs compilam sem erros
  - Serialização JSON funciona corretamente

---

## T17 — Criar StatisticsHandler
- **Objetivo:** Implementar handlers HTTP para endpoints de stats
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_statistics.go` (novo)
- **Implementação:**
  1. Criar struct `StatisticsHandler` com dependências dos 4 use cases
  2. Implementar `HandleGetOverview()`:
     - Extrair userID do JWT
     - Extrair query params (startDate, endDate)
     - Validar inputs
     - Chamar `getOverviewUC.Execute()`
     - Mapear para DTO
     - Retornar JSON 200
     - Tratar erros (400, 401, 500)
  3. Implementar `HandleGetProgression()`:
     - Extrair userID, query params (startDate, endDate, exerciseId)
     - Validar inputs
     - Chamar `getProgressionUC.Execute()`
     - Mapear para DTO
     - Retornar JSON 200
  4. Implementar `HandleGetPersonalRecords()`:
     - Extrair userID
     - Chamar `getPersonalRecordsUC.Execute()`
     - Mapear para DTO
     - Retornar JSON 200
  5. Implementar `HandleGetFrequency()`:
     - Extrair userID, query params (startDate, endDate)
     - Validar inputs
     - Chamar `getFrequencyUC.Execute()`
     - Mapear para DTO
     - Retornar JSON 200
- **Critério de aceite:**
  - Handlers compilam sem erros
  - Retornam status codes corretos
  - Mapeiam erros de domínio para HTTP corretamente

---

## T18 — Registrar rotas de statistics no router
- **Objetivo:** Adicionar endpoints de stats no Chi router
- **Arquivos:**
  - `internal/kinetria/gateways/http/router.go`
- **Implementação:**
  1. Adicionar rotas protegidas:
     - `GET /api/v1/stats/overview` → `statsHandler.HandleGetOverview`
     - `GET /api/v1/stats/progression` → `statsHandler.HandleGetProgression`
     - `GET /api/v1/stats/personal-records` → `statsHandler.HandleGetPersonalRecords`
     - `GET /api/v1/stats/frequency` → `statsHandler.HandleGetFrequency`
  2. Garantir que middleware de autenticação está aplicado
- **Critério de aceite:**
  - Rotas registradas corretamente
  - Middleware de autenticação aplicado
  - Endpoints acessíveis via HTTP

---

## T19 — Registrar dependências no Fx (main.go)
- **Objetivo:** Configurar injeção de dependências
- **Arquivos:**
  - `cmd/kinetria/api/main.go`
- **Implementação:**
  1. Adicionar `statistics.NewGetOverviewUC` no `fx.Provide`
  2. Adicionar `statistics.NewGetProgressionUC` no `fx.Provide`
  3. Adicionar `statistics.NewGetPersonalRecordsUC` no `fx.Provide`
  4. Adicionar `statistics.NewGetFrequencyUC` no `fx.Provide`
  5. Adicionar `http.NewStatisticsHandler` no `fx.Provide`
  6. Garantir que handler é injetado no router
- **Critério de aceite:**
  - Aplicação inicia sem erros
  - Dependências resolvidas corretamente pelo Fx

---

## T20 — Criar testes unitários para GetOverviewUC
- **Objetivo:** Testar lógica de negócio de overview
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_overview_test.go` (novo)
- **Implementação:**
  1. Criar mocks de repositories
  2. Testar cenários (table-driven):
     - Happy path: retorna stats corretamente
     - Calcula streak corretamente (dias consecutivos)
     - Streak quebrado por dia sem treino
     - Usuário sem treinos (retorna zeros)
     - Período inválido (startDate > endDate)
     - Período muito longo (> 2 anos)
     - Defaults aplicados (últimos 30 dias)
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case
  - Lógica de streak testada

---

## T21 — Criar testes unitários para GetProgressionUC
- **Objetivo:** Testar lógica de negócio de progressão
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_progression_test.go` (novo)
- **Implementação:**
  1. Criar mock de `SetRecordRepository`
  2. Testar cenários:
     - Happy path: retorna progressão corretamente
     - Calcula % de mudança corretamente
     - Filtra por exercício quando fornecido
     - Período sem treinos (retorna array vazio)
     - Período inválido
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case
  - % de mudança calculado corretamente

---

## T22 — Criar testes unitários para GetPersonalRecordsUC e GetFrequencyUC
- **Objetivo:** Testar lógica de negócio de PRs e frequência
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_personal_records_test.go` (novo)
  - `internal/kinetria/domain/statistics/uc_get_frequency_test.go` (novo)
- **Implementação:**
  1. Testar `GetPersonalRecordsUC`:
     - Happy path: retorna PRs ordenados
     - Usuário sem treinos (retorna array vazio)
  2. Testar `GetFrequencyUC`:
     - Happy path: retorna frequência com dias vazios preenchidos
     - Usuário sem treinos (todos dias com count=0)
     - Período customizado
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% nos use cases

---

## T23 — Criar testes de integração para GET /stats/overview
- **Objetivo:** Testar endpoint de overview com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_statistics_integration_test.go` (novo)
- **Implementação:**
  1. Setup: criar user, sessions, set_records
  2. Testar cenários:
     - GET /stats/overview com JWT: retorna 200 com stats corretos
     - Período customizado: retorna stats do período
     - Usuário sem treinos: retorna zeros
     - Período inválido: retorna 400
     - Sem JWT: retorna 401
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Cenários BDD cobertos
  - Stats calculados corretamente

---

## T24 — Criar testes de integração para demais endpoints de stats
- **Objetivo:** Testar endpoints de progression, PRs e frequency com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_statistics_integration_test.go`
- **Implementação:**
  1. Setup: criar dados de teste (múltiplas sessions, exercícios, set_records)
  2. Testar `GET /stats/progression`:
     - Com/sem exerciseId
     - % de mudança calculado corretamente
  3. Testar `GET /stats/personal-records`:
     - PRs ordenados corretamente
     - Apenas 1 exercício por grupo muscular
     - Desempate correto
  4. Testar `GET /stats/frequency`:
     - 365 dias retornados
     - Dias vazios com count=0
  5. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Todos os cenários BDD cobertos
  - Agregações e cálculos corretos

---

## T25 — Documentar endpoints no README
- **Objetivo:** Atualizar documentação da API
- **Arquivos:**
  - `README.md`
- **Implementação:**
  1. Adicionar seção "Statistics" na tabela de endpoints
  2. Adicionar exemplos de request/response para os 4 endpoints
  3. Documentar query params disponíveis
  4. Documentar estrutura de retorno (overview, progression, PRs, frequency)
  5. Documentar validações e erros possíveis
  6. Adicionar exemplos de uso com curl
- **Critério de aceite:**
  - Documentação clara e completa
  - Exemplos funcionam corretamente
  - Alinhada com comportamento implementado

---

## T26 — Adicionar comentários Godoc nos use cases
- **Objetivo:** Documentar código para geração de docs
- **Arquivos:**
  - `internal/kinetria/domain/statistics/uc_get_overview.go`
  - `internal/kinetria/domain/statistics/uc_get_progression.go`
  - `internal/kinetria/domain/statistics/uc_get_personal_records.go`
  - `internal/kinetria/domain/statistics/uc_get_frequency.go`
- **Implementação:**
  1. Adicionar comentário de pacote em `doc.go`
  2. Adicionar comentários Godoc em structs e métodos públicos
  3. Documentar parâmetros e retornos
  4. Documentar erros possíveis
  5. Documentar regras de negócio (streak, PR, etc)
- **Critério de aceite:**
  - Comentários seguem padrão Godoc
  - `go doc` exibe documentação corretamente

---

## T27 — Atualizar documentação Swagger
- **Objetivo:** Gerar documentação interativa da API
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_statistics.go`
- **Implementação:**
  1. Adicionar annotations Swagger nos handlers:
     - `@Summary`, `@Description`, `@Tags`
     - `@Accept`, `@Produce`
     - `@Param` (query params)
     - `@Success`, `@Failure`
     - `@Security` (JWT)
  2. Rodar `make swagger` para regenerar docs
  3. Testar endpoints no Swagger UI
- **Critério de aceite:**
  - Swagger UI exibe endpoints de statistics
  - Exemplos de request/response corretos
  - Query params documentados
  - Autenticação JWT funciona no Swagger UI

---

## T28 — (Opcional) Adicionar índices de performance
- **Objetivo:** Otimizar queries agregadas se necessário
- **Arquivos:**
  - `internal/kinetria/gateways/migrations/014_add_stats_indexes.sql` (novo)
- **Implementação:**
  1. Criar migration com índices:
     - `idx_set_records_user_stats` em (session_id, workout_exercise_id, weight DESC, reps DESC)
     - `idx_sessions_user_date` em (user_id, started_at) WHERE status = 'completed'
  2. Testar queries com EXPLAIN ANALYZE antes e depois
  3. Aplicar migration apenas se houver ganho significativo
- **Critério de aceite:**
  - Índices criados corretamente
  - Queries usam índices (verificar com EXPLAIN)
  - Performance melhorada (medir tempo de resposta)
