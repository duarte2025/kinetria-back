package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// LoginInput holds the credentials for authentication.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput holds the tokens returned after successful authentication.
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// LoginUC implements the use case for authenticating a user.
type LoginUC struct {
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	tokenManager     ports.TokenManager
	jwtExpiry        time.Duration
	tokenExpiry      time.Duration
}

// NewLoginUC creates a new LoginUC with all required dependencies.
func NewLoginUC(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	tokenManager ports.TokenManager,
	jwtExpiry time.Duration,
	tokenExpiry time.Duration,
) *LoginUC {
	return &LoginUC{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenManager:     tokenManager,
		jwtExpiry:        jwtExpiry,
		tokenExpiry:      tokenExpiry,
	}
}

// Execute authenticates a user and returns tokens.
// Returns ErrInvalidCredentials if the email does not exist or the password is incorrect.
func (uc *LoginUC) Execute(ctx context.Context, input LoginInput) (LoginOutput, error) {
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return LoginOutput{}, domainerrors.ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return LoginOutput{}, domainerrors.ErrInvalidCredentials
	}

	refreshTokenPlain, err := generateRefreshToken()
	if err != nil {
		return LoginOutput{}, err
	}

	now := time.Now()
	refreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     hashToken(refreshTokenPlain),
		ExpiresAt: now.Add(uc.tokenExpiry),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return LoginOutput{}, err
	}

	accessToken, err := uc.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPlain,
		ExpiresIn:    int(uc.jwtExpiry.Seconds()),
	}, nil
}
