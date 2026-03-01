package profile_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

// mockProfileUserRepo is an inline mock for ports.UserRepository used in profile tests.
type mockProfileUserRepo struct {
	byID       map[uuid.UUID]*entities.User
	getByIDErr error
	updateErr  error
}

func (m *mockProfileUserRepo) Create(_ context.Context, _ *entities.User) error {
	return nil
}

func (m *mockProfileUserRepo) GetByEmail(_ context.Context, _ string) (*entities.User, error) {
	return nil, domainerrors.ErrNotFound
}

func (m *mockProfileUserRepo) GetByID(_ context.Context, id uuid.UUID) (*entities.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if u, ok := m.byID[id]; ok {
		return u, nil
	}
	return nil, domainerrors.ErrNotFound
}

func (m *mockProfileUserRepo) Update(_ context.Context, user *entities.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	// Persist back so assertions can read the updated state
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*entities.User)
	}
	m.byID[user.ID] = user
	return nil
}
