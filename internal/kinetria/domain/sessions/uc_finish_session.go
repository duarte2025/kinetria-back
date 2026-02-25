package sessions

import (
	"context"
	"database/sql"
	stdErrors "errors"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

// FinishSessionInput represents input for finishing a session.
type FinishSessionInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Notes     string
}

// FinishSessionOutput represents output after finishing a session.
type FinishSessionOutput struct {
	Session entities.Session
}

// FinishSessionUseCase orchestrates finishing an active session.
type FinishSessionUseCase struct {
	sessionRepo  ports.SessionRepository
	auditLogRepo ports.AuditLogRepository
}

// NewFinishSessionUseCase creates a new instance of FinishSessionUseCase.
func NewFinishSessionUseCase(
	sessionRepo ports.SessionRepository,
	auditLogRepo ports.AuditLogRepository,
) *FinishSessionUseCase {
	return &FinishSessionUseCase{
		sessionRepo:  sessionRepo,
		auditLogRepo: auditLogRepo,
	}
}

// Execute finishes an active session.
func (uc *FinishSessionUseCase) Execute(ctx context.Context, input FinishSessionInput) (FinishSessionOutput, error) {
	// Validate input
	if input.SessionID == uuid.Nil {
		return FinishSessionOutput{}, errors.ErrMalformedParameters
	}

	// Find session and validate ownership
	session, err := uc.sessionRepo.FindByID(ctx, input.SessionID)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return FinishSessionOutput{}, errors.ErrNotFound
		}
		return FinishSessionOutput{}, fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil || session.UserID != input.UserID {
		return FinishSessionOutput{}, errors.ErrNotFound
	}

	// Validate session is active
	if session.Status != vos.SessionStatusActive {
		return FinishSessionOutput{}, errors.ErrSessionAlreadyClosed
	}

	// Update session
	now := time.Now()
	updated, err := uc.sessionRepo.UpdateStatus(ctx, input.SessionID, vos.SessionStatusCompleted.String(), &now, input.Notes)
	if err != nil {
		return FinishSessionOutput{}, fmt.Errorf("failed to update session: %w", err)
	}
	if !updated {
		return FinishSessionOutput{}, errors.ErrSessionAlreadyClosed
	}

	// Update local entity
	session.Status = vos.SessionStatusCompleted
	session.FinishedAt = &now
	session.Notes = input.Notes
	session.UpdatedAt = now

	// Audit log
	actionData, _ := json.Marshal(map[string]interface{}{
		"finishedAt": now,
		"notes":      input.Notes,
	})
	auditEntry := entities.AuditLog{
		ID:         uuid.New(),
		UserID:     input.UserID,
		EntityType: "session",
		EntityID:   session.ID,
		Action:     "completed",
		ActionData: actionData,
		OccurredAt: now,
	}
	_ = uc.auditLogRepo.Append(ctx, &auditEntry)

	return FinishSessionOutput{Session: *session}, nil
}
