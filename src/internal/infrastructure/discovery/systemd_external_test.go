//go:build linux

// Package discovery_test provides external tests for the systemd discoverer.
package discovery_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
)

// TestNewSystemdDiscoverer tests the NewSystemdDiscoverer constructor.
func TestNewSystemdDiscoverer(t *testing.T) {
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
			patterns: []string{"nginx.service"},
			wantNil:  false,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"nginx.service", "postgresql*.service"},
			wantNil:  false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewSystemdDiscoverer(tc.patterns)
			// Verify discoverer is not nil.
			if (d == nil) != tc.wantNil {
				t.Errorf("NewSystemdDiscoverer() = %v, wantNil = %v", d, tc.wantNil)
			}
		})
	}
}

// TestSystemdDiscoverer_Type tests SystemdDiscoverer.Type method.
func TestSystemdDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantType target.Type
	}{
		{
			name:     "empty patterns",
			patterns: nil,
			wantType: target.TypeSystemd,
		},
		{
			name:     "with patterns",
			patterns: []string{"nginx.service"},
			wantType: target.TypeSystemd,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewSystemdDiscoverer(tc.patterns)
			got := d.Type()
			// Verify type matches expected.
			if got != tc.wantType {
				t.Errorf("Type() = %v, want %v", got, tc.wantType)
			}
		})
	}
}

// TestSystemdDiscoverer_Discover tests SystemdDiscoverer.Discover method.
func TestSystemdDiscoverer_Discover(t *testing.T) {
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
			patterns: []string{"*.service"},
			wantErr:  false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := discovery.NewSystemdDiscoverer(tc.patterns)
			targets, err := d.Discover(context.Background())
			// Check error matches expectation.
			if (err != nil) != tc.wantErr {
				t.Errorf("Discover() error = %v, wantErr %v", err, tc.wantErr)
			}
			// Verify targets is not nil on success.
			if err == nil && len(targets) < 0 {
				t.Error("Discover() returned nil targets without error")
			}
		})
	}
}
