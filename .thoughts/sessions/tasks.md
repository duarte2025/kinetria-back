# Tasks — mvp-userflow (Start Workout Session)

> **Feature**: Iniciar Sessão de Treino  
> **Escopo**: Apenas `POST /api/v1/sessions` (StartSession)  
> **Dependências**: foundation-infrastructure (migrations, auth JWT, entidades básicas)

---

## 📋 Ordem de Implementação

**Recomendação**: seguir ordem sequencial (T01 → T02 → ... → T08) para minimizar bloqueios.

---

## T01 — Criar/atualizar entidades de domínio (Session, Workout, User)

### Objetivo
Definir as entidades de domínio necessárias para a feature StartSession.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/entities/entities.go`

### Implementação (passos)

1. **Criar entidade `Session`**:
   ```go
   // SessionID type alias for UUID
   type SessionID = uuid.UUID

   // Session representa uma sessão de treino ativa, finalizada ou abandonada
   type Session struct {
       ID         SessionID
       UserID     UserID
       WorkoutID  WorkoutID
       StartedAt  time.Time
       FinishedAt *time.Time  // nullable (null = sessão não finalizada)
       Status     vos.SessionStatus
       Notes      string      // max 1000 chars
       CreatedAt  time.Time
       UpdatedAt  time.Time
   }
   ```

2. **Criar entidade `Workout` (mínima)**:
   ```go
   type WorkoutID = uuid.UUID

   type Workout struct {
       ID          WorkoutID
       UserID      UserID
       Name        string
       Description string
       Type        string  // enum: FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO
       Intensity   string  // enum: BAIXA, MODERADA, ALTA
       Duration    int     // minutos (estimado)
       ImageURL    string
       CreatedAt   time.Time
       UpdatedAt   time.Time
   }
   ```

3. **Criar entidade `User` (mínima)**:
   ```go
   type UserID = uuid.UUID

   type User struct {
       ID              UserID
       Name            string
       Email           string
       PasswordHash    string
       ProfileImageURL string
       CreatedAt       time.Time
       UpdatedAt       time.Time
   }
   ```

4. **Criar entidade `AuditLog`**:
   ```go
   type AuditLogID = uuid.UUID

   type AuditLog struct {
       ID         AuditLogID
       UserID     UserID
       EntityType string          // "session", "workout", "set_record"
       EntityID   uuid.UUID       // ID da entidade afetada
       Action     string          // "created", "updated", "deleted", "completed"
       ActionData json.RawMessage // dados da ação em JSON
       OccurredAt time.Time
       IPAddress  string
       UserAgent  string
   }
   ```

### Critério de aceite (testes/checks)
- [ ] Entidades compilam sem erro
- [ ] Tipos UUID estão corretamente aliased
- [ ] Campos nullable usam ponteiros (`*time.Time`)
- [ ] Comentários Godoc em todas as entidades exportadas
- [ ] `make lint` passa sem warnings

---

## T02 — Criar Value Objects (SessionStatus)

### Objetivo
Definir enums type-safe para Status de sessão.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/vos/vos.go`

### Implementação (passos)

1. **Criar enum `SessionStatus`**:
   ```go
   package vos

   // SessionStatus representa o estado de uma sessão de treino
   type SessionStatus string

   const (
       SessionStatusActive    SessionStatus = "active"
       SessionStatusCompleted SessionStatus = "completed"
       SessionStatusAbandoned SessionStatus = "abandoned"
   )

   // String retorna representação string do status
   func (s SessionStatus) String() string {
       return string(s)
   }

   // IsValid valida se o status é um valor permitido
   func (s SessionStatus) IsValid() bool {
       switch s {
       case SessionStatusActive, SessionStatusCompleted, SessionStatusAbandoned:
           return true
       }
       return false
   }
   ```

2. **(Opcional) Criar enum `SetStatus`** (se necessário para futuras features):
   ```go
   type SetStatus string

   const (
       SetStatusCompleted SetStatus = "completed"
       SetStatusSkipped   SetStatus = "skipped"
   )

   func (s SetStatus) String() string {
       return string(s)
   }

   func (s SetStatus) IsValid() bool {
       switch s {
       case SetStatusCompleted, SetStatusSkipped:
           return true
       }
       return false
   }
   ```

### Critério de aceite (testes/checks)
- [ ] SessionStatus compila sem erro
- [ ] Método `IsValid()` retorna true apenas para valores válidos
- [ ] Testes unitários cobrem validação (`IsValid()`)
- [ ] Comentários Godoc em tipos exportados
- [ ] `make lint` passa sem warnings

**Exemplo de teste**:
```go
func TestSessionStatus_IsValid(t *testing.T) {
    tests := []struct {
        name   string
        status SessionStatus
        want   bool
    }{
        {"active válido", SessionStatusActive, true},
        {"completed válido", SessionStatusCompleted, true},
        {"abandoned válido", SessionStatusAbandoned, true},
        {"inválido", SessionStatus("invalid"), false},
        {"vazio", SessionStatus(""), false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.status.IsValid(); got != tt.want {
                t.Errorf("IsValid() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## T03 — Criar erros de domínio customizados

### Objetivo
Adicionar erros específicos para a feature StartSession (ex: `ErrActiveSessionExists`).

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/errors/errors.go`

### Implementação (passos)

