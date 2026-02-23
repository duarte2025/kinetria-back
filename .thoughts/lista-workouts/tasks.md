# Tasks — lista-workouts

## Visão Geral

Este backlog contém as tarefas necessárias para implementar o endpoint `GET /workouts` (lista de workouts do usuário autenticado).

**Pré-requisitos** (bloquantes):
- ✅ Feature **foundation-infrastructure** implementada (migrations, entidades Workout/Exercise, Docker Compose)
- ✅ Feature **AUTH** com middleware de autenticação JWT disponível

**Ordem de execução**: As tarefas devem ser executadas sequencialmente (T01 → T10).

---

## T01 — Criar port WorkoutRepository

### Objetivo
Definir a interface (port) do repositório de workouts no domínio.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/ports/workout_repository.go` (criar)

### Implementação (passos)

1. Criar arquivo `internal/kinetria/domain/ports/workout_repository.go`
2. Definir interface `WorkoutRepository`:
   ```go
   package ports
   
   import (
       "context"
       "github.com/google/uuid"
       "kinetria-back/internal/kinetria/domain/entities"
   )
   
   type WorkoutRepository interface {
       // ListByUserID retorna workouts do usuário com paginação
       // Params:
       //   - ctx: context para cancelamento/timeout
       //   - userID: UUID do usuário autenticado
       //   - offset: número de registros para pular
       //   - limit: número máximo de registros para retornar
       // Returns:
       //   - []entities.Workout: lista de workouts
       //   - int: total de workouts do usuário (para calcular totalPages)
       //   - error: erro se houver falha na consulta
       ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error)
   }
   ```

3. Adicionar documentação Godoc na interface

### Critério de aceite
- [ ] Interface `WorkoutRepository` criada com método `ListByUserID`
- [ ] Assinatura do método documentada com Godoc
- [ ] Arquivo compilável (`go build ./...` sem erros)
- [ ] Comentários explicando parâmetros e retornos

---

## T02 — Implementar queries SQLC

### Objetivo
Criar queries SQL tipadas para listar workouts e contar total.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/repositories/queries/workouts.sql` (criar)
- `sqlc.yaml` (já existe, configurado)

### Implementação (passos)

1. Criar arquivo `internal/kinetria/gateways/repositories/queries/workouts.sql`

2. Implementar query `ListWorkoutsByUserID`:
   ```sql
   -- name: ListWorkoutsByUserID :many
   SELECT 
       id, 
       user_id, 
       name, 
       description, 
       type, 
       intensity, 
       duration, 
       image_url, 
       created_at, 
       updated_at
   FROM workouts
   WHERE user_id = $1
   ORDER BY created_at DESC
   LIMIT $2 OFFSET $3;
   ```

3. Implementar query `CountWorkoutsByUserID`:
   ```sql
   -- name: CountWorkoutsByUserID :one
   SELECT COUNT(*) 
   FROM workouts
   WHERE user_id = $1;
   ```

4. Gerar código Go com SQLC:
   ```bash
   make sqlc-generate
   # ou
   sqlc generate
   ```

5. Verificar código gerado em `internal/kinetria/gateways/repositories/sqlc/`

### Critério de aceite
- [ ] Query `ListWorkoutsByUserID` criada com ORDER BY created_at DESC
- [ ] Query `CountWorkoutsByUserID` criada
- [ ] Código Go gerado com sucesso (`sqlc generate` sem erros)
- [ ] Tipos Go gerados correspondem à entidade `Workout` de domínio
- [ ] Query utiliza índice (verificar EXPLAIN se possível)

---

## T03 — Implementar WorkoutRepository (adapter)

