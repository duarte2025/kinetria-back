package vos_test

import (
	"errors"
	"testing"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

func TestWorkoutIntensity_Validate_ValidValues(t *testing.T) {
	tests := []struct {
		name string
		wi   vos.WorkoutIntensity
	}{
		{"baixa", vos.WorkoutIntensityBaixa},
		{"moderada", vos.WorkoutIntensityModerada},
		{"alta", vos.WorkoutIntensityAlta},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wi.Validate(); err != nil {
				t.Errorf("expected no error for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestWorkoutIntensity_Validate_InvalidValues(t *testing.T) {
	invalid := vos.WorkoutIntensity("invalid")
	err := invalid.Validate()
	if err == nil {
		t.Error("expected error for invalid intensity, got nil")
	}
	if !errors.Is(err, domerrors.ErrMalformedParameters) {
		t.Errorf("expected ErrMalformedParameters, got %v", err)
	}
}

func TestWorkoutIntensity_String(t *testing.T) {
	tests := []struct {
		wi       vos.WorkoutIntensity
		expected string
	}{
		{vos.WorkoutIntensityBaixa, "BAIXA"},
		{vos.WorkoutIntensityModerada, "MODERADA"},
		{vos.WorkoutIntensityAlta, "ALTA"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.wi.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
