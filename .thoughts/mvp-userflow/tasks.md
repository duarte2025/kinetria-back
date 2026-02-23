# Tasks — mvp-userflow

> **Pré-requisito**: A feature `foundation-infrastructure` deve ser concluída antes de iniciar este backlog (migrations, Docker Compose, módulos `pkg/`, conexão DB).

---

## T01 — Criar entidades de domínio e VOs do MVP

**Objetivo**: Definir as structs de domínio (`User`, `Workout`, `Exercise`, `Session`, `SetRecord`, `RefreshToken`, `AuditLog`) e seus Value Objects.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/entities/entities.go` — **ATUALIZAR** (substituir template)
- `internal/kinetria/domain/vos/vos.go` — **ATUALIZAR** (adicionar VOs reais)

**Implementação (passos)**:

1. Em `entities.go`, definir as structs:
   - `User` (ID, Name, Email, PasswordHash, ProfileImageURL, CreatedAt, UpdatedAt)
   - `Workout` (ID, UserID, Name, Description, Type, Intensity, Duration, ImageURL, CreatedAt, UpdatedAt)
   - `Exercise` (ID, WorkoutID, Name, ThumbnailURL, Sets, Reps, Muscles, RestTime, Weight, OrderIndex)
   - `Session` (ID, UserID, WorkoutID, StartedAt, FinishedAt `*time.Time`, Status, Notes, CreatedAt, UpdatedAt)
   - `SetRecord` (ID, SessionID, ExerciseID, SetNumber, Weight, Reps, Status, RecordedAt)
   - `RefreshToken` (ID, UserID, Token, ExpiresAt, RevokedAt `*time.Time`, CreatedAt)
   - `AuditLog` (ID, UserID, EntityType, EntityID, Action, ActionData `json.RawMessage`, OccurredAt, IPAddress, UserAgent)
   - Type aliases: `UserID = uuid.UUID`, `WorkoutID = uuid.UUID`, `SessionID = uuid.UUID`, `ExerciseID = uuid.UUID`
2. Em `vos.go`, definir:
   - `SessionStatus` string enum: `SessionStatusActive`, `SessionStatusCompleted`, `SessionStatusAbandoned`
   - `SetRecordStatus` string enum: `SetRecordStatusCompleted`, `SetRecordStatusSkipped`
   - `WorkoutType` string enum: `WorkoutTypeForca`, `WorkoutTypeHipertrofia`, `WorkoutTypeMobilidade`, `WorkoutTypeCondicionamento`
   - `WorkoutIntensity` string enum: `WorkoutIntensityBaixa`, `WorkoutIntensityModerada`, `WorkoutIntensityAlta`
   - `AuditAction` string enum: `AuditActionCreated`, `AuditActionUpdated`, `AuditActionCompleted`, `AuditActionAbandoned`
   - Implementar `IsValid()` e `String()` em todos os VOs
3. Definir constantes de assets default: `DefaultProfileImageURL`, `DefaultExerciseThumbnailURL`

**Critério de aceite**:
- [ ] Structs compilam sem erros (`go build ./...`)
- [ ] Todos os VOs possuem `IsValid()` retornando false para valores inválidos
- [ ] `go vet ./...` sem warnings

---

## T02 — Criar erros de domínio do MVP

**Objetivo**: Definir erros sentinela específicos de cada domínio (auth, workouts, sessions).

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/errors/errors.go` — **ATUALIZAR**

**Implementação (passos)**:

1. Adicionar erros específicos ao arquivo existente:
   ```go
   // Auth
   var ErrEmailAlreadyExists  = errors.New("email already exists")
   var ErrInvalidCredentials  = errors.New("invalid credentials")
   var ErrTokenExpired        = errors.New("token expired or revoked")

   // Workouts
   var ErrWorkoutNotFound     = errors.New("workout not found")

   // Sessions
   var ErrSessionNotFound         = errors.New("session not found")
   var ErrSessionAlreadyActive    = errors.New("user already has an active session")
   var ErrSessionAlreadyClosed    = errors.New("session is already closed")
   var ErrSetAlreadyRecorded      = errors.New("set already recorded for this exercise")
   var ErrExerciseNotInWorkout    = errors.New("exercise does not belong to the session workout")
   ```

