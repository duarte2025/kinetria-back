# Tasks — foundation-infrastructure

Este backlog detalha todas as tarefas atômicas necessárias para implementar a feature `foundation-infrastructure`. As tarefas estão ordenadas por dependência e podem ser executadas sequencialmente ou em paralelo quando não houver dependências.

---

## T01 — Criar Docker Compose para ambiente de desenvolvimento

**Objetivo**: Configurar Docker Compose com PostgreSQL e aplicação Go para desenvolvimento local.

**Arquivos/pacotes prováveis**:
- `docker-compose.yml` (raiz do projeto) - **CRIAR**
- `Dockerfile` (raiz do projeto) - **CRIAR**
- `.dockerignore` (raiz do projeto) - **CRIAR**

**Implementação (passos)**:

1. Criar `docker-compose.yml`:
   - Serviço `postgres`: imagem `postgres:16-alpine`
   - Environment variables: `POSTGRES_DB=kinetria`, `POSTGRES_USER=kinetria`, `POSTGRES_PASSWORD=kinetria_dev_pass`
   - Ports: `5432:5432`
   - Volume: `postgres_data:/var/lib/postgresql/data`
   - Volume para migrations: `./migrations:/docker-entrypoint-initdb.d`
   - Health check: `pg_isready -U kinetria`
   - Network: `kinetria-network`

2. Criar serviço `app` no `docker-compose.yml`:
   - Build context: `.`, dockerfile: `Dockerfile`
   - Environment variables: todas as vars de `.env.example` + DB vars
   - Ports: `8080:8080`
   - Depends on: `postgres` (condition: `service_healthy`)
   - Volume: `.:/app` (para hot reload no dev)
   - Command: `go run cmd/kinetria/api/main.go`

3. Criar `Dockerfile` (multi-stage):
   - Stage 1 (builder): `golang:1.25.0-alpine`, instalar git/make, copiar go.mod/go.sum, download deps, copiar código, `make build`
   - Stage 2 (runtime): `alpine:latest`, copiar binário de builder, expor porta 8080, CMD executar binário

4. Criar `.dockerignore`:
   - Adicionar: `.git`, `.github`, `.kiro`, `.thoughts`, `bin/`, `coverage.out`, `coverage.html`, `.env`, `.env.local`, `*.log`

**Critério de aceite (testes/checks)**:
- [ ] `docker-compose up -d` executa sem erros
- [ ] Container `kinetria-postgres` está rodando e healthy
- [ ] Container `kinetria-app` está rodando
- [ ] PostgreSQL aceita conexões na porta 5432 (`psql -h localhost -U kinetria -d kinetria` conecta)
- [ ] App está disponível na porta 8080
- [ ] Logs não mostram erros críticos (`docker-compose logs`)
- [ ] `docker-compose down` para tudo corretamente
- [ ] `docker-compose down -v` remove volumes

**Dependências**: Nenhuma (pode ser primeira tarefa)

---

## T02 — Criar migration 001 - users table

**Objetivo**: Criar tabela `users` com schema completo e índices.

**Arquivos/pacotes prováveis**:
- `migrations/001_create_users.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/001_create_users.sql`
2. Adicionar CREATE TABLE users com:
   - Colunas: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `email VARCHAR(255) NOT NULL UNIQUE`, `name VARCHAR(255) NOT NULL`, `password_hash VARCHAR(255) NOT NULL`, `profile_image_url VARCHAR(500)`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
3. Adicionar índice: `CREATE INDEX idx_users_email ON users(email);`
4. Adicionar comentários explicativos no SQL

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/001_create_users.sql`
- [ ] SQL é válido (sem erros de sintaxe)
- [ ] Após `docker-compose up`, tabela `users` existe no banco
- [ ] Query `\d users` no psql mostra todas as colunas corretas
- [ ] Índice `idx_users_email` existe (`\di` no psql)
- [ ] Primary key em `id` está configurada
- [ ] Constraint UNIQUE em `email` funciona (tentar inserir email duplicado falha)

**Dependências**: T01 (Docker Compose precisa existir para testar)

---

## T03 — Criar migration 002 - workouts table

**Objetivo**: Criar tabela `workouts` com campos de type, intensity, duration e image_url.

**Arquivos/pacotes prováveis**:
- `migrations/002_create_workouts.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/002_create_workouts.sql`
2. Criar tabela `workouts` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `name VARCHAR(255) NOT NULL`, `description VARCHAR(500) NOT NULL DEFAULT ''`, `type VARCHAR(50) NOT NULL`, `intensity VARCHAR(50) NOT NULL`, `duration INT NOT NULL DEFAULT 0`, `image_url VARCHAR(500) NOT NULL DEFAULT ''`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
3. Criar índices:
   - `CREATE INDEX idx_workouts_user_id ON workouts(user_id);`
   - `CREATE INDEX idx_workouts_type ON workouts(type);`
   - `CREATE INDEX idx_workouts_user_type ON workouts(user_id, type);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/002_create_workouts.sql`
- [ ] Tabela `workouts` existe com todas as colunas
- [ ] Foreign key para `users(id)` existe com ON DELETE CASCADE
- [ ] Colunas `type`, `intensity`, `duration`, `image_url` existem
- [ ] Default de `duration = 0` e `image_url = ''` funcionam
- [ ] Todos os 3 índices existem
- [ ] Sem ENUM — validação de type/intensity é responsabilidade do use case

**Dependências**: T02 (users table precisa existir)

---

## T04 — Criar migration 003 - exercises table

**Objetivo**: Criar tabela `exercises` vinculada a workouts, com muscles JSONB e campos de configuração de série.

**Arquivos/pacotes prováveis**:
- `migrations/003_create_exercises.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/003_create_exercises.sql`
2. Criar tabela `exercises` com:
   - Colunas: `id UUID PRIMARY KEY`, `workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE`, `name VARCHAR(255) NOT NULL`, `thumbnail_url VARCHAR(500) NOT NULL DEFAULT '/assets/exercises/generic.png'`, `sets INT NOT NULL DEFAULT 1 CHECK (sets >= 1)`, `reps VARCHAR(20) NOT NULL DEFAULT ''`, `muscles JSONB NOT NULL DEFAULT '[]'`, `rest_time INT NOT NULL DEFAULT 60 CHECK (rest_time >= 0)`, `weight DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (weight >= 0)`, `order_index INT NOT NULL DEFAULT 0 CHECK (order_index >= 0)`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
3. Criar índices:
   - `CREATE INDEX idx_exercises_workout_id ON exercises(workout_id);`
   - `CREATE INDEX idx_exercises_order ON exercises(workout_id, order_index);`
   - `CREATE INDEX idx_exercises_muscles ON exercises USING GIN(muscles);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/003_create_exercises.sql`
- [ ] Tabela `exercises` existe com todas as colunas
- [ ] Foreign key para `workouts(id)` existe com ON DELETE CASCADE
- [ ] `muscles` é JSONB (aceita arrays JSON)
- [ ] CHECK constraints funcionam (sets >= 1, rest_time >= 0, weight >= 0)
- [ ] Defaults funcionam: `thumbnail_url`, `sets=1`, `rest_time=60`, `weight=0`, `muscles='[]'`
- [ ] Índice GIN em `muscles` permite queries como `muscles @> '["chest"]'`
- [ ] Todos os 3 índices existem
- [ ] Sem ENUMs de categoria/músculo — muscles é JSONB livre

**Dependências**: T03 (workouts table precisa existir)

---

## T05 — Criar migration 004 - sessions table

**Objetivo**: Criar tabela `sessions` com status corretos, `finished_at` e UNIQUE parcial para sessão ativa.

**Arquivos/pacotes prováveis**:
- `migrations/004_create_sessions.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/004_create_sessions.sql`
2. Criar tabela `sessions` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT`, `status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned'))`, `notes VARCHAR(1000) NOT NULL DEFAULT ''`, `started_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `finished_at TIMESTAMPTZ`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
3. Criar índices:
   - `CREATE UNIQUE INDEX idx_sessions_active_user ON sessions(user_id) WHERE status = 'active';`
   - `CREATE INDEX idx_sessions_user_id ON sessions(user_id);`
   - `CREATE INDEX idx_sessions_status ON sessions(status);`
   - `CREATE INDEX idx_sessions_started_at ON sessions(started_at DESC);`
   - `CREATE INDEX idx_sessions_user_status ON sessions(user_id, status);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/004_create_sessions.sql`
- [ ] Tabela `sessions` existe com todas as colunas
- [ ] Status CHECK constraint: aceita `active|completed|abandoned`, rejeita outros
- [ ] `finished_at` é nullable (NULL = sessão ativa)
- [ ] Foreign key para `users(id)` com CASCADE
- [ ] Foreign key para `workouts(id)` com RESTRICT
- [ ] **UNIQUE parcial** funciona: tentar inserir 2 sessions `active` para o mesmo `user_id` falha
- [ ] Tentar deletar workout com session associada falha (RESTRICT)
- [ ] Todos os 5 índices existem

**Dependências**: T02 (users), T03 (workouts)

---

## T06 — Criar migration 005 - set_records table

**Objetivo**: Criar tabela `set_records` com weight em gramas, status, UNIQUE constraint e `recorded_at`.

**Arquivos/pacotes prováveis**:
- `migrations/005_create_set_records.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/005_create_set_records.sql`
2. Criar tabela `set_records` com:
   - Colunas: `id UUID PRIMARY KEY`, `session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE`, `exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT`, `set_number INT NOT NULL CHECK (set_number >= 1)`, `weight INT NOT NULL DEFAULT 0 CHECK (weight >= 0)`, `reps INT NOT NULL DEFAULT 0 CHECK (reps >= 0)`, `status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'skipped'))`, `recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
   - Constraint: `UNIQUE (session_id, exercise_id, set_number)`
