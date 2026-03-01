# Plan — Statistics Endpoints

## 1) Inputs usados
- `.thoughts/statistics-endpoints/research-report.md`

## 2) AS-IS (resumo)

### Estrutura atual
- Tabelas `sessions` e `set_records` possuem dados brutos de execução
- Índice existente: `idx_sessions_dashboard` (user_id, started_at DESC, status)
- `SessionRepository` possui método `GetCompletedSessionsByUserAndDateRange`
- `SetRecordRepository` possui apenas métodos básicos (Create, FindBySessionExerciseSet)
- **FALTA:** Queries agregadas para estatísticas (PRs, progressão, frequência)
- Dashboard atual calcula agregações simples em memória (Go)

### Fluxo atual do Dashboard (referência)
1. HTTP Request → Auth Middleware → Handler
2. Use Case agrega dados de múltiplas fontes
3. Repositories executam queries via SQLC
4. Cálculos simples em memória (Go)
5. Response mapeada para DTO

## 3) TO-BE (proposta)

### Interface HTTP
**Endpoints:**
- `GET /api/v1/stats/overview` — Visão geral (workouts, volume, tempo, streak)
- `GET /api/v1/stats/progression` — Progressão ao longo do tempo (gráfico)
- `GET /api/v1/stats/personal-records` — Lista de recordes pessoais
- `GET /api/v1/stats/frequency` — Heatmap de frequência (365 dias)

**Autenticação:** Obrigatória (JWT Bearer) em todos os endpoints

### Contratos

**GET /api/v1/stats/overview**
```
Query params:
- startDate (ISO 8601, opcional, default: 30 dias atrás)
- endDate (ISO 8601, opcional, default: hoje)

Response 200:
{
  "data": {
    "totalWorkouts": 45,
    "totalSets": 540,
    "totalReps": 6480,
    "totalVolume": 518400000,
    "totalTime": 2250,
    "currentStreak": 7,
    "longestStreak": 14,
    "averagePerWeek": 3.5
  }
}
```

**GET /api/v1/stats/progression**
```
Query params:
- startDate (ISO 8601, opcional, default: 30 dias atrás)
- endDate (ISO 8601, opcional, default: hoje)
- exerciseId (uuid, opcional) — Filtrar por exercício específico

Response 200:
{
  "data": {
    "exerciseId": "uuid|null",
    "exerciseName": "string|null",
    "dataPoints": [
      {
        "date": "2026-02-01",
        "value": 12000.5,
        "change": 5.2
      }
    ]
  }
}
```

**GET /api/v1/stats/personal-records**
```
Response 200:
{
  "data": {
    "records": [
      {
        "exerciseId": "uuid",
        "exerciseName": "Supino Reto",
        "weight": 80000,
        "reps": 12,
        "volume": 960000,
        "achievedAt": "2026-02-28T10:30:00Z",
        "previousBest": 75000
      }
    ]
  }
}
```

**GET /api/v1/stats/frequency**
```
Query params:
- startDate (ISO 8601, opcional, default: 365 dias atrás)
- endDate (ISO 8601, opcional, default: hoje)

Response 200:
{
  "data": [
    {
      "date": "2025-03-01",
      "count": 1
    },
    {
      "date": "2025-03-02",
      "count": 0
    }
  ]
}
```

**Erros comuns:**
```
400 Bad Request:
- Período inválido (startDate > endDate)
- Período muito longo (> 2 anos)
- ExerciseID inválido

401 Unauthorized:
- Token JWT ausente ou inválido
```

### Persistência

**Queries SQLC (sessions.sql):**
- `GetStatsByUserAndPeriod` — Agregação de workouts e tempo total
- `GetFrequencyByUserAndPeriod` — Group by date para heatmap
- `GetSessionsForStreak` — Últimos 365 dias para cálculo de streak

**Queries SQLC (set_records.sql):**
- `GetPersonalRecordsByUser` — Window function para PRs por grupo muscular
- `GetProgressionByUserAndExercise` — Agregação por dia (volume, peso máximo)
- `GetTotalSetsRepsVolume` — Agregação para overview

**Índices (opcional, se performance for problema):**
- `idx_set_records_user_stats` — (session_id, workout_exercise_id, weight DESC, reps DESC)
- `idx_sessions_user_date` — (user_id, started_at) WHERE status = 'completed'

### Domain Layer

**Structs de retorno:**
```go
type OverviewStats struct {
    TotalWorkouts  int
    TotalSets      int
    TotalReps      int
    TotalVolume    int64
    TotalTime      int
    CurrentStreak  int
    LongestStreak  int
    AveragePerWeek float64
}

type ProgressionData struct {
    ExerciseID   *uuid.UUID
    ExerciseName *string
    DataPoints   []ProgressionPoint
}

type ProgressionPoint struct {
    Date   time.Time
    Value  float64
    Change float64
}

type PersonalRecord struct {
    ExerciseID   uuid.UUID
    ExerciseName string
    Weight       int
    Reps         int
    Volume       int64
    AchievedAt   time.Time
    PreviousBest *int
}

type FrequencyData struct {
    Date  time.Time
    Count int
}
```