### Objetivo
Implementar o adapter do repositório que usa SQLC para acessar o banco de dados.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/repositories/workout_repository.go` (criar)

### Implementação (passos)

1. Criar arquivo `internal/kinetria/gateways/repositories/workout_repository.go`

2. Implementar struct `WorkoutRepository`:
   ```go
   package repositories
   
   import (
       "context"
       "github.com/google/uuid"
       "kinetria-back/internal/kinetria/domain/entities"
       "kinetria-back/internal/kinetria/domain/ports"
       "kinetria-back/internal/kinetria/gateways/repositories/sqlc"
   )
   
   type WorkoutRepository struct {
       queries *sqlc.Queries
   }
   
   func NewWorkoutRepository(queries *sqlc.Queries) ports.WorkoutRepository {
       return &WorkoutRepository{
           queries: queries,
       }
   }
   ```

3. Implementar método `ListByUserID`:
   ```go
   func (r *WorkoutRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error) {
       // 1. Buscar workouts com paginação
       rows, err := r.queries.ListWorkoutsByUserID(ctx, sqlc.ListWorkoutsByUserIDParams{
           UserID: userID,
           Limit:  int32(limit),
           Offset: int32(offset),
       })
       if err != nil {
           return nil, 0, fmt.Errorf("failed to list workouts: %w", err)
       }
       
       // 2. Buscar total de workouts
       total, err := r.queries.CountWorkoutsByUserID(ctx, userID)
       if err != nil {
           return nil, 0, fmt.Errorf("failed to count workouts: %w", err)
       }
       
       // 3. Mapear para entidades de domínio
       workouts := make([]entities.Workout, len(rows))
       for i, row := range rows {
           workouts[i] = mapSQLCWorkoutToEntity(row)
       }
       
       return workouts, int(total), nil
   }
   ```

4. Implementar função de mapeamento `mapSQLCWorkoutToEntity`:
   ```go
   func mapSQLCWorkoutToEntity(row sqlc.Workout) entities.Workout {
       return entities.Workout{
           ID:          row.ID,
           UserID:      row.UserID,
           Name:        row.Name,
           Description: row.Description,
           Type:        row.Type,
           Intensity:   row.Intensity,
           Duration:    int(row.Duration),
           ImageURL:    row.ImageURL,
           CreatedAt:   row.CreatedAt,
           UpdatedAt:   row.UpdatedAt,
       }
   }
   ```

5. Adicionar provider no `cmd/kinetria/api/main.go`:
   ```go
   fx.Provide(
       fx.Annotate(
           repositories.NewWorkoutRepository,
           fx.As(new(ports.WorkoutRepository)),
       ),
   ),
   ```

### Critério de aceite
- [ ] Struct `WorkoutRepository` implementada
- [ ] Método `ListByUserID` implementado com 2 queries (list + count)
- [ ] Mapeamento SQLC → entidade de domínio implementado
- [ ] Provider registrado no fx com interface `ports.WorkoutRepository`
- [ ] Código compilável (`go build ./...` sem erros)
- [ ] Implementação respeita interface `ports.WorkoutRepository`

---

## T04 — Implementar ListWorkoutsUC (use case)

### Objetivo
Implementar o caso de uso de listagem de workouts com validação de input e cálculo de paginação.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/workouts/uc_list_workouts.go` (criar)
- `internal/kinetria/domain/workouts/` (criar pasta se não existir)

### Implementação (passos)

1. Criar pasta `internal/kinetria/domain/workouts/` se não existir

2. Criar arquivo `uc_list_workouts.go`

3. Definir structs de input/output:
   ```go
   package workouts
   
   import (
       "context"
       "fmt"
       "math"
       "github.com/google/uuid"
       "kinetria-back/internal/kinetria/domain/entities"
       "kinetria-back/internal/kinetria/domain/ports"
   )
   
   type ListWorkoutsInput struct {
       UserID   uuid.UUID
       Page     int  // >= 1
       PageSize int  // >= 1, <= 100
   }
   
   type ListWorkoutsOutput struct {
       Workouts   []entities.Workout
       Total      int
       Page       int
       PageSize   int
       TotalPages int
   }
   ```