3. Criar índices:
   - `CREATE INDEX idx_set_records_session_id ON set_records(session_id);`
   - `CREATE INDEX idx_set_records_exercise_id ON set_records(exercise_id);`
   - `CREATE INDEX idx_set_records_session_exercise ON set_records(session_id, exercise_id);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/005_create_set_records.sql`
- [ ] Tabela `set_records` existe com todas as colunas
- [ ] `weight` é INT (gramas, não DECIMAL)
- [ ] `status` CHECK: aceita `completed|skipped`, rejeita outros
- [ ] **UNIQUE constraint** `(session_id, exercise_id, set_number)` funciona: inserir duplicata falha
- [ ] `recorded_at` existe (não `created_at`)
- [ ] CHECK constraints funcionam (weight >= 0, reps >= 0, set_number >= 1)
- [ ] Sem `duration_seconds` e `notes` — MVP simplificado
- [ ] Todos os 3 índices existem

**Dependências**: T04 (exercises), T05 (sessions)

---

## T07 — Criar migration 006 - refresh_tokens table

**Objetivo**: Criar tabela `refresh_tokens` com `revoked_at` (pointer/nullable) para suportar autenticação JWT.

**Arquivos/pacotes prováveis**:
- `migrations/006_create_refresh_tokens.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/006_create_refresh_tokens.sql`
2. Criar tabela `refresh_tokens` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `token VARCHAR(255) NOT NULL UNIQUE`, `expires_at TIMESTAMPTZ NOT NULL`, `revoked_at TIMESTAMPTZ`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
3. Criar índices:
   - `CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);`
   - `CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);`
   - `CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);`
   - `CREATE INDEX idx_refresh_tokens_user_revoked ON refresh_tokens(user_id, revoked_at) WHERE revoked_at IS NULL;`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/006_create_refresh_tokens.sql`
- [ ] Tabela `refresh_tokens` existe com todas as colunas
- [ ] Constraint UNIQUE em `token`
- [ ] `revoked_at` é TIMESTAMPTZ nullable (NULL = token válido)
- [ ] Foreign key para `users(id)` com CASCADE
- [ ] Índice parcial `WHERE revoked_at IS NULL` existe
- [ ] Todos os 4 índices existem

**Dependências**: T02 (users)

---

## T08 — Criar migration 007 - audit_log table

**Objetivo**: Criar tabela `audit_log` append-only com `action` como VARCHAR livre, `action_data` JSONB e `occurred_at`.

**Arquivos/pacotes prováveis**:
- `migrations/007_create_audit_log.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/007_create_audit_log.sql`
2. Criar tabela `audit_log` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT`, `entity_type VARCHAR(100) NOT NULL`, `entity_id UUID NOT NULL`, `action VARCHAR(100) NOT NULL`, `action_data JSONB`, `occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `ip_address VARCHAR(45)`, `user_agent TEXT`
3. Criar índices:
   - `CREATE INDEX idx_audit_log_user_occurred ON audit_log(user_id, occurred_at DESC);`
   - `CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);`
   - `CREATE INDEX idx_audit_log_action_data ON audit_log USING GIN(action_data);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/007_create_audit_log.sql`
- [ ] Tabela `audit_log` existe com todas as colunas
- [ ] `action` é VARCHAR livre (não ENUM) — aceita qualquer string
- [ ] `action_data` é JSONB (aceita JSON válido)
- [ ] `occurred_at` existe (não `created_at`)
- [ ] Foreign key para `users(id)` com RESTRICT (não SET NULL)
- [ ] `user_id NOT NULL` (sempre obrigatório)
- [ ] Índice composto `(user_id, occurred_at DESC)` existe
- [ ] Índice GIN em `action_data` permite queries JSON
- [ ] Todos os 3 índices existem

**Dependências**: T02 (users)

---

## T09 — Criar entidade User no domain

**Objetivo**: Definir entidade `User` com todos os campos do schema.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/user.go` - **CRIAR**
- `internal/kinetria/domain/entities/entities.go` - **ATUALIZAR** (remover comentários de exemplo)

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/user.go`
2. Definir `type UserID = uuid.UUID`
3. Criar struct `User` com campos:
   - `ID UserID`
   - `Email string`
   - `Name string`
   - `PasswordHash string`
   - `ProfileImageURL string`
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`
4. Adicionar imports necessários: `time`, `github.com/google/uuid`
5. Atualizar `entities.go`: remover comentários de exemplo e imports não usados

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `user.go` criado
- [ ] Struct `User` possui todos os 7 campos
- [ ] `UserID` é type alias de `uuid.UUID`
- [ ] Código compila sem erros (`go build ./internal/kinetria/domain/entities`)
- [ ] Arquivo `entities.go` não possui comentários de exemplo

