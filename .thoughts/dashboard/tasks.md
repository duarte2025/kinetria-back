# Tasks — Dashboard

**Feature**: Dashboard (agregação de dados do usuário)  
**Diretório**: `.thoughts/dashboard/`  
**Decisão arquitetural**: Agregação no handler HTTP com goroutines paralelas (seguindo `bff-aggregation-strategy.md`)

---

## Ordem de Implementação

**⚠️ BLOQUEIO**: Esta feature depende de repositories de `workouts` e `sessions`. Recomenda-se implementar apenas os métodos mínimos necessários ou aguardar as features completas.

**Ordem recomendada**:
1. T01-T03: Criar ports, queries SQLC e repositories
2. T04-T07: Implementar use cases atômicos (domain)
3. T08-T09: Implementar handler e rota
4. T10-T12: Testes e documentação

---

## T01 — Criar ports dos repositories necessários

**Objetivo**: Definir interfaces dos repositories que o dashboard precisa consumir.

**Arquivos/pacotes**:
- `internal/kinetria/domain/ports/repositories.go`

**Implementação**:

```go
// internal/kinetria/domain/ports/repositories.go

// WorkoutRepository (adicionar ao arquivo existente)
type WorkoutRepository interface {
	// GetFirstByUserID retorna o primeiro workout do usuário (ordenado por created_at ASC).
	// Retorna nil se o usuário não tiver workouts.
	GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error)
	
	// Outros métodos serão adicionados pela feature workouts
}

// SessionRepository (adicionar ao arquivo existente)
type SessionRepository interface {
	// GetCompletedSessionsByUserAndDateRange retorna todas as sessões completed do usuário
	// no intervalo de datas (inclusive).
	// Datas devem estar em UTC. Usa DATE(started_at) para determinar o dia.
	GetCompletedSessionsByUserAndDateRange(
		ctx context.Context,
		userID uuid.UUID,
		startDate time.Time,
		endDate time.Time,
	) ([]entities.Session, error)
	
	// Outros métodos serão adicionados pela feature sessions
}
```

**Critério de aceite**:
- [ ] Interfaces `WorkoutRepository` e `SessionRepository` declaradas em `ports/repositories.go`
- [ ] Métodos documentados com comentários Godoc
- [ ] Código compila sem erros

---

## T02 — Criar queries SQLC para dashboard

**Objetivo**: Escrever queries SQL que serão geradas pelo SQLC para suportar os repositories.

**Arquivos/pacotes**:
- `internal/kinetria/gateways/repositories/queries/workouts.sql` (novo arquivo)
- `internal/kinetria/gateways/repositories/queries/sessions.sql` (novo arquivo)

**Implementação**:

### `workouts.sql`
```sql
-- name: GetFirstWorkoutByUserID :one
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
ORDER BY created_at ASC
LIMIT 1;
```

### `sessions.sql`
```sql
-- name: GetCompletedSessionsByDateRange :many
SELECT 
    id, 
    user_id, 
    workout_id, 
    status, 
    notes,
    started_at, 
    finished_at, 
    created_at, 
    updated_at
FROM sessions
WHERE user_id = $1
  AND status = 'completed'
  AND DATE(started_at) BETWEEN $2 AND $3
ORDER BY started_at DESC;
```

**Decisões**:
- `DATE(started_at)`: sessão iniciada às 23:55 e terminada às 00:10 = conta no dia de início
- `status = 'completed'`: apenas sessões completadas contam (não "active" nem "abandoned")
- `ORDER BY started_at DESC`: mais recentes primeiro (facilita debug)

**Gerar código SQLC**:
```bash
make sqlc
```

**Critério de aceite**:
- [ ] Arquivos `.sql` criados em `queries/`
- [ ] `make sqlc` roda sem erros
- [ ] Código gerado em `internal/kinetria/gateways/repositories/sqlc_*.go`
- [ ] Queries retornam structs corretas (verificar tipos no código gerado)

---

## T03 — Implementar repositories (Workout e Session)

**Objetivo**: Implementar os repositories concretos usando SQLC.

