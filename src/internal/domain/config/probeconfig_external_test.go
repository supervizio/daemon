// Package config_test provides black-box tests for probe configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
)

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
