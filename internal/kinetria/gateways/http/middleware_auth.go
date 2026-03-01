package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// AuthMiddleware creates a middleware that validates JWT tokens and injects userID into the request context.
// It expects an "Authorization: Bearer <token>" header.
// Returns 401 Unauthorized if the token is missing, invalid, or expired.
func AuthMiddleware(jwtManager *gatewayauth.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := jwtManager.ParseToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
				return
			}

			// Inject userID into request context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// tryExtractUserIDFromJWT attempts to extract the userID from the JWT Bearer token
// in the Authorization header. Returns nil if the token is absent or invalid — no error is returned.
// This is used for endpoints with optional authentication.
func tryExtractUserIDFromJWT(r *http.Request, jwtManager *gatewayauth.JWTManager) *uuid.UUID {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := jwtManager.ParseToken(tokenString)
	if err != nil {
		return nil
	}
	return &userID
}
