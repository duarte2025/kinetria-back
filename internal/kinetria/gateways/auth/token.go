package auth

import (
"crypto/rand"
"crypto/sha256"
"fmt"

"encoding/base64"
)

const refreshTokenLength = 32 // 32 bytes = 256 bits

// GenerateRefreshToken generates a cryptographically random token of 256 bits, base64url encoded.
func GenerateRefreshToken() (string, error) {
b := make([]byte, refreshTokenLength)
if _, err := rand.Read(b); err != nil {
return "", fmt.Errorf("failed to generate refresh token: %w", err)
}
return base64.URLEncoding.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of the given token as a lowercase hex string (64 chars).
// Use this to hash refresh tokens before storing them in the database.
func HashToken(token string) string {
hash := sha256.Sum256([]byte(token))
return fmt.Sprintf("%x", hash)
}