**Critério de aceite**:
- [ ] `go build ./...` sem erros
- [ ] Erros genéricos (`ErrNotFound`, `ErrConflict`) mantidos para compatibilidade

---

## T03 — Criar ports (interfaces) de repositórios e serviços

**Objetivo**: Definir contratos de repositório e do serviço de autenticação JWT.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/ports/repositories.go` — **ATUALIZAR**
- `internal/kinetria/domain/ports/services.go` — **ATUALIZAR**

**Implementação (passos)**:

1. Em `repositories.go`, definir interfaces com `//go:generate moq`:
   - `UserRepository`: `FindByID`, `FindByEmail`, `Insert`, `Update`
   - `WorkoutRepository`: `FindByID`, `FindByUserID` (paginado), `FindByUserIDAndID`
   - `ExerciseRepository`: `FindByWorkoutID`
   - `SessionRepository`: `FindByID`, `FindByUserIDAndID`, `FindActiveByUserID`, `Insert`, `Update`
   - `SetRecordRepository`: `FindBySessionID`, `FindBySessionAndExerciseAndSet`, `Insert`
   - `RefreshTokenRepository`: `FindByToken`, `Insert`, `Revoke`, `RevokeAllByUserID`
   - `AuditLogRepository`: `Append`
2. Em `services.go`, definir:
   - `TokenService`: `GenerateAccessToken(userID) (string, error)`, `ValidateAccessToken(token) (userID uuid.UUID, error)`, `GenerateRefreshToken() (plain string, hash string, error)`, `HashRefreshToken(plain string) string`
   - `PasswordService`: `HashPassword(plain string) (string, error)`, `CheckPassword(plain, hash string) bool`

**Critério de aceite**:
- [ ] Todas as interfaces possuem `//go:generate moq`
- [ ] `go build ./...` sem erros
- [ ] `make mocks` gera arquivos em `internal/kinetria/domain/ports/mocks/`

---

## T04 — Criar migrations SQL completas

**Objetivo**: Criar 7 migrations SQL para todas as tabelas do MVP.

**Arquivos/pacotes prováveis**:
- `migrations/001_create_users.sql`
- `migrations/002_create_workouts.sql`
- `migrations/003_create_exercises.sql`
- `migrations/004_create_sessions.sql`
- `migrations/005_create_set_records.sql`
- `migrations/006_create_refresh_tokens.sql`
- `migrations/007_create_audit_log.sql`

**Implementação (passos)**:

1. `001_create_users.sql`:
   ```sql
   CREATE TABLE users (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     email VARCHAR(255) NOT NULL,
     name VARCHAR(255) NOT NULL,
     password_hash VARCHAR(255) NOT NULL,
     profile_image_url VARCHAR(500) NOT NULL DEFAULT '/assets/avatars/default.png',
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     CONSTRAINT users_email_unique UNIQUE (email)
   );
   CREATE INDEX idx_users_email ON users(email);
   ```

2. `002_create_workouts.sql`:
   ```sql
   CREATE TABLE workouts (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     name VARCHAR(255) NOT NULL,
     description VARCHAR(500) NOT NULL DEFAULT '',
     type VARCHAR(50) NOT NULL,
     intensity VARCHAR(50) NOT NULL,
     duration INT NOT NULL DEFAULT 0,
     image_url VARCHAR(500) NOT NULL DEFAULT '',
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   CREATE INDEX idx_workouts_user_id ON workouts(user_id);
   CREATE INDEX idx_workouts_user_type ON workouts(user_id, type);
   ```

