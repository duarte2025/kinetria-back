package profile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	domainprofile "github.com/kinetria/kinetria-back/internal/kinetria/domain/profile"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestGetProfileUC_Execute(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	userID := uuid.New()

	existingUser := &entities.User{
		ID:        userID,
		Email:     "alice@example.com",
		Name:      "Alice",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name        string
		setupRepo   func(*mockProfileUserRepo)
		input       domainprofile.GetProfileInput
		wantErr     error
		checkOutput func(t *testing.T, out *domainprofile.GetProfileOutput)
	}{
		{
			name: "happy path - user exists",
			setupRepo: func(r *mockProfileUserRepo) {
				r.byID[userID] = existingUser
			},
			input:   domainprofile.GetProfileInput{UserID: userID},
			wantErr: nil,
			checkOutput: func(t *testing.T, out *domainprofile.GetProfileOutput) {
				if out == nil {
					t.Fatal("output should not be nil")
				}
				if out.User == nil {
					t.Fatal("output.User should not be nil")
				}
				if out.User.ID != userID {
					t.Errorf("User.ID = %v, want %v", out.User.ID, userID)
				}
				if out.User.Email != "alice@example.com" {
					t.Errorf("User.Email = %v, want alice@example.com", out.User.Email)
				}
			},
		},
		{
			name: "user not found - GetByID returns ErrNotFound",
			setupRepo: func(r *mockProfileUserRepo) {
				r.getByIDErr = domainerrors.ErrNotFound
			},
			input:   domainprofile.GetProfileInput{UserID: userID},
			wantErr: domainerrors.ErrNotFound,
		},
		{
			name: "repository error - GetByID returns generic error",
			setupRepo: func(r *mockProfileUserRepo) {
				r.getByIDErr = errors.New("connection refused")
			},
			input:   domainprofile.GetProfileInput{UserID: userID},
			wantErr: errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProfileUserRepo{
				byID: make(map[uuid.UUID]*entities.User),
			}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			uc := domainprofile.NewGetProfileUC(tracer, repo)
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Execute() expected error %v, got nil", tt.wantErr)
					return
				}
				// For sentinel errors use errors.Is; for generic errors just check non-nil
				if errors.Is(tt.wantErr, domainerrors.ErrNotFound) {
					if !errors.Is(err, domainerrors.ErrNotFound) {
						t.Errorf("Execute() error = %v, wantErr wrapping %v", err, tt.wantErr)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, out)
			}
		})
	}
}
