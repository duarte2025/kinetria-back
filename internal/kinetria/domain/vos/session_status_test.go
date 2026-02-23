package vos_test

import (
	"errors"
	"testing"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

func TestSessionStatus_Validate_ValidValues(t *testing.T) {
	tests := []struct {
		name string
		ss   vos.SessionStatus
	}{
		{"active", vos.SessionStatusActive},
		{"completed", vos.SessionStatusCompleted},
		{"abandoned", vos.SessionStatusAbandoned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ss.Validate(); err != nil {
				t.Errorf("expected no error for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestSessionStatus_Validate_InvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		status vos.SessionStatus
	}{
		{"in_progress", vos.SessionStatus("in_progress")},
		{"cancelled", vos.SessionStatus("cancelled")},
		{"empty", vos.SessionStatus("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.name)
			}
			if !errors.Is(err, domerrors.ErrMalformedParameters) {
				t.Errorf("expected ErrMalformedParameters, got %v", err)
			}
		})
	}
}

func TestSessionStatus_String(t *testing.T) {
	tests := []struct {
		ss       vos.SessionStatus
		expected string
	}{
		{vos.SessionStatusActive, "active"},
		{vos.SessionStatusCompleted, "completed"},
		{vos.SessionStatusAbandoned, "abandoned"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.ss.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