**Dependências**: Nenhuma (pode ser feito em paralelo com migrations)

---

## T10 — Criar entidade Workout no domain

**Objetivo**: Definir entidade `Workout` com campos type, intensity, duration e image_url.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/workout.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/workout.go`
2. Definir `type WorkoutID = uuid.UUID`
3. Criar struct `Workout` com campos:
   - `ID WorkoutID`
   - `UserID UserID`
   - `Name string`
   - `Description string` (max 500 chars)
   - `Type string` (`"FORÇA"|"HIPERTROFIA"|"MOBILIDADE"|"CONDICIONAMENTO"`)
   - `Intensity string` (`"BAIXA"|"MODERADA"|"ALTA"`)
   - `Duration int` (minutos, calculado)
   - `ImageURL string` (default baseado no Type)
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`
4. Adicionar imports: `time`, `uuid`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `workout.go` criado
- [ ] Struct `Workout` possui todos os 10 campos
- [ ] `WorkoutID` é type alias de `uuid.UUID`
- [ ] Sem campo `Status` — workouts não têm status
- [ ] Código compila sem erros

**Dependências**: T09 (User entity para UserID)

---

## T11 — Criar entidade Exercise no domain

**Objetivo**: Definir entidade `Exercise` pertencente a um Workout, com muscles JSONB e configuração de série.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/exercise.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/exercise.go`
2. Definir `type ExerciseID = uuid.UUID`
3. Criar struct `Exercise` com campos:
   - `ID ExerciseID`
   - `WorkoutID WorkoutID`
   - `Name string`
   - `ThumbnailURL string` (default: `/assets/exercises/generic.png`)
   - `Sets int` (min 1)
   - `Reps string` (`"8-12"` ou `"10"`)
   - `Muscles []string` (JSONB, ex: `["chest", "triceps"]`)
   - `RestTime int` (segundos, default 60)
   - `Weight float64` (kg, 0 para bodyweight)
   - `OrderIndex int`
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `exercise.go` criado
- [ ] Struct `Exercise` possui todos os 12 campos
- [ ] `ExerciseID` é type alias de `uuid.UUID`
- [ ] `WorkoutID` presente (exercise pertence a workout)
- [ ] `Muscles` é `[]string` (mapeado do JSONB)
- [ ] Sem campos de catálogo global (category, difficulty, equipment)
- [ ] Código compila sem erros

**Dependências**: T09 (UserID), T10 (WorkoutID)

---

## T12 — Criar entidade Session no domain

**Objetivo**: Definir entidade `Session` com status corretos e `FinishedAt` como pointer.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/session.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/session.go`
2. Definir `type SessionID = uuid.UUID`
3. Criar struct `Session` com campos:
   - `ID SessionID`
   - `UserID UserID`
   - `WorkoutID WorkoutID`
   - `Status string` (`"active"|"completed"|"abandoned"`)
   - `Notes string` (max 1000 chars)
   - `StartedAt time.Time`
   - `FinishedAt *time.Time` (pointer — null = sessão ativa)
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `session.go` criado
- [ ] Struct `Session` possui todos os 9 campos
- [ ] `SessionID` é type alias de `uuid.UUID`
- [ ] `FinishedAt` é pointer (`*time.Time`) — null significa ativa
- [ ] Status é `string` simples (não VO — validação no use case)
- [ ] Sem campo `CompletedAt` — substituído por `FinishedAt`
- [ ] Código compila sem erros

**Dependências**: T09 (UserID), T10 (WorkoutID)

---

## T13 — Criar entidade SetRecord no domain

**Objetivo**: Definir entidade `SetRecord` com weight em gramas, status e `RecordedAt`.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/set_record.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/set_record.go`
2. Definir `type SetRecordID = uuid.UUID`
3. Criar struct `SetRecord` com campos:
   - `ID SetRecordID`
   - `SessionID SessionID`
   - `ExerciseID ExerciseID`
   - `SetNumber int` (min 1)
   - `Weight int` (gramas, min 0; use case converte de/para kg)
   - `Reps int` (min 0, 0 = falha)
   - `Status string` (`"completed"|"skipped"`)
   - `RecordedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `set_record.go` criado
- [ ] Struct `SetRecord` possui todos os 8 campos
- [ ] `SetRecordID` é type alias de `uuid.UUID`
- [ ] `Weight` é `int` (gramas) — não float64/DECIMAL
- [ ] `Status` é `string` (não VO)
- [ ] `RecordedAt` presente (não `CreatedAt`)
- [ ] Sem campos `DurationSeconds`, `Notes` — MVP simplificado
- [ ] Código compila sem erros

**Dependências**: T11 (ExerciseID), T12 (SessionID)

---

## T14 — Criar Value Objects WorkoutType e WorkoutIntensity

**Objetivo**: Criar VOs para os campos `type` e `intensity` do Workout.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/workout_type.go` - **CRIAR**
- `internal/kinetria/domain/vos/workout_intensity.go` - **CRIAR**

**Implementação (passos)**:

### workout_type.go
1. Definir `type WorkoutType string`
2. Criar constantes:
   - `WorkoutTypeForca WorkoutType = "FORÇA"`
   - `WorkoutTypeHipertrofia WorkoutType = "HIPERTROFIA"`
   - `WorkoutTypeMobilidade WorkoutType = "MOBILIDADE"`
   - `WorkoutTypeCondicionamento WorkoutType = "CONDICIONAMENTO"`
3. Implementar `Validate() error` e `String() string`

### workout_intensity.go
1. Definir `type WorkoutIntensity string`
2. Criar constantes:
   - `WorkoutIntensityBaixa WorkoutIntensity = "BAIXA"`
   - `WorkoutIntensityModerada WorkoutIntensity = "MODERADA"`
   - `WorkoutIntensityAlta WorkoutIntensity = "ALTA"`
3. Implementar `Validate() error` e `String() string`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `workout_type.go` criado com 4 constantes
- [ ] Arquivo `workout_intensity.go` criado com 3 constantes
- [ ] `Validate()` retorna `nil` para valores válidos
- [ ] `Validate()` retorna erro wrapping `errors.ErrMalformedParameters` para inválidos
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T15 — Criar Value Object SessionStatus

**Objetivo**: Criar VO `SessionStatus` com valores corretos: active, completed, abandoned.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/session_status.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/session_status.go`
2. Definir `type SessionStatus string`
3. Criar constantes:
   - `SessionStatusActive SessionStatus = "active"`
   - `SessionStatusCompleted SessionStatus = "completed"`
   - `SessionStatusAbandoned SessionStatus = "abandoned"`
