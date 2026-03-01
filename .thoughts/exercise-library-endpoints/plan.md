# Plan — Exercise Library Endpoints

## 1) Inputs usados
- `.thoughts/exercise-library-endpoints/research-report.md`

## 2) AS-IS (resumo)

### Estrutura atual
- Tabela `exercises` possui: `id`, `name`, `thumbnail_url`, `muscles`, `created_at`, `updated_at`
- **FALTA:** Campos `description`, `instructions`, `tips`, `difficulty`, `equipment`, `video_url`
- **FALTA:** Seed de exercícios (tabela vazia)
- Entity `Exercise` não possui campos completos
- `ExerciseRepository` possui apenas métodos auxiliares (`ExistsByIDAndWorkoutID`, `FindWorkoutExerciseID`)
- **FALTA:** Métodos `List`, `GetByID`, `GetUserStats`, `GetHistory`

### Relacionamento atual
```
exercises (1) ----< workout_exercises >---- (N) workouts
                         ↓
                   set_records (via workout_exercise_id)
                         ↓
                      sessions
```

### Fluxo atual
- Exercises são referenciados apenas em `workout_exercises`
- Não há endpoints públicos para listar/buscar exercises
- Frontend usa dados mockados

## 3) TO-BE (proposta)

### Interface HTTP
**Endpoints:**
- `GET /api/v1/exercises` — Listar exercícios com filtros (público ou autenticado)
- `GET /api/v1/exercises/:id` — Detalhes do exercício + stats do usuário (público ou autenticado)
- `GET /api/v1/exercises/:id/history` — Histórico de execução (requer autenticação)

**Autenticação:** 
- GET /exercises e GET /exercises/:id → Opcional (se autenticado, inclui stats)
- GET /exercises/:id/history → Obrigatória (JWT Bearer)

### Contratos

**GET /api/v1/exercises**
```
Query params:
- muscleGroup (string, opcional) — Filtrar por grupo muscular
- equipment (string, opcional) — Filtrar por equipamento
- difficulty (string, opcional) — Filtrar por dificuldade
- search (string, opcional) — Buscar por nome (ILIKE)
- page (int, opcional, default: 1, min: 1)
- pageSize (int, opcional, default: 20, min: 1, max: 100)

Response 200:
{
  "data": [
    {
      "id": "uuid",
      "name": "string",
      "description": "string|null",
      "instructions": "string|null",
      "tips": "string|null",
      "difficulty": "string|null",
      "equipment": "string|null",
      "thumbnailUrl": "string|null",
      "videoUrl": "string|null",
      "muscles": ["string"]
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 50,
    "totalPages": 3
  }
}
```

**GET /api/v1/exercises/:id**
```
Response 200 (não autenticado):
{
  "data": {
    "id": "uuid",
    "name": "string",
    "description": "string|null",
    "instructions": "string|null",
    "tips": "string|null",
    "difficulty": "string|null",
    "equipment": "string|null",
    "thumbnailUrl": "string|null",
    "videoUrl": "string|null",
    "muscles": ["string"]
  }
}

Response 200 (autenticado):
{
  "data": {
    "id": "uuid",
    "name": "string",
    "description": "string|null",
    "instructions": "string|null",
    "tips": "string|null",
    "difficulty": "string|null",
    "equipment": "string|null",
    "thumbnailUrl": "string|null",
    "videoUrl": "string|null",
    "muscles": ["string"],
    "userStats": {
      "lastPerformed": "2026-03-01T10:30:00Z|null",
      "bestWeight": 80000|null,
      "timesPerformed": 15,
      "averageWeight": 75000.5|null
    }
  }
}

Response 404:
{
  "error": {
    "code": "NOT_FOUND",
    "message": "exercise not found"
  }
}
```

**GET /api/v1/exercises/:id/history**
```
Query params:
- page (int, opcional, default: 1, min: 1)
- pageSize (int, opcional, default: 20, min: 1, max: 100)

Response 200:
{
  "data": [
    {
      "sessionId": "uuid",
      "workoutName": "Treino A",
      "performedAt": "2026-03-01T10:30:00Z",
      "sets": [
        {
          "setNumber": 1,
          "reps": 12,
          "weight": 80000,
          "status": "completed"
        },
        {
          "setNumber": 2,
          "reps": 10,
          "weight": 80000,
          "status": "completed"
        }
      ]
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 45,
    "totalPages": 3
  }
}

Response 401:
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "authentication required"
  }
}

Response 404:
{
  "error": {
    "code": "NOT_FOUND",
    "message": "exercise not found"
  }
}
```

### Persistência

**Migration 011:** `011_expand_exercises_table.sql`
```sql
ALTER TABLE exercises 
ADD COLUMN description TEXT,
ADD COLUMN instructions TEXT,
ADD COLUMN tips TEXT,
ADD COLUMN difficulty VARCHAR(50),
ADD COLUMN equipment VARCHAR(100),
ADD COLUMN video_url TEXT;

CREATE INDEX idx_exercises_name ON exercises USING gin(to_tsvector('portuguese', name));
CREATE INDEX idx_exercises_muscles ON exercises USING gin(muscles);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX idx_exercises_equipment ON exercises(equipment);
```

**Migration 012 (ou seed script):** `012_seed_exercises.sql`
- 30-40 exercícios mais comuns
- Conteúdo genérico/público
- URLs mock para thumbnails e vídeos

