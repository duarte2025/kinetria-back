package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// UserRepository implements ports.UserRepository using PostgreSQL via SQLC.
type UserRepository struct {
	q *queries.Queries
}

// NewUserRepository creates a new UserRepository backed by the provided *sql.DB.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{q: queries.New(db)}
}

// Create inserts a new user into the database.
// Returns ErrEmailAlreadyExists if the email is already taken.
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	_, err := r.q.CreateUser(ctx, queries.CreateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		ProfileImageUrl: sql.NullString{
			String: user.ProfileImageURL,
			Valid:  user.ProfileImageURL != "",
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

// GetByEmail retrieves a user by email address.
// Returns ErrNotFound if no user exists with the given email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return rowToUser(row.ID, row.Name, row.Email, row.PasswordHash, row.ProfileImageUrl, row.Preferences, row.CreatedAt, row.UpdatedAt), nil
}

// GetByID retrieves a user by ID.
// Returns ErrNotFound if no user exists with the given ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return rowToUser(row.ID, row.Name, row.Email, row.PasswordHash, row.ProfileImageUrl, row.Preferences, row.CreatedAt, row.UpdatedAt), nil
}

// Update updates an existing user's mutable fields.
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}
	return r.q.UpdateUser(ctx, queries.UpdateUserParams{
		ID:   user.ID,
		Name: user.Name,
		ProfileImageUrl: sql.NullString{
			String: user.ProfileImageURL,
			Valid:  user.ProfileImageURL != "",
		},
		Preferences: preferencesJSON,
	})
}

func rowToUser(id uuid.UUID, name, email, passwordHash string, profileImageUrl sql.NullString, preferencesJSON []byte, createdAt, updatedAt time.Time) *entities.User {
	profileURL := ""
	if profileImageUrl.Valid {
		profileURL = profileImageUrl.String
	}
	prefs := vos.DefaultUserPreferences()
	if len(preferencesJSON) > 0 {
		if err := json.Unmarshal(preferencesJSON, &prefs); err != nil {
			slog.Warn("failed to unmarshal user preferences, using defaults", "error", err)
		}
	}
	return &entities.User{
		ID:              id,
		Name:            name,
		Email:           email,
		PasswordHash:    passwordHash,
		ProfileImageURL: profileURL,
		Preferences:     prefs,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