4. Implementar `Validate()` e `String()` (mesmo padrão de T14)

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado com 3 constantes (active, completed, abandoned)
- [ ] **Sem** `in_progress` e `cancelled` — valores incorretos
- [ ] Método `Validate()` funciona corretamente
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T16 — Criar Value Object SetRecordStatus

**Objetivo**: Criar VO `SetRecordStatus` com valores completed e skipped.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/set_record_status.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/set_record_status.go`
2. Definir `type SetRecordStatus string`
3. Criar constantes:
   - `SetRecordStatusCompleted SetRecordStatus = "completed"`
   - `SetRecordStatusSkipped SetRecordStatus = "skipped"`
4. Implementar `Validate()` e `String()`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado com 2 constantes
- [ ] Validação funciona
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T17 — Criar package constants com defaults e validações

**Objetivo**: Centralizar constantes de defaults de assets e regras de validação alinhadas à arquitetura.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/constants/defaults.go` - **CRIAR**
- `internal/kinetria/domain/constants/validation.go` - **CRIAR**

**Implementação (passos)**:

### defaults.go
1. Criar arquivo `defaults.go`
2. Definir constantes:
   ```go
   const (
       DefaultUserAvatarURL         = "/assets/avatars/default.png"
       DefaultExerciseThumbnailURL  = "/assets/exercises/generic.png"
       DefaultWorkoutImageForca          = "/assets/workouts/forca.png"
       DefaultWorkoutImageHipertrofia    = "/assets/workouts/hipertrofia.png"
       DefaultWorkoutImageMobilidade     = "/assets/workouts/mobilidade.png"
       DefaultWorkoutImageCondicionamento = "/assets/workouts/condicionamento.png"
       DefaultExerciseRestTime      = 60   // segundos
       DefaultExerciseSets          = 1
       DefaultSetWeight             = 0    // gramas (bodyweight)
   )
   ```

### validation.go
1. Criar arquivo `validation.go`
2. Definir constantes de validação:
   ```go
   const (
       MinNameLength        = 1
       MaxNameLength        = 255
       MaxDescriptionLength = 500   // para Workout.Description
       MaxNotesLength       = 1000  // para Session.Notes
       MinSetNumber         = 1
       MaxSetNumber         = 20
       MaxWeight            = 500_000  // gramas (500kg)
       MaxReps              = 100
   )
   ```

**Critério de aceite (testes/checks)**:
- [ ] Ambos os arquivos criados
- [ ] `defaults.go` possui 9 constantes (avatares, thumbnails, imagens de workout, defaults de exercício)
- [ ] `validation.go` possui 8 constantes
- [ ] Constante de `DefaultUserAvatarURL` existe
- [ ] Constantes de `DefaultWorkoutImage*` por tipo existem (Forca, Hipertrofia, Mobilidade, Condicionamento)
- [ ] `MaxDescriptionLength = 500` e `MaxNotesLength = 1000` (não genérico)
- [ ] Código compila sem erros

**Dependências**: Nenhuma

---

## T18 — Criar entidades RefreshToken e AuditLog

**Objetivo**: Definir entidades `RefreshToken` com `RevokedAt *time.Time` e `AuditLog` alinhado à arquitetura.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/refresh_token.go` - **CRIAR**
- `internal/kinetria/domain/entities/audit_log.go` - **CRIAR**

**Implementação (passos)**:

### RefreshToken
1. Criar `refresh_token.go`
2. Definir `type RefreshTokenID = uuid.UUID`
3. Struct `RefreshToken`:
   - `ID RefreshTokenID`
   - `UserID UserID`
   - `Token string` (hash do token, nunca plaintext)
   - `ExpiresAt time.Time`
   - `RevokedAt *time.Time` (pointer — null = válido)
   - `CreatedAt time.Time`

### AuditLog
1. Criar `audit_log.go`
2. Definir `type AuditLogID = uuid.UUID`
3. Struct `AuditLog`:
   - `ID AuditLogID`
   - `UserID UserID` (não nullable — sempre obrigatório)
   - `EntityType string` (ex: `"session"`, `"set_record"`)
   - `EntityID uuid.UUID`
   - `Action string` (ex: `"created"`, `"updated"`, `"completed"`)
   - `ActionData json.RawMessage` (estado antes/depois)
   - `OccurredAt time.Time`
   - `IPAddress string`
   - `UserAgent string`

**Critério de aceite (testes/checks)**:
- [ ] Ambos os arquivos criados
- [ ] `RefreshToken.RevokedAt` é `*time.Time` (pointer, não bool)
- [ ] `RefreshToken.Token` (não `TokenHash`)
- [ ] `AuditLog.UserID` é não-nullable (`UserID`, não `*UserID`)
- [ ] `AuditLog.ActionData` é `json.RawMessage` (não `map[string]interface{}`)
- [ ] `AuditLog.OccurredAt` (não `CreatedAt`)
- [ ] Código compila sem erros

**Dependências**: T09 (UserID)

---

## T19 — Atualizar Config com variáveis de banco de dados

**Objetivo**: Adicionar configurações de DB e HTTP server no gateway config.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/config/config.go` - **ATUALIZAR**
- `.env.example` - **ATUALIZAR**

**Implementação (passos)**:

