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
   - Colunas: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `email VARCHAR(255) NOT NULL UNIQUE`, `name VARCHAR(255) NOT NULL`, `password_hash VARCHAR(255) NOT NULL`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
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

**Objetivo**: Criar ENUM `workout_status` e tabela `workouts` com relacionamento a `users`.

**Arquivos/pacotes prováveis**:
- `migrations/002_create_workouts.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/002_create_workouts.sql`
2. Criar ENUM: `CREATE TYPE workout_status AS ENUM ('draft', 'published', 'archived');`
3. Criar tabela `workouts` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `name VARCHAR(255) NOT NULL`, `description TEXT`, `status workout_status NOT NULL DEFAULT 'draft'`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
4. Criar índices:
   - `CREATE INDEX idx_workouts_user_id ON workouts(user_id);`
   - `CREATE INDEX idx_workouts_status ON workouts(status);`
   - `CREATE INDEX idx_workouts_user_status ON workouts(user_id, status);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/002_create_workouts.sql`
- [ ] ENUM `workout_status` existe (`\dT` no psql)
- [ ] Tabela `workouts` existe com todas as colunas
- [ ] Foreign key para `users(id)` existe com ON DELETE CASCADE
- [ ] Inserir workout com `user_id` inválido falha (constraint)
- [ ] Deletar user cascade deleta seus workouts (testar)
- [ ] Todos os 3 índices existem
- [ ] Default status é `draft` ao inserir sem especificar

**Dependências**: T02 (users table precisa existir)

---

## T04 — Criar migration 003 - exercises table

**Objetivo**: Criar ENUMs de categoria/músculo e tabela `exercises` para catálogo.

**Arquivos/pacotes prováveis**:
- `migrations/003_create_exercises.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/003_create_exercises.sql`
2. Criar ENUMs:
   - `CREATE TYPE exercise_category AS ENUM ('strength', 'cardio', 'flexibility', 'balance');`
   - `CREATE TYPE muscle_group AS ENUM ('chest', 'back', 'legs', 'shoulders', 'arms', 'core', 'full_body');`
3. Criar tabela `exercises` com:
   - Colunas: `id UUID PRIMARY KEY`, `name VARCHAR(255) NOT NULL`, `description TEXT`, `category exercise_category NOT NULL`, `primary_muscle_group muscle_group NOT NULL`, `equipment_required VARCHAR(255)`, `difficulty_level INT NOT NULL CHECK (difficulty_level BETWEEN 1 AND 5)`, `video_url VARCHAR(500)`, `thumbnail_url VARCHAR(500)`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
4. Criar índices:
   - `CREATE INDEX idx_exercises_category ON exercises(category);`
   - `CREATE INDEX idx_exercises_muscle_group ON exercises(primary_muscle_group);`
   - `CREATE INDEX idx_exercises_difficulty ON exercises(difficulty_level);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/003_create_exercises.sql`
- [ ] ENUMs `exercise_category` e `muscle_group` existem
- [ ] Tabela `exercises` existe com todas as colunas
- [ ] Constraint CHECK em `difficulty_level` funciona (tentar inserir 0 ou 6 falha)
- [ ] Inserir exercise com category inválida falha
- [ ] Todos os 3 índices existem
- [ ] Colunas nullable (equipment_required, video_url, thumbnail_url) aceitam NULL

**Dependências**: T01 (independente de outras migrations)

---

## T05 — Criar migration 004 - sessions table

**Objetivo**: Criar ENUM `session_status` e tabela `sessions` relacionada a `users` e `workouts`.

**Arquivos/pacotes prováveis**:
- `migrations/004_create_sessions.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/004_create_sessions.sql`
2. Criar ENUM: `CREATE TYPE session_status AS ENUM ('in_progress', 'completed', 'cancelled');`
3. Criar tabela `sessions` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT`, `status session_status NOT NULL DEFAULT 'in_progress'`, `started_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `completed_at TIMESTAMPTZ`, `notes TEXT`, `created_at TIMESTAMPTZ`, `updated_at TIMESTAMPTZ`
4. Criar índices:
   - `CREATE INDEX idx_sessions_user_id ON sessions(user_id);`
   - `CREATE INDEX idx_sessions_status ON sessions(status);`
   - `CREATE INDEX idx_sessions_started_at ON sessions(started_at DESC);`
   - `CREATE INDEX idx_sessions_user_status ON sessions(user_id, status);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/004_create_sessions.sql`