**Arquivos/pacotes**:
- `internal/kinetria/gateways/repositories/workout_repository.go` (novo)
- `internal/kinetria/gateways/repositories/session_repository.go` (novo)
- `internal/kinetria/gateways/repositories/module.go` (atualizar providers fx)

**Implementação**:

### `workout_repository.go`
```go
package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/sqlc"
)

type WorkoutRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewWorkoutRepository(queries *sqlc.Queries) ports.WorkoutRepository {
	return &WorkoutRepositoryImpl{queries: queries}
}

func (r *WorkoutRepositoryImpl) GetFirstByUserID(ctx context.Context, userID uuid.UUID) (*entities.Workout, error) {
	row, err := r.queries.GetFirstWorkoutByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Sem workouts = retornar nil (não é erro)
		}
		return nil, err
	}

	return &entities.Workout{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		Description: row.Description,
		Type:        row.Type,
		Intensity:   row.Intensity,
		Duration:    int(row.Duration),
		ImageURL:    row.ImageUrl,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
```

### `session_repository.go`
```go
package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/sqlc"
)

type SessionRepositoryImpl struct {
	queries *sqlc.Queries
}

func NewSessionRepository(queries *sqlc.Queries) ports.SessionRepository {
	return &SessionRepositoryImpl{queries: queries}
}

func (r *SessionRepositoryImpl) GetCompletedSessionsByUserAndDateRange(
	ctx context.Context,
	userID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) ([]entities.Session, error) {
	rows, err := r.queries.GetCompletedSessionsByDateRange(ctx, sqlc.GetCompletedSessionsByDateRangeParams{
		UserID:      userID,
		StartedAt:   startDate,
		StartedAt_2: endDate,
	})
	if err != nil {
		return nil, err
	}

	sessions := make([]entities.Session, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, entities.Session{
			ID:         row.ID,
			UserID:     row.UserID,
			WorkoutID:  row.WorkoutID,
			Status:     row.Status,
			Notes:      row.Notes,
			StartedAt:  row.StartedAt,
			FinishedAt: row.FinishedAt, // *time.Time (null se ativa)
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}

	return sessions, nil
}
```

### `module.go` (atualizar providers fx)
```go
// Adicionar no slice de providers:
fx.Provide(NewWorkoutRepository),
fx.Provide(NewSessionRepository),
```

**Critério de aceite**:
- [ ] Repositories implementam as interfaces de `ports/repositories.go`
- [ ] `GetFirstByUserID` retorna `nil` (não erro) quando não há workouts
- [ ] `GetCompletedSessionsByUserAndDateRange` retorna slice vazio quando não há sessões
- [ ] Código compila sem erros
- [ ] Providers registrados no fx

---

## T04 — Implementar use case `GetUserProfileUC`

**Objetivo**: Use case que retorna dados de perfil do usuário (para seção `user` do dashboard).

**Arquivos/pacotes**:
- `internal/kinetria/domain/dashboard/uc_get_user_profile.go` (novo)
- `internal/kinetria/domain/dashboard/module.go` (novo, providers fx)

**Implementação**:

### `uc_get_user_profile.go`
```go
package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetUserProfileInput struct {
	UserID uuid.UUID
}

type GetUserProfileOutput struct {
	ID              uuid.UUID
	Name            string
	Email           string
	ProfileImageURL string
}

type GetUserProfileUC struct {
	userRepo ports.UserRepository
}

func NewGetUserProfileUC(userRepo ports.UserRepository) *GetUserProfileUC {
	return &GetUserProfileUC{userRepo: userRepo}
}

func (uc *GetUserProfileUC) Execute(ctx context.Context, input GetUserProfileInput) (*GetUserProfileOutput, error) {
	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetUserProfileOutput{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		ProfileImageURL: user.ProfileImageURL,
	}, nil
}
```

### `module.go`
```go
package dashboard

import "go.uber.org/fx"

var Module = fx.Module("dashboard",
	fx.Provide(
		NewGetUserProfileUC,
		// outros use cases virão aqui
	),
)
```