4. Implementar use case:
   ```go
   type ListWorkoutsUC struct {
       repo ports.WorkoutRepository
   }
   
   func NewListWorkoutsUC(repo ports.WorkoutRepository) *ListWorkoutsUC {
       return &ListWorkoutsUC{
           repo: repo,
       }
   }
   
   func (uc *ListWorkoutsUC) Execute(ctx context.Context, input ListWorkoutsInput) (ListWorkoutsOutput, error) {
       // 1. Validar input
       if err := uc.validateInput(input); err != nil {
           return ListWorkoutsOutput{}, err
       }
       
       // 2. Aplicar defaults
       page := input.Page
       if page == 0 {
           page = 1
       }
       pageSize := input.PageSize
       if pageSize == 0 {
           pageSize = 20
       }
       
       // 3. Calcular offset
       offset := (page - 1) * pageSize
       
       // 4. Buscar workouts
       workouts, total, err := uc.repo.ListByUserID(ctx, input.UserID, offset, pageSize)
       if err != nil {
           return ListWorkoutsOutput{}, fmt.Errorf("failed to list workouts: %w", err)
       }
       
       // 5. Calcular totalPages
       totalPages := 0
       if total > 0 {
           totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
       }
       
       return ListWorkoutsOutput{
           Workouts:   workouts,
           Total:      total,
           Page:       page,
           PageSize:   pageSize,
           TotalPages: totalPages,
       }, nil
   }
   
   func (uc *ListWorkoutsUC) validateInput(input ListWorkoutsInput) error {
       if input.UserID == uuid.Nil {
           return fmt.Errorf("userID is required")
       }
       if input.Page < 0 {
           return fmt.Errorf("page must be greater than or equal to 1")
       }
       if input.PageSize < 0 || input.PageSize > 100 {
           return fmt.Errorf("pageSize must be between 1 and 100")
       }
       return nil
   }
   ```

5. Registrar provider no `main.go`:
   ```go
   fx.Provide(
       workouts.NewListWorkoutsUC,
   ),
   ```

### Critério de aceite
- [ ] Use case implementado com validação de input
- [ ] Defaults aplicados (page=1, pageSize=20)
- [ ] Cálculo de offset e totalPages correto
- [ ] Erros de validação retornam mensagens claras
- [ ] Provider registrado no fx
- [ ] Código compilável
- [ ] UserID validado (não pode ser uuid.Nil)

---

## T05 — Implementar DTOs (Data Transfer Objects)

### Objetivo
Criar DTOs para serialização JSON da resposta HTTP.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/http/dto_workouts.go` (criar)
- `internal/kinetria/gateways/http/dto_common.go` (criar se não existir)

### Implementação (passos)

1. Criar arquivo `dto_common.go` (se não existir):
   ```go
   package http
   
   type ApiResponseDTO struct {
       Data interface{}        `json:"data"`
       Meta *PaginationMetaDTO `json:"meta,omitempty"`
   }
   
   type ApiErrorDTO struct {
       Code    string                 `json:"code"`
       Message string                 `json:"message"`
       Details map[string]interface{} `json:"details,omitempty"`
   }
   
   type PaginationMetaDTO struct {
       Page       int `json:"page"`
       PageSize   int `json:"pageSize"`
       Total      int `json:"total"`
       TotalPages int `json:"totalPages"`
   }
   ```

2. Criar arquivo `dto_workouts.go`:
   ```go
   package http
   
   import (
       "kinetria-back/internal/kinetria/domain/entities"
   )
   
   type WorkoutSummaryDTO struct {
       ID          string  `json:"id"`
       Name        string  `json:"name"`
       Description *string `json:"description"` // nullable
       Type        *string `json:"type"`        // nullable
       Intensity   *string `json:"intensity"`   // nullable
       Duration    int     `json:"duration"`
       ImageURL    *string `json:"imageUrl"`    // nullable
   }
   
   func MapWorkoutToSummaryDTO(workout entities.Workout) WorkoutSummaryDTO {
       dto := WorkoutSummaryDTO{
           ID:       workout.ID.String(),
           Name:     workout.Name,
           Duration: workout.Duration,
       }
       
       // Mapear campos opcionais (string vazia -> nil)
       if workout.Description != "" {
           dto.Description = &workout.Description
       }
       if workout.Type != "" {
           dto.Type = &workout.Type
       }
       if workout.Intensity != "" {
           dto.Intensity = &workout.Intensity
       }
       if workout.ImageURL != "" {
           dto.ImageURL = &workout.ImageURL
       }
       
       return dto
   }
   
   func MapWorkoutsToSummaryDTOs(workouts []entities.Workout) []WorkoutSummaryDTO {
       dtos := make([]WorkoutSummaryDTO, len(workouts))
       for i, workout := range workouts {
           dtos[i] = MapWorkoutToSummaryDTO(workout)
       }
       return dtos
   }
   ```

### Critério de aceite
- [ ] DTOs criados: `ApiResponseDTO`, `ApiErrorDTO`, `PaginationMetaDTO`, `WorkoutSummaryDTO`
- [ ] Função de mapeamento `MapWorkoutToSummaryDTO` implementada
- [ ] Campos opcionais mapeados para ponteiros `nil` quando vazios
- [ ] JSON tags configuradas corretamente
- [ ] Código compilável

---

## T06 — Implementar WorkoutsHandler (HTTP handler)

### Objetivo
Implementar o handler HTTP que recebe a requisição, extrai parâmetros, chama o use case e retorna a resposta.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/http/handler_workouts.go` (criar)