- [ ] ENUM `session_status` existe
- [ ] Tabela `sessions` existe com todas as colunas
- [ ] Foreign key para `users(id)` com CASCADE
- [ ] Foreign key para `workouts(id)` com RESTRICT
- [ ] Tentar deletar workout com session ativa falha (RESTRICT)
- [ ] Deletar user cascade deleta sessions
- [ ] `completed_at` é nullable
- [ ] Todos os 4 índices existem
- [ ] Default status é `in_progress`

**Dependências**: T02 (users), T03 (workouts)

---

## T06 — Criar migration 005 - set_records table

**Objetivo**: Criar tabela `set_records` para registrar séries executadas em sessions.

**Arquivos/pacotes prováveis**:
- `migrations/005_create_set_records.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/005_create_set_records.sql`
2. Criar tabela `set_records` com:
   - Colunas: `id UUID PRIMARY KEY`, `session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE`, `exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT`, `set_number INT NOT NULL CHECK (set_number > 0)`, `reps INT CHECK (reps >= 0)`, `weight_kg DECIMAL(6,2) CHECK (weight_kg >= 0)`, `duration_seconds INT CHECK (duration_seconds >= 0)`, `notes TEXT`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
3. Criar índices:
   - `CREATE INDEX idx_set_records_session_id ON set_records(session_id);`
   - `CREATE INDEX idx_set_records_exercise_id ON set_records(exercise_id);`
   - `CREATE INDEX idx_set_records_session_exercise ON set_records(session_id, exercise_id);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/005_create_set_records.sql`
- [ ] Tabela `set_records` existe com todas as colunas
- [ ] Foreign key para `sessions(id)` com CASCADE
- [ ] Foreign key para `exercises(id)` com RESTRICT
- [ ] CHECK constraints funcionam (reps, weight_kg, duration_seconds >= 0, set_number > 0)
- [ ] Tentar inserir `set_number = 0` falha
- [ ] Tentar inserir `weight_kg = -5` falha
- [ ] Campos nullable (reps, weight_kg, duration_seconds) aceitam NULL
- [ ] Tipo DECIMAL(6,2) armazena peso corretamente (ex: 125.50)
- [ ] Todos os 3 índices existem

**Dependências**: T04 (sessions), T04 (exercises)

---

## T07 — Criar migration 006 - refresh_tokens table

**Objetivo**: Criar tabela `refresh_tokens` para suportar autenticação JWT.

**Arquivos/pacotes prováveis**:
- `migrations/006_create_refresh_tokens.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/006_create_refresh_tokens.sql`
2. Criar tabela `refresh_tokens` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `token_hash VARCHAR(255) NOT NULL UNIQUE`, `expires_at TIMESTAMPTZ NOT NULL`, `revoked BOOLEAN NOT NULL DEFAULT FALSE`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
3. Criar índices:
   - `CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);`
   - `CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);`
   - `CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);`
   - `CREATE INDEX idx_refresh_tokens_user_revoked ON refresh_tokens(user_id, revoked);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/006_create_refresh_tokens.sql`
- [ ] Tabela `refresh_tokens` existe com todas as colunas
- [ ] Constraint UNIQUE em `token_hash`
- [ ] Tentar inserir token_hash duplicado falha
- [ ] Foreign key para `users(id)` com CASCADE
- [ ] Default `revoked = FALSE`
- [ ] Todos os 4 índices existem

**Dependências**: T02 (users)

---

## T08 — Criar migration 007 - audit_log table

**Objetivo**: Criar ENUM `audit_action` e tabela `audit_log` para rastreabilidade.

**Arquivos/pacotes prováveis**:
- `migrations/007_create_audit_log.sql` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `migrations/007_create_audit_log.sql`
2. Criar ENUM `audit_action` com valores:
   - `'user_created', 'user_updated', 'user_deleted'`
   - `'workout_created', 'workout_updated', 'workout_deleted'`
   - `'session_started', 'session_completed', 'session_cancelled'`
   - `'set_recorded', 'set_updated', 'set_deleted'`
   - `'login', 'logout', 'token_refreshed'`
3. Criar tabela `audit_log` com:
   - Colunas: `id UUID PRIMARY KEY`, `user_id UUID REFERENCES users(id) ON DELETE SET NULL`, `action audit_action NOT NULL`, `entity_type VARCHAR(100)`, `entity_id UUID`, `metadata JSONB`, `ip_address INET`, `user_agent TEXT`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
