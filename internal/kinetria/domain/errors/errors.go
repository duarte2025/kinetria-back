package errors

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("data conflict")
	ErrMalformedParameters = errors.New("malformed parameters")
	ErrFailedDependency    = errors.New("failed dependency")

	// Auth errors
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenInvalid       = errors.New("token invalid")
)