3. `003_create_exercises.sql`:
   ```sql
   CREATE TABLE exercises (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
     name VARCHAR(255) NOT NULL,
     thumbnail_url VARCHAR(500) NOT NULL DEFAULT '/assets/exercises/generic.png',
     sets INT NOT NULL DEFAULT 1,
     reps VARCHAR(20) NOT NULL DEFAULT '10',
     muscles JSONB NOT NULL DEFAULT '[]',
     rest_time INT NOT NULL DEFAULT 60,
     weight FLOAT8 NOT NULL DEFAULT 0,
     order_index INT NOT NULL DEFAULT 0
   );
   CREATE INDEX idx_exercises_workout_id ON exercises(workout_id);
   CREATE INDEX idx_exercises_workout_order ON exercises(workout_id, order_index);
   ```

4. `004_create_sessions.sql`:
   ```sql
   CREATE TABLE sessions (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     workout_id UUID NOT NULL REFERENCES workouts(id),
     started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     finished_at TIMESTAMPTZ,
     status VARCHAR(20) NOT NULL DEFAULT 'active',
     notes VARCHAR(1000) NOT NULL DEFAULT '',
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     CONSTRAINT sessions_status_check CHECK (status IN ('active','completed','abandoned'))
   );
   CREATE INDEX idx_sessions_user_id ON sessions(user_id);
   CREATE INDEX idx_sessions_workout_id ON sessions(workout_id);
   CREATE INDEX idx_sessions_user_status ON sessions(user_id, status);
   CREATE UNIQUE INDEX idx_sessions_user_active ON sessions(user_id) WHERE status = 'active';
   ```

5. `005_create_set_records.sql`:
   ```sql
   CREATE TABLE set_records (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
     exercise_id UUID NOT NULL REFERENCES exercises(id),
     set_number INT NOT NULL,
     weight FLOAT8 NOT NULL DEFAULT 0,
     reps INT NOT NULL DEFAULT 0,
     status VARCHAR(20) NOT NULL DEFAULT 'completed',
     recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     CONSTRAINT set_records_status_check CHECK (status IN ('completed','skipped')),
     CONSTRAINT set_records_unique_set UNIQUE (session_id, exercise_id, set_number)
   );
   CREATE INDEX idx_set_records_session_id ON set_records(session_id);
   CREATE INDEX idx_set_records_session_exercise ON set_records(session_id, exercise_id);
   ```

6. `006_create_refresh_tokens.sql`:
   ```sql
   CREATE TABLE refresh_tokens (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     token VARCHAR(64) NOT NULL,
     expires_at TIMESTAMPTZ NOT NULL,
     revoked_at TIMESTAMPTZ,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     CONSTRAINT refresh_tokens_token_unique UNIQUE (token)
   );
   CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
   CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
   ```

7. `007_create_audit_log.sql`:
   ```sql
   CREATE TABLE audit_log (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID NOT NULL,
     entity_type VARCHAR(50) NOT NULL,
     entity_id UUID NOT NULL,
     action VARCHAR(50) NOT NULL,
     action_data JSONB,
     occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     ip_address VARCHAR(45) NOT NULL DEFAULT '',
     user_agent VARCHAR(500) NOT NULL DEFAULT ''
   );
   CREATE INDEX idx_audit_log_user_occurred ON audit_log(user_id, occurred_at);
   CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
   ```

**Critério de aceite**:
- [ ] Todas as migrations aplicam sem erro no PostgreSQL via Docker Compose
- [ ] Índices críticos existem: `idx_sessions_user_active` (partial unique), `idx_set_records_unique_set` (unique), `idx_audit_log_user_occurred`
- [ ] Constraints CHECK funcionam: inserir status inválido gera erro
- [ ] FKs funcionam: deletar usuário em cascata remove workouts, sessions, set_records

---

## T05 — Implementar serviços de infraestrutura (JWT + bcrypt)

