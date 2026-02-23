package vos_test

import (
	"errors"
	"testing"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

func TestWorkoutType_Validate_ValidValues(t *testing.T) {
	tests := []struct {
		name string
		wt   vos.WorkoutType
	}{
		{"forca", vos.WorkoutTypeForca},
		{"hipertrofia", vos.WorkoutTypeHipertrofia},
		{"mobilidade", vos.WorkoutTypeMobilidade},
		{"condicionamento", vos.WorkoutTypeCondicionamento},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wt.Validate(); err != nil {
				t.Errorf("expected no error for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestWorkoutType_Validate_InvalidValues(t *testing.T) {
	invalid := vos.WorkoutType("invalid_type")
	err := invalid.Validate()
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
	if !errors.Is(err, domerrors.ErrMalformedParameters) {
		t.Errorf("expected ErrMalformedParameters, got %v", err)
	}
}

func TestWorkoutType_String(t *testing.T) {
	tests := []struct {
		wt       vos.WorkoutType
		expected string
	}{
		{vos.WorkoutTypeForca, "FORÇA"},
		{vos.WorkoutTypeHipertrofia, "HIPERTROFIA"},
		{vos.WorkoutTypeMobilidade, "MOBILIDADE"},
		{vos.WorkoutTypeCondicionamento, "CONDICIONAMENTO"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.wt.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
