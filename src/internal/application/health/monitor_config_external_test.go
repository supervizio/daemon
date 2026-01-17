// Package health_test provides black-box tests for monitor_config.go.
package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphealth "github.com/kodflow/daemon/internal/application/health"
)

// TestNewProbeMonitorConfig tests the NewProbeMonitorConfig constructor.
func TestNewProbeMonitorConfig(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// factory is the factory to use.
		factory apphealth.Creator
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
			config := apphealth.NewProbeMonitorConfig(tt.factory)

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
	mockFactory := &mockCreator{}

	tests := []struct {
		// name is the test case name.
		name string
		// config is the configuration to test.
		config apphealth.ProbeMonitorConfig
		// hasFactory indicates whether the config has a factory.
		hasFactory bool
	}{
		{
			name:       "empty_config",
			config:     apphealth.ProbeMonitorConfig{},
			hasFactory: false,
		},
		{
			name: "config_with_timeouts",
			config: apphealth.ProbeMonitorConfig{
				DefaultTimeout:  5 * time.Second,
				DefaultInterval: 10 * time.Second,
			},
			hasFactory: false,
		},
		{
			name: "config_with_factory",
			config: apphealth.ProbeMonitorConfig{
				Factory: mockFactory,
			},
			hasFactory: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor from config.
			monitor := apphealth.NewProbeMonitor(tt.config)

			// Verify monitor was created successfully.
			require.NotNil(t, monitor)
			// Verify initial status is unhealthy (process not running).
			assert.False(t, monitor.IsHealthy())
			// Verify factory configuration.
			if tt.hasFactory {
				assert.Equal(t, mockFactory, tt.config.Factory)
			} else {
				assert.Nil(t, tt.config.Factory)
			}
		})
	}
}
