// Package health_test provides black-box tests for the health package.
package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewCheckConfig tests default config creation.
func TestNewCheckConfig(t *testing.T) {
	tests := []struct {
		name             string
		expectedTimeout  time.Duration
		expectedInterval time.Duration
		expectedSuccess  int
		expectedFailure  int
	}{
		{
			name:             "default_config_values",
			expectedTimeout:  health.DefaultTimeout,
			expectedInterval: health.DefaultInterval,
			expectedSuccess:  health.DefaultSuccessThreshold,
			expectedFailure:  health.DefaultFailureThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with defaults.
			cfg := health.NewCheckConfig()

			// Verify default values.
			assert.Equal(t, tt.expectedTimeout, cfg.Timeout)
			assert.Equal(t, tt.expectedInterval, cfg.Interval)
			assert.Equal(t, tt.expectedSuccess, cfg.SuccessThreshold)
			assert.Equal(t, tt.expectedFailure, cfg.FailureThreshold)
		})
	}
}

// TestCheckConfig_WithTimeout tests the WithTimeout method.
func TestCheckConfig_WithTimeout(t *testing.T) {
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
			cfg := health.NewCheckConfig().WithTimeout(tt.timeout)

			// Verify timeout.
			assert.Equal(t, tt.expected, cfg.Timeout)
		})
	}
}

// TestCheckConfig_WithInterval tests the WithInterval method.
func TestCheckConfig_WithInterval(t *testing.T) {
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
			cfg := health.NewCheckConfig().WithInterval(tt.interval)

			// Verify interval.
			assert.Equal(t, tt.expected, cfg.Interval)
		})
	}
}

// TestCheckConfig_WithSuccessThreshold tests the WithSuccessThreshold method.
func TestCheckConfig_WithSuccessThreshold(t *testing.T) {
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
			cfg := health.NewCheckConfig().WithSuccessThreshold(tt.threshold)

			// Verify threshold.
			assert.Equal(t, tt.expected, cfg.SuccessThreshold)
		})
	}
}

// TestCheckConfig_WithFailureThreshold tests the WithFailureThreshold method.
func TestCheckConfig_WithFailureThreshold(t *testing.T) {
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
			cfg := health.NewCheckConfig().WithFailureThreshold(tt.threshold)

			// Verify threshold.
			assert.Equal(t, tt.expected, cfg.FailureThreshold)
		})
	}
}

// TestCheckConfig_Validate tests configuration validation.
func TestCheckConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      health.CheckConfig
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid_default_config",
			config:      health.NewCheckConfig(),
			expectError: false,
		},
		{
			name: "invalid_timeout",
			config: health.CheckConfig{
				Timeout:          0,
				Interval:         time.Second,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			expectError: true,
			expectedErr: health.ErrInvalidTimeout,
		},
		{
			name: "invalid_interval",
			config: health.CheckConfig{
				Timeout:          time.Second,
				Interval:         0,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			expectError: true,
			expectedErr: health.ErrInvalidInterval,
		},
		{
			name: "invalid_success_threshold",
			config: health.CheckConfig{
				Timeout:          time.Second,
				Interval:         time.Second,
				SuccessThreshold: 0,
				FailureThreshold: 1,
			},
			expectError: true,
			expectedErr: health.ErrInvalidSuccessThreshold,
		},
		{
			name: "invalid_failure_threshold",
			config: health.CheckConfig{
				Timeout:          time.Second,
				Interval:         time.Second,
				SuccessThreshold: 1,
				FailureThreshold: 0,
			},
			expectError: true,
			expectedErr: health.ErrInvalidFailureThreshold,
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
