package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

func makeUserWithPassword(email, plainPassword string) *entities.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), 4) // low cost for tests
	return &entities.User{
		ID:           uuid.New(),
		Email:        email,
		Name:         "Test User",
		PasswordHash: string(hash),
	}
}

func TestLoginUC_Execute(t *testing.T) {
	tests := []struct {
		name        string
		input       domainauth.LoginInput
		setupMocks  func(*mockUserRepo)
		wantErr     error
		checkOutput func(t *testing.T, out domainauth.LoginOutput)
	}{
		{
			name:  "success - valid credentials",
			input: domainauth.LoginInput{Email: "user@example.com", Password: "correctpassword"},
			setupMocks: func(r *mockUserRepo) {
				r.users["user@example.com"] = makeUserWithPassword("user@example.com", "correctpassword")
			},
			wantErr: nil,
			checkOutput: func(t *testing.T, out domainauth.LoginOutput) {
				if out.AccessToken == "" {
					t.Error("AccessToken should not be empty")
				}
				if out.RefreshToken == "" {
					t.Error("RefreshToken should not be empty")
				}
				if out.ExpiresIn != 3600 {
					t.Errorf("ExpiresIn = %d, want 3600", out.ExpiresIn)
				}
			},
		},
		{
			name:       "error - user not found",
			input:      domainauth.LoginInput{Email: "nonexistent@example.com", Password: "password"},
			setupMocks: func(r *mockUserRepo) {},
			wantErr:    domainerrors.ErrInvalidCredentials,
		},
		{
			name:  "error - wrong password",
			input: domainauth.LoginInput{Email: "user@example.com", Password: "wrongpassword"},
			setupMocks: func(r *mockUserRepo) {
				r.users["user@example.com"] = makeUserWithPassword("user@example.com", "correctpassword")
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{users: make(map[string]*entities.User)}
			refreshTokenRepo := &mockRefreshTokenRepo{tokens: make(map[string]*entities.RefreshToken)}

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			uc := domainauth.NewLoginUC(userRepo, refreshTokenRepo, newJWTManager(), 720*time.Hour)
			out, err := uc.Execute(context.Background(), tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, out)
			}
		})
	}
}
