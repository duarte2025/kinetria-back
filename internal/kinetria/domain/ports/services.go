package ports

import "github.com/google/uuid"

// TokenManager handles JWT token generation and validation.
// This interface allows the domain layer to use token operations without depending on gateway implementations.
type TokenManager interface {
	GenerateToken(userID uuid.UUID) (string, error)
	ParseToken(tokenString string) (uuid.UUID, error)
}