### Implementação (passos)

1. Criar arquivo `handler_workouts.go`:
   ```go
   package http
   
   import (
       "context"
       "encoding/json"
       "net/http"
       "strconv"
       
       "github.com/google/uuid"
       "github.com/rs/zerolog/log"
       "kinetria-back/internal/kinetria/domain/workouts"
   )
   
   type WorkoutsHandler struct {
       listWorkoutsUC *workouts.ListWorkoutsUC
   }
   
   func NewWorkoutsHandler(listWorkoutsUC *workouts.ListWorkoutsUC) *WorkoutsHandler {
       return &WorkoutsHandler{
           listWorkoutsUC: listWorkoutsUC,
       }
   }
   ```

2. Implementar método `ListWorkouts`:
   ```go
   func (h *WorkoutsHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
       ctx := r.Context()
       
       // 1. Extrair userID do context (injetado pelo middleware de auth)
       userID, err := extractUserIDFromContext(ctx)
       if err != nil {
           respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid authentication")
           return
       }
       
       // 2. Parse query params
       page, err := parseIntQueryParam(r, "page", 1)
       if err != nil {
           respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "page must be a valid integer")
           return
       }
       
       pageSize, err := parseIntQueryParam(r, "pageSize", 20)
       if err != nil {
           respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "pageSize must be a valid integer")
           return
       }
       
       // 3. Chamar use case
       output, err := h.listWorkoutsUC.Execute(ctx, workouts.ListWorkoutsInput{
           UserID:   userID,
           Page:     page,
           PageSize: pageSize,
       })
       if err != nil {
           log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list workouts")
           respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
           return
       }
       
       // 4. Mapear para DTOs
       workoutDTOs := MapWorkoutsToSummaryDTOs(output.Workouts)
       
       // 5. Criar response
       response := ApiResponseDTO{
           Data: workoutDTOs,
           Meta: &PaginationMetaDTO{
               Page:       output.Page,
               PageSize:   output.PageSize,
               Total:      output.Total,
               TotalPages: output.TotalPages,
           },
       }
       
       // 6. Log de sucesso
       log.Info().
           Str("user_id", userID.String()).
           Int("page", page).
           Int("page_size", pageSize).
           Int("total", output.Total).
           Msg("list_workouts_success")
       
       // 7. Responder
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(http.StatusOK)
       json.NewEncoder(w).Encode(response)
   }
   ```

3. Implementar helpers:
   ```go
   func extractUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
       userIDValue := ctx.Value("userID")
       if userIDValue == nil {
           return uuid.Nil, fmt.Errorf("userID not found in context")
       }
       
       userID, ok := userIDValue.(uuid.UUID)
       if !ok {
           return uuid.Nil, fmt.Errorf("userID is not a valid UUID")
       }
       
       return userID, nil
   }
   
   func parseIntQueryParam(r *http.Request, key string, defaultValue int) (int, error) {
       valueStr := r.URL.Query().Get(key)
       if valueStr == "" {
           return defaultValue, nil
       }
       
       value, err := strconv.Atoi(valueStr)
       if err != nil {
           return 0, fmt.Errorf("invalid value for %s", key)
       }
       
       return value, nil
   }
   
   func respondError(w http.ResponseWriter, status int, code, message string) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(ApiErrorDTO{
           Code:    code,
           Message: message,
       })
   }
   ```

4. Registrar provider no `main.go`:
   ```go
   fx.Provide(
       httphandler.NewWorkoutsHandler,
   ),
   ```