4. Criar índices:
   - `CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);`
   - `CREATE INDEX idx_audit_log_action ON audit_log(action);`
   - `CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);`
   - `CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);`
   - `CREATE INDEX idx_audit_log_metadata ON audit_log USING GIN(metadata);`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado em `migrations/007_create_audit_log.sql`
- [ ] ENUM `audit_action` existe com todos os valores
- [ ] Tabela `audit_log` existe com todas as colunas
- [ ] Foreign key para `users(id)` com ON DELETE SET NULL
- [ ] Deletar user seta `user_id = NULL` no audit log (preserva histórico)
- [ ] Campo `metadata` é JSONB (aceita JSON válido)
- [ ] Índice GIN em `metadata` permite queries JSON (ex: `metadata @> '{"key": "value"}'`)
- [ ] Campo `ip_address` é tipo INET (aceita IPs válidos, rejeita inválidos)
- [ ] Todos os 5 índices existem

**Dependências**: T02 (users - para foreign key, mas com SET NULL pode ser independente)

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
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`
4. Adicionar imports necessários: `time`, `github.com/google/uuid`
5. Atualizar `entities.go`: remover comentários de exemplo e imports não usados

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `user.go` criado
- [ ] Struct `User` possui todos os 6 campos
- [ ] `UserID` é type alias de `uuid.UUID`
- [ ] Código compila sem erros (`go build ./internal/kinetria/domain/entities`)
- [ ] Arquivo `entities.go` não possui comentários de exemplo

**Dependências**: Nenhuma (pode ser feito em paralelo com migrations)

---

## T10 — Criar entidade Workout no domain

**Objetivo**: Definir entidade `Workout` com relacionamento a `User` e status.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/workout.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/workout.go`
2. Definir `type WorkoutID = uuid.UUID`
3. Criar struct `Workout` com campos:
   - `ID WorkoutID`
   - `UserID UserID`
   - `Name string`
   - `Description string`
   - `Status vos.WorkoutStatus` (importar de vos - criar stub se ainda não existir)
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`
4. Adicionar imports: `time`, `uuid`, `internal/kinetria/domain/vos`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `workout.go` criado
- [ ] Struct `Workout` possui todos os 7 campos
- [ ] `WorkoutID` é type alias de `uuid.UUID`
- [ ] Campo `Status` usa tipo `vos.WorkoutStatus`
- [ ] Código compila sem erros

**Dependências**: T09 (User entity para UserID), T14 (WorkoutStatus VO - ou criar stub)

---

## T11 — Criar entidade Exercise no domain

**Objetivo**: Definir entidade `Exercise` com categoria, músculo e metadados.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/exercise.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/exercise.go`
2. Definir `type ExerciseID = uuid.UUID`
3. Criar struct `Exercise` com campos:
   - `ID ExerciseID`
   - `Name string`
   - `Description string`
   - `Category vos.ExerciseCategory`
   - `PrimaryMuscleGroup vos.MuscleGroup`
   - `EquipmentRequired string`
   - `DifficultyLevel int`
   - `VideoURL string`
   - `ThumbnailURL string`
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `exercise.go` criado
- [ ] Struct `Exercise` possui todos os 11 campos
- [ ] `ExerciseID` é type alias de `uuid.UUID`
- [ ] Campos `Category` e `PrimaryMuscleGroup` usam VOs do pacote `vos`
- [ ] Código compila sem erros

**Dependências**: T16, T17 (VOs de ExerciseCategory e MuscleGroup)

---

## T12 — Criar entidade Session no domain

**Objetivo**: Definir entidade `Session` para rastreamento de treinos ativos.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/session.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/session.go`
2. Definir `type SessionID = uuid.UUID`
3. Criar struct `Session` com campos:
   - `ID SessionID`
   - `UserID UserID`
   - `WorkoutID WorkoutID`
   - `Status vos.SessionStatus`
   - `StartedAt time.Time`
   - `CompletedAt *time.Time` (pointer - nullable)
   - `Notes string`
   - `CreatedAt time.Time`
   - `UpdatedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `session.go` criado