1. **Adicionar novos erros**:
   ```go
   package errors

   import "errors"

   var (
       // Erros existentes
       ErrNotFound            = errors.New("not found")
       ErrConflict            = errors.New("data conflict")
       ErrMalformedParameters = errors.New("malformed parameters")
       ErrFailedDependency    = errors.New("failed dependency")

       // Novos erros para StartSession
       ErrActiveSessionExists = errors.New("user already has an active session")
       ErrWorkoutNotFound     = errors.New("workout not found")
   )
   ```

2. **(Opcional) Criar função helper para wrapping**:
   ```go
   import "fmt"

   // WrapNotFound retorna erro de not found com contexto
   func WrapNotFound(entity string, id interface{}) error {
       return fmt.Errorf("%w: %s with id '%v'", ErrNotFound, entity, id)
   }

   // Exemplo de uso:
   // return errors.WrapNotFound("workout", workoutID)
   ```

### Critério de aceite (testes/checks)
- [ ] Novos erros compilam sem erro
- [ ] Erros podem ser comparados com `errors.Is()`
- [ ] (Se wrapper criado) Testes cobrem unwrap de erros
- [ ] `make lint` passa sem warnings

---

## T04 — Criar interfaces de repositório (ports)

### Objetivo
Definir contratos (ports) para os repositórios de Session, Workout e AuditLog.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/ports/repositories.go`

### Implementação (passos)

1. **Criar interface `SessionRepository`**:
   ```go
   package ports

   import (
       "context"
       "github.com/google/uuid"
       "internal/kinetria/domain/entities"
   )

   //go:generate moq -stub -pkg mocks -out mocks/repositories.go . SessionRepository WorkoutRepository AuditLogRepository

   // SessionRepository gerencia persistência de sessões de treino
   type SessionRepository interface {
       // Create insere nova sessão no banco
       Create(ctx context.Context, session *entities.Session) error

       // FindActiveByUserID retorna sessão ativa do usuário (se existir)
       // Retorna (nil, nil) se não houver sessão ativa
       FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error)
   }
   ```

2. **Criar interface `WorkoutRepository`**:
   ```go
   // WorkoutRepository gerencia persistência de workouts
   type WorkoutRepository interface {
       // FindByID retorna workout por ID
       FindByID(ctx context.Context, workoutID uuid.UUID) (*entities.Workout, error)

       // ExistsByIDAndUserID verifica se workout existe e pertence ao usuário
       ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error)
   }
   ```

3. **Criar interface `AuditLogRepository`**:
   ```go
   // AuditLogRepository gerencia log de auditoria (append-only)
   type AuditLogRepository interface {
       // Append registra evento de auditoria
       Append(ctx context.Context, entry *entities.AuditLog) error
   }
   ```

### Critério de aceite (testes/checks)
- [ ] Interfaces compilam sem erro
- [ ] Comentários Godoc descrevem comportamento esperado
- [ ] Diretiva `//go:generate moq` está presente
- [ ] `make mocks` gera mocks sem erro
- [ ] Mocks gerados estão em `ports/mocks/repositories.go`
- [ ] `make lint` passa sem warnings

---

## T05 — Criar queries SQLC para persistência

### Objetivo
Escrever queries SQL type-safe usando SQLC para as operações de Session, Workout e AuditLog.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/repositories/queries.sql` (novo arquivo)

### Implementação (passos)

1. **Criar arquivo `queries.sql`** (se não existir):
   ```sql
   -- name: CreateSession :exec
   INSERT INTO sessions (
       id,
       user_id,
       workout_id,
       started_at,
       status,
       notes,
       created_at,
       updated_at
   ) VALUES (
       $1, $2, $3, $4, $5, $6, $7, $8
   );

   -- name: FindActiveSessionByUserID :one
   SELECT 
       id,
       user_id,
       workout_id,
       started_at,
       finished_at,
       status,
       notes,
       created_at,
       updated_at
   FROM sessions
   WHERE user_id = $1 AND status = 'active'
   LIMIT 1;

   -- name: ExistsWorkoutByIDAndUserID :one
   SELECT EXISTS(
       SELECT 1 
       FROM workouts 
       WHERE id = $1 AND user_id = $2
   ) AS exists;

   -- name: AppendAuditLog :exec
   INSERT INTO audit_log (
       id,
       user_id,
       entity_type,
       entity_id,
       action,
       action_data,
       occurred_at,
       ip_address,
       user_agent
   ) VALUES (
       $1, $2, $3, $4, $5, $6, $7, $8, $9
   );
   ```

2. **Gerar código SQLC**:
   ```bash
   make sqlc
   ```

3. **Verificar arquivos gerados**:
   - `internal/kinetria/gateways/repositories/models.go`
   - `internal/kinetria/gateways/repositories/queries.sql.go`
   - `internal/kinetria/gateways/repositories/db.go`

### Critério de aceite (testes/checks)
- [ ] Arquivo `queries.sql` criado com todas as queries
- [ ] `make sqlc` executa sem erro
- [ ] Código Go gerado compila sem erro
- [ ] Tipos gerados correspondem às entidades de domínio
- [ ] `make lint` passa sem warnings

**Nota**: Se migrations não existirem, este passo vai falhar. Verificar dependência de foundation-infrastructure.

---

## T06 — Implementar Use Case: StartSession

### Objetivo
Criar o use case de domínio que orquestra as validações e criação de sessão.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/sessions/uc_start_session.go` (novo arquivo)

### Implementação (passos)

