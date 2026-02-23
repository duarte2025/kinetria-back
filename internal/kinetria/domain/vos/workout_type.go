package vos

import (
	"fmt"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

type WorkoutType string

const (
	WorkoutTypeForca          WorkoutType = "FORÇA"
	WorkoutTypeHipertrofia    WorkoutType = "HIPERTROFIA"
	WorkoutTypeMobilidade     WorkoutType = "MOBILIDADE"
	WorkoutTypeCondicionamento WorkoutType = "CONDICIONAMENTO"
)

func (w WorkoutType) String() string {
	return string(w)
}

func (w WorkoutType) Validate() error {
	switch w {
	case WorkoutTypeForca, WorkoutTypeHipertrofia, WorkoutTypeMobilidade, WorkoutTypeCondicionamento:
		return nil
	}
	return fmt.Errorf("invalid workout type %q: %w", string(w), domerrors.ErrMalformedParameters)
}