**Queries SQLC:**
- `ListExercises` — Listar com filtros e paginação
- `CountExercises` — Contar total com filtros
- `GetExerciseByID` — Buscar por ID
- `GetExerciseUserStats` — Calcular stats do usuário (JOIN com sessions e set_records)
- `GetExerciseHistory` — Buscar histórico de execução (JOIN com sessions, workouts, set_records)
- `CountExerciseHistory` — Contar total de sessões no histórico

### Domain Layer

**Entity Exercise (adicionar campos):**
```go
Description  *string
Instructions *string
Tips         *string
Difficulty   *string
Equipment    *string
VideoURL     *string
```

**Structs auxiliares:**
```go
type ExerciseFilters struct {
    MuscleGroup *string
    Equipment   *string
    Difficulty  *string
    Search      *string
}

type ExerciseWithStats struct {
    Exercise       *entities.Exercise
    UserStats      *ExerciseUserStats
}

type ExerciseUserStats struct {
    LastPerformed  *time.Time
    BestWeight     *int
    TimesPerformed int
    AverageWeight  *float64
}

type ExerciseHistoryEntry struct {
    SessionID   uuid.UUID
    WorkoutName string
    PerformedAt time.Time
    Sets        []SetDetail
}

type SetDetail struct {
    SetNumber int
    Reps      int
    Weight    *int
    Status    string
}
```

**Use Cases:**
- `ListExercisesUC` — Valida filtros, chama repository
- `GetExerciseUC` — Busca exercise, opcionalmente busca stats do usuário
- `GetExerciseHistoryUC` — Busca histórico, agrupa sets por session

### Validações

**Filtros:**
- `page` >= 1
- `pageSize` >= 1 e <= 100
- `muscleGroup`, `equipment`, `difficulty`, `search` — Aceitar qualquer string (validação de enum pode vir depois)

**ExerciseID:**
- Validar UUID válido
- Retornar 404 se não encontrado

### Observabilidade
- Logs de erro em caso de falha no repository
- Retornar erros de domínio apropriados (validation, not found, internal)

## 4) Decisões e Assunções

1. **Biblioteca read-only:** Usuários não criam/editam exercícios (apenas admin via seed/migration)
2. **Autenticação opcional em GET /exercises e GET /exercises/:id:** Se autenticado, inclui user stats
3. **Autenticação obrigatória em GET /exercises/:id/history:** Requer JWT
4. **Seed inicial:** 30-40 exercícios mais comuns (conteúdo genérico)
5. **URLs mock:** Thumbnails e vídeos usam URLs mock por enquanto (CDN/S3 fica para v2)
6. **Campo `muscles` como TEXT[]:** Valores livres (enum pode vir depois)
7. **Search apenas em `name`:** ILIKE simples (full-text search em description fica para v2)
8. **History completo:** Mostrar todas execuções, não apenas best sets
9. **User stats:** lastPerformed, bestWeight, timesPerformed, averageWeight (últimas 10 execuções)

## 5) Riscos / Edge Cases

### Seed de Exercícios
- **Risco:** Criar conteúdo para 30-40 exercícios (descrições, instruções, dicas) é trabalhoso
- **Mitigação:** Começar com 20 exercícios mais comuns, expandir depois. Usar conteúdo genérico.

### Performance
- **Risco:** Query de user stats faz JOIN de 4 tabelas (exercises, workout_exercises, set_records, sessions)
- **Mitigação:** Índices compostos, limitar a últimas 10 execuções. Cache de stats se necessário.
- **Risco:** Busca full-text com índice GIN pode ser lenta
- **Mitigação:** Limitar resultados (max 100), usar paginação

### Dados vazios
- **Risco:** Biblioteca vazia se não rodar seed
- **Mitigação:** Retornar array vazio, não erro. Documentar necessidade de seed.
- **Risco:** User sem histórico
- **Mitigação:** Stats retornam null/zero, não erro

### Validações
- **Risco:** Filtros inválidos (difficulty, equipment com valores não esperados)
- **Mitigação:** Aceitar qualquer string por enquanto (enum pode vir depois)

### Autenticação opcional
- **Risco:** Complexidade no handler (verificar JWT mas não exigir)
- **Mitigação:** Usar helper que tenta extrair userID mas não retorna erro se falhar

## 6) Rollout / Compatibilidade

### Fase 1: Migration e Seed
1. Aplicar migration `011_expand_exercises_table.sql`
2. Aplicar seed `012_seed_exercises.sql` (ou rodar script Go)
3. Verificar que exercises existem no banco

### Fase 2: Backend
1. Atualizar entity `Exercise` com novos campos
2. Implementar queries SQLC
3. Implementar métodos no `ExerciseRepository`
4. Criar use cases
5. Criar handler com rotas

### Fase 3: Testes
1. Unit tests para use cases
2. Integration tests para endpoints (com seed)
3. Edge cases (filtros, paginação, user sem histórico)

### Compatibilidade
- ✅ Migration é aditiva (não quebra código existente)
- ✅ Novos campos têm default NULL (exercises antigos funcionam)
- ✅ Endpoints novos não afetam fluxos existentes
- ✅ Seed pode ser aplicado depois (endpoints retornam array vazio)