1. **Criar estrutura do use case**:
   ```go
   package sessions

   import (
       "context"
       "encoding/json"
       "fmt"
       "time"

       "github.com/google/uuid"
       "internal/kinetria/domain/entities"
       "internal/kinetria/domain/errors"
       "internal/kinetria/domain/ports"
       "internal/kinetria/domain/vos"
   )

   // StartSessionInput representa os dados de entrada para iniciar sessão
   type StartSessionInput struct {
       UserID    uuid.UUID
       WorkoutID uuid.UUID
   }

   // StartSessionOutput representa os dados de saída
   type StartSessionOutput struct {
       Session entities.Session
   }

   // StartSessionUseCase orquestra a criação de uma nova sessão de treino
   type StartSessionUseCase struct {
       sessionRepo   ports.SessionRepository
       workoutRepo   ports.WorkoutRepository
       auditLogRepo  ports.AuditLogRepository
   }

   // NewStartSessionUseCase cria nova instância do use case
   func NewStartSessionUseCase(
       sessionRepo ports.SessionRepository,
       workoutRepo ports.WorkoutRepository,
       auditLogRepo ports.AuditLogRepository,
   ) *StartSessionUseCase {
       return &StartSessionUseCase{
           sessionRepo:  sessionRepo,
           workoutRepo:  workoutRepo,
           auditLogRepo: auditLogRepo,
       }
   }
   ```

2. **Implementar método `Execute`**:
   ```go
   // Execute executa o use case de iniciar sessão
   func (uc *StartSessionUseCase) Execute(
       ctx context.Context,
       input StartSessionInput,
   ) (StartSessionOutput, error) {
       // 1. Validar WorkoutID (formato)
       if input.WorkoutID == uuid.Nil {
           return StartSessionOutput{}, errors.ErrMalformedParameters
       }

       // 2. Verificar ownership do workout
       exists, err := uc.workoutRepo.ExistsByIDAndUserID(ctx, input.WorkoutID, input.UserID)
       if err != nil {
           return StartSessionOutput{}, fmt.Errorf("failed to check workout ownership: %w", err)
       }
       if !exists {
           return StartSessionOutput{}, errors.WrapNotFound("workout", input.WorkoutID)
       }

       // 3. Verificar sessão ativa duplicada
       activeSession, err := uc.sessionRepo.FindActiveByUserID(ctx, input.UserID)
       if err != nil {
           return StartSessionOutput{}, fmt.Errorf("failed to check active session: %w", err)
       }
       if activeSession != nil {
           return StartSessionOutput{}, errors.ErrActiveSessionExists
       }

       // 4. Criar entidade Session
       now := time.Now()
       session := entities.Session{
           ID:         uuid.New(),
           UserID:     input.UserID,
           WorkoutID:  input.WorkoutID,
           StartedAt:  now,
           FinishedAt: nil,
           Status:     vos.SessionStatusActive,
           Notes:      "",
           CreatedAt:  now,
           UpdatedAt:  now,
       }

       // 5. Persistir sessão
       if err := uc.sessionRepo.Create(ctx, &session); err != nil {
           return StartSessionOutput{}, fmt.Errorf("failed to create session: %w", err)
       }

       // 6. Registrar audit log
       actionData, _ := json.Marshal(map[string]interface{}{
           "workoutId": session.WorkoutID.String(),
           "startedAt": session.StartedAt.Format(time.RFC3339),
       })

       auditEntry := entities.AuditLog{
           ID:         uuid.New(),
           UserID:     session.UserID,
           EntityType: "session",
           EntityID:   session.ID,
           Action:     "created",
           ActionData: actionData,
           OccurredAt: now,
           IPAddress:  extractIPFromContext(ctx),
           UserAgent:  extractUserAgentFromContext(ctx),
       }

       // Fire-and-forget (não bloqueia se falhar)
       _ = uc.auditLogRepo.Append(ctx, &auditEntry)

       return StartSessionOutput{Session: session}, nil
   }
   ```

3. **Criar helpers para extrair contexto**:
   ```go
   // extractIPFromContext extrai IP do request context
   func extractIPFromContext(ctx context.Context) string {
       // TODO: implementar extração real do contexto HTTP
       return "0.0.0.0"
   }

   // extractUserAgentFromContext extrai User-Agent do request context
   func extractUserAgentFromContext(ctx context.Context) string {
       // TODO: implementar extração real do contexto HTTP
       return ""
   }
   ```

### Critério de aceite (testes/checks)
- [ ] Use case compila sem erro
- [ ] Todas as validações implementadas (ownership, duplicação)
- [ ] Audit log é registrado após criação bem-sucedida
- [ ] Comentários Godoc em funções exportadas
- [ ] Testes unitários (ver T07)
- [ ] `make lint` passa sem warnings

---

## T07 — Criar testes unitários do Use Case (table-driven)