### Critério de aceite
- [ ] Handler implementado com extração de userID do context
- [ ] Query params `page` e `pageSize` parseados corretamente
- [ ] Validação de input com respostas 422 apropriadas
- [ ] Chamada ao use case implementada
- [ ] Mapeamento para DTOs e response wrapper implementado
- [ ] Logs estruturados de sucesso e erro implementados
- [ ] Helpers de parsing e error response implementados
- [ ] Provider registrado no fx
- [ ] Código compilável

---

## T07 — Registrar rota no router

### Objetivo
Registrar a rota `GET /api/v1/workouts` no router HTTP com middleware de autenticação.

### Arquivos/pacotes prováveis
- `internal/kinetria/gateways/http/router.go` (editar ou criar)

### Implementação (passos)

1. Localizar ou criar arquivo `router.go`

2. Implementar função `NewServiceRouter`:
   ```go
   package http
   
   import (
       "net/http"
       "github.com/go-chi/chi/v5"
       "github.com/go-chi/chi/v5/middleware"
   )
   
   func NewServiceRouter(
       workoutsHandler *WorkoutsHandler,
       authMiddleware *AuthMiddleware, // Assumindo que existe
   ) http.Handler {
       r := chi.NewRouter()
       
       // Middlewares globais
       r.Use(middleware.RequestID)
       r.Use(middleware.RealIP)
       r.Use(middleware.Logger)
       r.Use(middleware.Recoverer)
       
       // Rotas públicas (se houver)
       r.Route("/api/v1", func(r chi.Router) {
           // Rotas protegidas (requerem autenticação)
           r.Group(func(r chi.Router) {
               r.Use(authMiddleware.Require) // Middleware que extrai userID do JWT
               
               // Workouts
               r.Get("/workouts", workoutsHandler.ListWorkouts)
           })
       })
       
       return r
   }
   ```

3. Registrar router no `main.go`:
   ```go
   fx.Provide(
       xhttp.AsRouter(httphandler.NewServiceRouter),
   ),
   ```

### Critério de aceite
- [ ] Rota `GET /api/v1/workouts` registrada
- [ ] Middleware de autenticação aplicado à rota
- [ ] Middlewares globais configurados (RequestID, Logger, Recoverer)
- [ ] Router registrado como provider no fx
- [ ] Servidor HTTP inicia sem erros (`go run cmd/kinetria/api/main.go`)

---

## T08 — Implementar testes unitários

### Objetivo
Criar testes unitários para use case, repository mock e handler.

### Arquivos/pacotes prováveis
- `internal/kinetria/domain/workouts/uc_list_workouts_test.go` (criar)
- `internal/kinetria/gateways/http/handler_workouts_test.go` (criar)

### Implementação (passos)

1. **Teste do Use Case** (`uc_list_workouts_test.go`):
   ```go
   package workouts_test
   
   import (
       "context"
       "testing"
       
       "github.com/google/uuid"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
       
       "kinetria-back/internal/kinetria/domain/entities"
       "kinetria-back/internal/kinetria/domain/workouts"
   )
   
   // Mock do repository
   type MockWorkoutRepository struct {
       mock.Mock
   }
   
   func (m *MockWorkoutRepository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]entities.Workout, int, error) {
       args := m.Called(ctx, userID, offset, limit)
       return args.Get(0).([]entities.Workout), args.Int(1), args.Error(2)
   }
   
   func TestListWorkoutsUC_Success(t *testing.T) {
       // Setup
       repo := new(MockWorkoutRepository)
       uc := workouts.NewListWorkoutsUC(repo)
       
       userID := uuid.New()
       mockWorkouts := []entities.Workout{
           {ID: uuid.New(), Name: "Treino A"},
           {ID: uuid.New(), Name: "Treino B"},
       }
       
       repo.On("ListByUserID", mock.Anything, userID, 0, 20).Return(mockWorkouts, 2, nil)
       
       // Execute
       output, err := uc.Execute(context.Background(), workouts.ListWorkoutsInput{
           UserID:   userID,
           Page:     1,
           PageSize: 20,
       })
       
       // Assert
       assert.NoError(t, err)
       assert.Len(t, output.Workouts, 2)
       assert.Equal(t, 2, output.Total)
       assert.Equal(t, 1, output.Page)
       assert.Equal(t, 20, output.PageSize)
       assert.Equal(t, 1, output.TotalPages)
       repo.AssertExpectations(t)
   }
   
   func TestListWorkoutsUC_ValidationError_InvalidPage(t *testing.T) {
       repo := new(MockWorkoutRepository)
       uc := workouts.NewListWorkoutsUC(repo)
       
       _, err := uc.Execute(context.Background(), workouts.ListWorkoutsInput{
           UserID:   uuid.New(),
           Page:     -1,
           PageSize: 20,
       })
       
       assert.Error(t, err)
       assert.Contains(t, err.Error(), "page must be greater than or equal to 1")
   }
   
   func TestListWorkoutsUC_ValidationError_InvalidPageSize(t *testing.T) {
       repo := new(MockWorkoutRepository)
       uc := workouts.NewListWorkoutsUC(repo)
       
       _, err := uc.Execute(context.Background(), workouts.ListWorkoutsInput{
           UserID:   uuid.New(),
           Page:     1,
           PageSize: 101,
       })
       
       assert.Error(t, err)
       assert.Contains(t, err.Error(), "pageSize must be between 1 and 100")
   }
   
   // Adicionar mais cenários: defaults, cálculo de offset, totalPages, etc.
   ```