1. Editar `config/config.go`:
   - Adicionar campos no struct `Config`:
     ```go
     // Database
     DBHost     string `envconfig:"DB_HOST" required:"true"`
     DBPort     int    `envconfig:"DB_PORT" default:"5432"`
     DBUser     string `envconfig:"DB_USER" required:"true"`
     DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
     DBName     string `envconfig:"DB_NAME" required:"true"`
     DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"require"`
     
     // HTTP Server
     HTTPPort int `envconfig:"HTTP_PORT" default:"8080"`
     ```

2. Atualizar `.env.example`:
   ```env
   APP_NAME=kinetria
   ENVIRONMENT=development
   REQUEST_TIMEOUT=5s
   
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=kinetria
   DB_PASSWORD=kinetria_dev_pass
   DB_NAME=kinetria
   DB_SSL_MODE=disable
   
   HTTP_PORT=8080
   ```

**Critério de aceite (testes/checks)**:
- [ ] Struct `Config` possui 7 novos campos (6 DB + 1 HTTP)
- [ ] Tags `envconfig` estão corretas (required/default)
- [ ] `.env.example` possui todas as variáveis documentadas
- [ ] `ParseConfigFromEnv()` funciona sem erros quando vars estão setadas
- [ ] `ParseConfigFromEnv()` retorna erro quando variável required falta
- [ ] Defaults funcionam (ex: `DB_PORT=5432` se não setado)
- [ ] Código compila sem erros

**Dependências**: Nenhuma (pode ser paralelo)

---

## T20 — Criar Database Pool Provider

**Objetivo**: Criar provider Fx para pool de conexões PostgreSQL.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/repositories/pool.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `pool.go` no pacote `repositories`
2. Implementar função `NewDatabasePool`:
   ```go
   func NewDatabasePool(cfg config.Config) (*pgxpool.Pool, error) {
       dsn := fmt.Sprintf(
           "postgres://%s:%s@%s:%d/%s?sslmode=%s",
           cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
       )
       
       pool, err := pgxpool.New(context.Background(), dsn)
       if err != nil {
           return nil, fmt.Errorf("failed to create database pool: %w", err)
       }
       
       if err := pool.Ping(context.Background()); err != nil {
           return nil, fmt.Errorf("failed to ping database: %w", err)
       }
       
       return pool, nil
   }
   ```
3. Adicionar imports: `context`, `fmt`, `pgxpool`, `config`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `pool.go` criado
- [ ] Função `NewDatabasePool` aceita `config.Config` e retorna `(*pgxpool.Pool, error)`
- [ ] DSN é montado corretamente com todas as variáveis de config
- [ ] `pgxpool.New()` é chamado
- [ ] `pool.Ping()` é chamado para validar conexão
- [ ] Erros são wrapped com contexto (`fmt.Errorf` com `%w`)
- [ ] Código compila sem erros
- [ ] **Teste de integração** (ver T26): com DB real, pool conecta; com DB offline, retorna erro

**Dependências**: T19 (Config com DB vars)

---

## T21 — Criar Health Check Handler

**Objetivo**: Implementar handler HTTP para endpoint `/health`.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/http/health/handler.go` - **CRIAR** (novo pacote `health`)

**Implementação (passos)**:

1. Criar diretório `internal/kinetria/gateways/http/health/`
2. Criar arquivo `handler.go`
3. Definir struct de response:
   ```go
   type HealthResponse struct {
       Status  string `json:"status"`
       Service string `json:"service"`
       Version string `json:"version"`
   }
   ```
4. Implementar função provider:
   ```go
   func NewHealthHandler(cfg config.Config) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           resp := HealthResponse{
               Status:  "healthy",
               Service: cfg.AppName,
               Version: "undefined", // TODO: pegar de build vars
           }
           w.Header().Set("Content-Type", "application/json")
           w.WriteHeader(http.StatusOK)
           json.NewEncoder(w).Encode(resp)
       }
   }
   ```
5. Adicionar imports: `net/http`, `encoding/json`, `config`

**Critério de aceite (testes/checks)**:
- [ ] Diretório `health/` criado
- [ ] Arquivo `handler.go` criado
- [ ] Struct `HealthResponse` possui 3 campos com tags JSON
- [ ] Função `NewHealthHandler` retorna `http.HandlerFunc`
- [ ] Response status é 200 OK
- [ ] Content-Type é `application/json`
- [ ] JSON encode funciona sem erros
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T27)

**Dependências**: T19 (Config)

---

## T22 — Registrar providers e rotas no main.go

**Objetivo**: Integrar todos os componentes no Fx DI container e registrar rota `/health`.

**Arquivos/pacotes prováveis**:
- `cmd/kinetria/api/main.go` - **ATUALIZAR**

**Implementação (passos)**:

1. Atualizar imports no `main.go`:
   - `github.com/kinetria/kinetria-back/internal/kinetria/gateways/config`
   - `github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories`
   - `github.com/kinetria/kinetria-back/internal/kinetria/gateways/http/health`
   - `github.com/go-chi/chi/v5`
   - `net/http`

2. Atualizar `fx.New()`:
   ```go
   fx.New(
       fx.Provide(
           config.ParseConfigFromEnv,
           repositories.NewDatabasePool,
           health.NewHealthHandler,
           chi.NewRouter, // Chi router
       ),
       fx.Invoke(
           startHTTPServer,
       ),
   ).Run()
   ```

3. Implementar função `startHTTPServer`:
   ```go
   func startHTTPServer(lc fx.Lifecycle, cfg config.Config, router chi.Router, healthHandler http.HandlerFunc) {
       router.Get("/health", healthHandler)
       
       server := &http.Server{
           Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
           Handler: router,
       }
       
       lc.Append(fx.Hook{
           OnStart: func(ctx context.Context) error {
               go server.ListenAndServe()
               return nil
           },
           OnStop: func(ctx context.Context) error {
               return server.Shutdown(ctx)
           },
       })
   }
   ```

**Critério de aceite (testes/checks)**:
- [ ] Imports atualizados corretamente
- [ ] `fx.Provide` registra config, pool, health handler, router
- [ ] `fx.Invoke` chama `startHTTPServer`
- [ ] Função `startHTTPServer` registra rota `GET /health`
- [ ] Server sobe na porta configurada (default 8080)
- [ ] Lifecycle hooks funcionam (OnStart inicia server, OnStop graceful shutdown)
- [ ] Código compila sem erros
- [ ] **Teste E2E** (ver T28): `go run cmd/kinetria/api/main.go` → `curl /health` → 200 OK

**Dependências**: T19 (Config), T20 (Pool), T21 (Health Handler)

---

## T23 — Criar testes unitários para Value Objects