**Registrar módulo no `cmd/kinetria/main.go`**:
```go
import "github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"

// No fx.New():
dashboard.Module,
```

**Critério de aceite**:
- [ ] Use case retorna dados do usuário
- [ ] Retorna erro se usuário não existir
- [ ] Código compila e módulo registrado no fx

---

## T05 — Implementar use case `GetTodayWorkoutUC`

**Objetivo**: Retornar o "workout de hoje" (MVP: primeiro workout do usuário, ou null se não tiver).

**Arquivos/pacotes**:
- `internal/kinetria/domain/dashboard/uc_get_today_workout.go` (novo)
- `internal/kinetria/domain/dashboard/module.go` (atualizar)

**Implementação**:

```go
package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetTodayWorkoutInput struct {
	UserID uuid.UUID
}

type GetTodayWorkoutOutput struct {
	Workout *entities.Workout // null se usuário não tem workouts
}

type GetTodayWorkoutUC struct {
	workoutRepo ports.WorkoutRepository
}

func NewGetTodayWorkoutUC(workoutRepo ports.WorkoutRepository) *GetTodayWorkoutUC {
	return &GetTodayWorkoutUC{workoutRepo: workoutRepo}
}

func (uc *GetTodayWorkoutUC) Execute(ctx context.Context, input GetTodayWorkoutInput) (*GetTodayWorkoutOutput, error) {
	workout, err := uc.workoutRepo.GetFirstByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetTodayWorkoutOutput{Workout: workout}, nil // null se não encontrar
}
```

**Atualizar `module.go`**:
```go
fx.Provide(
	NewGetUserProfileUC,
	NewGetTodayWorkoutUC, // adicionar
),
```

**Critério de aceite**:
- [ ] Retorna `Workout != nil` se usuário tem workouts
- [ ] Retorna `Workout == nil` se usuário não tem workouts
- [ ] Código compila

---

## T06 — Implementar use case `GetWeekProgressUC`

**Objetivo**: Retornar array de 7 dias (últimos 7 dias incluindo hoje) com status de progresso.

**Arquivos/pacotes**:
- `internal/kinetria/domain/dashboard/uc_get_week_progress.go` (novo)
- `internal/kinetria/domain/dashboard/module.go` (atualizar)

**Implementação**:

```go
package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetWeekProgressInput struct {
	UserID uuid.UUID
}

type DayProgress struct {
	Day    string // "S", "T", "Q", "Q", "S", "S", "D"
	Date   string // "2026-02-17" (formato ISO)
	Status string // "completed", "missed", "future"
}

type GetWeekProgressOutput struct {
	Days []DayProgress // sempre 7 itens
}

type GetWeekProgressUC struct {
	sessionRepo ports.SessionRepository
}

func NewGetWeekProgressUC(sessionRepo ports.SessionRepository) *GetWeekProgressUC {
	return &GetWeekProgressUC{sessionRepo: sessionRepo}
}

func (uc *GetWeekProgressUC) Execute(ctx context.Context, input GetWeekProgressInput) (*GetWeekProgressOutput, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDate := today.AddDate(0, 0, -6) // 6 dias atrás
	endDate := today

	// Buscar sessões completed na semana
	sessions, err := uc.sessionRepo.GetCompletedSessionsByUserAndDateRange(ctx, input.UserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Mapear datas de sessões completed
	completedDates := make(map[string]bool)
	for _, s := range sessions {
		dateStr := s.StartedAt.Format("2006-01-02")
		completedDates[dateStr] = true
	}

	// Gerar array de 7 dias
	days := make([]DayProgress, 7)
	dayLabels := []string{"D", "S", "T", "Q", "Q", "S", "S"} // domingo=0, segunda=1, ...

	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		weekday := int(date.Weekday()) // 0=Sunday, 1=Monday, ...

		status := "missed"
		if date.After(today) {
			status = "future"
		} else if completedDates[dateStr] {
			status = "completed"
		}

		days[i] = DayProgress{
			Day:    dayLabels[weekday],
			Date:   dateStr,
			Status: status,
		}
	}

	return &GetWeekProgressOutput{Days: days}, nil
}
```