**Objetivo**: Implementar `TokenService` e `PasswordService` usados pelo use case de auth.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/auth/token_service.go` — **CRIAR**
- `internal/kinetria/gateways/auth/password_service.go` — **CRIAR**
- `internal/kinetria/gateways/config/config.go` — **ATUALIZAR** (adicionar `JWTSecret`, `JWTTTLSeconds`, `RefreshTokenTTLDays`)

**Implementação (passos)**:

1. Em `token_service.go`, implementar `TokenService` com `golang-jwt/jwt/v5`:
   - `GenerateAccessToken(userID uuid.UUID)` → JWT HS256, claims: `sub`, `exp`, `iat`; TTL 1h
   - `ValidateAccessToken(token string)` → parse JWT, validar exp; retornar `userID` ou erro
   - `GenerateRefreshToken()` → 32 bytes aleatórios via `crypto/rand`; retornar (plain, hash SHA-256)
   - `HashRefreshToken(plain string)` → SHA-256 hex string
2. Em `password_service.go`, implementar `PasswordService`:
   - `HashPassword(plain string)` → bcrypt cost 12
   - `CheckPassword(plain, hash string)` → `bcrypt.CompareHashAndPassword`
3. Em `config.go`, adicionar:
   ```go
   JWTSecret             string        `envconfig:"JWT_SECRET" required:"true"`
   JWTTTLSeconds         int           `envconfig:"JWT_TTL_SECONDS" default:"3600"`
   RefreshTokenTTLDays   int           `envconfig:"REFRESH_TOKEN_TTL_DAYS" default:"30"`
   ```

**Critério de aceite**:
- [ ] `go test ./internal/kinetria/gateways/auth/...` passando
- [ ] Testes unitários: geração e validação de token; expiração; hash de senha
- [ ] Token gerado pode ser inspecionado em jwt.io com a secret correta
- [ ] `crypto/rand` usado (não `math/rand`)

---

## T06 — Implementar queries SQLC e repositórios

**Objetivo**: Escrever queries SQL (SQLC), gerar código type-safe e implementar os repositórios.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/gateways/repositories/queries/queries.sql` — **ATUALIZAR**
- `internal/kinetria/gateways/repositories/queries/` — código gerado pelo SQLC
- `internal/kinetria/gateways/repositories/repository.go` — **ATUALIZAR**
- `internal/kinetria/gateways/repositories/repo_users.go` — **CRIAR**
- `internal/kinetria/gateways/repositories/repo_workouts.go` — **CRIAR**
- `internal/kinetria/gateways/repositories/repo_exercises.go` — **CRIAR**
- `internal/kinetria/gateways/repositories/repo_sessions.go` — **CRIAR**
- `internal/kinetria/gateways/repositories/repo_set_records.go` — **CRIAR**
- `internal/kinetria/gateways/repositories/repo_refresh_tokens.go` — **CRIAR**
- `internal/kinetria/gateways/repositories/repo_audit_log.go` — **CRIAR**

**Implementação (passos)**:

1. Escrever queries SQLC em `queries.sql` para cada entidade:
   - `users`: GetUserByID, GetUserByEmail, InsertUser
   - `workouts`: GetWorkoutByID, GetWorkoutByUserIDAndID, ListWorkoutsByUserID (com LIMIT/OFFSET), CountWorkoutsByUserID
   - `exercises`: ListExercisesByWorkoutID
   - `sessions`: GetSessionByID, GetSessionByUserIDAndID, GetActiveSessionByUserID, InsertSession, UpdateSession
   - `set_records`: GetSetRecord (by session+exercise+set), ListSetRecordsBySession, InsertSetRecord
   - `refresh_tokens`: GetRefreshTokenByToken, InsertRefreshToken, RevokeRefreshToken, RevokeAllRefreshTokensByUser
   - `audit_log`: InsertAuditLog
2. Executar `make sqlc` para gerar código
3. Implementar cada arquivo de repositório mapeando `queries.Model` → `entities.X` e convertendo `pgx.ErrNoRows` → `domerr.ErrXXXNotFound`
4. Mapear erro de violação de constraint única do PostgreSQL (código `23505`) para erros de domínio específicos

