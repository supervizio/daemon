// Package probe_test provides black-box tests for the probe package.
package probe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// TestNewConfig tests default config creation.
func TestNewConfig(t *testing.T) {
	tests := []struct {
		name             string
		expectedTimeout  time.Duration
		expectedInterval time.Duration
		expectedSuccess  int
		expectedFailure  int
	}{
		{
			name:             "default_config_values",
			expectedTimeout:  probe.DefaultTimeout,
			expectedInterval: probe.DefaultInterval,
			expectedSuccess:  probe.DefaultSuccessThreshold,
			expectedFailure:  probe.DefaultFailureThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with defaults.
			cfg := probe.NewConfig()

			// Verify default values.
			assert.Equal(t, tt.expectedTimeout, cfg.Timeout)
			assert.Equal(t, tt.expectedInterval, cfg.Interval)
			assert.Equal(t, tt.expectedSuccess, cfg.SuccessThreshold)
			assert.Equal(t, tt.expectedFailure, cfg.FailureThreshold)
		})
	}
}

// TestConfig_WithTimeout tests the WithTimeout method.
func TestConfig_WithTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "short_timeout",
			timeout:  time.Second,
			expected: time.Second,
		},
		{
			name:     "long_timeout",
			timeout:  30 * time.Second,
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with custom timeout.
			cfg := probe.NewConfig().WithTimeout(tt.timeout)

			// Verify timeout.
			assert.Equal(t, tt.expected, cfg.Timeout)
		})
	}
}

// TestConfig_WithInterval tests the WithInterval method.
func TestConfig_WithInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		expected time.Duration
	}{
		{
			name:     "short_interval",
			interval: 5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "long_interval",
			interval: time.Minute,
			expected: time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with custom interval.
			cfg := probe.NewConfig().WithInterval(tt.interval)

			// Verify interval.
			assert.Equal(t, tt.expected, cfg.Interval)
		})
	}
}

// TestConfig_WithSuccessThreshold tests the WithSuccessThreshold method.
func TestConfig_WithSuccessThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		expected  int
	}{
		{
			name:      "threshold_1",
			threshold: 1,
			expected:  1,
		},
		{
			name:      "threshold_5",
			threshold: 5,
			expected:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with custom threshold.
			cfg := probe.NewConfig().WithSuccessThreshold(tt.threshold)

			// Verify threshold.
			assert.Equal(t, tt.expected, cfg.SuccessThreshold)
		})
	}
}

// TestConfig_WithFailureThreshold tests the WithFailureThreshold method.
func TestConfig_WithFailureThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		expected  int
	}{
		{
			name:      "threshold_1",
			threshold: 1,
			expected:  1,
		},
		{
			name:      "threshold_3",
			threshold: 3,
			expected:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with custom threshold.
			cfg := probe.NewConfig().WithFailureThreshold(tt.threshold)

			// Verify threshold.
			assert.Equal(t, tt.expected, cfg.FailureThreshold)
		})
	}
}

// TestConfig_Validate tests configuration validation.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      probe.Config
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid_default_config",
			config:      probe.NewConfig(),
			expectError: false,
		},
		{
			name: "invalid_timeout",
			config: probe.Config{
				Timeout:          0,
				Interval:         time.Second,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			expectError: true,
			expectedErr: probe.ErrInvalidTimeout,
		},
		{
			name: "invalid_interval",
			config: probe.Config{
				Timeout:          time.Second,
				Interval:         0,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			expectError: true,
			expectedErr: probe.ErrInvalidInterval,
		},
		{
			name: "invalid_success_threshold",
			config: probe.Config{
				Timeout:          time.Second,
				Interval:         time.Second,
				SuccessThreshold: 0,
				FailureThreshold: 1,
			},
			expectError: true,
			expectedErr: probe.ErrInvalidSuccessThreshold,
		},
		{
			name: "invalid_failure_threshold",
			config: probe.Config{
				Timeout:          time.Second,
				Interval:         time.Second,
				SuccessThreshold: 1,
				FailureThreshold: 0,
			},
			expectError: true,
			expectedErr: probe.ErrInvalidFailureThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate config.
			err := tt.config.Validate()

			// Verify result.
			if tt.expectError {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
