package vos

import (
	"fmt"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

type SetRecordStatus string

const (
	SetRecordStatusCompleted SetRecordStatus = "completed"
	SetRecordStatusSkipped   SetRecordStatus = "skipped"
)

func (s SetRecordStatus) String() string {
	return string(s)
}

func (s SetRecordStatus) Validate() error {
	switch s {
	case SetRecordStatusCompleted, SetRecordStatusSkipped:
		return nil
	}
	return fmt.Errorf("invalid set record status %q: %w", string(s), domerrors.ErrMalformedParameters)
}