- [ ] Struct `Session` possui todos os 9 campos
- [ ] `SessionID` é type alias de `uuid.UUID`
- [ ] `CompletedAt` é pointer (`*time.Time`) para representar nullable
- [ ] Campo `Status` usa `vos.SessionStatus`
- [ ] Código compila sem erros

**Dependências**: T09 (UserID), T10 (WorkoutID), T15 (SessionStatus VO)

---

## T13 — Criar entidade SetRecord no domain

**Objetivo**: Definir entidade `SetRecord` para registros de séries.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/set_record.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/entities/set_record.go`
2. Definir `type SetRecordID = uuid.UUID`
3. Criar struct `SetRecord` com campos:
   - `ID SetRecordID`
   - `SessionID SessionID`
   - `ExerciseID ExerciseID`
   - `SetNumber int`
   - `Reps *int` (pointer - nullable)
   - `WeightKg *float64` (pointer - nullable, usar float64 para DECIMAL)
   - `DurationSeconds *int` (pointer - nullable)
   - `Notes string`
   - `CreatedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `set_record.go` criado
- [ ] Struct `SetRecord` possui todos os 9 campos
- [ ] `SetRecordID` é type alias de `uuid.UUID`
- [ ] `Reps`, `WeightKg`, `DurationSeconds` são pointers (nullable)
- [ ] `WeightKg` é `*float64` (precisão decimal)
- [ ] `SetNumber` é `int` não-nullable
- [ ] Código compila sem erros

**Dependências**: T11 (ExerciseID), T12 (SessionID)

---

## T14 — Criar Value Object WorkoutStatus

**Objetivo**: Criar VO `WorkoutStatus` com enum e validação.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/workout_status.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/workout_status.go`
2. Definir `type WorkoutStatus string`
3. Criar constantes:
   - `WorkoutStatusDraft WorkoutStatus = "draft"`
   - `WorkoutStatusPublished WorkoutStatus = "published"`
   - `WorkoutStatusArchived WorkoutStatus = "archived"`
4. Implementar método `func (s WorkoutStatus) Validate() error`:
   - Switch case para cada valor válido: retornar `nil`
   - Default: retornar erro wrapping `errors.ErrMalformedParameters`
5. Implementar método `func (s WorkoutStatus) String() string`: retornar `string(s)`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo `workout_status.go` criado
- [ ] Type `WorkoutStatus` definido como string
- [ ] 3 constantes criadas (Draft, Published, Archived)
- [ ] Método `Validate()` retorna `nil` para valores válidos
- [ ] Método `Validate()` retorna erro para valor inválido (ex: "invalid_status")
- [ ] Erro wraps `errors.ErrMalformedParameters`
- [ ] Método `String()` retorna string corretamente
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma (pode ser paralelo)

---

## T15 — Criar Value Object SessionStatus

**Objetivo**: Criar VO `SessionStatus` com enum e validação.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/session_status.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/session_status.go`
2. Definir `type SessionStatus string`
3. Criar constantes:
   - `SessionStatusInProgress SessionStatus = "in_progress"`
   - `SessionStatusCompleted SessionStatus = "completed"`
   - `SessionStatusCancelled SessionStatus = "cancelled"`
4. Implementar `Validate()` e `String()` (mesmo padrão de T14)

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado com 3 constantes
- [ ] Método `Validate()` funciona corretamente
- [ ] Método `String()` funciona corretamente
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T16 — Criar Value Object ExerciseCategory

**Objetivo**: Criar VO `ExerciseCategory` com enum e validação.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/exercise_category.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/exercise_category.go`
2. Definir `type ExerciseCategory string`
3. Criar constantes:
   - `ExerciseCategoryStrength ExerciseCategory = "strength"`
   - `ExerciseCategoryCardio ExerciseCategory = "cardio"`
   - `ExerciseCategoryFlexibility ExerciseCategory = "flexibility"`
   - `ExerciseCategoryBalance ExerciseCategory = "balance"`
4. Implementar `Validate()` e `String()`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado com 4 constantes
- [ ] Validação funciona
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T17 — Criar Value Object MuscleGroup

**Objetivo**: Criar VO `MuscleGroup` com enum e validação.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/muscle_group.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/muscle_group.go`
2. Definir `type MuscleGroup string`
3. Criar constantes (7 valores):
   - `MuscleGroupChest MuscleGroup = "chest"`
   - `MuscleGroupBack MuscleGroup = "back"`
   - `MuscleGroupLegs MuscleGroup = "legs"`
   - `MuscleGroupShoulders MuscleGroup = "shoulders"`
   - `MuscleGroupArms MuscleGroup = "arms"`
   - `MuscleGroupCore MuscleGroup = "core"`
   - `MuscleGroupFullBody MuscleGroup = "full_body"`
