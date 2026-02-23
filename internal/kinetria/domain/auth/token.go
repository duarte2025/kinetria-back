package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const refreshTokenLength = 32 // 32 bytes = 256 bits

// generateRefreshToken generates a cryptographically random token of 256 bits, base64url encoded.
func generateRefreshToken() (string, error) {
	b := make([]byte, refreshTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hash of the given token as a lowercase hex string (64 chars).
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