**Critério de aceite**:
- [ ] `make sqlc` executa sem erros
- [ ] `go build ./...` sem erros
- [ ] Repositórios implementam exatamente as interfaces dos ports
- [ ] Erros de constraint do Postgres são convertidos para `ErrEmailAlreadyExists`, `ErrSetAlreadyRecorded`, `ErrSessionAlreadyActive`

---

## T07 — Implementar feature AUTH (use cases + handler + router)

**Objetivo**: Implementar register, login, refresh e logout com testes.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/auth/uc_register.go` — **CRIAR**
- `internal/kinetria/domain/auth/uc_login.go` — **CRIAR**
- `internal/kinetria/domain/auth/uc_refresh_token.go` — **CRIAR**
- `internal/kinetria/domain/auth/uc_logout.go` — **CRIAR**
- `internal/kinetria/domain/auth/uc_register_test.go` — **CRIAR**
- `internal/kinetria/domain/auth/uc_login_test.go` — **CRIAR**
- `internal/kinetria/domain/auth/uc_refresh_token_test.go` — **CRIAR**
- `internal/kinetria/gateways/http/handler_auth.go` — **CRIAR**
- `internal/kinetria/gateways/http/router_auth.go` — **CRIAR**

**Implementação (passos)**:

1. `RegisterUC.Execute(ctx, RegisterInput{Name, Email, Password})`:
   - Verificar email único (→ `ErrEmailAlreadyExists` se duplicado)
   - Hash de senha via `PasswordService`
   - Inserir usuário
   - Gerar access + refresh tokens; persistir hash do refresh
   - Retornar `RegisterOutput{AccessToken, RefreshToken, ExpiresIn}`

2. `LoginUC.Execute(ctx, LoginInput{Email, Password})`:
   - Buscar usuário por email (→ `ErrInvalidCredentials` se não encontrado)
   - Verificar senha (→ `ErrInvalidCredentials` se inválida)
   - Gerar access + refresh tokens; persistir hash do refresh
   - Retornar `LoginOutput{AccessToken, RefreshToken, ExpiresIn}`

3. `RefreshTokenUC.Execute(ctx, RefreshInput{RefreshToken})`:
   - Hash do token recebido; buscar no banco
   - Verificar expiração e revogação (→ `ErrTokenExpired`)
   - Revogar token atual; gerar novo par de tokens; persistir novo hash
   - Retornar `RefreshOutput{AccessToken, RefreshToken, ExpiresIn}`

4. `LogoutUC.Execute(ctx, LogoutInput{UserID, RefreshToken})`:
   - Hash do token; revogar no banco
   - Retornar `LogoutOutput{}`

5. Handler HTTP: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
   - Decodificar JSON, validar struct com `validator/v10`
   - Mapear erros de domínio para HTTP (409, 401, 422)
   - Retornar wrapper `{ "data": {...} }`

6. Middleware JWT:
   - Extrair `Authorization: Bearer <token>` do header
   - Validar via `TokenService.ValidateAccessToken`
   - Injetar `userID` no contexto via key tipada
   - Retornar 401 se inválido/expirado

**Critério de aceite**:
- [ ] Testes unitários (table-driven) para todos os use cases cobrindo happy + sad paths
- [ ] `go test ./internal/kinetria/domain/auth/...` passando
- [ ] Handler retorna status HTTP corretos por cenário
- [ ] Middleware JWT bloqueia requests sem token válido com 401
- [ ] Senhas nunca são logadas

---

## T08 — Implementar feature WORKOUTS (use cases + handler + router)

**Objetivo**: Implementar listagem paginada e detalhe de workout com exercises.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/workouts/uc_list_workouts.go` — **CRIAR**
- `internal/kinetria/domain/workouts/uc_get_workout.go` — **CRIAR**
- `internal/kinetria/domain/workouts/uc_list_workouts_test.go` — **CRIAR**
- `internal/kinetria/domain/workouts/uc_get_workout_test.go` — **CRIAR**
- `internal/kinetria/gateways/http/handler_workouts.go` — **CRIAR**
- `internal/kinetria/gateways/http/router_workouts.go` — **CRIAR**

