package sessions

import (
"context"
"encoding/json"
"fmt"
"time"

"github.com/google/uuid"
"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

// StartSessionInput holds the input for starting a workout session.
type StartSessionInput struct {
UserID    uuid.UUID
WorkoutID uuid.UUID
}

// StartSessionOutput holds the result of starting a session.
type StartSessionOutput struct {
Session entities.Session
}

// StartSessionUC orchestrates creating a new workout session.
type StartSessionUC struct {
sessionRepo  ports.SessionRepository
workoutRepo  ports.WorkoutRepository
auditLogRepo ports.AuditLogRepository
}

// NewStartSessionUC creates a new StartSessionUC.
func NewStartSessionUC(
sessionRepo ports.SessionRepository,
workoutRepo ports.WorkoutRepository,
auditLogRepo ports.AuditLogRepository,
) *StartSessionUC {
return &StartSessionUC{
sessionRepo:  sessionRepo,
workoutRepo:  workoutRepo,
auditLogRepo: auditLogRepo,
}
}

// Execute runs the start session use case.
func (uc *StartSessionUC) Execute(ctx context.Context, input StartSessionInput) (StartSessionOutput, error) {
if input.WorkoutID == uuid.Nil {
return StartSessionOutput{}, domainerrors.ErrMalformedParameters
}

exists, err := uc.workoutRepo.ExistsByIDAndUserID(ctx, input.WorkoutID, input.UserID)
if err != nil {
return StartSessionOutput{}, fmt.Errorf("failed to check workout ownership: %w", err)
}
if !exists {
return StartSessionOutput{}, fmt.Errorf("%w: workout with id '%s' was not found", domainerrors.ErrWorkoutNotFound, input.WorkoutID)
}

activeSession, err := uc.sessionRepo.FindActiveByUserID(ctx, input.UserID)
if err != nil {
return StartSessionOutput{}, fmt.Errorf("failed to check active session: %w", err)
}
if activeSession != nil {
return StartSessionOutput{}, domainerrors.ErrActiveSessionExists
}

now := time.Now()
session := entities.Session{
ID:        uuid.New(),
UserID:    input.UserID,
WorkoutID: input.WorkoutID,
StartedAt: now,
Status:    vos.SessionStatusActive,
Notes:     "",
CreatedAt: now,
UpdatedAt: now,
}

if err := uc.sessionRepo.Create(ctx, &session); err != nil {
return StartSessionOutput{}, fmt.Errorf("failed to create session: %w", err)
}

actionData, _ := json.Marshal(map[string]string{
"workoutId": session.WorkoutID.String(),
"startedAt": session.StartedAt.Format(time.RFC3339),
})
auditEntry := &entities.AuditLog{
ID:         uuid.New(),
UserID:     session.UserID,
EntityType: "session",
EntityID:   session.ID,
Action:     "created",
ActionData: actionData,
OccurredAt: now,
}
_ = uc.auditLogRepo.Append(ctx, auditEntry)

return StartSessionOutput{Session: session}, nil
}