**Use Cases:**
- `GetOverviewUC` — Agrega dados de sessions e set_records, calcula streak
- `GetProgressionUC` — Busca progressão, calcula % de mudança
- `GetPersonalRecordsUC` — Busca PRs ordenados
- `GetFrequencyUC` — Busca frequência, preenche dias vazios

### Validações

**Período:**
- `startDate` <= `endDate`
- Período máximo: 2 anos (730 dias)
- Default: últimos 30 dias se não informado

**ExerciseID:**
- Validar UUID válido
- Verificar se exercício existe (opcional, pode retornar vazio)

### Regras de Negócio

**Personal Record:**
- Maior peso para o mesmo exercício
- Desempate: mais reps, depois mais recente
- Retornar apenas exercício mais usado por grupo muscular (top 15 PRs)

**Streak:**
- Dias consecutivos (sem permitir folga)
- Streak quebra se passar 1 dia sem treinar
- Considerar timezone UTC

**Progression:**
- Métrica principal: volume total (peso × reps)
- Métrica secundária: peso máximo
- Calcular % de mudança entre pontos consecutivos

**Frequency:**
- Últimos 365 dias
- Preencher dias sem treino com count=0

### Observabilidade
- Logs de erro em caso de falha no repository
- Retornar erros de domínio apropriados (validation, not found, internal)

## 4) Decisões e Assunções

1. **Personal Record:** Maior peso, desempate por reps e data. Filtrar por exercício mais usado por grupo muscular (top 15).
2. **Streak:** Dias consecutivos, sem permitir folga. Timezone UTC.
3. **Progression metric:** Volume total (peso × reps) como principal.
4. **Período padrão:** Últimos 30 dias se não informar startDate/endDate.
5. **Limite de período:** Máximo 2 anos (730 dias). Retornar 400 se maior.
6. **Frequency:** Últimos 365 dias, preencher dias vazios com count=0.
7. **Cache:** Não implementar na v1 (adicionar depois se necessário, TTL 5min).
8. **Volumetria esperada:** ~100 sessions por usuário ativo, ~50 set_records por session.

## 5) Riscos / Edge Cases

### Performance
- **Risco:** Personal Records query com JOIN de 3 tabelas + window function pode ser lento
- **Mitigação:** Índices compostos, limitar a top 15 PRs
- **Risco:** Progression query com agregação por dia pode gerar muitos registros
- **Mitigação:** Limitar período máximo (2 anos), paginar se necessário
- **Risco:** Frequency (365 dias) preencher dias vazios em Go pode ser custoso
- **Mitigação:** Fazer em SQL (generate_series) ou cache

### Cálculo de Streak
- **Risco:** Lógica complexa de dias consecutivos
- **Mitigação:** Definir regra clara (sem folga), documentar, testar edge cases
- **Risco:** Timezone (UTC vs timezone do usuário)
- **Mitigação:** Usar UTC por enquanto, adicionar timezone do usuário em v2

### Personal Record
- **Risco:** Empates (mesmo peso, mesmas reps em datas diferentes)
- **Mitigação:** Desempate por data (mais recente ganha)
- **Risco:** Múltiplos exercícios do mesmo grupo muscular
- **Mitigação:** Retornar apenas o mais usado (window function com rank)

### Dados vazios
- **Risco:** Usuário novo sem sessions
- **Mitigação:** Retornar stats com valores zero, não erro
- **Risco:** Período sem treinos
- **Mitigação:** Retornar arrays vazios, não erro

### Validações
- **Risco:** Período inválido (startDate > endDate)
- **Mitigação:** Validar no handler, retornar 400
- **Risco:** Período muito longo (> 2 anos)
- **Mitigação:** Validar no handler, retornar 400
- **Risco:** ExerciseID inválido
- **Mitigação:** Validar UUID, retornar 400 se inválido

## 6) Rollout / Compatibilidade

### Fase 1: Queries e Repository
1. Criar queries agregadas em `sessions.sql` e `set_records.sql`
2. Rodar `make sqlc` para gerar código
3. Implementar métodos nos repositories
4. Testar queries manualmente no psql

### Fase 2: Use Cases
1. Criar use cases para os 4 endpoints
2. Implementar lógica de agregação e cálculos
3. Testar com dados mockados

### Fase 3: HTTP Layer
1. Criar handler com DTOs
2. Implementar 4 handlers
3. Registrar rotas e DI

### Fase 4: Testes e Otimizações
1. Unit tests para use cases
2. Integration tests para endpoints
3. Performance tests (simular 1000+ sessions)
4. Adicionar índices se necessário

### Compatibilidade
- ✅ Endpoints novos não afetam fluxos existentes
- ✅ Queries agregadas não modificam dados
- ✅ Índices são opcionais (adicionar se performance for problema)
