package vos_test

import (
	"errors"
	"testing"

	domerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
)

func TestSetRecordStatus_Validate_ValidValues(t *testing.T) {
	tests := []struct {
		name string
		srs  vos.SetRecordStatus
	}{
		{"completed", vos.SetRecordStatusCompleted},
		{"skipped", vos.SetRecordStatusSkipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.srs.Validate(); err != nil {
				t.Errorf("expected no error for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestSetRecordStatus_Validate_InvalidValues(t *testing.T) {
	invalid := vos.SetRecordStatus("pending")
	err := invalid.Validate()
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
	if !errors.Is(err, domerrors.ErrMalformedParameters) {
		t.Errorf("expected ErrMalformedParameters, got %v", err)
	}
}

func TestSetRecordStatus_String(t *testing.T) {
	tests := []struct {
		srs      vos.SetRecordStatus
		expected string
	}{
		{vos.SetRecordStatusCompleted, "completed"},
		{vos.SetRecordStatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.srs.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
