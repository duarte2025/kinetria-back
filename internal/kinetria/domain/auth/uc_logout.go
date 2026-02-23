package auth

import (
	"context"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// LogoutInput holds the refresh token to revoke on logout.
type LogoutInput struct {
	RefreshToken string
}

// LogoutOutput is empty since logout returns 204 No Content.
type LogoutOutput struct{}

// LogoutUC implements the use case for revoking a refresh token (logout).
type LogoutUC struct {
	refreshTokenRepo ports.RefreshTokenRepository
}

// NewLogoutUC creates a new LogoutUC with all required dependencies.
func NewLogoutUC(refreshTokenRepo ports.RefreshTokenRepository) *LogoutUC {
	return &LogoutUC{refreshTokenRepo: refreshTokenRepo}
}

// Execute revokes the given refresh token.
// This operation is idempotent - it succeeds even if the token is already revoked or does not exist.
func (uc *LogoutUC) Execute(ctx context.Context, input LogoutInput) (LogoutOutput, error) {
	tokenHash := gatewayauth.HashToken(input.RefreshToken)
	if err := uc.refreshTokenRepo.RevokeByToken(ctx, tokenHash); err != nil {
		return LogoutOutput{}, err
	}
	return LogoutOutput{}, nil
}
