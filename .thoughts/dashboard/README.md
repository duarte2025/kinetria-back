# Dashboard — Artefatos de Planejamento (Fase Plan)

**Feature**: Dashboard (agregação de dados do usuário)  
**Status**: ✅ Planejamento concluído — Aguardando implementação  
**Data**: 2026-02-23

---

## 📋 Artefatos Criados

### 1. `plan.md` (555 linhas)
Plano completo da feature dashboard incluindo:

- **AS-IS**: Levantamento do estado atual (entidades, repositories, HTTP layer)
- **TO-BE**: Arquitetura proposta com agregação no handler HTTP (seguindo `bff-aggregation-strategy.md`)
- **Use Cases atômicos**: GetUserProfileUC, GetTodayWorkoutUC, GetWeekProgressUC, GetWeekStatsUC
- **Decisões confirmadas**: 
  - ✅ Agregação paralela no handler (4 goroutines)
  - ✅ "Today's workout" = primeiro workout do usuário (MVP sem agendamento)
  - ✅ Calorias estimadas: 7 kcal/min (sem sensor no MVP)
  - ✅ WeekProgress: últimos 7 dias incluindo hoje
- **Riscos mitigados**: N+1 query, timezone mismatch, goroutine leak
- **Observabilidade**: métricas, tracing e logs estruturados

### 2. `test-scenarios.feature` (384 linhas, 38 cenários BDD)
Cenários de teste em Gherkin cobrindo:

- **Happy paths**: dashboard completo, semana zerada, dias futuros
- **Edge cases**: 
  - Usuário sem workouts (todayWorkout = null)
  - Sessões ativas/abandonadas não contam
  - Múltiplas sessões no mesmo dia
  - Sessão iniciada em um dia e terminada em outro
- **Sad paths**: sem autenticação, token expirado, database down
- **Performance**: latência < 500ms
- **Observabilidade**: tracing e logs estruturados

### 3. `tasks.md` (1228 linhas, 12 tasks)
Backlog executável com tarefas detalhadas:

| Task | Descrição | Arquivos afetados |
|------|-----------|-------------------|
| T01 | Criar ports dos repositories | `domain/ports/repositories.go` |
| T02 | Criar queries SQLC | `queries/workouts.sql`, `queries/sessions.sql` |
| T03 | Implementar repositories | `repositories/workout_repository.go`, `repositories/session_repository.go` |
| T04 | Implementar GetUserProfileUC | `domain/dashboard/uc_get_user_profile.go` |
| T05 | Implementar GetTodayWorkoutUC | `domain/dashboard/uc_get_today_workout.go` |
| T06 | Implementar GetWeekProgressUC | `domain/dashboard/uc_get_week_progress.go` |
| T07 | Implementar GetWeekStatsUC | `domain/dashboard/uc_get_week_stats.go` |
| T08 | Implementar DashboardHandler | `gateways/http/handler_dashboard.go` |
| T09 | Registrar rota GET /dashboard | `gateways/http/router.go` |
| T10 | Testes unitários (use cases) | `domain/dashboard/*_test.go` |
| T11 | Testes de integração (handler) | `gateways/http/handler_dashboard_test.go` |
| T12 | Documentar API | `api-contract.yaml`, Godoc, README |

---

## 🎯 Decisões Arquiteturais Principais

### 1. Agregação no Handler HTTP (não no domain)
**Justificativa**: Seguir princípios de Arquitetura Hexagonal — domain deve ser agnóstico ao cliente. Use cases atômicos podem ser reutilizados por diferentes handlers (mobile, web, GraphQL).

**Referência**: `.thoughts/mvp-userflow/bff-aggregation-strategy.md`

### 2. Agregação Paralela com Goroutines
**Benefícios**:
- ⚡ Reduz latência total (4 queries paralelas vs sequenciais)
- 🔒 Compartilha mesmo context (trace, timeout, cancelamento)
- 🧪 Use cases testáveis independentemente

