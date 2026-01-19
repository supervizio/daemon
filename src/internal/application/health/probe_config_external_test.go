package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/application/health"
)

// TestDefaultProbeConfig tests the DefaultProbeConfig function.
func TestDefaultProbeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		wantInterval         time.Duration
		wantTimeout          time.Duration
		wantSuccessThreshold int
		wantFailureThreshold int
	}{
		{
			name:                 "default_values",
			wantInterval:         10 * time.Second,
			wantTimeout:          5 * time.Second,
			wantSuccessThreshold: 1,
			wantFailureThreshold: 3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Get default config.
			config := health.DefaultProbeConfig()

			// Verify all default values.
			assert.Equal(t, tt.wantInterval, config.Interval)
			assert.Equal(t, tt.wantTimeout, config.Timeout)
			assert.Equal(t, tt.wantSuccessThreshold, config.SuccessThreshold)
			assert.Equal(t, tt.wantFailureThreshold, config.FailureThreshold)
		})
	}
}

// TestProbeConfig_FieldsAccess tests that ProbeConfig fields are accessible.
func TestProbeConfig_FieldsAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		interval         time.Duration
		timeout          time.Duration
		successThreshold int
		failureThreshold int
	}{
		{
			name:             "custom_values",
			interval:         15 * time.Second,
			timeout:          3 * time.Second,
			successThreshold: 2,
			failureThreshold: 5,
		},
		{
			name:             "minimal_values",
			interval:         1 * time.Second,
			timeout:          100 * time.Millisecond,
			successThreshold: 1,
			failureThreshold: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create config with test values.
			config := health.ProbeConfig{
				Interval:         tt.interval,
				Timeout:          tt.timeout,
				SuccessThreshold: tt.successThreshold,
				FailureThreshold: tt.failureThreshold,
			}

			// Verify fields are set correctly.
			assert.Equal(t, tt.interval, config.Interval)
			assert.Equal(t, tt.timeout, config.Timeout)
			assert.Equal(t, tt.successThreshold, config.SuccessThreshold)
			assert.Equal(t, tt.failureThreshold, config.FailureThreshold)
		})
	}
}
