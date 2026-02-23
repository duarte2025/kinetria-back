package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
	"github.com/google/uuid"
)

func TestLogoutUC_Execute(t *testing.T) {
	tests := []struct {
		name       string
		plainToken string
		setupMock  func(*mockRefreshTokenRepo)
		wantErr    error
	}{
		{
			name:       "success - revoke existing token",
			plainToken: "valid-logout-token",
			setupMock: func(r *mockRefreshTokenRepo) {
				hash := gatewayauth.HashToken("valid-logout-token")
				r.tokens[hash] = &entities.RefreshToken{
					ID:        uuid.New(),
					Token:     hash,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					CreatedAt: time.Now(),
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - idempotent (token already revoked)",
			plainToken: "already-revoked-token",
			setupMock: func(r *mockRefreshTokenRepo) {
				now := time.Now()
				hash := gatewayauth.HashToken("already-revoked-token")
				r.tokens[hash] = &entities.RefreshToken{
					ID:        uuid.New(),
					Token:     hash,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					RevokedAt: &now,
					CreatedAt: time.Now(),
				}
			},
			wantErr: nil,
		},
		{
			name:       "success - idempotent (token not found)",
			plainToken: "nonexistent-token",
			setupMock:  func(r *mockRefreshTokenRepo) {},
			wantErr:    nil, // LogoutUC should be idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshTokenRepo := &mockRefreshTokenRepo{tokens: make(map[string]*entities.RefreshToken)}
			if tt.setupMock != nil {
				tt.setupMock(refreshTokenRepo)
			}

			uc := domainauth.NewLogoutUC(refreshTokenRepo)
			_, err := uc.Execute(context.Background(), domainauth.LogoutInput{
				RefreshToken: tt.plainToken,
			})

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogoutUC_RevokesToken(t *testing.T) {
	plainToken := "token-to-revoke"
	hash := gatewayauth.HashToken(plainToken)
	refreshTokenRepo := &mockRefreshTokenRepo{
		tokens: map[string]*entities.RefreshToken{
			hash: {
				ID:        uuid.New(),
				Token:     hash,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
		},
	}

	uc := domainauth.NewLogoutUC(refreshTokenRepo)
	_, err := uc.Execute(context.Background(), domainauth.LogoutInput{RefreshToken: plainToken})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if refreshTokenRepo.tokens[hash].RevokedAt == nil {
		t.Error("token should be revoked after logout")
	}
}