**Atualizar `module.go`**:
```go
fx.Provide(
	NewGetUserProfileUC,
	NewGetTodayWorkoutUC,
	NewGetWeekProgressUC, // adicionar
),
```

**Critério de aceite**:
- [ ] Retorna exatamente 7 itens
- [ ] Primeiro item = hoje - 6 dias, último item = hoje
- [ ] Status "completed" apenas para dias com sessão completed
- [ ] Status "future" para dias > hoje
- [ ] Status "missed" para dias ≤ hoje sem sessão
- [ ] Labels em português corretos (D, S, T, Q, Q, S, S)

---

## T07 — Implementar use case `GetWeekStatsUC`

**Objetivo**: Calcular estatísticas da semana (calorias e total de minutos).

**Arquivos/pacotes**:
- `internal/kinetria/domain/dashboard/uc_get_week_stats.go` (novo)
- `internal/kinetria/domain/dashboard/module.go` (atualizar)

**Implementação**:

```go
package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

type GetWeekStatsInput struct {
	UserID uuid.UUID
}

type GetWeekStatsOutput struct {
	Calories          int
	TotalTimeMinutes  int
}

type GetWeekStatsUC struct {
	sessionRepo ports.SessionRepository
}

func NewGetWeekStatsUC(sessionRepo ports.SessionRepository) *GetWeekStatsUC {
	return &GetWeekStatsUC{sessionRepo: sessionRepo}
}

func (uc *GetWeekStatsUC) Execute(ctx context.Context, input GetWeekStatsInput) (*GetWeekStatsOutput, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDate := today.AddDate(0, 0, -6)
	endDate := today

	sessions, err := uc.sessionRepo.GetCompletedSessionsByUserAndDateRange(ctx, input.UserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	totalMinutes := 0
	for _, s := range sessions {
		if s.FinishedAt != nil {
			duration := s.FinishedAt.Sub(s.StartedAt)
			totalMinutes += int(duration.Minutes())
		}
		// Se FinishedAt == nil mas status == completed: bug! Logar warning e ignorar.
	}

	// Calorias estimadas: 7 kcal/min (ACSM guideline para exercício moderado)
	calories := totalMinutes * 7

	return &GetWeekStatsOutput{
		Calories:         calories,
		TotalTimeMinutes: totalMinutes,
	}, nil
}
```

**Atualizar `module.go`**:
```go
fx.Provide(
	NewGetUserProfileUC,
	NewGetTodayWorkoutUC,
	NewGetWeekProgressUC,
	NewGetWeekStatsUC, // adicionar
),
```

**Critério de aceite**:
- [ ] Retorna 0 se não houver sessões
- [ ] Calcula duration corretamente (finished_at - started_at)
- [ ] Ignora sessões com `FinishedAt == nil` (não deveria existir para status=completed)
- [ ] Calorias = totalMinutes * 7
- [ ] Código compila

---

## T08 — Implementar `DashboardHandler` com agregação paralela

**Objetivo**: Handler HTTP que agrega os 4 use cases em paralelo e retorna DTO.

**Arquivos/pacotes**:
- `internal/kinetria/gateways/http/handler_dashboard.go` (novo)
- `internal/kinetria/gateways/http/module.go` (atualizar provider fx)

**Implementação**:

```go
package service

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

type DashboardHandler struct {
	getUserProfileUC  *dashboard.GetUserProfileUC
	getTodayWorkoutUC *dashboard.GetTodayWorkoutUC
	getWeekProgressUC *dashboard.GetWeekProgressUC
	getWeekStatsUC    *dashboard.GetWeekStatsUC
}

func NewDashboardHandler(
	getUserProfileUC *dashboard.GetUserProfileUC,
	getTodayWorkoutUC *dashboard.GetTodayWorkoutUC,
	getWeekProgressUC *dashboard.GetWeekProgressUC,
	getWeekStatsUC *dashboard.GetWeekStatsUC,
) *DashboardHandler {
	return &DashboardHandler{
		getUserProfileUC:  getUserProfileUC,
		getTodayWorkoutUC: getTodayWorkoutUC,
		getWeekProgressUC: getWeekProgressUC,
		getWeekStatsUC:    getWeekStatsUC,
	}
}

// GetDashboard handles GET /api/v1/dashboard
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extrair userID do JWT (middleware injeta no context)
	userID, ok := ctx.Value(gatewayauth.UserIDKey).(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication")
		return
	}

	// Estrutura para coletar resultados das goroutines
	type result struct {
		user         *dashboard.GetUserProfileOutput
		todayWorkout *dashboard.GetTodayWorkoutOutput
		weekProgress *dashboard.GetWeekProgressOutput
		weekStats    *dashboard.GetWeekStatsOutput
		err          error
		source       string // para debug
	}

	ch := make(chan result, 4)

	// Executar use cases em paralelo
	go func() {
		out, err := h.getUserProfileUC.Execute(ctx, dashboard.GetUserProfileInput{UserID: userID})
		ch <- result{user: out, err: err, source: "user"}
	}()

	go func() {
		out, err := h.getTodayWorkoutUC.Execute(ctx, dashboard.GetTodayWorkoutInput{UserID: userID})
		ch <- result{todayWorkout: out, err: err, source: "todayWorkout"}
	}()

	go func() {
		out, err := h.getWeekProgressUC.Execute(ctx, dashboard.GetWeekProgressInput{UserID: userID})
		ch <- result{weekProgress: out, err: err, source: "weekProgress"}
	}()

	go func() {
		out, err := h.getWeekStatsUC.Execute(ctx, dashboard.GetWeekStatsInput{UserID: userID})
		ch <- result{weekStats: out, err: err, source: "weekStats"}
	}()

	// Coletar resultados
	var res result
	for i := 0; i < 4; i++ {
		r := <-ch
		if r.err != nil {
			// Fail-fast: se qualquer use case falhar, retornar erro
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load dashboard data")
			return
		}

		// Merge resultados
		if r.user != nil {
			res.user = r.user
		}
		if r.todayWorkout != nil {
			res.todayWorkout = r.todayWorkout
		}
		if r.weekProgress != nil {
			res.weekProgress = r.weekProgress
		}
		if r.weekStats != nil {
			res.weekStats = r.weekStats
		}
	}

	// Montar DTO de resposta
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":              res.user.ID.String(),
			"name":            res.user.Name,
			"email":           res.user.Email,
			"profileImageUrl": res.user.ProfileImageURL,
		},
		"todayWorkout": nil, // default null
		"weekProgress": mapWeekProgressToDTO(res.weekProgress.Days),
		"stats": map[string]interface{}{
			"calories":          res.weekStats.Calories,
			"totalTimeMinutes":  res.weekStats.TotalTimeMinutes,
		},
	}

	// TodayWorkout pode ser null
	if res.todayWorkout.Workout != nil {
		w := res.todayWorkout.Workout
		response["todayWorkout"] = map[string]interface{}{
			"id":          w.ID.String(),
			"name":        w.Name,
			"description": w.Description,
			"type":        w.Type,
			"intensity":   w.Intensity,
			"duration":    w.Duration,
			"imageUrl":    w.ImageURL,
		}
	}

	writeSuccess(w, http.StatusOK, response)
}

// Helper: mapear weekProgress para DTO
func mapWeekProgressToDTO(days []dashboard.DayProgress) []map[string]interface{} {
	result := make([]map[string]interface{}, len(days))
	for i, d := range days {
		result[i] = map[string]interface{}{
			"day":    d.Day,
			"date":   d.Date,
			"status": d.Status,
		}
	}
	return result
}
```

**Atualizar `module.go`**:
```go
fx.Provide(
	NewAuthHandler,
	NewDashboardHandler, // adicionar
),
```

**Critério de aceite**:
- [ ] Handler registrado no fx
- [ ] Extrai `userID` do context corretamente
- [ ] Executa 4 use cases em paralelo
- [ ] Retorna 500 se qualquer use case falhar (fail-fast)
- [ ] Retorna DTO correto com todos os campos
- [ ] `todayWorkout = null` se usuário não tem workouts
- [ ] Código compila

