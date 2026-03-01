package auth_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

// mockUserRepo is a mock implementation of ports.UserRepository for testing.
type mockUserRepo struct {
	users         map[string]*entities.User // keyed by email
	createErr     error
	getByEmailErr error
	getByIDErr    error
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.users == nil {
		m.users = make(map[string]*entities.User)
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, domainerrors.ErrNotFound
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return nil, domainerrors.ErrNotFound
}

func (m *mockUserRepo) Update(_ context.Context, _ *entities.User) error {
	return nil
}

// mockRefreshTokenRepo is a mock implementation of ports.RefreshTokenRepository for testing.
type mockRefreshTokenRepo struct {
	tokens    map[string]*entities.RefreshToken // keyed by token hash
	createErr error
	revokeErr error
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *entities.RefreshToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.tokens == nil {
		m.tokens = make(map[string]*entities.RefreshToken)
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockRefreshTokenRepo) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t, nil
	}
	return nil, domainerrors.ErrTokenInvalid
}

func (m *mockRefreshTokenRepo) RevokeByToken(ctx context.Context, tokenHash string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	if t, ok := m.tokens[tokenHash]; ok {
		now := time.Now()
		t.RevokedAt = &now
	}
	return nil // idempotent - no error if not found
}

func (m *mockRefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// mockTokenManager is a mock implementation of ports.TokenManager for testing.
type mockTokenManager struct {
	token    string
	parseErr error
}

func (m *mockTokenManager) GenerateToken(userID uuid.UUID) (string, error) {
	if m.token != "" {
		return m.token, nil
	}
	return "mock-jwt-token", nil
}

func (m *mockTokenManager) ParseToken(tokenString string) (uuid.UUID, error) {
	if m.parseErr != nil {
		return uuid.Nil, m.parseErr
	}
	return uuid.New(), nil
}