**Implementação (passos)**:

1. `ListWorkoutsUC.Execute(ctx, ListWorkoutsInput{UserID, Page, PageSize})`:
   - Buscar workouts paginados por `user_id`
   - Retornar `ListWorkoutsOutput{Workouts []Workout, Total int}`

2. `GetWorkoutUC.Execute(ctx, GetWorkoutInput{UserID, WorkoutID})`:
   - Buscar workout por `user_id AND id` (→ `ErrWorkoutNotFound` se não encontrado)
   - Buscar exercises do workout
   - Retornar `GetWorkoutOutput{Workout}` com exercises populados

3. Handler HTTP:
   - `GET /workouts?page=1&pageSize=20` → resposta com `data[]` e `meta` de paginação
   - `GET /workouts/{workoutId}` → 404 se não encontrado
   - Extrair `userID` do contexto (injetado pelo middleware JWT)

**Critério de aceite**:
- [ ] Testes unitários (table-driven) para ambos os use cases
- [ ] `go test ./internal/kinetria/domain/workouts/...` passando
- [ ] Usuário A não consegue ver workouts de usuário B (isolamento por userID)
- [ ] Paginação funciona corretamente (meta.totalPages calculado)

---

## T09 — Implementar feature SESSIONS (use cases + handler + router)

**Objetivo**: Implementar os 4 use cases de sessão (start, record set, finish, abandon) com audit log.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/sessions/uc_start_session.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_record_set.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_finish_session.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_abandon_session.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_start_session_test.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_record_set_test.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_finish_session_test.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_abandon_session_test.go` — **CRIAR**
- `internal/kinetria/gateways/http/handler_sessions.go` — **CRIAR**
- `internal/kinetria/gateways/http/router_sessions.go` — **CRIAR**

**Implementação (passos)**:

1. `StartSessionUC.Execute(ctx, StartSessionInput{UserID, WorkoutID})`:
   - Verificar se workout pertence ao usuário (→ `ErrWorkoutNotFound`)
   - Verificar sessão ativa existente (→ `ErrSessionAlreadyActive`)
   - Criar sessão com status `active`
   - Registrar audit log: `entity_type="session"`, `action="created"`
   - Retornar `StartSessionOutput{Session}`

2. `RecordSetUC.Execute(ctx, RecordSetInput{UserID, SessionID, ExerciseID, SetNumber, Weight, Reps, Status})`:
   - Verificar sessão pertence ao usuário e está `active` (→ `ErrSessionNotFound` ou `ErrSessionAlreadyClosed`)
   - Verificar exercício pertence ao workout da sessão (→ `ErrExerciseNotInWorkout`)
   - Verificar set não foi registrado (→ `ErrSetAlreadyRecorded` se duplicado)
   - Inserir SetRecord
   - Registrar audit log: `entity_type="set_record"`, `action="created"`
   - Retornar `RecordSetOutput{SetRecord}`

3. `FinishSessionUC.Execute(ctx, FinishSessionInput{UserID, SessionID, Notes})`:
   - Verificar sessão pertence ao usuário (→ `ErrSessionNotFound`)
   - Verificar status `active` (→ `ErrSessionAlreadyClosed` se completed/abandoned)
   - Atualizar status para `completed`, `finished_at=now()`, `notes`
   - Registrar audit log: `action="completed"`
   - Retornar `FinishSessionOutput{Session}`

4. `AbandonSessionUC.Execute(ctx, AbandonSessionInput{UserID, SessionID})`:
   - Verificar sessão pertence ao usuário e está `active`
   - Atualizar status para `abandoned`, `finished_at=now()`
   - Registrar audit log: `action="abandoned"`
   - Retornar `AbandonSessionOutput{Session}`

5. Handler HTTP:
   - `POST /sessions` → 201 ou 409 (sessão ativa já existe)
   - `POST /sessions/{id}/sets` → 201, 404 ou 409
   - `PATCH /sessions/{id}/finish` → 200, 404 ou 409
   - `PATCH /sessions/{id}/abandon` → 200, 404 ou 409

**Critério de aceite**:
- [ ] Testes unitários (table-driven) para todos os 4 use cases
- [ ] `go test ./internal/kinetria/domain/sessions/...` passando
- [ ] Audit log é registrado em todas as mutações
- [ ] Constraint `UNIQUE(session_id, exercise_id, set_number)` retorna 409 ao invés de 500
- [ ] Sessão de outro usuário retorna 404 (não revela existência)

---

## T10 — Implementar feature DASHBOARD (use cases + handler + router)

**Objetivo**: Implementar endpoint de agregação de dados com chamadas paralelas.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/domain/users/uc_get_user.go` — **CRIAR**
- `internal/kinetria/domain/users/uc_get_week_stats.go` — **CRIAR**
- `internal/kinetria/domain/sessions/uc_get_week_progress.go` — **CRIAR**
- `internal/kinetria/gateways/http/handler_dashboard.go` — **CRIAR**
- `internal/kinetria/gateways/http/router_dashboard.go` — **CRIAR**

