package vos

import (
	"fmt"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

type WorkoutIntensity string

const (
	WorkoutIntensityBaixa    WorkoutIntensity = "BAIXA"
	WorkoutIntensityModerada WorkoutIntensity = "MODERADA"
	WorkoutIntensityAlta     WorkoutIntensity = "ALTA"
)

func (w WorkoutIntensity) String() string {
	return string(w)
}

func (w WorkoutIntensity) Validate() error {
	switch w {
	case WorkoutIntensityBaixa, WorkoutIntensityModerada, WorkoutIntensityAlta:
		return nil
	}
	return fmt.Errorf("invalid workout intensity %q: %w", string(w), domerrors.ErrMalformedParameters)
}