---

## T09 — Registrar rota `GET /dashboard` no router

**Objetivo**: Expor o endpoint `/api/v1/dashboard` protegido por JWT.

**Arquivos/pacotes**:
- `internal/kinetria/gateways/http/router.go`

**Implementação**:

```go
// router.go

func NewServiceRouter(
	authHandler *AuthHandler,
	dashboardHandler *DashboardHandler, // adicionar dependência
	jwtManager *gatewayauth.JWTManager,
) *ServiceRouter {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Rotas públicas (auth)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
		r.Post("/logout", authHandler.Logout)
	})

	// Rotas protegidas
	r.Route("/api/v1", func(r chi.Router) {
		// Middleware JWT para rotas protegidas
		r.Use(jwtManager.AuthMiddleware)

		r.Get("/dashboard", dashboardHandler.GetDashboard) // adicionar
	})

	return &ServiceRouter{router: r}
}
```

**Critério de aceite**:
- [ ] Rota registrada em `/api/v1/dashboard`
- [ ] Middleware JWT aplicado (retorna 401 se não autenticado)
- [ ] Endpoint acessível via `curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/dashboard`
- [ ] Código compila

---

## T10 — Testes unitários dos use cases

**Objetivo**: Testar cada use case isolado com mocks de repositories (table-driven tests).

**Arquivos/pacotes**:
- `internal/kinetria/domain/dashboard/uc_get_user_profile_test.go`
- `internal/kinetria/domain/dashboard/uc_get_today_workout_test.go`
- `internal/kinetria/domain/dashboard/uc_get_week_progress_test.go`
- `internal/kinetria/domain/dashboard/uc_get_week_stats_test.go`

**Ferramentas**:
- `testify/assert`
- `moq` para gerar mocks (ou criar mocks manualmente)

**Exemplo: `uc_get_user_profile_test.go`**
```go
package dashboard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/stretchr/testify/assert"
)

// Mock manual (ou usar moq)
type mockUserRepo struct {
	user *entities.User
	err  error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) { return nil, nil }
func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error                { return nil }

func TestGetUserProfileUC_Execute(t *testing.T) {
	tests := []struct {
		name       string
		userID     uuid.UUID
		mockUser   *entities.User
		mockErr    error
		wantOutput *dashboard.GetUserProfileOutput
		wantErr    bool
	}{
		{
			name:   "sucesso - retorna perfil do usuário",
			userID: uuid.New(),
			mockUser: &entities.User{
				ID:              uuid.New(),
				Name:            "Bruno Costa",
				Email:           "bruno@example.com",
				ProfileImageURL: "https://cdn.example.com/avatar.jpg",
			},
			mockErr: nil,
			wantOutput: &dashboard.GetUserProfileOutput{
				Name:            "Bruno Costa",
				Email:           "bruno@example.com",
				ProfileImageURL: "https://cdn.example.com/avatar.jpg",
			},
			wantErr: false,
		},
		{
			name:       "erro - usuário não encontrado",
			userID:     uuid.New(),
			mockUser:   nil,
			mockErr:    errors.New("user not found"),
			wantOutput: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{user: tt.mockUser, err: tt.mockErr}
			uc := dashboard.NewGetUserProfileUC(repo)

			output, err := uc.Execute(context.Background(), dashboard.GetUserProfileInput{UserID: tt.userID})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput.Name, output.Name)
				assert.Equal(t, tt.wantOutput.Email, output.Email)
			}
		})
	}
}
```

**Outros testes (seguir mesmo padrão)**:
- `GetTodayWorkoutUC`: testar com workout existente, sem workout, erro no repo
- `GetWeekProgressUC`: testar dias completed/missed/future, semana vazia, múltiplas sessões no mesmo dia
- `GetWeekStatsUC`: testar cálculo de calorias, sessões sem FinishedAt, semana vazia

