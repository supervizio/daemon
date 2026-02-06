//go:build linux

// Package discovery_test provides external tests for the OpenRC discoverer.
package discovery_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
)

// hasRCStatus checks if the rc-status command is available.
func hasRCStatus() bool {
	_, err := exec.LookPath("rc-status")
	return err == nil
}

// TestNewOpenRCDiscoverer tests the NewOpenRCDiscoverer constructor.
func TestNewOpenRCDiscoverer(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantNil  bool
	}{
		{
			name:     "nil patterns",
			patterns: nil,
			wantNil:  false,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			wantNil:  false,
		},
		{
			name:     "single pattern",
			patterns: []string{"nginx"},
			wantNil:  false,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"nginx", "postgresql*"},
			wantNil:  false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewOpenRCDiscoverer(tc.patterns)
			// Verify discoverer is not nil.
			if (d == nil) != tc.wantNil {
				t.Errorf("NewOpenRCDiscoverer() = %v, wantNil = %v", d, tc.wantNil)
			}
		})
	}
}

// TestOpenRCDiscoverer_Type tests OpenRCDiscoverer.Type method.
func TestOpenRCDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantType target.Type
	}{
		{
			name:     "empty patterns",
			patterns: nil,
			wantType: target.TypeOpenRC,
		},
		{
			name:     "with patterns",
			patterns: []string{"nginx"},
			wantType: target.TypeOpenRC,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewOpenRCDiscoverer(tc.patterns)
			got := d.Type()
			// Verify type matches expected.
			if got != tc.wantType {
				t.Errorf("Type() = %v, want %v", got, tc.wantType)
			}
		})
	}
}

// TestOpenRCDiscoverer_Discover tests OpenRCDiscoverer.Discover method.
func TestOpenRCDiscoverer_Discover(t *testing.T) {
	// Determine expected behavior based on rc-status availability.
	rcStatusAvailable := hasRCStatus()

	tests := []struct {
		name     string
		patterns []string
	}{
		{
			name:     "no patterns lists all",
			patterns: nil,
		},
		{
			name:     "empty patterns lists all",
			patterns: []string{},
		},
		{
			name:     "with pattern filter",
			patterns: []string{"*"},
		},
		{
			name:     "specific service pattern",
			patterns: []string{"nginx"},
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewOpenRCDiscoverer(tc.patterns)
			targets, err := d.Discover(context.Background())

			// When rc-status is not available, expect an error.
			if !rcStatusAvailable {
				if err == nil {
					t.Error("Discover() expected error when rc-status not available")
				}
				// Return early since we can't proceed without rc-status.
				return
			}

			// When rc-status is available, expect success.
			if err != nil {
				t.Errorf("Discover() unexpected error = %v", err)
			}
			// Verify targets is not nil on success.
			if err == nil && targets == nil {
				t.Error("Discover() returned nil targets without error")
			}
		})
	}
}
