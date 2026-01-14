// Package healthcheck_test provides black-box tests for probe_monitor_config.go.
package healthcheck_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apphc "github.com/kodflow/daemon/internal/application/healthcheck"
)

// TestNewProbeMonitorConfig tests the NewProbeMonitorConfig constructor.
func TestNewProbeMonitorConfig(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// factory is the factory to use.
		factory apphc.Creator
	}{
		{
			name:    "creates_with_nil_factory",
			factory: nil,
		},
		{
			name:    "creates_with_valid_factory",
			factory: &mockCreator{},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			config := apphc.NewProbeMonitorConfig(tt.factory)

			// Verify factory is set.
			assert.Equal(t, tt.factory, config.Factory)
			// Verify events channel is nil by default.
			assert.Nil(t, config.Events)
			// Verify default timeout is zero by default.
			assert.Equal(t, time.Duration(0), config.DefaultTimeout)
			// Verify default interval is zero by default.
			assert.Equal(t, time.Duration(0), config.DefaultInterval)
		})
	}
}

// TestProbeMonitorConfig_struct tests the ProbeMonitorConfig struct fields.
func TestProbeMonitorConfig_struct(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// config is the configuration to test.
		config apphc.ProbeMonitorConfig
	}{
		{
			name:   "empty_config",
			config: apphc.ProbeMonitorConfig{},
		},
		{
			name: "config_with_timeouts",
			config: apphc.ProbeMonitorConfig{
				DefaultTimeout:  5 * time.Second,
				DefaultInterval: 10 * time.Second,
			},
		},
		{
			name: "config_with_factory",
			config: apphc.ProbeMonitorConfig{
				Factory: &mockCreator{},
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify config can be created without panic.
			assert.NotPanics(t, func() {
				_ = apphc.NewProbeMonitor(tt.config)
			})
		})
	}
}