**Implementação (passos)**:

1. `GetUserUC.Execute(ctx, GetUserInput{UserID})`:
   - Retornar `GetUserOutput{User}` com id, name, email, profileImageUrl

2. `GetWeekStatsUC.Execute(ctx, GetWeekStatsInput{UserID, WeekStart, WeekEnd})`:
   - Calcular calorias queimadas (soma de duração × MET estimado ou valor fixo por enquanto: `duration_minutes × 5` cal)
   - Calcular tempo total em minutos de sessões `completed` na semana
   - Retornar `GetWeekStatsOutput{CaloriesBurned int, TotalTimeMinutes int}`

3. `GetWeekProgressUC.Execute(ctx, GetWeekProgressInput{UserID, WeekStart})`:
   - Para cada dia dos últimos 7 dias: verificar se há sessão `completed` (→ "completed"), `active/abandoned` sem `completed` (→ "missed"), ou data futura (→ "future")
   - Retornar `GetWeekProgressOutput{Days []DayProgress}`

4. Handler `GET /dashboard`:
   - Executar em paralelo via goroutines: `GetUserUC`, `GetWeekStatsUC`, `GetWeekProgressUC`, `GetTodayWorkoutUC` (buscar próximo workout não-concluído)
   - Aguardar resultados; retornar 500 se algum falhar
   - Montar `HomeData{User, TodayWorkout, WeekProgress, Stats}` e retornar 200

**Critério de aceite**:
- [ ] `go test ./internal/kinetria/domain/users/...` e `sessions/...` passando
- [ ] Chamadas paralelas com timeout de contexto funcionam
- [ ] `todayWorkout` é null quando não há workout agendado (sem panic)
- [ ] Dados de outro usuário não aparecem no dashboard

---

## T11 — Registrar todos os módulos no main.go com Fx

**Objetivo**: Conectar todos os providers, use cases e handlers no main.go via Fx.

**Arquivos/pacotes prováveis**:
- `cmd/kinetria/api/main.go` — **ATUALIZAR**
- `internal/kinetria/gateways/config/config.go` — **ATUALIZAR** (garantir todos os campos)

**Implementação (passos)**:

1. Adicionar `fx.Provide` na ordem:
   ```
   config.ParseConfigFromEnv →
   pgx pool (NewPgxPool) →
   repositories (fx.Annotate com fx.As) →
   gateways/auth.NewTokenService, NewPasswordService →
   use cases auth →
   use cases workouts →
   use cases sessions →
   use cases users (para dashboard) →
   handlers (auth, workouts, sessions, dashboard) →
   routers (xhttp.AsRouter para cada handler)
   ```
