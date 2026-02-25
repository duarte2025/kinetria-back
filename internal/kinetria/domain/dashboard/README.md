# Dashboard Module

## Overview

The Dashboard module provides aggregated data for the authenticated user's home screen, including:
- User profile information
- Today's recommended workout
- Week progress (last 7 days)
- Weekly statistics (calories burned, total time)

## Architecture

This module follows the **BFF Aggregation Strategy** (see `.kiro/decisions/bff-aggregation-strategy.md`):
- Aggregation happens at the HTTP handler level using goroutines
- Each use case is independent and can be executed in parallel
- Fail-fast error handling: if any use case fails, the entire request fails

## Use Cases

### GetUserProfileUC
Returns basic user profile information (name, email, avatar).

**Input**: `UserID`  
**Output**: User profile data

### GetTodayWorkoutUC
Returns the workout recommended for today. 

**MVP Implementation**: Returns the user's first workout (ordered by `created_at ASC`).  
**Future**: Will implement smart workout rotation based on schedule.

**Input**: `UserID`  
**Output**: Workout entity or `nil` if user has no workouts

### GetWeekProgressUC
Returns an array of 7 days (today - 6 to today) with completion status.

**Status values**:
- `completed`: User completed a session on this day
- `missed`: Day is in the past with no completed session
- `future`: Day is in the future

**Input**: `UserID`  
**Output**: Array of 7 `DayProgress` items

**Day labels**: `["D", "S", "T", "Q", "Q", "S", "S"]` (Portuguese weekday abbreviations)

### GetWeekStatsUC
Calculates weekly statistics based on completed sessions.

**Calorie calculation**: `totalMinutes * 7 kcal/min` (ACSM guideline for moderate exercise)

**Input**: `UserID`  
**Output**: `Calories` (int), `TotalTimeMinutes` (int)

## API Endpoint

### GET /api/v1/dashboard

**Authentication**: Required (JWT Bearer token)

**Response**:
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "string",
      "email": "string",
      "profileImageUrl": "string"
    },
    "todayWorkout": {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "type": "string",
      "intensity": "string",
      "duration": 45,
      "imageUrl": "string"
    } | null,
    "weekProgress": [
      {
        "day": "S",
        "date": "2026-02-19",
        "status": "completed"
      }
    ],
    "stats": {
      "calories": 420,
      "totalTimeMinutes": 60
    }
  }
}
```

**Error Responses**:
- `401 UNAUTHORIZED`: Missing or invalid JWT token
- `500 INTERNAL_ERROR`: Failed to load dashboard data

## Implementation Details

### Parallel Aggregation

The `DashboardHandler` executes all 4 use cases in parallel using goroutines:

```go
ch := make(chan result, 4)

go func() { /* GetUserProfileUC */ }()
go func() { /* GetTodayWorkoutUC */ }()
go func() { /* GetWeekProgressUC */ }()
go func() { /* GetWeekStatsUC */ }()

// Collect results with fail-fast error handling
```

### Date Range Logic

Week progress uses `DATE(started_at)` to determine which day a session belongs to:
- A session started at 23:55 and finished at 00:10 counts for the start date
- Only sessions with `status = 'completed'` are counted
- Date range is inclusive: `[today - 6, today]`

### Null Handling

`todayWorkout` is `null` when:
- User has no workouts in the database
- Repository returns `nil` (not an error)

## Testing

### Manual Testing

```bash
# Register and login
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"Password123!"}' \
  | jq -r '.data.accessToken')

# Get dashboard
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/dashboard | jq
```

### Unit Tests

**Status**: Not implemented yet (deferred to follow-up PR)

Planned test coverage:
- Each use case with mocked repositories
- Handler with mocked use cases
- Edge cases: no workouts, no sessions, future dates

## Future Enhancements

1. **Smart Workout Rotation**: Implement workout scheduling based on user preferences
2. **Caching**: Add Redis cache for dashboard data (TTL: 5 minutes)
3. **Partial Failure Handling**: Return partial data if non-critical use cases fail
4. **Performance Monitoring**: Add metrics for aggregation latency
5. **Personalized Recommendations**: ML-based workout suggestions

## Dependencies

- `ports.UserRepository`: User data access
- `ports.WorkoutRepository`: Workout data access
- `ports.SessionRepository`: Session data access

## Related Documentation

- [BFF Aggregation Strategy](.kiro/decisions/bff-aggregation-strategy.md)
- [Dashboard Planning](.thoughts/dashboard/plan.md)
- [Test Scenarios](.thoughts/dashboard/test-scenarios.feature)