2. **Teste do Handler** (`handler_workouts_test.go`):
   ```go
   package http_test
   
   import (
       "context"
       "encoding/json"
       "net/http"
       "net/http/httptest"
       "testing"
       
       "github.com/google/uuid"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
       
       httphandler "kinetria-back/internal/kinetria/gateways/http"
       "kinetria-back/internal/kinetria/domain/workouts"
   )
   
   type MockListWorkoutsUC struct {
       mock.Mock
   }
   
   func (m *MockListWorkoutsUC) Execute(ctx context.Context, input workouts.ListWorkoutsInput) (workouts.ListWorkoutsOutput, error) {
       args := m.Called(ctx, input)
       return args.Get(0).(workouts.ListWorkoutsOutput), args.Error(1)
   }
   
   func TestWorkoutsHandler_ListWorkouts_Success(t *testing.T) {
       // Setup
       mockUC := new(MockListWorkoutsUC)
       handler := httphandler.NewWorkoutsHandler(mockUC)
       
       userID := uuid.New()
       mockOutput := workouts.ListWorkoutsOutput{
           Workouts:   []entities.Workout{{Name: "Treino A"}},
           Total:      1,
           Page:       1,
           PageSize:   20,
           TotalPages: 1,
       }
       
       mockUC.On("Execute", mock.Anything, mock.Anything).Return(mockOutput, nil)
       
       // Request
       req := httptest.NewRequest("GET", "/workouts", nil)
       req = req.WithContext(context.WithValue(req.Context(), "userID", userID))
       w := httptest.NewRecorder()
       
       // Execute
       handler.ListWorkouts(w, req)
       
       // Assert
       assert.Equal(t, http.StatusOK, w.Code)
       
       var response httphandler.ApiResponseDTO
       json.Unmarshal(w.Body.Bytes(), &response)
       
       assert.NotNil(t, response.Data)
       assert.NotNil(t, response.Meta)
       assert.Equal(t, 1, response.Meta.Total)
   }
   
   // Adicionar mais cenários: erro de autenticação, validação, etc.
   ```

3. Rodar testes:
   ```bash
   go test ./internal/kinetria/domain/workouts/... -v
   go test ./internal/kinetria/gateways/http/... -v
   ```

### Critério de aceite
- [ ] Testes unitários do use case criados (happy path + validações)
- [ ] Testes do handler criados (sucesso, erro de auth, erro de validação)
- [ ] Mocks implementados para repository e use case
- [ ] Todos os testes passam (`go test ./...`)
- [ ] Cobertura > 80% do código da feature (`go test -cover`)
- [ ] Testes cobrem cenários BDD relevantes

---

## T09 — Implementar testes de integração

### Objetivo
Criar testes de integração que validam o fluxo completo (HTTP → Use Case → Repository → DB).