**Critério de aceite**:
- [ ] Todos os use cases têm testes table-driven
- [ ] Cobertura > 80% (verificar com `go test -cover`)
- [ ] Testes passam: `go test ./internal/kinetria/domain/dashboard/...`
- [ ] Casos de erro cobertos (repo retorna erro, dados vazios)

---

## T11 — Testes de integração do handler

**Objetivo**: Testar o handler completo com mocks dos use cases (ou banco real com testcontainers).

**Arquivos/pacotes**:
- `internal/kinetria/gateways/http/handler_dashboard_test.go`

**Estratégia**:
- Opção A: Mockar use cases e testar apenas agregação paralela no handler
- Opção B: Usar testcontainers + banco real e testar endpoint completo (E2E)

**Exemplo (Opção A - mocks)**:
```go
package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	service "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/dashboard"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
	"github.com/stretchr/testify/assert"
)

// Mocks dos use cases (implementar cada um)
type mockGetUserProfileUC struct {
	output *dashboard.GetUserProfileOutput
	err    error
}

func (m *mockGetUserProfileUC) Execute(ctx context.Context, input dashboard.GetUserProfileInput) (*dashboard.GetUserProfileOutput, error) {
	return m.output, m.err
}

// ... outros mocks

func TestDashboardHandler_GetDashboard(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		mockUser       *dashboard.GetUserProfileOutput
		mockWorkout    *dashboard.GetTodayWorkoutOutput
		mockProgress   *dashboard.GetWeekProgressOutput
		mockStats      *dashboard.GetWeekStatsOutput
		wantStatusCode int
	}{
		{
			name: "sucesso - dashboard completo",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), gatewayauth.UserIDKey, userID)
			},
			mockUser: &dashboard.GetUserProfileOutput{
				ID:    userID,
				Name:  "Test User",
				Email: "test@example.com",
			},
			mockWorkout: &dashboard.GetTodayWorkoutOutput{Workout: &entities.Workout{Name: "Treino A"}},
			mockProgress: &dashboard.GetWeekProgressOutput{Days: []dashboard.DayProgress{
				{Day: "S", Date: "2026-02-17", Status: "completed"},
			}},
			mockStats: &dashboard.GetWeekStatsOutput{Calories: 350, TotalTimeMinutes: 50},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "erro - sem autenticação",
			setupContext: func() context.Context {
				return context.Background() // sem userID no context
			},
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup handler com mocks
			handler := service.NewDashboardHandler(
				&mockGetUserProfileUC{output: tt.mockUser},
				&mockGetTodayWorkoutUC{output: tt.mockWorkout},
				&mockGetWeekProgressUC{output: tt.mockProgress},
				&mockGetWeekStatsUC{output: tt.mockStats},
			)

			req := httptest.NewRequest("GET", "/api/v1/dashboard", nil)
			req = req.WithContext(tt.setupContext())
			w := httptest.NewRecorder()

			handler.GetDashboard(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantStatusCode == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "data")
				// Validar campos específicos
			}
		})
	}
}
```

**Critério de aceite**:
- [ ] Testes cobrem: sucesso, sem autenticação, erro em use case
- [ ] Testes validam status code e estrutura da resposta
- [ ] Testes passam: `go test ./internal/kinetria/gateways/http/...`

---

## T12 — Documentar API

**Objetivo**: Atualizar documentação da API com endpoint `/dashboard` (contrato, exemplos, códigos de erro).

**Arquivos/pacotes**:
- `.thoughts/mvp-userflow/api-contract.yaml` (atualizar)
- `README.md` (ou criar `internal/kinetria/domain/dashboard/README.md`)

**Implementação**:

### Atualizar `api-contract.yaml`
```yaml
# Adicionar rota em paths:
paths:
  /dashboard:
    get:
      summary: Get user dashboard data
      description: |
        Returns aggregated data for the authenticated user's dashboard.
        Includes user profile, today's workout, week progress, and stats.
      operationId: getDashboard
      tags:
        - dashboard
      security:
        - BearerAuth: []
      responses:
        '200':
          description: Dashboard data loaded successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'
              examples:
                success:
                  value:
                    data:
                      user:
                        id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
                        name: "Bruno Costa"
                        email: "bruno@example.com"
                        profileImageUrl: null
                      todayWorkout:
                        id: "b2c3d4e5-f6a7-8901-bcde-f12345678901"
                        name: "Treino Peito/Tríceps"
                        description: "Foco em hipertrofia"
                        type: "FORÇA"
                        intensity: "Alta"
                        duration: 45
                        imageUrl: ""
                      weekProgress:
                        - { day: "S", date: "2026-02-17", status: "completed" }
                        - { day: "T", date: "2026-02-18", status: "missed" }
                        - { day: "Q", date: "2026-02-19", status: "missed" }
                        - { day: "Q", date: "2026-02-20", status: "completed" }
                        - { day: "S", date: "2026-02-21", status: "missed" }
                        - { day: "S", date: "2026-02-22", status: "missed" }
                        - { day: "D", date: "2026-02-23", status: "missed" }
                      stats:
                        calories: 420
                        totalTimeMinutes: 60
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalError'
```

### Criar `dashboard/README.md` (opcional)
```markdown
# Dashboard — Use Cases

## Visão Geral

Módulo responsável por agregar dados do usuário para exibição no dashboard.

## Use Cases

### GetUserProfileUC
Retorna dados de perfil do usuário (nome, email, avatar).

### GetTodayWorkoutUC
Retorna o workout de hoje (MVP: primeiro workout do usuário).

### GetWeekProgressUC
Retorna array de 7 dias com status de progresso (completed/missed/future).

### GetWeekStatsUC
Calcula estatísticas da semana (calorias e total de minutos).

## Agregação

A agregação ocorre no `DashboardHandler` (camada HTTP) usando goroutines paralelas.
Veja `bff-aggregation-strategy.md` para detalhes arquiteturais.

## Cálculo de Calorias

Calorias estimadas: `totalMinutes * 7 kcal/min` (ACSM guideline para exercício moderado).

## Exemplo de Resposta

```json
{
  "data": {
    "user": { ... },
    "todayWorkout": { ... } | null,
    "weekProgress": [ ... ],
    "stats": { "calories": 420, "totalTimeMinutes": 60 }
  }
}
```
```

### Adicionar Godoc nos use cases
Exemplo:
```go
// GetUserProfileUC retorna os dados de perfil do usuário autenticado.
// Usado pelo endpoint GET /dashboard para exibir informações básicas do usuário.
type GetUserProfileUC struct { ... }
```

**Critério de aceite**:
- [ ] Endpoint `/dashboard` documentado em `api-contract.yaml`
- [ ] Exemplos de request/response incluídos
- [ ] Códigos de erro documentados (401, 500)
- [ ] Godoc adicionado em todos os use cases exportados
- [ ] README do módulo dashboard criado (opcional mas recomendado)

---

## ✅ Checklist Final

Após implementar todas as tasks:

- [ ] T01: Ports criados
- [ ] T02: Queries SQLC geradas
- [ ] T03: Repositories implementados
- [ ] T04-T07: Use cases implementados
- [ ] T08: Handler com agregação paralela implementado
- [ ] T09: Rota registrada no router
- [ ] T10: Testes unitários passando
- [ ] T11: Testes de integração passando
- [ ] T12: Documentação atualizada

**Validações**:
```bash
# Compilar
make build

# Testes
go test ./internal/kinetria/domain/dashboard/... -v -cover
go test ./internal/kinetria/gateways/http/... -run TestDashboard -v

# Rodar servidor
make run

# Testar endpoint manualmente
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123!"}' \
  | jq -r '.data.accessToken')

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/dashboard | jq
```

**Critérios de aceite gerais**:
- [ ] Endpoint retorna 200 com JWT válido
- [ ] Endpoint retorna 401 sem JWT
- [ ] `todayWorkout = null` se usuário sem workouts
- [ ] `weekProgress` tem 7 itens
- [ ] `stats` calculados corretamente
- [ ] Agregação paralela < 500ms
- [ ] Logs estruturados presentes
- [ ] Sem regressões em testes existentes