### Objetivo
Testar o use case `StartSessionUseCase` com mocks dos repositórios.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/sessions/uc_start_session_test.go` (novo arquivo)

### Implementação (passos)

1. **Gerar mocks**:
   ```bash
   make mocks
   ```

2. **Criar testes table-driven**:
   ```go
   package sessions_test

   import (
       "context"
       "testing"

       "github.com/google/uuid"
       "github.com/stretchr/testify/assert"
       "internal/kinetria/domain/entities"
       "internal/kinetria/domain/errors"
       "internal/kinetria/domain/ports/mocks"
       "internal/kinetria/domain/sessions"
       "internal/kinetria/domain/vos"
   )

   func TestStartSessionUseCase_Execute(t *testing.T) {
       userID := uuid.New()
       workoutID := uuid.New()

       tests := []struct {
           name          string
           input         sessions.StartSessionInput
           mockSetup     func(*mocks.SessionRepositoryMock, *mocks.WorkoutRepositoryMock)
           expectedError error
       }{
           {
               name: "sucesso - cria sessão sem sessão ativa",
               input: sessions.StartSessionInput{
                   UserID:    userID,
                   WorkoutID: workoutID,
               },
               mockSetup: func(sr *mocks.SessionRepositoryMock, wr *mocks.WorkoutRepositoryMock) {
                   wr.ExistsByIDAndUserIDFunc = func(ctx context.Context, wid, uid uuid.UUID) (bool, error) {
                       return true, nil // workout existe e pertence ao usuário
                   }
                   sr.FindActiveByUserIDFunc = func(ctx context.Context, uid uuid.UUID) (*entities.Session, error) {
                       return nil, nil // sem sessão ativa
                   }
                   sr.CreateFunc = func(ctx context.Context, s *entities.Session) error {
                       return nil
                   }
               },
               expectedError: nil,
           },
           {
               name: "erro - workout não pertence ao usuário",
               input: sessions.StartSessionInput{
                   UserID:    userID,
                   WorkoutID: workoutID,
               },
               mockSetup: func(sr *mocks.SessionRepositoryMock, wr *mocks.WorkoutRepositoryMock) {
                   wr.ExistsByIDAndUserIDFunc = func(ctx context.Context, wid, uid uuid.UUID) (bool, error) {
                       return false, nil // workout não pertence ao usuário
                   }
               },
               expectedError: errors.ErrNotFound,
           },
           {
               name: "erro - usuário já tem sessão ativa",
               input: sessions.StartSessionInput{
                   UserID:    userID,
                   WorkoutID: workoutID,
               },
               mockSetup: func(sr *mocks.SessionRepositoryMock, wr *mocks.WorkoutRepositoryMock) {
                   wr.ExistsByIDAndUserIDFunc = func(ctx context.Context, wid, uid uuid.UUID) (bool, error) {
                       return true, nil
                   }
                   sr.FindActiveByUserIDFunc = func(ctx context.Context, uid uuid.UUID) (*entities.Session, error) {
                       return &entities.Session{ID: uuid.New()}, nil // sessão ativa existe
                   }
               },
               expectedError: errors.ErrActiveSessionExists,
           },
           {
               name: "erro - workoutID vazio",
               input: sessions.StartSessionInput{
                   UserID:    userID,
                   WorkoutID: uuid.Nil,
               },
               mockSetup:     func(sr *mocks.SessionRepositoryMock, wr *mocks.WorkoutRepositoryMock) {},
               expectedError: errors.ErrMalformedParameters,
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Arrange
               sessionRepoMock := &mocks.SessionRepositoryMock{}
               workoutRepoMock := &mocks.WorkoutRepositoryMock{}
               auditRepoMock := &mocks.AuditLogRepositoryMock{
                   AppendFunc: func(ctx context.Context, entry *entities.AuditLog) error {
                       return nil // audit sempre sucede em testes
                   },
               }

               tt.mockSetup(sessionRepoMock, workoutRepoMock)

               uc := sessions.NewStartSessionUseCase(sessionRepoMock, workoutRepoMock, auditRepoMock)

               // Act
               output, err := uc.Execute(context.Background(), tt.input)

               // Assert
               if tt.expectedError != nil {
                   assert.ErrorIs(t, err, tt.expectedError)
                   assert.Empty(t, output.Session.ID)
               } else {
                   assert.NoError(t, err)
                   assert.NotEqual(t, uuid.Nil, output.Session.ID)
                   assert.Equal(t, tt.input.UserID, output.Session.UserID)
                   assert.Equal(t, tt.input.WorkoutID, output.Session.WorkoutID)
                   assert.Equal(t, vos.SessionStatusActive, output.Session.Status)
               }
           })
       }
   }
   ```

### Critério de aceite (testes/checks)
- [ ] Testes cobrem happy path (sucesso)
- [ ] Testes cobrem sad paths (workout não existe, ownership, sessão duplicada, workoutID vazio)
- [ ] Testes usam mocks dos repositórios
- [ ] `make test` passa com todos os testes
- [ ] Cobertura > 80% no use case
- [ ] `make lint` passa sem warnings

---

## T08 — Implementar Handler HTTP (POST /sessions)

### Objetivo
Criar endpoint HTTP que expõe o use case StartSession.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/http/handler_sessions.go` (novo arquivo)
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/http/dto.go` (novo arquivo para DTOs)

### Implementação (passos)

1. **Criar DTOs**:
   ```go
   package http

   import (
       "time"
       "github.com/google/uuid"
   )

   // StartSessionRequestDTO representa o request body de POST /sessions
   type StartSessionRequestDTO struct {
       WorkoutID uuid.UUID `json:"workoutId" validate:"required,uuid"`
   }

   // SessionResponseDTO representa a sessão na resposta
   type SessionResponseDTO struct {
       ID         uuid.UUID  `json:"id"`
       WorkoutID  uuid.UUID  `json:"workoutId"`
       UserID     uuid.UUID  `json:"userId"`
       StartedAt  time.Time  `json:"startedAt"`
       FinishedAt *time.Time `json:"finishedAt"`
       Status     string     `json:"status"`
   }

   // ApiResponse wrapper genérico
   type ApiResponse struct {
       Data interface{} `json:"data,omitempty"`
       Meta interface{} `json:"meta,omitempty"`
   }

   // ApiError wrapper de erro
   type ApiError struct {
       Code    string                 `json:"code"`
       Message string                 `json:"message"`
       Details map[string]interface{} `json:"details,omitempty"`
   }
   ```

2. **Criar handler**:
   ```go
   package http

   import (
       "encoding/json"
       "net/http"

       "github.com/go-playground/validator/v10"
       "internal/kinetria/domain/errors"
       "internal/kinetria/domain/sessions"
   )

   // SessionsHandler gerencia endpoints de sessões
   type SessionsHandler struct {
       startSessionUC *sessions.StartSessionUseCase
       validator      *validator.Validate
   }

   // NewSessionsHandler cria nova instância do handler
   func NewSessionsHandler(
       startSessionUC *sessions.StartSessionUseCase,
       validator *validator.Validate,
   ) *SessionsHandler {
       return &SessionsHandler{
           startSessionUC: startSessionUC,
           validator:      validator,
       }
   }

   // StartSession manipula POST /api/v1/sessions
   func (h *SessionsHandler) StartSession(w http.ResponseWriter, r *http.Request) {
       ctx := r.Context()

       // 1. Parse request body
       var reqDTO StartSessionRequestDTO
       if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
           respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.", nil)
           return
       }

       // 2. Validar request
       if err := h.validator.Struct(reqDTO); err != nil {
           details := extractValidationErrors(err)
           respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.", details)
           return
       }

       // 3. Extrair userID do contexto (injetado pelo middleware JWT)
       userID, err := extractUserIDFromContext(ctx)
       if err != nil {
           respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.", nil)
           return
       }

       // 4. Chamar use case
       input := sessions.StartSessionInput{
           UserID:    userID,
           WorkoutID: reqDTO.WorkoutID,
       }

       output, err := h.startSessionUC.Execute(ctx, input)
       if err != nil {
           handleUseCaseError(w, err)
           return
       }

       // 5. Mapear para DTO
       responseDTO := SessionResponseDTO{
           ID:         output.Session.ID,
           WorkoutID:  output.Session.WorkoutID,
           UserID:     output.Session.UserID,
           StartedAt:  output.Session.StartedAt,
           FinishedAt: output.Session.FinishedAt,
           Status:     output.Session.Status.String(),
       }

       // 6. Responder 201 Created
       respondJSON(w, http.StatusCreated, ApiResponse{Data: responseDTO})
   }
   ```

3. **Criar helpers**:
   ```go
   // respondJSON envia resposta JSON
   func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(payload)
   }

   // respondError envia erro padronizado
   func respondError(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
       respondJSON(w, status, ApiError{
           Code:    code,
           Message: message,
           Details: details,
       })
   }

   // handleUseCaseError mapeia erros de domínio para HTTP
   func handleUseCaseError(w http.ResponseWriter, err error) {
       switch {
       case errors.Is(err, errors.ErrNotFound):
           respondError(w, http.StatusNotFound, "WORKOUT_NOT_FOUND", err.Error(), nil)
       case errors.Is(err, errors.ErrActiveSessionExists):
           respondError(w, http.StatusConflict, "ACTIVE_SESSION_EXISTS", "User already has an active session. Finish or abandon it before starting a new one.", nil)
       case errors.Is(err, errors.ErrMalformedParameters):
           respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid parameters.", nil)
       default:
           respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.", nil)
       }
   }

   // extractUserIDFromContext extrai userID do contexto JWT
   func extractUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
       // TODO: implementar extração real do contexto JWT
       return uuid.Nil, errors.New("not implemented")
   }

   // extractValidationErrors extrai detalhes de erros de validação
   func extractValidationErrors(err error) map[string]interface{} {
       details := make(map[string]interface{})
       if validationErrs, ok := err.(validator.ValidationErrors); ok {
           for _, e := range validationErrs {
               details[e.Field()] = fmt.Sprintf("validation failed on tag '%s'", e.Tag())
           }
       }
       return details
   }
   ```

4. **Registrar rota**:
   ```go
   // Em algum arquivo de setup de rotas (ex: router.go)
   func SetupRoutes(r chi.Router, handlers *SessionsHandler) {
       r.Route("/api/v1", func(r chi.Router) {
           r.Use(JWTMiddleware) // middleware de autenticação

           r.Post("/sessions", handlers.StartSession)
       })
   }
   ```

### Critério de aceite (testes/checks)
- [ ] Handler compila sem erro
- [ ] DTOs têm tags JSON e validate corretas
- [ ] Validação de request funciona (validator)
- [ ] Erros de domínio são mapeados para status HTTP corretos
- [ ] Response segue formato da API (ApiResponse wrapper)
- [ ] Testes de integração (ver T09)
- [ ] `make lint` passa sem warnings

---

## T09 — Criar testes de integração do Handler (HTTP + DB)

### Objetivo
Testar endpoint completo com database real (Docker Compose).

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/http/handler_sessions_integration_test.go` (novo arquivo)