### Arquivos/pacotes prováveis
- `internal/kinetria/tests/integration/workouts_test.go` (criar)
- `internal/kinetria/tests/integration/setup_test.go` (setup de DB de teste)

### Implementação (passos)

1. Criar setup de DB de teste (`setup_test.go`):
   ```go
   package integration_test
   
   import (
       "context"
       "database/sql"
       "testing"
       
       _ "github.com/lib/pq"
       "kinetria-back/internal/kinetria/gateways/repositories/sqlc"
   )
   
   func setupTestDB(t *testing.T) (*sql.DB, *sqlc.Queries) {
       // Conectar ao DB de teste (PostgreSQL em Docker)
       connStr := "postgres://user:pass@localhost:5432/kinetria_test?sslmode=disable"
       db, err := sql.Open("postgres", connStr)
       if err != nil {
           t.Fatalf("Failed to connect to test DB: %v", err)
       }
       
       // Aplicar migrations (ou truncar tabelas)
       truncateTables(t, db)
       
       queries := sqlc.New(db)
       return db, queries
   }
   
   func truncateTables(t *testing.T, db *sql.DB) {
       _, err := db.Exec("TRUNCATE workouts, exercises, users RESTART IDENTITY CASCADE")
       if err != nil {
           t.Fatalf("Failed to truncate tables: %v", err)
       }
   }
   ```

2. Criar teste de integração (`workouts_test.go`):
   ```go
   package integration_test
   
   import (
       "context"
       "testing"
       
       "github.com/google/uuid"
       "github.com/stretchr/testify/assert"
       
       "kinetria-back/internal/kinetria/domain/workouts"
       "kinetria-back/internal/kinetria/gateways/repositories"
   )
   
   func TestListWorkouts_Integration(t *testing.T) {
       if testing.Short() {
           t.Skip("Skipping integration test")
       }
       
       // Setup
       db, queries := setupTestDB(t)
       defer db.Close()
       
       repo := repositories.NewWorkoutRepository(queries)
       uc := workouts.NewListWorkoutsUC(repo)
       
       userID := uuid.New()
       ctx := context.Background()
       
       // Seed data
       seedWorkouts(t, queries, userID, 5)
       
       // Execute
       output, err := uc.Execute(ctx, workouts.ListWorkoutsInput{
           UserID:   userID,
           Page:     1,
           PageSize: 20,
       })
       
       // Assert
       assert.NoError(t, err)
       assert.Len(t, output.Workouts, 5)
       assert.Equal(t, 5, output.Total)
       assert.Equal(t, 1, output.TotalPages)
   }
   
   func seedWorkouts(t *testing.T, queries *sqlc.Queries, userID uuid.UUID, count int) {
       for i := 0; i < count; i++ {
           _, err := queries.CreateWorkout(context.Background(), sqlc.CreateWorkoutParams{
               ID:       uuid.New(),
               UserID:   userID,
               Name:     fmt.Sprintf("Treino %d", i+1),
               Duration: 30,
           })
           if err != nil {
               t.Fatalf("Failed to seed workout: %v", err)
           }
       }
   }
   ```

3. Rodar testes de integração:
   ```bash
   # Subir DB de teste
   docker-compose -f docker-compose.test.yml up -d
   
   # Rodar testes
   go test ./internal/kinetria/tests/integration/... -v
   
   # Derrubar DB
   docker-compose -f docker-compose.test.yml down
   ```

### Critério de aceite
- [ ] Setup de DB de teste criado (Docker Compose com PostgreSQL)
- [ ] Testes de integração implementados (happy path, paginação, isolamento de usuários)
- [ ] Seed de dados de teste implementado
- [ ] Testes passam com DB real (`go test ./internal/kinetria/tests/integration/...`)
- [ ] Cleanup de tabelas entre testes (TRUNCATE ou transactions)
- [ ] Documentação de como rodar testes de integração (README ou comentários)

---

## T10 — Documentar API

### Objetivo
Atualizar a documentação da API com exemplos de requisição/resposta do endpoint `GET /workouts`.

### Arquivos/pacotes prováveis
- `README.md` (editar)
- `docs/api/workouts.md` (criar se houver pasta de docs)

### Implementação (passos)

