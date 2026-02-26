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
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/sessions"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

func TestAbandonSessionUC_Execute(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	workoutID := uuid.New()

	activeSession := &entities.Session{
		ID:        sessionID,
		UserID:    userID,
		WorkoutID: workoutID,
		Status:    vos.SessionStatusActive,
		StartedAt: time.Now(),
	}

	tests := []struct {
		name          string
		input         sessions.AbandonSessionInput
		mockSetup     func(*mockAbandonSessionRepo)
		expectedError error
	}{
		{
			name: "success - abandons active session",
			input: sessions.AbandonSessionInput{
				UserID:    userID,
				SessionID: sessionID,
			},
			mockSetup: func(r *mockAbandonSessionRepo) {
				r.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return activeSession, nil
				}
				r.updateStatus = func(ctx context.Context, id uuid.UUID, status string, finishedAt *time.Time, notes string) (bool, error) {
					return true, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - malformed parameters (nil sessionID)",
			input: sessions.AbandonSessionInput{
				UserID:    userID,
				SessionID: uuid.Nil,
			},
			mockSetup:     func(r *mockAbandonSessionRepo) {},
			expectedError: domainerrors.ErrMalformedParameters,
		},
		{
			name: "error - session not found (sql.ErrNoRows)",
			input: sessions.AbandonSessionInput{
				UserID:    userID,
				SessionID: sessionID,
			},
			mockSetup: func(r *mockAbandonSessionRepo) {
				r.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return nil, sql.ErrNoRows
				}
			},
			expectedError: domainerrors.ErrNotFound,
		},
		{
			name: "error - session not found (nil session)",
			input: sessions.AbandonSessionInput{
				UserID:    userID,
				SessionID: sessionID,
			},
			mockSetup: func(r *mockAbandonSessionRepo) {
				r.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return nil, nil
				}
			},
			expectedError: domainerrors.ErrNotFound,
		},
		{
			name: "error - ownership mismatch",
			input: sessions.AbandonSessionInput{
				UserID:    uuid.New(),
				SessionID: sessionID,
			},
			mockSetup: func(r *mockAbandonSessionRepo) {
				r.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return activeSession, nil
				}
			},
			expectedError: domainerrors.ErrNotFound,
		},
		{
			name: "error - session already closed",
			input: sessions.AbandonSessionInput{
				UserID:    userID,
				SessionID: sessionID,
			},
			mockSetup: func(r *mockAbandonSessionRepo) {
				r.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return &entities.Session{
						ID:        sessionID,
						UserID:    userID,
						WorkoutID: workoutID,
						Status:    vos.SessionStatusAbandoned,
					}, nil
				}
			},
			expectedError: domainerrors.ErrSessionAlreadyClosed,
		},
		{
			name: "error - concurrent close (update returned 0 rows)",
			input: sessions.AbandonSessionInput{
				UserID:    userID,
				SessionID: sessionID,
			},
			mockSetup: func(r *mockAbandonSessionRepo) {
				r.findByID = func(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
					return activeSession, nil
				}
				r.updateStatus = func(ctx context.Context, id uuid.UUID, status string, finishedAt *time.Time, notes string) (bool, error) {
					return false, nil
				}
			},
			expectedError: domainerrors.ErrSessionAlreadyClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAbandonSessionRepo{}
			auditRepo := &mockAuditRepo{append: func(ctx context.Context, entry *entities.AuditLog) error { return nil }}

			tt.mockSetup(repo)

			uc := sessions.NewAbandonSessionUseCase(repo, auditRepo)
			output, err := uc.Execute(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil || !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output.Session.ID == uuid.Nil {
					t.Error("expected session to be populated")
				}
			}
		})
	}
}

// mockAbandonSessionRepo is a mock SessionRepository for AbandonSession tests.
type mockAbandonSessionRepo struct {
	findByID     func(context.Context, uuid.UUID) (*entities.Session, error)
	updateStatus func(context.Context, uuid.UUID, string, *time.Time, string) (bool, error)
}

func (m *mockAbandonSessionRepo) Create(ctx context.Context, session *entities.Session) error {
	return nil
}

func (m *mockAbandonSessionRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Session, error) {
	return nil, nil
}

func (m *mockAbandonSessionRepo) FindByID(ctx context.Context, sessionID uuid.UUID) (*entities.Session, error) {
	if m.findByID != nil {
		return m.findByID(ctx, sessionID)
	}
	return nil, nil
}

func (m *mockAbandonSessionRepo) UpdateStatus(ctx context.Context, sessionID uuid.UUID, status string, finishedAt *time.Time, notes string) (bool, error) {
	if m.updateStatus != nil {
		return m.updateStatus(ctx, sessionID, status, finishedAt, notes)
	}
	return true, nil
}

func (m *mockAbandonSessionRepo) GetCompletedSessionsByUserAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]entities.Session, error) {
	return nil, nil
}
