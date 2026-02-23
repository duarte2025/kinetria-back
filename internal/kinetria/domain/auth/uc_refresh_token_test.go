package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

func makeRefreshToken(plainToken string, userID uuid.UUID, expiresAt time.Time, revokedAt *time.Time) (*entities.RefreshToken, string) {
	hash := gatewayauth.HashToken(plainToken)
	return &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     hash,
		ExpiresAt: expiresAt,
		RevokedAt: revokedAt,
		CreatedAt: time.Now(),
	}, hash
}

func TestRefreshTokenUC_Execute(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		plainToken  string
		setupMock   func(*mockRefreshTokenRepo, string)
		wantErr     error
		checkOutput func(t *testing.T, out domainauth.RefreshTokenOutput, oldPlainToken string)
	}{
		{
			name:       "success - valid token",
			plainToken: "valid-plain-token",
			setupMock: func(r *mockRefreshTokenRepo, hash string) {
				tok, _ := makeRefreshToken("valid-plain-token", userID, time.Now().Add(24*time.Hour), nil)
				r.tokens[tok.Token] = tok
			},
			wantErr: nil,
			checkOutput: func(t *testing.T, out domainauth.RefreshTokenOutput, oldPlainToken string) {
				if out.AccessToken == "" {
					t.Error("AccessToken should not be empty")
				}
				if out.RefreshToken == "" {
					t.Error("RefreshToken should not be empty")
				}
				if out.RefreshToken == oldPlainToken {
					t.Error("New refresh token should differ from old one")
				}
			},
		},
		{
			name:       "error - token not found",
			plainToken: "nonexistent-token",
			setupMock:  func(r *mockRefreshTokenRepo, hash string) {},
			wantErr:    domainerrors.ErrTokenInvalid,
		},
		{
			name:       "error - token revoked",
			plainToken: "revoked-token",
			setupMock: func(r *mockRefreshTokenRepo, hash string) {
				now := time.Now()
				tok, _ := makeRefreshToken("revoked-token", userID, time.Now().Add(24*time.Hour), &now)
				r.tokens[tok.Token] = tok
			},
			wantErr: domainerrors.ErrTokenRevoked,
		},
		{
			name:       "error - token expired",
			plainToken: "expired-token",
			setupMock: func(r *mockRefreshTokenRepo, hash string) {
				tok, _ := makeRefreshToken("expired-token", userID, time.Now().Add(-1*time.Hour), nil)
				r.tokens[tok.Token] = tok
			},
			wantErr: domainerrors.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshTokenRepo := &mockRefreshTokenRepo{tokens: make(map[string]*entities.RefreshToken)}
			hash := gatewayauth.HashToken(tt.plainToken)
			tt.setupMock(refreshTokenRepo, hash)

			uc := domainauth.NewRefreshTokenUC(refreshTokenRepo, newJWTManager(), 720*time.Hour)
			out, err := uc.Execute(context.Background(), domainauth.RefreshTokenInput{
				RefreshToken: tt.plainToken,
			})

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, out, tt.plainToken)
			}
		})
	}
}

func TestRefreshTokenUC_TokenRotation(t *testing.T) {
	userID := uuid.New()
	plainToken := "rotation-test-token"
	refreshTokenRepo := &mockRefreshTokenRepo{tokens: make(map[string]*entities.RefreshToken)}

	tok, _ := makeRefreshToken(plainToken, userID, time.Now().Add(24*time.Hour), nil)
	refreshTokenRepo.tokens[tok.Token] = tok

	uc := domainauth.NewRefreshTokenUC(refreshTokenRepo, newJWTManager(), 720*time.Hour)

	out, err := uc.Execute(context.Background(), domainauth.RefreshTokenInput{
		RefreshToken: plainToken,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Old token should be revoked
	oldHash := gatewayauth.HashToken(plainToken)
	oldToken := refreshTokenRepo.tokens[oldHash]
	if oldToken == nil || oldToken.RevokedAt == nil {
		t.Error("old token should be revoked after rotation")
	}

	// New token should be different
	if out.RefreshToken == plainToken {
		t.Error("new token should differ from old")
	}
}
