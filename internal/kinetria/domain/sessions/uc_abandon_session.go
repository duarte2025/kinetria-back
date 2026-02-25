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

// AbandonSessionInput represents input for abandoning a session.
type AbandonSessionInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// AbandonSessionOutput represents output after abandoning a session.
type AbandonSessionOutput struct {
	Session entities.Session
}

// AbandonSessionUseCase orchestrates abandoning an active session.
type AbandonSessionUseCase struct {
	sessionRepo  ports.SessionRepository
	auditLogRepo ports.AuditLogRepository
}

// NewAbandonSessionUseCase creates a new instance of AbandonSessionUseCase.
func NewAbandonSessionUseCase(
	sessionRepo ports.SessionRepository,
	auditLogRepo ports.AuditLogRepository,
) *AbandonSessionUseCase {
	return &AbandonSessionUseCase{
		sessionRepo:  sessionRepo,
		auditLogRepo: auditLogRepo,
	}
}

// Execute abandons an active session.
func (uc *AbandonSessionUseCase) Execute(ctx context.Context, input AbandonSessionInput) (AbandonSessionOutput, error) {
	// Validate input
	if input.SessionID == uuid.Nil {
		return AbandonSessionOutput{}, errors.ErrMalformedParameters
	}

	// Find session and validate ownership
	session, err := uc.sessionRepo.FindByID(ctx, input.SessionID)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return AbandonSessionOutput{}, errors.ErrNotFound
		}
		return AbandonSessionOutput{}, fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil || session.UserID != input.UserID {
		return AbandonSessionOutput{}, errors.ErrNotFound
	}

	// Validate session is active
	if session.Status != vos.SessionStatusActive {
		return AbandonSessionOutput{}, errors.ErrSessionAlreadyClosed
	}

	// Update session
	now := time.Now()
	updated, err := uc.sessionRepo.UpdateStatus(ctx, input.SessionID, vos.SessionStatusAbandoned.String(), &now, "")
	if err != nil {
		return AbandonSessionOutput{}, fmt.Errorf("failed to update session: %w", err)
	}
	if !updated {
		return AbandonSessionOutput{}, errors.ErrSessionAlreadyClosed
	}

	// Update local entity
	session.Status = vos.SessionStatusAbandoned
	session.FinishedAt = &now
	session.UpdatedAt = now

	// Audit log
	actionData, _ := json.Marshal(map[string]interface{}{
		"finishedAt": now,
	})
	auditEntry := entities.AuditLog{
		ID:         uuid.New(),
		UserID:     input.UserID,
		EntityType: "session",
		EntityID:   session.ID,
		Action:     "abandoned",
		ActionData: actionData,
		OccurredAt: now,
	}
	_ = uc.auditLogRepo.Append(ctx, &auditEntry)

	return AbandonSessionOutput{Session: *session}, nil
}
