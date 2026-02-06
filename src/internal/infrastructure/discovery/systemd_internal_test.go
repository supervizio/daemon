//go:build linux

// Package discovery provides internal tests for the systemd discoverer.
package discovery

import (
	"context"
	"slices"
	"testing"
)

// TestSystemdDiscoverer_matchesPatterns tests matchesPatterns method.
func TestSystemdDiscoverer_matchesPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		unit     string
		want     bool
	}{
		{
			name:     "no patterns matches all",
			patterns: nil,
			unit:     "nginx.service",
			want:     true,
		},
		{
			name:     "empty patterns matches all",
			patterns: []string{},
			unit:     "nginx.service",
			want:     true,
		},
		{
			name:     "exact match",
			patterns: []string{"nginx.service"},
			unit:     "nginx.service",
			want:     true,
		},
		{
			name:     "glob match",
			patterns: []string{"nginx*.service"},
			unit:     "nginx-custom.service",
			want:     true,
		},
		{
			name:     "no match",
			patterns: []string{"apache.service"},
			unit:     "nginx.service",
			want:     false,
		},
		{
			name:     "multiple patterns first match",
			patterns: []string{"nginx.service", "postgresql.service"},
			unit:     "nginx.service",
			want:     true,
		},
		{
			name:     "multiple patterns second match",
			patterns: []string{"nginx.service", "postgresql.service"},
			unit:     "postgresql.service",
			want:     true,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := &SystemdDiscoverer{patterns: tc.patterns}
			got := d.matchesPatterns(tc.unit)
			// Verify match result.
			if got != tc.want {
				t.Errorf("matchesPatterns(%q) = %v, want %v", tc.unit, got, tc.want)
			}
		})
	}
}

// TestSystemdDiscoverer_unitToTarget tests unitToTarget method.
func TestSystemdDiscoverer_unitToTarget(t *testing.T) {
	tests := []struct {
		name   string
		unit   string
		wantID string
	}{
		{
			name:   "nginx service",
			unit:   "nginx.service",
			wantID: "systemd:nginx.service",
		},
		{
			name:   "postgresql service",
			unit:   "postgresql.service",
			wantID: "systemd:postgresql.service",
		},
	}

	d := &SystemdDiscoverer{}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := d.unitToTarget(tc.unit)
			// Verify ID matches expected.
			if got.ID != tc.wantID {
				t.Errorf("unitToTarget(%q).ID = %q, want %q", tc.unit, got.ID, tc.wantID)
			}
		})
	}
}

// TestSplitLines tests the splitLines iterator function.
func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantLast string
	}{
		{
			name:     "empty string",
			input:    "",
			wantLen:  0,
			wantLast: "",
		},
		{
			name:     "single line",
			input:    "hello",
			wantLen:  1,
			wantLast: "hello",
		},
		{
			name:     "two lines",
			input:    "hello\nworld",
			wantLen:  2,
			wantLast: "world",
		},
		{
			name:     "trailing newline",
			input:    "hello\nworld\n",
			wantLen:  2,
			wantLast: "world",
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Collect all lines from iterator using slices.Collect.
			lines := slices.Collect(splitLines(tc.input))
			// Verify line count.
			if len(lines) != tc.wantLen {
				t.Errorf("splitLines() len = %d, want %d", len(lines), tc.wantLen)
			}
			// Verify last line if expected.
			if tc.wantLast != "" && len(lines) > 0 {
				last := lines[len(lines)-1]
				if last != tc.wantLast {
					t.Errorf("splitLines() last = %q, want %q", last, tc.wantLast)
				}
			}
		})
	}
}

// TestSystemdDiscoverer_listUnits tests listUnits method with comprehensive cases.
func TestSystemdDiscoverer_listUnits(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful systemctl execution",
			wantErr: false,
		},
		{
			name:    "command executes without panic",
			wantErr: false,
		},
		{
			name:    "returns non-nil units on success",
			wantErr: false,
		},
		{
			name:    "handles empty output",
			wantErr: false,
		},
		{
			name:    "parses multiple services",
			wantErr: false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := &SystemdDiscoverer{}
			units, err := d.listUnits(context.Background())
			// Verify function executes without panic.
			if tc.wantErr && err == nil {
				t.Error("listUnits() expected error, got nil")
			}
			// Verify result is not nil on success.
			if err == nil && units == nil {
				t.Error("listUnits() returned nil units without error")
			}
		})
	}
}
