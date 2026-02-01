//go:build linux

// Package discovery_test provides external tests for the OpenRC discoverer.
package discovery_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
)

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
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
	}{
		{
			name:     "no patterns lists all",
			patterns: nil,
			wantErr:  false,
		},
		{
			name:     "empty patterns lists all",
			patterns: []string{},
			wantErr:  false,
		},
		{
			name:     "with pattern filter",
			patterns: []string{"*"},
			wantErr:  false,
		},
		{
			name:     "specific service pattern",
			patterns: []string{"nginx"},
			wantErr:  false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewOpenRCDiscoverer(tc.patterns)
			targets, err := d.Discover(context.Background())
			// Check error matches expectation.
			if (err != nil) != tc.wantErr {
				t.Errorf("Discover() error = %v, wantErr %v", err, tc.wantErr)
			}
			// Verify targets is not nil on success.
			if err == nil && targets == nil {
				t.Error("Discover() returned nil targets without error")
			}
		})
	}
}