**Objetivo**: Testar validação de todos os VOs (WorkoutType, WorkoutIntensity, SessionStatus, SetRecordStatus).

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/workout_type_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/workout_intensity_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/session_status_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/set_record_status_test.go` - **CRIAR**

**Implementação (passos)**:

Para cada VO, criar arquivo de teste com:

1. **Teste de valores válidos** (table-driven):
   ```go
   func TestWorkoutType_Validate_ValidValues(t *testing.T) {
       tests := []struct{
           name string
           wt   WorkoutType
       }{
           {"forca", WorkoutTypeForca},
           {"hipertrofia", WorkoutTypeHipertrofia},
           {"mobilidade", WorkoutTypeMobilidade},
           {"condicionamento", WorkoutTypeCondicionamento},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := tt.wt.Validate()
               if err != nil {
                   t.Errorf("expected no error for %s, got %v", tt.name, err)
               }
           })
       }
   }
   ```

2. **Teste de valores inválidos**:
   ```go
   func TestWorkoutType_Validate_InvalidValues(t *testing.T) {
       invalid := WorkoutType("invalid_type")
       err := invalid.Validate()
       if err == nil {
           t.Error("expected error for invalid type, got nil")
       }
       if !errors.Is(err, errors.ErrMalformedParameters) {
           t.Errorf("expected ErrMalformedParameters, got %v", err)
       }
   }
   ```

3. **Teste de String()**:
   ```go
   func TestWorkoutType_String(t *testing.T) {
       wt := WorkoutTypeForca
       if wt.String() != "FORÇA" {
           t.Errorf("expected 'FORÇA', got '%s'", wt.String())
       }
   }
   ```

Repetir para todos os 4 VOs.

**Critério de aceite (testes/checks)**:
- [ ] 4 arquivos de teste criados (*_test.go)
- [ ] Cada arquivo testa: WorkoutType, WorkoutIntensity, SessionStatus, SetRecordStatus
- [ ] Cada arquivo testa valores válidos (tabela com todos os casos)
- [ ] Cada arquivo testa valores inválidos
- [ ] Cada arquivo testa método `String()`
- [ ] `make test` executa todos os testes
- [ ] Todos os testes passam (0 failures)
- [ ] Coverage dos VOs >= 80% (`go test -cover ./internal/kinetria/domain/vos`)

**Dependências**: T14-T16 (VOs implementados)

---

## T24 — Criar teste de integração para Database Pool

**Objetivo**: Testar conexão real com PostgreSQL via Docker.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/repositories/pool_test.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `pool_test.go`
2. Implementar teste que:
   - Verifica se `INTEGRATION_TEST=1` está setado (skip se não for integração)
   - Cria `config.Config` com vars de ambiente ou hardcoded (localhost:5432)
   - Chama `NewDatabasePool(cfg)`
   - Verifica se pool != nil e err == nil
   - Chama `pool.Ping()` e verifica sucesso
   - Fecha pool (`pool.Close()`)

3. Adicionar teste de falha:
   - Config com porta errada (ex: 9999)
   - Espera erro "failed to ping database"

**Exemplo**:
```go
func TestNewDatabasePool_Success(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") != "1" {
        t.Skip("Skipping integration test")
    }
    
    cfg := config.Config{
        DBHost:     "localhost",
        DBPort:     5432,
        DBUser:     "kinetria",
        DBPassword: "kinetria_dev_pass",
        DBName:     "kinetria",
        DBSSLMode:  "disable",
    }
    
    pool, err := NewDatabasePool(cfg)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    defer pool.Close()
    
    if err := pool.Ping(context.Background()); err != nil {
        t.Errorf("expected ping to succeed, got %v", err)
    }
}
```

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `pool_test.go` criado
- [ ] Teste de sucesso (com DB rodando) passa
- [ ] Teste de falha (porta errada) verifica erro esperado
- [ ] Testes usam `INTEGRATION_TEST=1` para gate (skip por padrão)
- [ ] `INTEGRATION_TEST=1 make test` executa e passa (com Docker Compose rodando)
- [ ] Teste limpa recursos (defer pool.Close())

**Dependências**: T20 (Pool implementado), T01 (Docker Compose para rodar teste)

---

## T25 — Criar teste unitário para Health Handler

**Objetivo**: Testar handler `/health` com httptest.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/http/health/handler_test.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `handler_test.go`
2. Implementar teste:
   - Criar `config.Config` com `AppName = "test-service"`
   - Chamar `NewHealthHandler(cfg)`
   - Criar `httptest.NewRequest("GET", "/health", nil)`
   - Criar `httptest.NewRecorder()`
   - Chamar `handler.ServeHTTP(recorder, request)`
   - Verificar status code = 200
   - Verificar Content-Type = "application/json"
   - Decodificar JSON response
   - Verificar campos: `status == "healthy"`, `service == "test-service"`, `version` existe

**Exemplo**:
```go
func TestHealthHandler(t *testing.T) {
    cfg := config.Config{AppName: "test-service"}
    handler := NewHealthHandler(cfg)
    
    req := httptest.NewRequest("GET", "/health", nil)
    rec := httptest.NewRecorder()
    
    handler.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", rec.Code)
    }
    
    if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
        t.Errorf("expected Content-Type application/json, got %s", ct)
    }
    
    var resp HealthResponse
    if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    
    if resp.Status != "healthy" {
        t.Errorf("expected status 'healthy', got '%s'", resp.Status)
    }
    if resp.Service != "test-service" {
        t.Errorf("expected service 'test-service', got '%s'", resp.Service)
    }
}
```

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `handler_test.go` criado
- [ ] Teste verifica status code 200
- [ ] Teste verifica Content-Type
- [ ] Teste decodifica JSON response corretamente
- [ ] Teste verifica todos os campos do response
- [ ] `make test` executa o teste
- [ ] Teste passa

**Dependências**: T23 (Health Handler implementado)

---

## T26 — Teste E2E: Docker Compose + Health Check

**Objetivo**: Validar end-to-end que Docker sobe e `/health` responde.

**Arquivos/pacotes prováveis**:
- Script manual ou documentação em README

**Implementação (passos)**:

1. Com Docker Compose parado, executar:
   ```bash
   docker-compose down -v  # Limpar ambiente
   docker-compose up -d    # Subir serviços
   sleep 10                # Aguardar inicialização
   curl http://localhost:8080/health  # Testar endpoint
   ```

2. Verificar:
   - Response status 200 OK
   - JSON response válido com campos esperados
   - Logs do app não mostram erros (`docker-compose logs app`)
   - PostgreSQL healthy (`docker-compose ps`)

3. Conectar ao banco e verificar tabelas:
   ```bash
   docker exec -it kinetria-postgres psql -U kinetria -d kinetria -c "\dt"
   ```
   - Deve listar 7 tabelas

4. Documentar no README.md:
   - Seção "Como testar localmente"
   - Comandos Docker Compose
   - Comando de teste do health check
   - Comando para verificar migrations

**Critério de aceite (testes/checks)**:
- [ ] `docker-compose up -d` sobe sem erros
- [ ] `curl http://localhost:8080/health` retorna 200 OK
- [ ] Response JSON é válido e contém `{"status":"healthy",...}`
- [ ] `docker-compose logs app` não mostra erros de conexão
- [ ] `docker exec ... \dt` lista 7 tabelas (users, workouts, exercises, sessions, set_records, refresh_tokens, audit_log)
- [ ] README.md possui seção com instruções de teste local

**Dependências**: Todas as tarefas anteriores (T01-T24)

---

## T27 — Documentar setup e arquitetura no README.md

**Objetivo**: Atualizar README.md com instruções completas de Docker, migrations, e estrutura.

**Arquivos/pacotes prováveis**:
- `README.md` - **ATUALIZAR**

**Implementação (passos)**:

