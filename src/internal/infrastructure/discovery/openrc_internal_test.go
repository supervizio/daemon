//go:build linux

package discovery

import (
	"slices"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

// TestSplitLinesOpenRC verifies the line splitting iterator.
func TestSplitLinesOpenRC(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLines []string
		wantCount int
	}{
		{
			name:      "empty string",
			input:     "",
			wantLines: []string{},
			wantCount: 0,
		},
		{
			name:      "single line no newline",
			input:     "nginx [started]",
			wantLines: []string{"nginx [started]"},
			wantCount: 1,
		},
		{
			name:      "single line with newline",
			input:     "nginx [started]\n",
			wantLines: []string{"nginx [started]"},
			wantCount: 1,
		},
		{
			name:      "multiple lines",
			input:     "nginx [started]\npostgres [started]\nredis [stopped]\n",
			wantLines: []string{"nginx [started]", "postgres [started]", "redis [stopped]"},
			wantCount: 3,
		},
		{
			name:      "multiple lines no trailing newline",
			input:     "service1 [started]\nservice2 [started]",
			wantLines: []string{"service1 [started]", "service2 [started]"},
			wantCount: 2,
		},
		{
			name:      "empty lines in middle",
			input:     "nginx [started]\n\npostgres [started]\n",
			wantLines: []string{"nginx [started]", "", "postgres [started]"},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := slices.Collect(splitLinesOpenRC(tt.input))

			assert.Len(t, lines, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantLines, lines)
			}
		})
	}
}

// TestSplitLinesOpenRC_EarlyTermination verifies iterator can be stopped early.
func TestSplitLinesOpenRC_EarlyTermination(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		stopAfter int
		wantCount int
	}{
		{
			name:      "stop after first line",
			input:     "line1\nline2\nline3\n",
			stopAfter: 1,
			wantCount: 1,
		},
		{
			name:      "stop after second line",
			input:     "line1\nline2\nline3\n",
			stopAfter: 2,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			for range splitLinesOpenRC(tt.input) {
				count++
				if count >= tt.stopAfter {
					break
				}
			}

			assert.Equal(t, tt.wantCount, count)
		})
	}
}