4. Implementar `Validate()` e `String()`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado com 7 constantes
- [ ] Validação funciona
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T18 — Criar Value Object AuditAction

**Objetivo**: Criar VO `AuditAction` com enum extenso de ações auditáveis.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/audit_action.go` - **CRIAR**

**Implementação (passos)**:

1. Criar arquivo `internal/kinetria/domain/vos/audit_action.go`
2. Definir `type AuditAction string`
3. Criar constantes (15 valores):
   - User: `AuditActionUserCreated`, `AuditActionUserUpdated`, `AuditActionUserDeleted`
   - Workout: `AuditActionWorkoutCreated`, `AuditActionWorkoutUpdated`, `AuditActionWorkoutDeleted`
   - Session: `AuditActionSessionStarted`, `AuditActionSessionCompleted`, `AuditActionSessionCancelled`
   - Set: `AuditActionSetRecorded`, `AuditActionSetUpdated`, `AuditActionSetDeleted`
   - Auth: `AuditActionLogin`, `AuditActionLogout`, `AuditActionTokenRefreshed`
4. Implementar `Validate()` e `String()`

**Critério de aceite (testes/checks)**:
- [ ] Arquivo criado com 15 constantes
- [ ] Validação funciona para todos os valores
- [ ] Código compila sem erros
- [ ] **Teste unitário criado** (ver T25)

**Dependências**: Nenhuma

---

## T19 — Criar entidades RefreshToken e AuditLog

**Objetivo**: Definir entidades `RefreshToken` e `AuditLog`.

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
   - `TokenHash string`
   - `ExpiresAt time.Time`
   - `Revoked bool`
   - `CreatedAt time.Time`

### AuditLog
1. Criar `audit_log.go`
2. Definir `type AuditLogID = uuid.UUID`
3. Struct `AuditLog`:
   - `ID AuditLogID`
   - `UserID *UserID` (pointer - nullable)
   - `Action vos.AuditAction`
   - `EntityType string`
   - `EntityID *uuid.UUID` (pointer - nullable)
   - `Metadata map[string]interface{}` (para JSONB)
   - `IPAddress string`
   - `UserAgent string`
   - `CreatedAt time.Time`

**Critério de aceite (testes/checks)**:
- [ ] Ambos os arquivos criados
- [ ] `RefreshToken` possui 6 campos
- [ ] `AuditLog` possui 9 campos
- [ ] `UserID` e `EntityID` em AuditLog são pointers (nullable)
- [ ] `Metadata` é `map[string]interface{}` (compatível com JSONB)
- [ ] Código compila sem erros

**Dependências**: T09 (UserID), T18 (AuditAction VO)

---

## T20 — Criar package constants com defaults e validações

**Objetivo**: Centralizar constantes de defaults e regras de validação.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/constants/defaults.go` - **CRIAR**
- `internal/kinetria/domain/constants/validation.go` - **CRIAR**

**Implementação (passos)**:

### defaults.go
1. Criar arquivo `defaults.go`
2. Definir constantes:
   ```go
   const (
       DefaultExerciseThumbnail = "/assets/defaults/exercise-placeholder.png"
       DefaultExerciseVideo     = ""
       DefaultDifficultyLevel   = 3
   )
   ```

### validation.go
1. Criar arquivo `validation.go`
2. Definir constantes de validação:
   ```go
   const (
       MinNameLength         = 1
       MaxNameLength         = 255
       MinDescriptionLength  = 0
       MaxDescriptionLength  = 5000
       MinDifficultyLevel    = 1
       MaxDifficultyLevel    = 5
       MinSetNumber          = 1
       MaxSetNumber          = 100
   )
   ```

**Critério de aceite (testes/checks)**:
- [ ] Ambos os arquivos criados
- [ ] `defaults.go` possui 3 constantes (Thumbnail, Video, DifficultyLevel)
- [ ] `validation.go` possui 8 constantes (Min/Max para Name, Description, Difficulty, SetNumber)
- [ ] Código compila sem erros
- [ ] Constantes são usáveis em outros pacotes (ex: `constants.DefaultDifficultyLevel`)