1. Atualizar `README.md` com seção de endpoints implementados:
   ```markdown
   ## Endpoints Disponíveis
   
   ### GET /api/v1/workouts
   
   Lista todos os workouts do usuário autenticado com paginação.
   
   **Autenticação**: Obrigatória (JWT Bearer)
   
   **Query Parameters**:
   - `page` (int, opcional, default: 1) - Número da página
   - `pageSize` (int, opcional, default: 20, max: 100) - Itens por página
   
   **Exemplo de Requisição**:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/workouts?page=1&pageSize=10" \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   ```
   
   **Exemplo de Resposta (200 OK)**:
   ```json
   {
     "data": [
       {
         "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
         "name": "Treino de Peito e Tríceps",
         "description": "Foco em hipertrofia com exercícios compostos",
         "type": "FORÇA",
         "intensity": "Alta",
         "duration": 45,
         "imageUrl": "https://cdn.kinetria.app/workouts/chest.jpg"
       }
     ],
     "meta": {
       "page": 1,
       "pageSize": 10,
       "total": 1,
       "totalPages": 1
     }
   }
   ```
   
   **Possíveis Erros**:
   - `401 Unauthorized` - Token JWT inválido ou ausente
   - `422 Validation Error` - Parâmetros de paginação inválidos
   - `500 Internal Error` - Erro interno do servidor
   ```

2. Criar documentação detalhada (opcional):
   ```bash
   mkdir -p docs/api
   ```
   
   Criar `docs/api/workouts.md` com:
   - Descrição detalhada do endpoint
   - Exemplos de todos os cenários (sucesso, erros)
   - Schemas de request/response
   - Códigos de erro possíveis

3. Adicionar comentários Godoc nas funções públicas:
   ```go
   // ListWorkouts retorna a lista de workouts do usuário autenticado com paginação.
   //
   // Query Parameters:
   //   - page: número da página (default: 1, min: 1)
   //   - pageSize: quantidade de itens por página (default: 20, min: 1, max: 100)
   //
   // Requer autenticação JWT Bearer.
   //
   // Responses:
   //   - 200: lista de workouts retornada com sucesso
   //   - 401: token JWT inválido ou ausente
   //   - 422: parâmetros de validação inválidos
   //   - 500: erro interno do servidor
   func (h *WorkoutsHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
       // ...
   }
   ```

### Critério de aceite
- [ ] README.md atualizado com seção do endpoint `GET /workouts`
- [ ] Exemplos de curl com request e response documentados
- [ ] Códigos de erro possíveis documentados
- [ ] Comentários Godoc adicionados nas funções públicas
- [ ] (Opcional) Documentação detalhada em `docs/api/workouts.md`
- [ ] Validação: seguir formato de documentação do OpenAPI spec (`.thoughts/mvp-userflow/api-contract.yaml`)

---

## Resumo do Backlog

| Task | Descrição | Arquivos Principais | Bloqueante |
|------|-----------|---------------------|------------|
| T01  | Criar port WorkoutRepository | `domain/ports/workout_repository.go` | - |
| T02  | Implementar queries SQLC | `gateways/repositories/queries/workouts.sql` | T01 |
| T03  | Implementar WorkoutRepository | `gateways/repositories/workout_repository.go` | T02 |
| T04  | Implementar ListWorkoutsUC | `domain/workouts/uc_list_workouts.go` | T01 |
| T05  | Implementar DTOs | `gateways/http/dto_*.go` | - |
| T06  | Implementar WorkoutsHandler | `gateways/http/handler_workouts.go` | T04, T05 |
| T07  | Registrar rota no router | `gateways/http/router.go` | T06 |
| T08  | Implementar testes unitários | `*_test.go` | T04, T06 |
| T09  | Implementar testes de integração | `tests/integration/*_test.go` | T03, T04 |
| T10  | Documentar API | `README.md`, `docs/` | T07 |

**Estimativa total**: 8-12 horas (1-2 dias de desenvolvimento)

**Dependências externas** (bloquantes):
- ✅ Foundation-infrastructure (migrations, entidades)
- ✅ Feature AUTH (middleware de autenticação)

---

**Documento gerado em**: 2026-02-23  
**Versão**: 1.0  
**Status**: ✅ Pronto para execução  
**Próximo passo**: Iniciar implementação com T01
