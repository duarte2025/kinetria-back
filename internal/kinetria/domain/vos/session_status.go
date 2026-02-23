package vos

import (
	"fmt"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusAbandoned SessionStatus = "abandoned"
)

func (s SessionStatus) String() string {
	return string(s)
}

func (s SessionStatus) Validate() error {
	switch s {
	case SessionStatusActive, SessionStatusCompleted, SessionStatusAbandoned:
		return nil
	}
	return fmt.Errorf("invalid session status %q: %w", string(s), domerrors.ErrMalformedParameters)
}
