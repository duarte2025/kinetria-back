package auth

import (
"errors"
"time"

"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
)

// JWTManager handles JWT token generation and validation.
type JWTManager struct {
secret []byte
expiry time.Duration
}

// NewJWTManager creates a new JWTManager with the given secret and token expiry.
func NewJWTManager(secret string, expiry time.Duration) *JWTManager {
return &JWTManager{
secret: []byte(secret),
expiry: expiry,
}
}

// GenerateToken generates a JWT with the user ID in the "sub" claim.
// The token is signed with HS256 and expires after the configured duration.
func (m *JWTManager) GenerateToken(userID uuid.UUID) (string, error) {
now := time.Now()
claims := jwt.RegisteredClaims{
Subject:   userID.String(),
IssuedAt:  jwt.NewNumericDate(now),
ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString(m.secret)
}

// ParseToken validates a JWT and returns the user ID from the "sub" claim.
// Returns an error if the token is invalid, expired, or uses an unexpected signing method.
func (m *JWTManager) ParseToken(tokenString string) (uuid.UUID, error) {
token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, errors.New("invalid signing method")
}
return m.secret, nil
})
if err != nil {
return uuid.Nil, err
}

claims, ok := token.Claims.(*jwt.RegisteredClaims)
if !ok || !token.Valid {
return uuid.Nil, errors.New("invalid token claims")
}

userID, err := uuid.Parse(claims.Subject)
if err != nil {
return uuid.Nil, err
}
return userID, nil
}