**Dependências**: Nenhuma

---

## T21 — Atualizar Config com variáveis de banco de dados

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

## T22 — Criar Database Pool Provider

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

**Dependências**: T21 (Config com DB vars)

---

## T23 — Criar Health Check Handler

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

**Dependências**: T21 (Config)

---

## T24 — Registrar providers e rotas no main.go

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

**Dependências**: T21 (Config), T22 (Pool), T23 (Health Handler)

---

## T25 — Criar testes unitários para Value Objects

**Objetivo**: Testar validação de todos os VOs (WorkoutStatus, SessionStatus, ExerciseCategory, MuscleGroup, AuditAction).

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/vos/workout_status_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/session_status_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/exercise_category_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/muscle_group_test.go` - **CRIAR**
- `internal/kinetria/domain/vos/audit_action_test.go` - **CRIAR**

**Implementação (passos)**:

Para cada VO, criar arquivo de teste com:

1. **Teste de valores válidos** (table-driven):
   ```go
   func TestWorkoutStatus_Validate_ValidValues(t *testing.T) {
       tests := []struct{
           name   string
           status WorkoutStatus
       }{
           {"draft", WorkoutStatusDraft},
           {"published", WorkoutStatusPublished},
           {"archived", WorkoutStatusArchived},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := tt.status.Validate()
               if err != nil {
                   t.Errorf("expected no error for %s, got %v", tt.name, err)
               }
           })
       }
   }
   ```

2. **Teste de valores inválidos**:
   ```go
   func TestWorkoutStatus_Validate_InvalidValues(t *testing.T) {
       invalid := WorkoutStatus("invalid_status")
       err := invalid.Validate()
       if err == nil {
           t.Error("expected error for invalid status, got nil")
       }
       if !errors.Is(err, errors.ErrMalformedParameters) {
           t.Errorf("expected ErrMalformedParameters, got %v", err)
       }
   }
   ```

3. **Teste de String()**:
   ```go
   func TestWorkoutStatus_String(t *testing.T) {
       status := WorkoutStatusDraft
       if status.String() != "draft" {
           t.Errorf("expected 'draft', got '%s'", status.String())
       }
   }
   ```

Repetir para todos os 5 VOs.

**Critério de aceite (testes/checks)**:
- [ ] 5 arquivos de teste criados (*_test.go)
- [ ] Cada arquivo testa valores válidos (tabela com todos os casos)
- [ ] Cada arquivo testa valores inválidos
- [ ] Cada arquivo testa método `String()`
- [ ] `make test` executa todos os testes
- [ ] Todos os testes passam (0 failures)
- [ ] Coverage dos VOs >= 80% (`go test -cover ./internal/kinetria/domain/vos`)

**Dependências**: T14-T18 (VOs implementados)

---

## T26 — Criar teste de integração para Database Pool

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

**Dependências**: T22 (Pool implementado), T01 (Docker Compose para rodar teste)

---

## T27 — Criar teste unitário para Health Handler

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

## T28 — Teste E2E: Docker Compose + Health Check

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

## T29 — Documentar setup e arquitetura no README.md

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
   - `WorkoutStatus` - draft, published, archived
   - `SessionStatus` - in_progress, completed, cancelled
   - `ExerciseCategory` - strength, cardio, flexibility, balance
   - `MuscleGroup` - chest, back, legs, shoulders, arms, core, full_body
   - `AuditAction` - Ações auditáveis (15 tipos)

   ### Constants
   - `defaults.go` - Valores padrão (thumbnails, difficulty)
   - `validation.go` - Regras de validação (min/max lengths)
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

## T30 — Documentar API (health endpoint)

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

## T31 — Adicionar comentários Godoc nas entidades e VOs

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
   // WorkoutStatus representa o status de um plano de treino.
   // Valores possíveis: draft (rascunho), published (publicado), archived (arquivado).
   type WorkoutStatus string
   
   const (
       // WorkoutStatusDraft indica que o workout está em rascunho (não visível para treino).
       WorkoutStatusDraft WorkoutStatus = "draft"
       
       // WorkoutStatusPublished indica que o workout está publicado (pronto para uso).
       WorkoutStatusPublished WorkoutStatus = "published"
       
       // WorkoutStatusArchived indica que o workout foi arquivado (não mais ativo).
       WorkoutStatusArchived WorkoutStatus = "archived"
   )
   
   // Validate verifica se o WorkoutStatus possui um valor válido.
   // Retorna ErrMalformedParameters se o valor for inválido.
   func (s WorkoutStatus) Validate() error {
       // ...
   }
   ```