### Implementação (passos)

1. **Setup de teste com DB real**:
   ```go
   package http_test

   import (
       "bytes"
       "context"
       "encoding/json"
       "net/http"
       "net/http/httptest"
       "testing"

       "github.com/go-chi/chi/v5"
       "github.com/google/uuid"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/require"
   )

   func TestStartSession_Integration(t *testing.T) {
       if testing.Short() {
           t.Skip("skipping integration test")
       }

       // Setup DB (assumindo helper de teste)
       db := setupTestDB(t)
       defer cleanupTestDB(t, db)

       // Setup repositories
       sessionRepo := repositories.NewSessionRepository(db)
       workoutRepo := repositories.NewWorkoutRepository(db)
       auditRepo := repositories.NewAuditLogRepository(db)

       // Setup use case
       startSessionUC := sessions.NewStartSessionUseCase(sessionRepo, workoutRepo, auditRepo)

       // Setup handler
       validator := validator.New()
       handler := http.NewSessionsHandler(startSessionUC, validator)

       // Setup router
       r := chi.NewRouter()
       r.Post("/api/v1/sessions", handler.StartSession)

       // Seed data
       userID := uuid.New()
       workoutID := seedWorkout(t, db, userID, "Treino de Peito")

       t.Run("sucesso - cria sessão", func(t *testing.T) {
           reqBody := map[string]interface{}{
               "workoutId": workoutID.String(),
           }
           bodyBytes, _ := json.Marshal(reqBody)

           req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewReader(bodyBytes))
           req = req.WithContext(contextWithUserID(req.Context(), userID)) // mock JWT context

           w := httptest.NewRecorder()
           r.ServeHTTP(w, req)

           assert.Equal(t, http.StatusCreated, w.Code)

           var response http.ApiResponse
           json.NewDecoder(w.Body).Decode(&response)

           sessionData := response.Data.(map[string]interface{})
           assert.NotEmpty(t, sessionData["id"])
           assert.Equal(t, workoutID.String(), sessionData["workoutId"])
           assert.Equal(t, "active", sessionData["status"])
       })

       t.Run("erro 409 - sessão ativa duplicada", func(t *testing.T) {
           // Criar sessão ativa primeiro
           seedActiveSession(t, db, userID, workoutID)

           reqBody := map[string]interface{}{
               "workoutId": workoutID.String(),
           }
           bodyBytes, _ := json.Marshal(reqBody)

           req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewReader(bodyBytes))
           req = req.WithContext(contextWithUserID(req.Context(), userID))

           w := httptest.NewRecorder()
           r.ServeHTTP(w, req)

           assert.Equal(t, http.StatusConflict, w.Code)

           var errorResp http.ApiError
           json.NewDecoder(w.Body).Decode(&errorResp)
           assert.Equal(t, "ACTIVE_SESSION_EXISTS", errorResp.Code)
       })
   }
   ```