**Padrão**:
```go
// 4 goroutines executando use cases em paralelo
// Canal para sincronizar resultados
// Fail-fast se qualquer use case falhar
```

### 3. Use Cases Atômicos
| Use Case | Responsabilidade |
|----------|------------------|
| GetUserProfileUC | Retornar dados de perfil (nome, email, avatar) |
| GetTodayWorkoutUC | Retornar primeiro workout do usuário (MVP sem agendamento) |
| GetWeekProgressUC | Gerar array de 7 dias (completed/missed/future) |
| GetWeekStatsUC | Calcular calorias e minutos totais da semana |

### 4. Cálculo de Calorias (Estimativa)
**Fórmula**: `totalMinutes * 7 kcal/min`  
**Justificativa**: Sem sensor/wearable no MVP → usar estimativa ACSM para exercício moderado  
**Evolução futura**: Substituir por dados reais de sensores quando disponível

---

## 📦 Dependências

### ⚠️ BLOQUEIO
Esta feature depende de:
- ✅ **Auth** (JWT middleware) — já implementado
- ⏳ **Workouts** (WorkoutRepository + queries SQLC) — planejado mas não implementado
- ⏳ **Sessions** (SessionRepository + queries SQLC) — planejado mas não implementado

### Estratégias de desbloqueio:
- **Opção A (recomendado)**: Implementar features `workouts` e `sessions` primeiro (completas)
- **Opção B**: Implementar stubs dos repositories com queries mínimas para dashboard funcionar isoladamente

---

## 🧪 Estratégia de Testes

| Nível | O que testar | Ferramenta |
|-------|--------------|------------|
| **Use cases (unitário)** | Testar cada use case com mocks de repositories | testify + moq |
| **Handler (integração)** | Testar agregação paralela com mocks dos use cases | httptest |
| **Queries SQLC (integração)** | Testar queries com banco real | testcontainers |
| **E2E** | Endpoint completo com JWT + banco real | curl + jq |

**Cobertura esperada**: > 80%

---

## 🚀 Próximos Passos

### Para o time de implementação:

1. **Revisar artefatos**:
   - [ ] Validar schema `DashboardData` no contrato OpenAPI
   - [ ] Confirmar estimativa de calorias (7 kcal/min) com stakeholders
   - [ ] Confirmar timezone padrão (UTC recomendado)

2. **Escolher estratégia de implementação**:
   - Opção A: Implementar workouts + sessions primeiro
   - Opção B: Implementar dashboard com stubs

3. **Executar tasks em ordem**:
   - T01-T03: Repositories
   - T04-T07: Use cases
   - T08-T09: Handler + rota
   - T10-T12: Testes + docs

4. **Validar implementação**:
   - Rodar testes: `go test ./internal/kinetria/domain/dashboard/... -v -cover`
   - Testar endpoint manualmente com curl
   - Verificar métricas de latência (< 500ms)

---

## 📚 Referências

- `.thoughts/mvp-userflow/api-contract.yaml` — contrato OpenAPI
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — arquitetura geral
- `.thoughts/mvp-userflow/bff-aggregation-strategy.md` — decisão de agregação no handler
- `.thoughts/workouts/plan.md` — dependência: feature workouts
- `.thoughts/sessions/plan.md` — dependência: feature sessions

---

## 📊 Métricas de Planejamento

- **Artefatos criados**: 3 (plan, test-scenarios, tasks)
- **Total de linhas**: 2167 linhas
- **Cenários BDD**: 38 cenários
- **Tasks no backlog**: 12 tasks
- **Tempo estimado**: ~16-24h (2-3 sprints) — depende das features workouts/sessions

---

## ✅ Status de Aprovação

- [ ] **Arquiteto**: revisar decisões arquiteturais (agregação, use cases)
- [ ] **Tech Lead**: revisar backlog de tasks
- [ ] **QA**: revisar cenários BDD
- [ ] **Product Owner**: confirmar estimativa de calorias e "today's workout" sem agendamento

**Pronto para implementação**: Aguardando aprovação + resolução de dependências (workouts/sessions).
