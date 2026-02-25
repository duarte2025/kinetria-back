package sessions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

// mockSessionRepository is a mock implementation of ports.SessionRepository for testing.
type mockSessionRepository struct {
	createErr          error
	findActiveErr      error
	findActiveResponse *entities.Session
}

func (m *mockSessionRepository) Create(ctx context.Context, session *entities.Session) error {
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}

func (m *mockSessionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error) {
	if m.findActiveErr != nil {
		return nil, m.findActiveErr
	}
	return m.findActiveResponse, nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error) {
	return nil, nil
}

func (m *mockSessionRepository) UpdateStatus(ctx context.Context, sessionID uuid.UUID, status string, finishedAt *time.Time, notes string) error {
	return nil
}

// mockWorkoutRepository is a mock implementation of ports.WorkoutRepository for testing.
type mockWorkoutRepository struct {
	existsResponse bool
	existsErr      error
}

func (m *mockWorkoutRepository) ExistsByIDAndUserID(ctx context.Context, workoutID, userID uuid.UUID) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	return m.existsResponse, nil
}

func (m *mockWorkoutRepository) ListByUserID(_ context.Context, _ uuid.UUID, _, _ int) ([]entities.Workout, int, error) {
	return nil, 0, nil
}

// mockAuditLogRepository is a mock implementation of ports.AuditLogRepository for testing.
type mockAuditLogRepository struct {
	appendErr    error
	appendCalled bool
}

func (m *mockAuditLogRepository) Append(ctx context.Context, entry *entities.AuditLog) error {
	m.appendCalled = true
	if m.appendErr != nil {
		return m.appendErr
	}
	return nil
}

func TestStartSessionUC_Execute(t *testing.T) {
	userID := uuid.New()
	workoutID := uuid.New()

	tests := []struct {
		name        string
		input       sessions.StartSessionInput
		setupMocks  func(*mockSessionRepository, *mockWorkoutRepository, *mockAuditLogRepository)
		wantErr     error
		checkOutput func(t *testing.T, out sessions.StartSessionOutput, auditRepo *mockAuditLogRepository)
	}{
		{
			name: "success - creates session when no active session exists",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: workoutID,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {
				wr.existsResponse = true // workout exists
				sr.findActiveResponse = nil // no active session
			},
			wantErr: nil,
			checkOutput: func(t *testing.T, out sessions.StartSessionOutput, auditRepo *mockAuditLogRepository) {
				if out.Session.ID == uuid.Nil {
					t.Error("Session ID should not be nil")
				}
				if out.Session.UserID != userID {
					t.Errorf("Session UserID = %v, want %v", out.Session.UserID, userID)
				}
				if out.Session.WorkoutID != workoutID {
					t.Errorf("Session WorkoutID = %v, want %v", out.Session.WorkoutID, workoutID)
				}
				if out.Session.Status != vos.SessionStatusActive {
					t.Errorf("Session Status = %v, want %v", out.Session.Status, vos.SessionStatusActive)
				}
				if out.Session.StartedAt.IsZero() {
					t.Error("Session StartedAt should not be zero")
				}
				if out.Session.CreatedAt.IsZero() {
					t.Error("Session CreatedAt should not be zero")
				}
				if out.Session.UpdatedAt.IsZero() {
					t.Error("Session UpdatedAt should not be zero")
				}
				if !auditRepo.appendCalled {
					t.Error("Audit log should be called on success")
				}
			},
		},
		{
			name: "error - workoutID is uuid.Nil",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: uuid.Nil,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {},
			wantErr:    domainerrors.ErrMalformedParameters,
		},
		{
			name: "error - workout not found (ExistsByIDAndUserID returns false)",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: workoutID,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {
				wr.existsResponse = false // workout doesn't exist
			},
			wantErr: domainerrors.ErrWorkoutNotFound,
		},
		{
			name: "error - workout repo error",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: workoutID,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {
				wr.existsErr = errors.New("database connection error")
			},
			wantErr: nil, // we check error wrapping below
			checkOutput: func(t *testing.T, out sessions.StartSessionOutput, auditRepo *mockAuditLogRepository) {
				// This test case will fail in the main test logic
				// Just ensuring we don't panic on error
			},
		},
		{
			name: "error - active session exists",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: workoutID,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {
				wr.existsResponse = true // workout exists
				// active session exists
				sr.findActiveResponse = &entities.Session{
					ID:        uuid.New(),
					UserID:    userID,
					WorkoutID: workoutID,
					Status:    vos.SessionStatusActive,
					StartedAt: time.Now(),
				}
			},
			wantErr: domainerrors.ErrActiveSessionExists,
		},
		{
			name: "error - session repo FindActive error",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: workoutID,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {
				wr.existsResponse = true // workout exists
				sr.findActiveErr = errors.New("database query failed")
			},
			wantErr: nil, // we check error wrapping below
			checkOutput: func(t *testing.T, out sessions.StartSessionOutput, auditRepo *mockAuditLogRepository) {
				// This test case will fail in the main test logic
			},
		},
		{
			name: "error - session Create error",
			input: sessions.StartSessionInput{
				UserID:    userID,
				WorkoutID: workoutID,
			},
			setupMocks: func(sr *mockSessionRepository, wr *mockWorkoutRepository, ar *mockAuditLogRepository) {
				wr.existsResponse = true    // workout exists
				sr.findActiveResponse = nil // no active session
				sr.createErr = errors.New("insert failed")
			},
			wantErr: nil, // we check error wrapping below
			checkOutput: func(t *testing.T, out sessions.StartSessionOutput, auditRepo *mockAuditLogRepository) {
				// This test case will fail in the main test logic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := &mockSessionRepository{}
			workoutRepo := &mockWorkoutRepository{}
			auditRepo := &mockAuditLogRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(sessionRepo, workoutRepo, auditRepo)
			}

			uc := sessions.NewStartSessionUC(sessionRepo, workoutRepo, auditRepo)
			out, err := uc.Execute(context.Background(), tt.input)

			// Special handling for wrapped errors
			if tt.name == "error - workout repo error" {
				if err == nil {
					t.Error("Execute() should return an error when workout repo fails")
					return
				}
				if !errors.Is(err, workoutRepo.existsErr) {
					// Check that error is wrapped
					unwrapped := errors.Unwrap(err)
					if unwrapped == nil {
						t.Error("Execute() error should wrap the underlying error")
					}
				}
				return
			}

			if tt.name == "error - session repo FindActive error" {
				if err == nil {
					t.Error("Execute() should return an error when FindActive fails")
					return
				}
				// Check that error is wrapped
				unwrapped := errors.Unwrap(err)
				if unwrapped == nil {
					t.Error("Execute() error should wrap the underlying error")
				}
				return
			}

			if tt.name == "error - session Create error" {
				if err == nil {
					t.Error("Execute() should return an error when Create fails")
					return
				}
				// Check that error is wrapped
				unwrapped := errors.Unwrap(err)
				if unwrapped == nil {
					t.Error("Execute() error should wrap the underlying error")
				}
				// Verify audit log was NOT called on failure
				if auditRepo.appendCalled {
					t.Error("Audit log should not be called when session creation fails")
				}
				return
			}

			// Standard error checking for domain errors
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// Success case
			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, out, auditRepo)
			}
		})
	}
}