1. Adicionar seção "Como rodar localmente com Docker":
   ```markdown
   ## Desenvolvimento com Docker

   ### Pré-requisitos
   - Docker e Docker Compose instalados

   ### Subir ambiente
   ```bash
   docker-compose up -d
   ```

   ### Verificar status
   ```bash
   docker-compose ps
   curl http://localhost:8080/health
   ```

   ### Ver logs
   ```bash
   docker-compose logs -f app
   ```

   ### Conectar ao banco
   ```bash
   docker exec -it kinetria-postgres psql -U kinetria -d kinetria
   ```

   ### Parar ambiente
   ```bash
   docker-compose down
   ```

   ### Resetar banco (apagar volumes)
   ```bash
   docker-compose down -v
   ```
   ```

2. Adicionar seção "Migrations":
   ```markdown
   ## Migrations

   As migrations SQL estão em `migrations/` e são aplicadas automaticamente quando o container PostgreSQL inicia pela primeira vez.

   - `001_create_users.sql` - Tabela de usuários
   - `002_create_workouts.sql` - Planos de treino
   - `003_create_exercises.sql` - Catálogo de exercícios
   - `004_create_sessions.sql` - Sessões de treino
   - `005_create_set_records.sql` - Registros de séries
   - `006_create_refresh_tokens.sql` - Tokens de autenticação
   - `007_create_audit_log.sql` - Log de auditoria

   Para reaplicar migrations, delete o volume: `docker-compose down -v && docker-compose up -d`
   ```

3. Adicionar seção "Estrutura de Domínio":
   ```markdown
   ## Estrutura de Domínio

   ### Entidades
   - `User` - Usuários do sistema
   - `Workout` - Planos de treino personalizados
   - `Exercise` - Catálogo de exercícios
   - `Session` - Sessão de treino ativa
   - `SetRecord` - Registro de série executada
   - `RefreshToken` - Tokens para renovação de autenticação
   - `AuditLog` - Log de auditoria de ações

   ### Value Objects
   - `WorkoutType` - FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO
   - `WorkoutIntensity` - BAIXA, MODERADA, ALTA
   - `SessionStatus` - active, completed, abandoned
   - `SetRecordStatus` - completed, skipped

   ### Constants
   - `defaults.go` - Valores padrão de assets (avatares, thumbnails, imagens de workout)
   - `validation.go` - Regras de validação (min/max lengths, limits)
   ```

4. Atualizar seção de "Testes":
   ```markdown
   ## Testes

   ### Testes unitários
   ```bash
   make test
   ```

   ### Testes de integração (requer Docker)
   ```bash
   docker-compose up -d
   INTEGRATION_TEST=1 make test
   ```

   ### Coverage
   ```bash
   make test-coverage
   open coverage.html
   ```
   ```

**Critério de aceite (testes/checks)**:
- [ ] README.md possui seção "Desenvolvimento com Docker" com comandos
- [ ] README.md lista todas as 7 migrations com descrição
- [ ] README.md documenta estrutura de domínio (entidades, VOs, constants)
- [ ] README.md explica como rodar testes (unitários e integração)
- [ ] Comandos documentados funcionam quando executados
- [ ] Markdown está bem formatado (preview no GitHub)

**Dependências**: T28 (validação E2E completa)

---

## T28 — Documentar API (health endpoint)

**Objetivo**: Documentar o endpoint `/health` no padrão do projeto.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/docs/health.md` - **CRIAR** (ou adicionar no README)

**Implementação (passos)**:

1. Criar diretório `internal/kinetria/docs/` se não existir
2. Criar arquivo `health.md` (ou adicionar seção no README principal)
3. Documentar endpoint:

```markdown
# Health Check API

## GET /health

Endpoint público para verificação de saúde do serviço.

### Request

```http
GET /health HTTP/1.1
```

Sem parâmetros. Sem autenticação necessária.

### Response

**Status:** `200 OK`

**Content-Type:** `application/json`

**Body:**
```json
{
  "status": "healthy",
  "service": "kinetria",
  "version": "undefined"
}
```

### Campos

| Campo | Tipo | Descrição |
|-------|------|-----------|
| status | string | Status do serviço (`healthy` ou `unhealthy`) |
| service | string | Nome do serviço (configurado via `APP_NAME`) |
| version | string | Versão do build (a ser implementado) |

### Exemplos

**cURL:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "kinetria",
  "version": "undefined"
}
```

### Uso

Este endpoint é utilizado por:
- Load balancers (health checks)
- Monitoramento (Prometheus, Datadog, etc.)
- CI/CD pipelines (validar deploy)
- Desenvolvedores (verificar se app está rodando)

### Melhorias Futuras

- Adicionar verificação de conexão com banco de dados
- Incluir métricas de uptime
- Adicionar versão real do build (git commit, tag)
```

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `health.md` criado (ou seção no README)
- [ ] Documentação cobre: método HTTP, path, request, response, campos, exemplos
- [ ] Exemplo de cURL está correto e funcional
- [ ] Response JSON no doc corresponde ao código real
- [ ] Mencionado uso (load balancers, monitoramento, etc.)
- [ ] Markdown bem formatado

**Dependências**: T23 (Health Handler implementado)

---

## T29 — Adicionar comentários Godoc nas entidades e VOs

**Objetivo**: Documentar entidades e VOs exportados com comentários Godoc.

**Arquivos/pacotes prováveis**:
- Todos os arquivos em `internal/kinetria/domain/entities/*.go` - **ATUALIZAR**
- Todos os arquivos em `internal/kinetria/domain/vos/*.go` - **ATUALIZAR**

**Implementação (passos)**:

1. Para cada entidade, adicionar comentário antes do struct:
   ```go
   // User representa um usuário do sistema Kinetria.
   // Contém informações de autenticação (email, password_hash) e metadados (created_at, updated_at).
   type User struct {
       ID           UserID
       Email        string
       Name         string
       PasswordHash string
       CreatedAt    time.Time
       UpdatedAt    time.Time
   }
   ```

2. Para cada VO, adicionar comentário antes do type e constantes:
   ```go
   // WorkoutType representa o tipo de um plano de treino.
   // Valores possíveis: FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO.
   type WorkoutType string
   
   const (
       // WorkoutTypeForca indica treino focado em força máxima.
       WorkoutTypeForca WorkoutType = "FORÇA"
       
       // WorkoutTypeHipertrofia indica treino focado em hipertrofia muscular.
       WorkoutTypeHipertrofia WorkoutType = "HIPERTROFIA"
       
       // WorkoutTypeMobilidade indica treino focado em mobilidade e flexibilidade.
       WorkoutTypeMobilidade WorkoutType = "MOBILIDADE"
       
       // WorkoutTypeCondicionamento indica treino focado em condicionamento cardiovascular.
       WorkoutTypeCondicionamento WorkoutType = "CONDICIONAMENTO"
   )
   
   // Validate verifica se o WorkoutType possui um valor válido.
   // Retorna ErrMalformedParameters se o valor for inválido.
   func (t WorkoutType) Validate() error {
       // ...
   }
   ```

3. Aplicar para todas as entidades (7) e VOs (4)

