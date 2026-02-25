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

// RecordSetInput represents input for recording a set.
type RecordSetInput struct {
	UserID     uuid.UUID
	SessionID  uuid.UUID
	ExerciseID uuid.UUID
	SetNumber  int
	Weight     int // grams
	Reps       int
	Status     vos.SetRecordStatus
}

// RecordSetOutput represents output after recording a set.
type RecordSetOutput struct {
	SetRecord entities.SetRecord
}

// RecordSetUseCase orchestrates recording a set during an active session.
type RecordSetUseCase struct {
	sessionRepo   ports.SessionRepository
	setRecordRepo ports.SetRecordRepository
	exerciseRepo  ports.ExerciseRepository
	auditLogRepo  ports.AuditLogRepository
}

// NewRecordSetUseCase creates a new instance of RecordSetUseCase.
func NewRecordSetUseCase(
	sessionRepo ports.SessionRepository,
	setRecordRepo ports.SetRecordRepository,
	exerciseRepo ports.ExerciseRepository,
	auditLogRepo ports.AuditLogRepository,
) *RecordSetUseCase {
	return &RecordSetUseCase{
		sessionRepo:   sessionRepo,
		setRecordRepo: setRecordRepo,
		exerciseRepo:  exerciseRepo,
		auditLogRepo:  auditLogRepo,
	}
}

// Execute records a set for an active session.
func (uc *RecordSetUseCase) Execute(ctx context.Context, input RecordSetInput) (RecordSetOutput, error) {
	// Validate inputs
	if input.SessionID == uuid.Nil || input.ExerciseID == uuid.Nil {
		return RecordSetOutput{}, errors.ErrMalformedParameters
	}
	if input.SetNumber < 1 || input.Weight < 0 || input.Reps < 0 {
		return RecordSetOutput{}, errors.ErrMalformedParameters
	}
	if err := input.Status.Validate(); err != nil {
		return RecordSetOutput{}, errors.ErrMalformedParameters
	}

	// Find session and validate ownership
	session, err := uc.sessionRepo.FindByID(ctx, input.SessionID)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return RecordSetOutput{}, errors.ErrNotFound
		}
		return RecordSetOutput{}, fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil || session.UserID != input.UserID {
		return RecordSetOutput{}, errors.ErrNotFound
	}

	// Validate session is active
	if session.Status != vos.SessionStatusActive {
		return RecordSetOutput{}, errors.ErrSessionNotActive
	}

	// Validate exercise belongs to workout
	exists, err := uc.exerciseRepo.ExistsByIDAndWorkoutID(ctx, input.ExerciseID, session.WorkoutID)
	if err != nil {
		return RecordSetOutput{}, fmt.Errorf("failed to check exercise: %w", err)
	}
	if !exists {
		return RecordSetOutput{}, errors.ErrExerciseNotFound
	}

	// Check for duplicate
	existing, err := uc.setRecordRepo.FindBySessionExerciseSet(ctx, input.SessionID, input.ExerciseID, input.SetNumber)
	if err != nil && err != sql.ErrNoRows {
		return RecordSetOutput{}, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if existing != nil {
		return RecordSetOutput{}, errors.ErrSetAlreadyRecorded
	}

	// Create SetRecord
	now := time.Now()
	setRecord := entities.SetRecord{
		ID:         uuid.New(),
		SessionID:  input.SessionID,
		ExerciseID: input.ExerciseID,
		SetNumber:  input.SetNumber,
		Weight:     input.Weight,
		Reps:       input.Reps,
		Status:     input.Status.String(),
		RecordedAt: now,
	}

	// Persist
	if err := uc.setRecordRepo.Create(ctx, &setRecord); err != nil {
		return RecordSetOutput{}, fmt.Errorf("failed to create set record: %w", err)
	}

	// Audit log
	actionData, _ := json.Marshal(setRecord)
	auditEntry := entities.AuditLog{
		ID:         uuid.New(),
		UserID:     input.UserID,
		EntityType: "set_record",
		EntityID:   setRecord.ID,
		Action:     "created",
		ActionData: actionData,
		OccurredAt: now,
	}
	_ = uc.auditLogRepo.Append(ctx, &auditEntry)

	return RecordSetOutput{SetRecord: setRecord}, nil
}