2. Adicionar middleware JWT no router principal
3. Configurar `NewPgxPool` em `gateways/config` com connection string do env

**Critério de aceite**:
- [ ] `go build ./...` sem erros
- [ ] `make run` sobe sem panic
- [ ] `GET /health` retorna 200 com DB conectado
- [ ] `POST /auth/register` funciona end-to-end

---

## T12 — Implementar testes de integração

**Objetivo**: Criar testes de integração cobrindo os cenários BDD com banco de dados real (testcontainers ou banco de teste).

**Arquivos/pacotes prováveis**:
- `internal/kinetria/tests/integration_auth_test.go` — **CRIAR**
- `internal/kinetria/tests/integration_sessions_test.go` — **CRIAR**
- `internal/kinetria/tests/integration_dashboard_test.go` — **CRIAR**
- `internal/kinetria/tests/helpers.go` — **CRIAR** (setup/teardown do banco)

**Implementação (passos)**:

1. Configurar `testcontainers-go` (ou `pgx` com banco de teste) para subir PostgreSQL em memória
2. Aplicar migrations via `golang-migrate` ou script SQL direto
3. Implementar helpers: `createTestUser()`, `createTestWorkout()`, `createTestSession()`
4. Cenários de integração obrigatórios:
   - Fluxo completo auth: register → login → refresh → logout
   - Fluxo completo treino: start session → record sets → finish session
   - Isolation: usuário A não acessa dados do usuário B
   - Idempotência: registrar mesmo set duas vezes retorna 409

**Critério de aceite**:
- [ ] `go test ./internal/kinetria/tests/...` passando (requer Docker)
- [ ] Cobertura dos cenários BDD happy paths
- [ ] Banco é limpo entre testes (transações ou truncate)
- [ ] `make test-integration` executa os testes de integração separadamente

---

## T13 — Documentar API

**Objetivo**: Documentar os endpoints, payloads e exemplos de forma acessível.

**Arquivos/pacotes prováveis**:
- `internal/kinetria/docs/api.md` — **CRIAR** (documentação de referência)
- `.thoughts/mvp-userflow/api-contract.yaml` — **MANTER** (já existente como referência OpenAPI)
- `README.md` — **ATUALIZAR** (adicionar seção de endpoints e como rodar)

**Implementação (passos)**:

1. Criar `internal/kinetria/docs/api.md` com:
   - Visão geral dos endpoints agrupados por tag (auth, dashboard, workouts, sessions)
   - Para cada endpoint: método, path, auth, request body, response 2xx, response de erro
   - Exemplos de curl para os fluxos principais
2. Adicionar Godoc nas funções/tipos exportados dos use cases e ports principais
3. Atualizar `README.md` com instruções de:
   - Como rodar localmente (`docker-compose up`, `make run`)
   - Como rodar testes (`make test`, `make test-integration`)
   - Variáveis de ambiente obrigatórias

**Critério de aceite**:
- [ ] `internal/kinetria/docs/api.md` cobre todos os 11 endpoints
- [ ] README atualizado com fluxo de setup local
- [ ] Ports e use cases exportados possuem Godoc

---

## Ordem de execução recomendada

```
T01 → T02 → T03  (domínio base: entidades, erros, ports)
T04              (migrations SQL)
T05              (serviços JWT + bcrypt)
T06              (SQLC + repositórios)
T07              (AUTH — primeiro fluxo e2e)
T08              (WORKOUTS)
T09              (SESSIONS — mais complexo)
T10              (DASHBOARD)
T11              (main.go: wiring final)
T12              (testes de integração)
T13              (documentação)
```

**Dependências críticas**:
- `foundation-infrastructure` deve estar concluída antes de T04 e T06
- T03 (ports) deve estar concluída antes de T05, T06, T07, T08, T09, T10
- T07 deve estar concluída antes de T08 e T09 (auth é pré-requisito para todas as rotas protegidas)
