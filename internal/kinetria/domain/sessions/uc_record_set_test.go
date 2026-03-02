package sessions_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

func TestRecordSetUC_Execute(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	workoutID := uuid.New()
	exerciseID := uuid.New()

	tests := []struct {
		name          string
		input         sessions.RecordSetInput
		mockSetup     func(*mockSessionRepo, *mockSetRecordRepo, *mockExerciseRepo)
		expectedError error
	}{
		{
			name: "success - records set",
			input: sessions.RecordSetInput{
				UserID:     userID,
				SessionID:  sessionID,
				ExerciseID: exerciseID,
				SetNumber:  1,
				Weight:     82500,
				Reps:       10,
				Status:     vos.SetRecordStatusCompleted,
			},
			mockSetup: func(sr *mockSessionRepo, srr *mockSetRecordRepo, er *mockExerciseRepo) {
				sr.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return &entities.Session{
						ID:        sessionID,
						UserID:    userID,
						WorkoutID: workoutID,
						Status:    vos.SessionStatusActive,
					}, nil
				}
				er.findWorkoutExerciseID = func(ctx context.Context, eid, wid uuid.UUID) (uuid.UUID, error) {
					return uuid.New(), nil
				}
				srr.findBySessionExerciseSet = func(ctx context.Context, sid, eid uuid.UUID, setNum int) (*entities.SetRecord, error) {
					return nil, sql.ErrNoRows
				}
				srr.create = func(ctx context.Context, sr *entities.SetRecord) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - session not found",
			input: sessions.RecordSetInput{
				UserID:     userID,
				SessionID:  sessionID,
				ExerciseID: exerciseID,
				SetNumber:  1,
				Weight:     82500,
				Reps:       10,
				Status:     vos.SetRecordStatusCompleted,
			},
			mockSetup: func(sr *mockSessionRepo, srr *mockSetRecordRepo, er *mockExerciseRepo) {
				sr.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: domainerrors.ErrNotFound,
		},
		{
			name: "error - session not active",
			input: sessions.RecordSetInput{
				UserID:     userID,
				SessionID:  sessionID,
				ExerciseID: exerciseID,
				SetNumber:  1,
				Weight:     82500,
				Reps:       10,
				Status:     vos.SetRecordStatusCompleted,
			},
			mockSetup: func(sr *mockSessionRepo, srr *mockSetRecordRepo, er *mockExerciseRepo) {
				sr.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return &entities.Session{
						ID:        sessionID,
						UserID:    userID,
						WorkoutID: workoutID,
						Status:    vos.SessionStatusCompleted,
					}, nil
				}
			},
			expectedError: domainerrors.ErrSessionNotActive,
		},
		{
			name: "error - exercise not found",
			input: sessions.RecordSetInput{
				UserID:     userID,
				SessionID:  sessionID,
				ExerciseID: exerciseID,
				SetNumber:  1,
				Weight:     82500,
				Reps:       10,
				Status:     vos.SetRecordStatusCompleted,
			},
			mockSetup: func(sr *mockSessionRepo, srr *mockSetRecordRepo, er *mockExerciseRepo) {
				sr.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return &entities.Session{
						ID:        sessionID,
						UserID:    userID,
						WorkoutID: workoutID,
						Status:    vos.SessionStatusActive,
					}, nil
				}
				er.findWorkoutExerciseID = func(ctx context.Context, eid, wid uuid.UUID) (uuid.UUID, error) {
					return uuid.Nil, sql.ErrNoRows
				}
			},
			expectedError: domainerrors.ErrExerciseNotFound,
		},
		{
			name: "error - set already recorded",
			input: sessions.RecordSetInput{
				UserID:     userID,
				SessionID:  sessionID,
				ExerciseID: exerciseID,
				SetNumber:  1,
				Weight:     82500,
				Reps:       10,
				Status:     vos.SetRecordStatusCompleted,
			},
			mockSetup: func(sr *mockSessionRepo, srr *mockSetRecordRepo, er *mockExerciseRepo) {
				sr.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return &entities.Session{
						ID:        sessionID,
						UserID:    userID,
						WorkoutID: workoutID,
						Status:    vos.SessionStatusActive,
					}, nil
				}
				er.findWorkoutExerciseID = func(ctx context.Context, eid, wid uuid.UUID) (uuid.UUID, error) {
					return uuid.New(), nil
				}
				srr.findBySessionExerciseSet = func(ctx context.Context, sid, eid uuid.UUID, setNum int) (*entities.SetRecord, error) {
					return &entities.SetRecord{ID: uuid.New()}, nil
				}
			},
			expectedError: domainerrors.ErrSetAlreadyRecorded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := &mockSessionRepo{}
			setRecordRepo := &mockSetRecordRepo{}
			exerciseRepo := &mockExerciseRepo{}
			auditRepo := &mockAuditRepo{append: func(ctx context.Context, entry *entities.AuditLog) error { return nil }}

			tt.mockSetup(sessionRepo, setRecordRepo, exerciseRepo)

			uc := sessions.NewRecordSetUseCase(sessionRepo, setRecordRepo, exerciseRepo, auditRepo)
			output, err := uc.Execute(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil || !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output.SetRecord.ID == uuid.Nil {
					t.Error("expected set record to be created")
				}
			}
		})
	}
}