2. **Helpers de teste**:
   ```go
   func setupTestDB(t *testing.T) *sql.DB {
       // TODO: conectar com PostgreSQL de teste
       // docker-compose up -d postgres-test
       return nil
   }

   func cleanupTestDB(t *testing.T, db *sql.DB) {
       // Limpar tabelas após teste
   }

   func seedWorkout(t *testing.T, db *sql.DB, userID uuid.UUID, name string) uuid.UUID {
       // Inserir workout no DB
       return uuid.New()
   }

   func seedActiveSession(t *testing.T, db *sql.DB, userID, workoutID uuid.UUID) {
       // Inserir sessão ativa no DB
   }

   func contextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
       // Mock de contexto com userID (simula middleware JWT)
       return context.WithValue(ctx, "userID", userID)
   }
   ```

### Critério de aceite (testes/checks)
- [ ] Testes de integração cobrem happy path (201 Created)
- [ ] Testes cobrem sad paths (401, 404, 409, 422)
- [ ] Testes usam database real (PostgreSQL via Docker)
- [ ] Limpeza de dados entre testes (transações rollback ou TRUNCATE)
- [ ] `make test-integration` passa (ou `go test -tags=integration`)
- [ ] Cobertura end-to-end > 70%

---

## T10 — Documentar API no código (Godoc)

### Objetivo
Adicionar comentários Godoc em todas as funções/tipos exportados.

### Arquivos/pacotes prováveis
- Todos os arquivos criados nas tasks anteriores

### Implementação (passos)

1. **Documentar entidades**:
   ```go
   // Session representa uma sessão de treino ativa, finalizada ou abandonada.
   // Uma sessão rastreia a execução de um workout específico por um usuário,
   // incluindo timestamp de início, status e notas opcionais.
   type Session struct { ... }
   ```

2. **Documentar use cases**:
   ```go
   // StartSessionUseCase orquestra a criação de uma nova sessão de treino.
   // Valida ownership do workout, previne duplicação de sessão ativa e
   // registra evento de auditoria.
   type StartSessionUseCase struct { ... }

   // Execute inicia uma nova sessão de treino para o usuário.
   //
   // Validações aplicadas:
   //   - Workout deve existir e pertencer ao usuário
   //   - Usuário não pode ter mais de uma sessão ativa
   //
   // Retorna:
   //   - StartSessionOutput com a sessão criada
   //   - errors.ErrNotFound se workout não existe ou não pertence ao usuário
   //   - errors.ErrActiveSessionExists se usuário já tem sessão ativa
   func (uc *StartSessionUseCase) Execute(...) (StartSessionOutput, error) { ... }
   ```