// TestOpenRCDiscoverer_matchesPatterns verifies pattern matching logic.
func TestOpenRCDiscoverer_matchesPatterns(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		service     string
		shouldMatch bool
	}{
		{
			name:        "no patterns matches all",
			patterns:    nil,
			service:     "nginx",
			shouldMatch: true,
		},
		{
			name:        "empty patterns matches all",
			patterns:    []string{},
			service:     "postgresql",
			shouldMatch: true,
		},
		{
			name:        "exact match",
			patterns:    []string{"nginx"},
			service:     "nginx",
			shouldMatch: true,
		},
		{
			name:        "exact no match",
			patterns:    []string{"nginx"},
			service:     "redis",
			shouldMatch: false,
		},
		{
			name:        "wildcard suffix match",
			patterns:    []string{"postgres*"},
			service:     "postgresql",
			shouldMatch: true,
		},
		{
			name:        "wildcard suffix no match",
			patterns:    []string{"postgres*"},
			service:     "nginx",
			shouldMatch: false,
		},
		{
			name:        "wildcard prefix match",
			patterns:    []string{"*sql"},
			service:     "postgresql",
			shouldMatch: true,
		},
		{
			name:        "multiple patterns first matches",
			patterns:    []string{"nginx", "redis"},
			service:     "nginx",
			shouldMatch: true,
		},
		{
			name:        "multiple patterns second matches",
			patterns:    []string{"nginx", "redis"},
			service:     "redis",
			shouldMatch: true,
		},
		{
			name:        "multiple patterns none match",
			patterns:    []string{"nginx", "redis"},
			service:     "postgres",
			shouldMatch: false,
		},
		{
			name:        "question mark wildcard",
			patterns:    []string{"redis?"},
			service:     "redis1",
			shouldMatch: true,
		},
		{
			name:        "character class match",
			patterns:    []string{"redis[123]"},
			service:     "redis1",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &OpenRCDiscoverer{
				patterns: tt.patterns,
			}

			result := d.matchesPatterns(tt.service)

			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestOpenRCDiscoverer_serviceToTarget verifies service to target conversion.
func TestOpenRCDiscoverer_serviceToTarget(t *testing.T) {
	tests := []struct {
		name          string
		service       string
		wantID        string
		wantName      string
		wantType      target.Type
		wantProbeType string
		wantLabel     string
	}{
		{
			name:          "nginx service",
			service:       "nginx",
			wantID:        "openrc:nginx",
			wantName:      "nginx",
			wantType:      target.TypeOpenRC,
			wantProbeType: "exec",
			wantLabel:     "nginx",
		},
		{
			name:          "postgresql service",
			service:       "postgresql",
			wantID:        "openrc:postgresql",
			wantName:      "postgresql",
			wantType:      target.TypeOpenRC,
			wantProbeType: "exec",
			wantLabel:     "postgresql",
		},
		{
			name:          "service with dash",
			service:       "php-fpm",
			wantID:        "openrc:php-fpm",
			wantName:      "php-fpm",
			wantType:      target.TypeOpenRC,
			wantProbeType: "exec",
			wantLabel:     "php-fpm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &OpenRCDiscoverer{}

			tgt := d.serviceToTarget(tt.service)

			assert.Equal(t, tt.wantID, tgt.ID)
			assert.Equal(t, tt.wantName, tgt.Name)
			assert.Equal(t, tt.wantType, tgt.Type)
			assert.Equal(t, tt.wantProbeType, tgt.ProbeType)
			assert.Equal(t, target.SourceDiscovered, tgt.Source)
			assert.Equal(t, tt.wantLabel, tgt.Labels["openrc.service"])
			assert.NotNil(t, tgt.ProbeTarget)
		})
	}
}

// TestOpenRCDiscoverer_serviceToTarget_ProbeConfig verifies probe configuration.
func TestOpenRCDiscoverer_serviceToTarget_ProbeConfig(t *testing.T) {
	tests := []struct {
		name              string
		service           string
		wantInterval      bool
		wantTimeout       bool
		wantSuccessThresh bool
		wantFailureThresh bool
	}{
		{
			name:              "has default probe configuration",
			service:           "nginx",
			wantInterval:      true,
			wantTimeout:       true,
			wantSuccessThresh: true,
			wantFailureThresh: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &OpenRCDiscoverer{}

			tgt := d.serviceToTarget(tt.service)

			if tt.wantInterval {
				assert.NotZero(t, tgt.Interval)
			}
			if tt.wantTimeout {
				assert.NotZero(t, tgt.Timeout)
			}
			if tt.wantSuccessThresh {
				assert.NotZero(t, tgt.SuccessThreshold)
			}
			if tt.wantFailureThresh {
				assert.NotZero(t, tgt.FailureThreshold)
			}
		})
	}
}

// TestOpenRCDiscoverer_listServices verifies service listing error handling.
// Note: This test cannot call rc-status directly as it requires OpenRC.
func TestOpenRCDiscoverer_listServices(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
	}{
		{
			name:     "fails when rc-status not available with wildcard pattern",
			patterns: []string{"*"},
			wantErr:  true, // rc-status is not available in test environment
		},
		{
			name:     "fails when rc-status not available with specific pattern",
			patterns: []string{"nginx"},
			wantErr:  true, // rc-status is not available in test environment
		},
		{
			name:     "fails when rc-status not available with empty patterns",
			patterns: []string{},
			wantErr:  true, // rc-status is not available in test environment
		},
		{
			name:     "fails when rc-status not available with multiple patterns",
			patterns: []string{"nginx", "redis*", "postgres?"},
			wantErr:  true, // rc-status is not available in test environment
		},
		{
			name:     "fails when rc-status not available with nil patterns",
			patterns: nil,
			wantErr:  true, // rc-status is not available in test environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &OpenRCDiscoverer{patterns: tt.patterns}

			_, err := d.listServices(t.Context())

			if tt.wantErr {
				// Expected to fail because rc-status is not installed.
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