// Mock repositories
type mockSessionRepo struct {
	findByID func(context.Context, uuid.UUID) (*entities.Session, error)
}

func (m *mockSessionRepo) Create(ctx context.Context, session *entities.Session) error {
	return nil
}

func (m *mockSessionRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error) {
	if m.findByID != nil {
		return m.findByID(ctx, sessionID)
	}
	return nil, nil
}

func (m *mockSessionRepo) UpdateStatus(ctx context.Context, sessionID uuid.UUID, status string, finishedAt *time.Time, notes string) (bool, error) {
	return true, nil
}

func (m *mockSessionRepo) GetCompletedSessionsByUserAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]entities.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) GetStatsByUserAndPeriod(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SessionStats, error) {
	return &ports.SessionStats{}, nil
}

func (m *mockSessionRepo) GetFrequencyByUserAndPeriod(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]ports.FrequencyData, error) {
	return nil, nil
}

func (m *mockSessionRepo) GetSessionsForStreak(_ context.Context, _ uuid.UUID) ([]time.Time, error) {
	return nil, nil
}

type mockSetRecordRepo struct {
	create                   func(context.Context, *entities.SetRecord) error
	findBySessionExerciseSet func(context.Context, uuid.UUID, uuid.UUID, int) (*entities.SetRecord, error)
}

func (m *mockSetRecordRepo) Create(ctx context.Context, setRecord *entities.SetRecord) error {
	if m.create != nil {
		return m.create(ctx, setRecord)
	}
	return nil
}

func (m *mockSetRecordRepo) FindBySessionExerciseSet(ctx context.Context, sessionID, exerciseID uuid.UUID, setNumber int) (*entities.SetRecord, error) {
	if m.findBySessionExerciseSet != nil {
		return m.findBySessionExerciseSet(ctx, sessionID, exerciseID, setNumber)
	}
	return nil, nil
}

func (m *mockSetRecordRepo) GetTotalSetsRepsVolume(_ context.Context, _ uuid.UUID, _, _ time.Time) (*ports.SetRecordStats, error) {
	return &ports.SetRecordStats{}, nil
}

func (m *mockSetRecordRepo) GetPersonalRecordsByUser(_ context.Context, _ uuid.UUID) ([]ports.PersonalRecord, error) {
	return nil, nil
}

func (m *mockSetRecordRepo) GetProgressionByUserAndExercise(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _, _ time.Time) ([]ports.ProgressionPoint, error) {
	return nil, nil
}

type mockExerciseRepo struct {
	existsByIDAndWorkoutID func(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	findWorkoutExerciseID  func(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
}

func (m *mockExerciseRepo) ExistsByIDAndWorkoutID(ctx context.Context, exerciseID, workoutID uuid.UUID) (bool, error) {
	if m.existsByIDAndWorkoutID != nil {
		return m.existsByIDAndWorkoutID(ctx, exerciseID, workoutID)
	}
	return false, nil
}

func (m *mockExerciseRepo) FindWorkoutExerciseID(ctx context.Context, exerciseID, workoutID uuid.UUID) (uuid.UUID, error) {
	if m.findWorkoutExerciseID != nil {
		return m.findWorkoutExerciseID(ctx, exerciseID, workoutID)
	}
	return uuid.New(), nil
}

func (m *mockExerciseRepo) List(_ context.Context, _ ports.ExerciseFilters, _, _ int) ([]*entities.Exercise, int, error) {
	return nil, 0, nil
}

func (m *mockExerciseRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Exercise, error) {
	return nil, nil
}

func (m *mockExerciseRepo) GetUserStats(_ context.Context, _, _ uuid.UUID) (*ports.ExerciseUserStats, error) {
	return nil, nil
}

func (m *mockExerciseRepo) GetHistory(_ context.Context, _, _ uuid.UUID, _, _ int) ([]*ports.ExerciseHistoryEntry, int, error) {
	return nil, 0, nil
}

type mockAuditRepo struct {
	append func(context.Context, *entities.AuditLog) error
}

func (m *mockAuditRepo) Append(ctx context.Context, entry *entities.AuditLog) error {
	if m.append != nil {
		return m.append(ctx, entry)
	}
	return nil
}
