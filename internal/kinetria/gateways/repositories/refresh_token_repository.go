package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/queries"
)

// RefreshTokenRepository implements ports.RefreshTokenRepository using PostgreSQL via SQLC.
type RefreshTokenRepository struct {
	q *queries.Queries
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository backed by the provided *sql.DB.
func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{q: queries.New(db)}
}

// Create inserts a new refresh token into the database.
// The Token field should contain the SHA-256 hash of the actual token.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	_, err := r.q.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	})
	return err
}

// GetByToken retrieves a refresh token by its hash.
// Returns ErrTokenInvalid if no token exists with the given hash.
func (r *RefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrTokenInvalid
		}
		return nil, err
	}

	var revokedAt *time.Time
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		revokedAt = &t
	}

	return &entities.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		Token:     row.Token,
		ExpiresAt: row.ExpiresAt,
		RevokedAt: revokedAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

// RevokeByToken marks the refresh token with the given hash as revoked.
func (r *RefreshTokenRepository) RevokeByToken(ctx context.Context, tokenHash string) error {
	return r.q.RevokeRefreshToken(ctx, tokenHash)
}

// RevokeAllByUserID marks all active refresh tokens for the user as revoked.
func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.q.RevokeAllUserTokens(ctx, userID)
}
