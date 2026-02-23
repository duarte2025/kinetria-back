package errors

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("data conflict")
	ErrMalformedParameters = errors.New("malformed parameters")
	ErrFailedDependency    = errors.New("failed dependency")
)
