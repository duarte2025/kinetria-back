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
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

func newJWTManager() *gatewayauth.JWTManager {
	return gatewayauth.NewJWTManager("test-secret-key-that-is-at-least-32bytes", time.Hour)
}

func TestRegisterUC_Execute(t *testing.T) {
	tests := []struct {
		name        string
		input       domainauth.RegisterInput
		setupMocks  func(*mockUserRepo)
		wantErr     error
		checkOutput func(t *testing.T, out domainauth.RegisterOutput)
	}{
		{
			name: "success - valid registration",
			input: domainauth.RegisterInput{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(r *mockUserRepo) {},
			wantErr:    nil,
			checkOutput: func(t *testing.T, out domainauth.RegisterOutput) {
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
			name: "error - email already exists",
			input: domainauth.RegisterInput{
				Name:     "Jane Doe",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(r *mockUserRepo) {
				r.users["existing@example.com"] = &entities.User{
					ID:    uuid.New(),
					Email: "existing@example.com",
				}
			},
			wantErr: domainerrors.ErrEmailAlreadyExists,
		},
		{
			name: "error - password too short",
			input: domainauth.RegisterInput{
				Name:     "Short Pass",
				Email:    "short@example.com",
				Password: "short",
			},
			setupMocks: func(r *mockUserRepo) {},
			wantErr:    domainerrors.ErrMalformedParameters,
		},
		{
			name: "error - password exactly 7 chars (too short)",
			input: domainauth.RegisterInput{
				Name:     "Almost",
				Email:    "almost@example.com",
				Password: "1234567",
			},
			setupMocks: func(r *mockUserRepo) {},
			wantErr:    domainerrors.ErrMalformedParameters,
		},
		{
			name: "success - password exactly 8 chars (minimum)",
			input: domainauth.RegisterInput{
				Name:     "Min Pass",
				Email:    "minpass@example.com",
				Password: "12345678",
			},
			setupMocks: func(r *mockUserRepo) {},
			wantErr:    nil,
			checkOutput: func(t *testing.T, out domainauth.RegisterOutput) {
				if out.AccessToken == "" {
					t.Error("AccessToken should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{users: make(map[string]*entities.User)}
			refreshTokenRepo := &mockRefreshTokenRepo{tokens: make(map[string]*entities.RefreshToken)}
			jwtMgr := newJWTManager()

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			uc := domainauth.NewRegisterUC(userRepo, refreshTokenRepo, jwtMgr, 720*time.Hour)
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

func TestRegisterUC_PasswordIsHashed(t *testing.T) {
	userRepo := &mockUserRepo{users: make(map[string]*entities.User)}
	refreshTokenRepo := &mockRefreshTokenRepo{tokens: make(map[string]*entities.RefreshToken)}
	uc := domainauth.NewRegisterUC(userRepo, refreshTokenRepo, newJWTManager(), 720*time.Hour)

	_, err := uc.Execute(context.Background(), domainauth.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "plainpassword",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	storedUser := userRepo.users["test@example.com"]
	if storedUser == nil {
		t.Fatal("user not stored")
	}
	if storedUser.PasswordHash == "plainpassword" {
		t.Error("password should be hashed, not stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.PasswordHash), []byte("plainpassword")); err != nil {
		t.Errorf("stored hash does not match original password: %v", err)
	}
}