4. Gerar e verificar godoc:
   ```bash
   go doc internal/kinetria/domain/entities.User
   go doc internal/kinetria/domain/vos.WorkoutType
   ```

**Critério de aceite (testes/checks)**:
- [ ] Todas as 7 entidades possuem comentário Godoc no struct
- [ ] Todos os 4 VOs possuem comentário Godoc no type
- [ ] Constantes exportadas possuem comentários explicativos
- [ ] Métodos exportados (Validate, String) possuem comentários
- [ ] `go doc` exibe documentação corretamente (testar algumas entidades/VOs)
- [ ] Comentários seguem convenção Godoc (começam com nome do tipo/função)
- [ ] Código compila sem erros

**Dependências**: T09-T18 (entidades e VOs implementados)

---

## T30 — Validação final e checklist de aceite da feature

**Objetivo**: Executar todos os testes e validar checklist de aceite da feature completa.

**Arquivos/pacotes prováveis**:
- N/A (validação manual/automática)

**Implementação (passos)**:

1. **Executar validação completa**:
   ```bash
   # Limpar ambiente
   docker-compose down -v
   
   # Subir ambiente
   docker-compose up -d
   
   # Aguardar inicialização
   sleep 15
   
   # Testar health check
   curl http://localhost:8080/health | jq
   
   # Verificar logs (sem erros)
   docker-compose logs app | grep -i error
   
   # Verificar banco
   docker exec -it kinetria-postgres psql -U kinetria -d kinetria -c "\dt"
   docker exec -it kinetria-postgres psql -U kinetria -d kinetria -c "\di"
   docker exec -it kinetria-postgres psql -U kinetria -d kinetria -c "\dT"
   
   # Rodar testes unitários
   make test
   
   # Rodar testes de integração
   INTEGRATION_TEST=1 make test
   
   # Verificar coverage
   make test-coverage
   
   # Rodar linter
   make lint
   ```

2. **Validar checklist do plan.md (seção 9)**:
   - [ ] Infraestrutura: Docker funcional, Postgres conectável, App rodando
   - [ ] Migrations: 7 arquivos criados, aplicados, tabelas/índices/ENUMs existem
   - [ ] Domain: 7 entidades, 5 VOs, 2 arquivos constants
   - [ ] Config: atualizado com DB vars, .env.example completo, pool registrado
   - [ ] Health check: handler implementado, rota registrada, responde 200
   - [ ] Testes: unitários passam, integração passa, E2E validado
   - [ ] Documentação: README atualizado, health.md criado, Godoc adicionado

3. **Gerar relatório**:
   - Criar arquivo `.thoughts/foundation-infrastructure/validation-report.md`
   - Listar todos os checks executados e resultados
   - Incluir outputs de comandos (logs, tabelas, tests)
   - Screenshot ou paste do `curl /health` response
   - Lista de arquivos criados/modificados (git status)

**Critério de aceite (testes/checks)**:
- [ ] Todos os comandos de validação executam sem erros
- [ ] `docker-compose up` funciona
- [ ] `curl /health` retorna 200 com JSON correto
- [ ] Banco possui 7 tabelas
- [ ] `make test` passa (0 failures)
- [ ] `INTEGRATION_TEST=1 make test` passa
- [ ] `make lint` passa (0 issues)
- [ ] Coverage >= 70% (ou meta definida)
- [ ] Checklist do plan.md 100% completo
- [ ] Relatório de validação criado

**Dependências**: Todas as tarefas anteriores (T01-T31)

---

## Resumo de Tarefas

| ID | Tarefa | Tipo | Dependências |
|----|--------|------|--------------|
| T01 | Docker Compose + Dockerfile | Infra | - |
| T02 | Migration 001 - users | Migration | T01 |
| T03 | Migration 002 - workouts | Migration | T02 |
| T04 | Migration 003 - exercises | Migration | T03 |
| T05 | Migration 004 - sessions | Migration | T02, T03 |
| T06 | Migration 005 - set_records | Migration | T04, T05 |
| T07 | Migration 006 - refresh_tokens | Migration | T02 |
| T08 | Migration 007 - audit_log | Migration | T02 |
| T09 | Entity User | Domain | - |
| T10 | Entity Workout | Domain | T09 |
| T11 | Entity Exercise | Domain | T10 |
| T12 | Entity Session | Domain | T09, T10 |
| T13 | Entity SetRecord | Domain | T11, T12 |
| T14 | VOs WorkoutType + WorkoutIntensity | Domain | - |
| T15 | VO SessionStatus | Domain | - |
| T16 | VO SetRecordStatus | Domain | - |
| T17 | Constants (defaults + validation) | Domain | - |
| T18 | Entities RefreshToken + AuditLog | Domain | T09 |
| T19 | Update Config (DB vars) | Gateway | - |
| T20 | Database Pool Provider | Gateway | T19 |
| T21 | Health Check Handler | Gateway | T19 |
| T22 | Register providers in main.go | Integration | T19, T20, T21 |
| T23 | Unit tests - VOs | Tests | T14-T16 |
| T24 | Integration test - Pool | Tests | T20, T01 |
| T25 | Unit test - Health Handler | Tests | T21 |
| T26 | E2E test - Docker + Health | Tests | T01-T22 |
| T27 | Update README - Docker/Migrations | Docs | T26 |
| T28 | Document /health endpoint | Docs | T21 |
| T29 | Add Godoc comments | Docs | T09-T18 |
| T30 | Final validation checklist | Validation | T01-T29 |

**Total: 30 tarefas**

---

## Ordem Sugerida de Implementação

### Fase 1: Infraestrutura Base (paralelo possível)
- T01: Docker Compose
- T19: Config (DB vars)
- T17: Constants

### Fase 2: Migrations (sequencial devido a dependências FK)
- T02: Users table
- T03: Workouts table
- T04: Exercises table (pertence a workouts)
- T05: Sessions table
- T06: Set records table
- T07: Refresh tokens table
- T08: Audit log table

### Fase 3: Domain Layer (paralelo possível)
- T14-T16: Value Objects (em paralelo: WorkoutType/Intensity, SessionStatus, SetRecordStatus)
- T09: User entity
- T10: Workout entity
- T11: Exercise entity
- T12: Session entity
- T13: SetRecord entity
- T18: RefreshToken + AuditLog entities

### Fase 4: Gateway Layer
- T20: Database Pool Provider
- T21: Health Check Handler
- T22: Main.go integration

### Fase 5: Testes
- T23: Unit tests (VOs)
- T25: Unit test (Health)
- T24: Integration test (Pool)
- T26: E2E test

### Fase 6: Documentação
- T27: README
- T28: API docs
- T29: Godoc comments
- T30: Final validation

---

**Estimativa total de esforço**: ~14-18 horas (considerando desenvolvimento + testes + documentação)

**Feature pronta quando**: T30 (validação final) estiver completa com todos os checks passando.
