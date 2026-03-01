# Execution Report — Exercise Library Endpoints

## Status Summary

| Task | Status | Notes |
|------|--------|-------|
| T01 | done | Migration 011 - add nullable columns to exercises |
| T02 | done | Migration 012 - seed 30 exercises |
| T03 | done | Entity Exercise - added 6 new *string fields |
| T04 | done | domain/exercises/types.go - ExerciseWithStats |
| T05 | done | ports/repositories.go - ExerciseFilters, ExerciseUserStats, ExerciseHistoryEntry, SetDetail; ExerciseRepository 4 new methods |
| T06 | done | SQL queries (manual - sqlc not available); models.go updated; exercises.sql.go updated |
| T07 | done | ExerciseRepository - List, GetByID, GetUserStats, GetHistory implemented |
| T08 | done | ListExercisesUC with validation |
| T09 | done | GetExerciseUC with optional auth |
| T10 | done | GetExerciseHistoryUC with pagination |
| T11 | done | DTOs in handler_exercises.go |
| T12 | done | ExercisesHandler with 3 endpoints |
| T13 | done | tryExtractUserIDFromJWT in middleware_auth.go |
| T14 | done | Routes: GET /exercises, GET /exercises/{id}, GET /exercises/{id}/history |
| T15 | done | main.go DI wiring for exercise use cases and handler |
| T16 | done | 14 unit tests for ListExercisesUC - all pass |
| T17 | done | 6 unit tests for GetExerciseUC - all pass |
| T18 | done | 11 unit tests for GetExerciseHistoryUC - all pass |
| T19 | skipped | Integration tests require running DB |
| T20 | skipped | Integration tests require running DB |
| T21 | skipped | Integration tests require running DB |
| T22 | skipped | Documentation task |
| T23 | skipped | Documentation task |
| T24 | skipped | Documentation task |

## Changes Summary

### Files Created
- `internal/kinetria/gateways/migrations/011_expand_exercises_table.sql`
- `internal/kinetria/gateways/migrations/012_seed_exercises.sql` (30 exercises)
- `internal/kinetria/domain/exercises/types.go`
- `internal/kinetria/domain/exercises/uc_list_exercises.go`
- `internal/kinetria/domain/exercises/uc_get_exercise.go`
- `internal/kinetria/domain/exercises/uc_get_exercise_history.go`
- `internal/kinetria/domain/exercises/uc_list_exercises_test.go` (14 tests)
- `internal/kinetria/domain/exercises/uc_get_exercise_test.go` (6 tests)
- `internal/kinetria/domain/exercises/uc_get_exercise_history_test.go` (11 tests)
- `internal/kinetria/gateways/http/handler_exercises.go`

### Files Modified
- `internal/kinetria/domain/entities/exercise.go` - added Description, Instructions, Tips, Difficulty, Equipment, VideoURL as *string
- `internal/kinetria/domain/ports/repositories.go` - added ExerciseFilters, ExerciseUserStats, SetDetail, ExerciseHistoryEntry; expanded ExerciseRepository interface
- `internal/kinetria/gateways/repositories/queries/models.go` - added nullable fields to Exercise SQLC model
- `internal/kinetria/gateways/repositories/queries/exercises.sql` - added 6 new queries
- `internal/kinetria/gateways/repositories/queries/exercises.sql.go` - added Go implementations for 6 new queries
- `internal/kinetria/gateways/repositories/exercise_repository.go` - implemented List, GetByID, GetUserStats, GetHistory
- `internal/kinetria/gateways/http/middleware_auth.go` - added tryExtractUserIDFromJWT helper
- `internal/kinetria/gateways/http/router.go` - added ExercisesHandler field and 3 new routes
- `cmd/kinetria/api/main.go` - registered exercise use cases and handler in Fx DI
- `internal/kinetria/domain/sessions/uc_record_set_test.go` - updated mock to implement new ExerciseRepository methods
- `internal/kinetria/tests/setup_test.go` - wired exercise handler in integration test setup

## Test Results

```
go test ./internal/kinetria/domain/exercises/... -v
--- PASS: TestGetExerciseHistoryUC_Execute (11 subtests)
--- PASS: TestGetExerciseUC_Execute (6 subtests)
--- PASS: TestListExercisesUC_Execute (14 subtests)
PASS  ok  github.com/kinetria/kinetria-back/internal/kinetria/domain/exercises  0.004s

go test ./internal/...
ok  domain/auth, domain/dashboard, domain/exercises, domain/profile, domain/sessions, domain/vos, domain/workouts
ok  gateways/http/health, gateways/repositories, tests
```

## Build

```
go build ./... ✅
go build ./cmd/kinetria/api/... ✅
```

## API Endpoints Added

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /api/v1/exercises | Optional | List exercises with filters |
| GET | /api/v1/exercises/{id} | Optional | Get exercise details + optional stats |
| GET | /api/v1/exercises/{id}/history | Required | Get user's exercise history |

## Query Params for GET /exercises
- `muscleGroup` - filter by muscle group (JSONB containment)
- `equipment` - filter by equipment
- `difficulty` - filter by difficulty
- `search` - full-text search on name (ILIKE)
- `page` (default: 1, min: 1)
- `pageSize` (default: 20, min: 1, max: 100)