3. Aplicar para todas as entidades (7) e VOs (5)

4. Gerar e verificar godoc:
   ```bash
   go doc internal/kinetria/domain/entities.User
   go doc internal/kinetria/domain/vos.WorkoutStatus
   ```

**Critério de aceite (testes/checks)**:
- [ ] Todas as 7 entidades possuem comentário Godoc no struct
- [ ] Todos os 5 VOs possuem comentário Godoc no type
- [ ] Constantes exportadas possuem comentários explicativos
- [ ] Métodos exportados (Validate, String) possuem comentários
- [ ] `go doc` exibe documentação corretamente (testar algumas entidades/VOs)
- [ ] Comentários seguem convenção Godoc (começam com nome do tipo/função)
- [ ] Código compila sem erros

**Dependências**: T09-T20 (entidades e VOs implementados)

---

## T32 — Validação final e checklist de aceite da feature

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
| T04 | Migration 003 - exercises | Migration | T01 |
| T05 | Migration 004 - sessions | Migration | T02, T03 |
| T06 | Migration 005 - set_records | Migration | T04, T05 |
| T07 | Migration 006 - refresh_tokens | Migration | T02 |
| T08 | Migration 007 - audit_log | Migration | T02 |
| T09 | Entity User | Domain | - |
| T10 | Entity Workout | Domain | T09, T14 |
| T11 | Entity Exercise | Domain | T16, T17 |
| T12 | Entity Session | Domain | T09, T10, T15 |
| T13 | Entity SetRecord | Domain | T11, T12 |
| T14 | VO WorkoutStatus | Domain | - |
| T15 | VO SessionStatus | Domain | - |
| T16 | VO ExerciseCategory | Domain | - |
| T17 | VO MuscleGroup | Domain | - |
| T18 | VO AuditAction | Domain | - |
| T19 | Entities RefreshToken + AuditLog | Domain | T09, T18 |
| T20 | Constants (defaults + validation) | Domain | - |
| T21 | Update Config (DB vars) | Gateway | - |
| T22 | Database Pool Provider | Gateway | T21 |
| T23 | Health Check Handler | Gateway | T21 |
| T24 | Register providers in main.go | Integration | T21, T22, T23 |
| T25 | Unit tests - VOs | Tests | T14-T18 |
| T26 | Integration test - Pool | Tests | T22, T01 |
| T27 | Unit test - Health Handler | Tests | T23 |
| T28 | E2E test - Docker + Health | Tests | T01-T24 |
| T29 | Update README - Docker/Migrations | Docs | T28 |
| T30 | Document /health endpoint | Docs | T23 |
| T31 | Add Godoc comments | Docs | T09-T20 |
| T32 | Final validation checklist | Validation | T01-T31 |

**Total: 32 tarefas**

---

## Ordem Sugerida de Implementação

### Fase 1: Infraestrutura Base (paralelo possível)
- T01: Docker Compose
- T21: Config (DB vars)
- T20: Constants

### Fase 2: Migrations (sequencial devido a dependências FK)
- T02: Users table
- T03: Workouts table
- T04: Exercises table
- T05: Sessions table
- T06: Set records table
- T07: Refresh tokens table
- T08: Audit log table

### Fase 3: Domain Layer (paralelo possível)
- T14-T18: Value Objects (todos em paralelo)
- T09: User entity
- T10: Workout entity
- T11: Exercise entity
- T12: Session entity
- T13: SetRecord entity
- T19: RefreshToken + AuditLog entities

### Fase 4: Gateway Layer
- T22: Database Pool Provider
- T23: Health Check Handler
- T24: Main.go integration

### Fase 5: Testes
- T25: Unit tests (VOs)
- T27: Unit test (Health)
- T26: Integration test (Pool)
- T28: E2E test

### Fase 6: Documentação
- T29: README
- T30: API docs
- T31: Godoc comments
- T32: Final validation

---

**Estimativa total de esforço**: ~16-20 horas (considerando desenvolvimento + testes + documentação)

**Feature pronta quando**: T32 (validação final) estiver completa com todos os checks passando.