3. **Documentar handlers**:
   ```go
   // StartSession manipula POST /api/v1/sessions.
   // Cria nova sessão de treino para o usuário autenticado.
   //
   // Request body:
   //   {"workoutId": "uuid"}
   //
   // Responses:
   //   201 Created - sessão criada com sucesso
   //   401 Unauthorized - token JWT inválido
   //   404 Not Found - workout não existe
   //   409 Conflict - usuário já tem sessão ativa
   //   422 Unprocessable Entity - validação falhou
   func (h *SessionsHandler) StartSession(w http.ResponseWriter, r *http.Request) { ... }
   ```

### Critério de aceite (testes/checks)
- [ ] Todos os tipos exportados têm comentário Godoc
- [ ] Todas as funções exportadas têm comentário Godoc
- [ ] Comentários seguem formato: "NomeTipo faz X" ou "NomeFuncao faz Y"
- [ ] Comentários descrevem validações e comportamentos esperados
- [ ] `go doc` exibe documentação corretamente
- [ ] `make lint` passa sem warnings (golangci-lint verifica godoc)

---

## T11 — Documentar endpoint no README do domínio

### Objetivo
Atualizar documentação técnica da feature no repositório.

### Arquivos/pacotes prováveis
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/docs/sessions.md` (novo arquivo)
  OU
- `/home/runner/work/kinetria-back/kinetria-back/README.md` (seção de API)

### Implementação (passos)

1. **Criar arquivo de documentação**:
   ```markdown
   # Sessions API

   ## POST /api/v1/sessions

   Inicia uma nova sessão de treino para o usuário autenticado.

   ### Request

   **Headers**:
   - `Authorization: Bearer <JWT>`

   **Body**:
   ```json
   {
     "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
   }
   ```

   **Validações**:
   - `workoutId` (required, UUID): ID do workout a ser executado

   ### Response Success (201 Created)

   ```json
   {
     "data": {
       "id": "d4e5f6a7-b8c9-0123-defa-234567890123",
       "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
       "userId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
       "startedAt": "2026-02-23T14:00:00Z",
       "finishedAt": null,
       "status": "active"
     }
   }
   ```

   ### Error Responses

   | Status | Code | Descrição |
   |--------|------|-----------|
   | 401 | UNAUTHORIZED | Token JWT inválido ou expirado |
   | 404 | WORKOUT_NOT_FOUND | Workout não existe ou não pertence ao usuário |
   | 409 | ACTIVE_SESSION_EXISTS | Usuário já tem sessão ativa |
   | 422 | VALIDATION_ERROR | Request body inválido |

   ### Regras de Negócio

   1. Workout deve existir e pertencer ao usuário autenticado
   2. Usuário só pode ter 1 sessão ativa por vez
   3. Toda criação de sessão é registrada no audit log

   ### Exemplos

   **cURL**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/sessions \
     -H "Authorization: Bearer <TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{"workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"}'
   ```

   **httpie**:
   ```bash
   http POST :8080/api/v1/sessions \
     Authorization:"Bearer <TOKEN>" \
     workoutId="b2c3d4e5-f6a7-8901-bcde-f12345678901"
   ```
   ```

### Critério de aceite (testes/checks)
- [ ] Documentação criada com exemplos de request/response
- [ ] Todos os códigos de erro documentados
- [ ] Regras de negócio listadas
- [ ] Exemplos cURL/httpie funcionam
- [ ] Documentação revisada (typos, formato)

---

## T12 — Validar conformidade com OpenAPI spec

### Objetivo
Garantir que o endpoint implementado segue exatamente o contrato OpenAPI.

### Arquivos/pacotes prováveis
- `.thoughts/mvp-userflow/api-contract.yaml`
- Testes de contrato (opcional: usar ferramentas como Prism, Dredd)

### Implementação (passos)

1. **Validar schemas manualmente**:
   - Request body: `StartSessionRequestDTO` ✅ `workoutId: uuid`
   - Response 201: `SessionResponseDTO` ✅ campos corretos
   - Response 401/404/409/422: `ApiError` ✅ code + message

2. **Validar status codes**:
   - 201 Created ✅
   - 401 Unauthorized ✅
   - 404 Not Found ✅
   - 409 Conflict ✅
   - 422 Unprocessable Entity ✅

3. **(Opcional) Usar Prism para validação automatizada**:
   ```bash
   # Mock server baseado em OpenAPI
   npx @stoplight/prism-cli mock api-contract.yaml

   # Validar requests reais contra spec
   npx @stoplight/prism-cli proxy api-contract.yaml http://localhost:8080
   ```

### Critério de aceite (testes/checks)
- [ ] Todos os campos do request DTO batem com OpenAPI spec
- [ ] Todos os campos do response DTO batem com OpenAPI spec
- [ ] Status codes batem com spec
- [ ] Mensagens de erro seguem formato `ApiError`
- [ ] (Opcional) Validação automatizada com Prism passa

---

## T13 — Adicionar logs estruturados e métricas

### Objetivo
Instrumentar o endpoint com logs (zerolog) e métricas (Prometheus).

### Arquivos/pacotes prováveis
- Handler HTTP (`handler_sessions.go`)
- Use Case (`uc_start_session.go`)

### Implementação (passos)

