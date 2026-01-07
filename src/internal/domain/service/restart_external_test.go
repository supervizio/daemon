// Package service provides domain value objects for service configuration.
package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestRestartPolicy_String tests the String method of RestartPolicy.
//
// Params:
//   - t: testing context
//
// Test cases verify string representation for all restart policies.
func TestRestartPolicy_String(t *testing.T) {
	tests := []struct {
		name   string
		policy service.RestartPolicy
		want   string
	}{
		{"always", service.RestartAlways, "always"},
		{"on-failure", service.RestartOnFailure, "on-failure"},
		{"never", service.RestartNever, "never"},
		{"unless-stopped", service.RestartUnless, "unless-stopped"},
	}

	// Iterate through all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.policy.String())
		})
	}
}

// TestRestartConfig_ShouldRestartOnExit tests the ShouldRestartOnExit method.
//
// Params:
//   - t: testing context
//
// Test cases cover all restart policies and edge conditions.
func TestRestartConfig_ShouldRestartOnExit(t *testing.T) {
	tests := []struct {
		name       string
		cfg        service.RestartConfig
		exitCode   int
		retryCount int
		want       bool
	}{
		// RestartAlways policy tests
		{
			name:       "always_restart_on_success",
			cfg:        service.RestartConfig{Policy: service.RestartAlways, MaxRetries: 3},
			exitCode:   0,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "always_restart_on_failure",
			cfg:        service.RestartConfig{Policy: service.RestartAlways, MaxRetries: 3},
			exitCode:   1,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "always_restart_on_exit_127",
			cfg:        service.RestartConfig{Policy: service.RestartAlways, MaxRetries: 3},
			exitCode:   127,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "always_stop_after_max_retries_success",
			cfg:        service.RestartConfig{Policy: service.RestartAlways, MaxRetries: 3},
			exitCode:   0,
			retryCount: 3,
			want:       false,
		},
		{
			name:       "always_stop_after_max_retries_failure",
			cfg:        service.RestartConfig{Policy: service.RestartAlways, MaxRetries: 3},
			exitCode:   1,
			retryCount: 3,
			want:       false,
		},
		// RestartOnFailure policy tests
		{
			name:       "on_failure_no_restart_on_success",
			cfg:        service.RestartConfig{Policy: service.RestartOnFailure, MaxRetries: 3},
			exitCode:   0,
			retryCount: 0,
			want:       false,
		},
		{
			name:       "on_failure_no_restart_on_success_with_retries",
			cfg:        service.RestartConfig{Policy: service.RestartOnFailure, MaxRetries: 3},
			exitCode:   0,
			retryCount: 1,
			want:       false,
		},
		{
			name:       "on_failure_restart_on_exit_1",
			cfg:        service.RestartConfig{Policy: service.RestartOnFailure, MaxRetries: 3},
			exitCode:   1,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "on_failure_restart_on_exit_127",
			cfg:        service.RestartConfig{Policy: service.RestartOnFailure, MaxRetries: 3},
			exitCode:   127,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "on_failure_stop_after_max_retries",
			cfg:        service.RestartConfig{Policy: service.RestartOnFailure, MaxRetries: 3},
			exitCode:   1,
			retryCount: 3,
			want:       false,
		},
		// RestartNever policy tests
		{
			name:       "never_no_restart_on_success",
			cfg:        service.RestartConfig{Policy: service.RestartNever, MaxRetries: 3},
			exitCode:   0,
			retryCount: 0,
			want:       false,
		},
		{
			name:       "never_no_restart_on_exit_1",
			cfg:        service.RestartConfig{Policy: service.RestartNever, MaxRetries: 3},
			exitCode:   1,
			retryCount: 0,
			want:       false,
		},
		{
			name:       "never_no_restart_on_exit_127",
			cfg:        service.RestartConfig{Policy: service.RestartNever, MaxRetries: 3},
			exitCode:   127,
			retryCount: 0,
			want:       false,
		},
		// RestartUnless policy tests
		{
			name:       "unless_restart_on_success",
			cfg:        service.RestartConfig{Policy: service.RestartUnless, MaxRetries: 3},
			exitCode:   0,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "unless_restart_on_failure",
			cfg:        service.RestartConfig{Policy: service.RestartUnless, MaxRetries: 3},
			exitCode:   1,
			retryCount: 0,
			want:       true,
		},
		{
			name:       "unless_ignores_max_retries",
			cfg:        service.RestartConfig{Policy: service.RestartUnless, MaxRetries: 3},
			exitCode:   0,
			retryCount: 100,
			want:       true,
		},
		// Unknown policy tests
		{
			name:       "unknown_policy_no_restart_on_success",
			cfg:        service.RestartConfig{Policy: "unknown", MaxRetries: 3},
			exitCode:   0,
			retryCount: 0,
			want:       false,
		},
		{
			name:       "unknown_policy_no_restart_on_failure",
			cfg:        service.RestartConfig{Policy: "unknown", MaxRetries: 3},
			exitCode:   1,
			retryCount: 0,
			want:       false,
		},
	}

	// Iterate through all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.ShouldRestartOnExit(tt.exitCode, tt.retryCount)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNewRestartConfig tests the NewRestartConfig constructor function.
//
// Params:
//   - t: testing context
//
// Test cases verify the NewRestartConfig function creates correct configurations.
func TestNewRestartConfig(t *testing.T) {
	tests := []struct {
		name        string
		policy      service.RestartPolicy
		wantPolicy  service.RestartPolicy
		wantRetries int
		wantDelay   shared.Duration
	}{
		{
			name:        "creates config with always policy",
			policy:      service.RestartAlways,
			wantPolicy:  service.RestartAlways,
			wantRetries: 3,
			wantDelay:   shared.Seconds(5),
		},
		{
			name:        "creates config with on-failure policy",
			policy:      service.RestartOnFailure,
			wantPolicy:  service.RestartOnFailure,
			wantRetries: 3,
			wantDelay:   shared.Seconds(5),
		},
		{
			name:        "creates config with never policy",
			policy:      service.RestartNever,
			wantPolicy:  service.RestartNever,
			wantRetries: 3,
			wantDelay:   shared.Seconds(5),
		},
		{
			name:        "creates config with unless-stopped policy",
			policy:      service.RestartUnless,
			wantPolicy:  service.RestartUnless,
			wantRetries: 3,
			wantDelay:   shared.Seconds(5),
		},
	}

	// Iterate through all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.NewRestartConfig(tt.policy)
			assert.Equal(t, tt.wantPolicy, cfg.Policy)
			assert.Equal(t, tt.wantRetries, cfg.MaxRetries)
			assert.Equal(t, tt.wantDelay, cfg.Delay)
		})
	}
}

// TestDefaultRestartConfig tests the DefaultRestartConfig function.
//
// Params:
//   - t: testing context
//
// Test cases verify default values for restart configuration.
func TestDefaultRestartConfig(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected any
		actual   any
	}{
		{
			name:     "default_policy",
			field:    "Policy",
			expected: service.RestartOnFailure,
			actual:   service.DefaultRestartConfig().Policy,
		},
		{
			name:     "default_max_retries",
			field:    "MaxRetries",
			expected: 3,
			actual:   service.DefaultRestartConfig().MaxRetries,
		},
		{
			name:     "default_delay",
			field:    "Delay",
			expected: shared.Seconds(5),
			actual:   service.DefaultRestartConfig().Delay,
		},
	}

	// Iterate through all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.actual)
		})
	}
}
