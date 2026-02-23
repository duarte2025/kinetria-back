package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/constants"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
)

// RegisterInput holds the input data for user registration.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

// RegisterOutput holds the tokens returned after successful registration.
type RegisterOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds
}

// RegisterUC implements the use case for registering a new user.
type RegisterUC struct {
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	tokenManager     ports.TokenManager
	jwtExpiry        time.Duration
	tokenExpiry      time.Duration
}

// NewRegisterUC creates a new RegisterUC with all required dependencies.
func NewRegisterUC(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	tokenManager ports.TokenManager,
	jwtExpiry time.Duration,
	tokenExpiry time.Duration,
) *RegisterUC {
	return &RegisterUC{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenManager:     tokenManager,
		jwtExpiry:        jwtExpiry,
		tokenExpiry:      tokenExpiry,
	}
}

// Execute registers a new user and returns authentication tokens.
// Returns ErrEmailAlreadyExists if the email is already registered.
// Returns ErrMalformedParameters if the password is shorter than 8 characters.
func (uc *RegisterUC) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	if len(input.Password) < 8 {
		return RegisterOutput{}, domainerrors.ErrMalformedParameters
	}

	_, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err == nil {
		return RegisterOutput{}, domainerrors.ErrEmailAlreadyExists
	}
	if err != domainerrors.ErrNotFound {
		return RegisterOutput{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return RegisterOutput{}, err
	}

	now := time.Now()
	user := &entities.User{
		ID:              uuid.New(),
		Name:            input.Name,
		Email:           input.Email,
		PasswordHash:    string(passwordHash),
		ProfileImageURL: constants.DefaultUserAvatarURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return RegisterOutput{}, err
	}

	refreshTokenPlain, err := generateRefreshToken()
	if err != nil {
		return RegisterOutput{}, err
	}

	refreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     hashToken(refreshTokenPlain),
		ExpiresAt: now.Add(uc.tokenExpiry),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return RegisterOutput{}, err
	}

	accessToken, err := uc.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPlain,
		ExpiresIn:    int(uc.jwtExpiry.Seconds()),
	}, nil
}
