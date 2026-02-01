// Package config_test provides black-box tests for probe configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
)

// TestNewProbeConfig tests the NewProbeConfig constructor function.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that NewProbeConfig creates a properly initialized
// ProbeConfig with correct default values for all probe types.
func TestNewProbeConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		probeType string
	}{
		{
			name:      "tcp_probe",
			probeType: config.ProbeTypeTCP,
		},
		{
			name:      "udp_probe",
			probeType: config.ProbeTypeUDP,
		},
		{
			name:      "http_probe",
			probeType: config.ProbeTypeHTTP,
		},
		{
			name:      "grpc_probe",
			probeType: config.ProbeTypeGRPC,
		},
		{
			name:      "exec_probe",
			probeType: config.ProbeTypeExec,
		},
		{
			name:      "icmp_probe",
			probeType: config.ProbeTypeICMP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create probe config.
			cfg := config.NewProbeConfig(tt.probeType)

			// Verify type is set correctly.
			assert.Equal(t, tt.probeType, cfg.Type)
			// Verify default interval is set.
			assert.Greater(t, int64(cfg.Interval), int64(0))
			// Verify default timeout is set.
			assert.Greater(t, int64(cfg.Timeout), int64(0))
			// Verify thresholds are set.
			assert.GreaterOrEqual(t, cfg.SuccessThreshold, 1)
			assert.GreaterOrEqual(t, cfg.FailureThreshold, 1)
			// Verify HTTP defaults.
			assert.NotEmpty(t, cfg.Method)
			assert.Equal(t, 200, cfg.StatusCode)
			// Verify ICMP default mode.
			assert.Equal(t, config.ICMPModeAuto, cfg.ICMPMode)
		})
	}
}

// TestDefaultProbeConfig tests the DefaultProbeConfig constructor function.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that DefaultProbeConfig returns the same values
// as NewProbeConfig for all probe types.
func TestDefaultProbeConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		probeType string
	}{
		{
			name:      "tcp_default",
			probeType: config.ProbeTypeTCP,
		},
		{
			name:      "http_default",
			probeType: config.ProbeTypeHTTP,
		},
		{
			name:      "icmp_default",
			probeType: config.ProbeTypeICMP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create default probe config.
			cfg := config.DefaultProbeConfig(tt.probeType)

			// Verify type is set correctly.
			assert.Equal(t, tt.probeType, cfg.Type)
			// Verify defaults match NewProbeConfig.
			expected := config.NewProbeConfig(tt.probeType)
			assert.Equal(t, expected.Interval, cfg.Interval)
			assert.Equal(t, expected.Timeout, cfg.Timeout)
			assert.Equal(t, expected.SuccessThreshold, cfg.SuccessThreshold)
			assert.Equal(t, expected.FailureThreshold, cfg.FailureThreshold)
			assert.Equal(t, expected.Method, cfg.Method)
			assert.Equal(t, expected.StatusCode, cfg.StatusCode)
			assert.Equal(t, expected.ICMPMode, cfg.ICMPMode)
		})
	}
}

// TestICMPMode_constants tests ICMP mode constant values.
func TestICMPMode_constants(t *testing.T) {
	tests := []struct {
		name     string
		mode     config.ICMPMode
		expected string
	}{
		{
			name:     "native_mode",
			mode:     config.ICMPModeNative,
			expected: "native",
		},
		{
			name:     "fallback_mode",
			mode:     config.ICMPModeFallback,
			expected: "fallback",
		},
		{
			name:     "auto_mode",
			mode:     config.ICMPModeAuto,
			expected: "auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant value.
			assert.Equal(t, tt.expected, string(tt.mode))
		})
	}
}

// TestNewProbeConfig_ICMPMode tests default ICMP mode in probe config.
func TestNewProbeConfig_ICMPMode(t *testing.T) {
	tests := []struct {
		name         string
		probeType    string
		expectedMode config.ICMPMode
	}{
		{
			name:         "icmp_probe_defaults_to_auto",
			probeType:    config.ProbeTypeICMP,
			expectedMode: config.ICMPModeAuto,
		},
		{
			name:         "tcp_probe_has_auto_mode",
			probeType:    config.ProbeTypeTCP,
			expectedMode: config.ICMPModeAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create probe config.
			cfg := config.NewProbeConfig(tt.probeType)

			// Verify default ICMP mode.
			assert.Equal(t, tt.expectedMode, cfg.ICMPMode)
		})
	}
}

// TestProbeConfig_ICMPMode_customization tests setting custom ICMP mode.
func TestProbeConfig_ICMPMode_customization(t *testing.T) {
	tests := []struct {
		name       string
		probeType  string
		customMode config.ICMPMode
	}{
		{
			name:       "set_native_mode",
			probeType:  config.ProbeTypeICMP,
			customMode: config.ICMPModeNative,
		},
		{
			name:       "set_fallback_mode",
			probeType:  config.ProbeTypeICMP,
			customMode: config.ICMPModeFallback,
		},
		{
			name:       "set_auto_mode",
			probeType:  config.ProbeTypeICMP,
			customMode: config.ICMPModeAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create probe config.
			cfg := config.NewProbeConfig(tt.probeType)

			// Set custom ICMP mode.
			cfg.ICMPMode = tt.customMode

			// Verify custom mode is set.
			assert.Equal(t, tt.customMode, cfg.ICMPMode)
		})
	}
}

// TestDefaultProbeConfig_ICMPMode tests default ICMP mode via DefaultProbeConfig.
func TestDefaultProbeConfig_ICMPMode(t *testing.T) {
	tests := []struct {
		name         string
		probeType    string
		expectedMode config.ICMPMode
	}{
		{
			name:         "default_config_auto_mode",
			probeType:    config.ProbeTypeICMP,
			expectedMode: config.ICMPModeAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create default probe config.
			cfg := config.DefaultProbeConfig(tt.probeType)

			// Verify default ICMP mode.
			assert.Equal(t, tt.expectedMode, cfg.ICMPMode)
		})
	}
}
