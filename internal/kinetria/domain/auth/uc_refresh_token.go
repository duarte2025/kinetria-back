package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// RefreshTokenInput holds the refresh token for renewal.
type RefreshTokenInput struct {
	RefreshToken string
}

// RefreshTokenOutput holds the new tokens after a successful refresh.
type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// RefreshTokenUC implements the use case for renewing an access token.
type RefreshTokenUC struct {
	refreshTokenRepo ports.RefreshTokenRepository
	jwtManager       *gatewayauth.JWTManager
	tokenExpiry      time.Duration
}

// NewRefreshTokenUC creates a new RefreshTokenUC with all required dependencies.
func NewRefreshTokenUC(
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtManager *gatewayauth.JWTManager,
	tokenExpiry time.Duration,
) *RefreshTokenUC {
	return &RefreshTokenUC{
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		tokenExpiry:      tokenExpiry,
	}
}

// Execute validates the refresh token, revokes it, and issues new tokens (rotation).
// Returns ErrTokenInvalid if the token does not exist.
// Returns ErrTokenRevoked if the token has been revoked.
// Returns ErrTokenExpired if the token has expired.
func (uc *RefreshTokenUC) Execute(ctx context.Context, input RefreshTokenInput) (RefreshTokenOutput, error) {
	tokenHash := gatewayauth.HashToken(input.RefreshToken)

	storedToken, err := uc.refreshTokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return RefreshTokenOutput{}, err
	}

	if storedToken.RevokedAt != nil {
		return RefreshTokenOutput{}, domainerrors.ErrTokenRevoked
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return RefreshTokenOutput{}, domainerrors.ErrTokenExpired
	}

	if err := uc.refreshTokenRepo.RevokeByToken(ctx, tokenHash); err != nil {
		return RefreshTokenOutput{}, err
	}

	newRefreshTokenPlain, err := gatewayauth.GenerateRefreshToken()
	if err != nil {
		return RefreshTokenOutput{}, err
	}

	now := time.Now()
	newRefreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    storedToken.UserID,
		Token:     gatewayauth.HashToken(newRefreshTokenPlain),
		ExpiresAt: now.Add(uc.tokenExpiry),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return RefreshTokenOutput{}, err
	}

	accessToken, err := uc.jwtManager.GenerateToken(storedToken.UserID)
	if err != nil {
		return RefreshTokenOutput{}, err
	}

	return RefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenPlain,
		ExpiresIn:    3600,
	}, nil
}