1. **Adicionar logs no handler**:
   ```go
   import "github.com/rs/zerolog/log"

   func (h *SessionsHandler) StartSession(w http.ResponseWriter, r *http.Request) {
       start := time.Now()
       ctx := r.Context()

       // ... lógica do handler ...

       duration := time.Since(start)

       // Log de sucesso
       log.Info().
           Str("method", "POST").
           Str("path", "/api/v1/sessions").
           Str("user_id", userID.String()).
           Str("workout_id", reqDTO.WorkoutID.String()).
           Str("session_id", output.Session.ID.String()).
           Int("status", http.StatusCreated).
           Dur("duration_ms", duration).
           Msg("session_created")

       // Log de erro (no caso de falha)
       log.Error().
           Str("method", "POST").
           Str("path", "/api/v1/sessions").
           Str("user_id", userID.String()).
           Str("error", err.Error()).
           Int("status", status).
           Dur("duration_ms", duration).
           Msg("session_creation_failed")
   }
   ```

2. **Adicionar métricas Prometheus**:
   ```go
   import (
       "github.com/prometheus/client_golang/prometheus"
       "github.com/prometheus/client_golang/prometheus/promauto"
   )

   var (
       sessionsStartedTotal = promauto.NewCounterVec(
           prometheus.CounterOpts{
               Name: "sessions_started_total",
               Help: "Total number of sessions started",
           },
           []string{"user_id"},
       )

       sessionsStartErrorsTotal = promauto.NewCounterVec(
           prometheus.CounterOpts{
               Name: "sessions_start_errors_total",
               Help: "Total number of session start errors",
           },
           []string{"error_type"},
       )

       sessionStartDuration = promauto.NewHistogram(
           prometheus.HistogramOpts{
               Name:    "session_start_duration_seconds",
               Help:    "Duration of session start requests",
               Buckets: prometheus.DefBuckets,
           },
       )
   )

   // No handler
   defer func(start time.Time) {
       sessionStartDuration.Observe(time.Since(start).Seconds())
   }(time.Now())

   // Após sucesso
   sessionsStartedTotal.WithLabelValues(userID.String()).Inc()

   // Após erro
   sessionsStartErrorsTotal.WithLabelValues("conflict").Inc()
   ```

### Critério de aceite (testes/checks)
- [ ] Logs estruturados em JSON (zerolog)
- [ ] Logs contêm user_id, workout_id, session_id, status, duration
- [ ] Logs NÃO contêm dados sensíveis (tokens, senhas)
- [ ] Métricas Prometheus instrumentadas
- [ ] Endpoint `/metrics` expõe métricas corretamente
- [ ] Teste manual: `curl localhost:8080/metrics | grep sessions_started_total`

---

## Resumo das Tasks

| Task | Título | Arquivos Principais | Dependências |
|------|--------|---------------------|--------------|
| T01  | Criar entidades de domínio | `entities/entities.go` | - |
| T02  | Criar Value Objects | `vos/vos.go` | - |
| T03  | Criar erros customizados | `errors/errors.go` | - |
| T04  | Criar interfaces de repositório | `ports/repositories.go` | T01, T02 |
| T05  | Criar queries SQLC | `gateways/repositories/queries.sql` | T01, T04 |
| T06  | Implementar Use Case | `domain/sessions/uc_start_session.go` | T01-T05 |
| T07  | Testes unitários Use Case | `domain/sessions/uc_start_session_test.go` | T06 |
| T08  | Implementar Handler HTTP | `gateways/http/handler_sessions.go` | T06 |
| T09  | Testes integração Handler | `gateways/http/handler_sessions_integration_test.go` | T08 |
| T10  | Documentar código (Godoc) | Todos os arquivos | T01-T09 |
| T11  | Documentar API (README) | `docs/sessions.md` | T08 |
| T12  | Validar conformidade OpenAPI | Testes de contrato | T08 |
| T13  | Logs e métricas | Handler + Use Case | T08 |

---

## Checklist Final (Critérios de Aceite da Feature)

Antes de considerar a feature **completa**:

### Código
- [ ] Todas as tasks (T01-T13) concluídas
- [ ] `make build` compila sem erro
- [ ] `make lint` passa sem warnings
- [ ] `make test` passa com cobertura > 70%
- [ ] `make test-integration` passa (se aplicável)

### Funcionalidade
- [ ] Endpoint `POST /api/v1/sessions` responde 201 Created
- [ ] Validação de JWT funciona (401 sem token)
- [ ] Validação de ownership funciona (404 para workout de outro usuário)
- [ ] Constraint de sessão única funciona (409 em duplicação)
- [ ] Audit log é criado em toda sessão iniciada

### Documentação
- [ ] Godoc em todas as funções/tipos exportados
- [ ] Documentação da API atualizada (README ou docs/)
- [ ] Exemplos cURL/httpie funcionam

### Observabilidade
- [ ] Logs estruturados (zerolog) em JSON
- [ ] Métricas Prometheus (`sessions_started_total`, `sessions_start_errors_total`)
- [ ] Endpoint `/metrics` expõe métricas

### Conformidade
- [ ] Request/response seguem OpenAPI spec
- [ ] Status codes corretos (201, 401, 404, 409, 422)
- [ ] Formato de erro padronizado (`ApiError`)

---

**Documento gerado em**: 2026-02-23  
**Feature**: mvp-userflow (Start Workout Session)  
**Total de tasks**: 13  
**Estimativa**: 3-5 dias (1 dev experiente)  
**Próxima feature**: RecordSet, FinishSession, AbandonSession
