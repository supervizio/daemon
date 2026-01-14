// Package health provides internal tests for probe_monitor_config.go.
// It tests internal implementation details using white-box testing.
package healthcheck

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domain "github.com/kodflow/daemon/internal/domain/health"
)

// Test_ProbeMonitorConfig_struct tests the ProbeMonitorConfig struct fields.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitorConfig_struct(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// factory is the prober factory.
		factory Creator
		// events is the events channel.
		events chan<- domain.Event
		// defaultTimeout is the default timeout.
		defaultTimeout time.Duration
		// defaultInterval is the default interval.
		defaultInterval time.Duration
	}{
		{
			name:            "empty_config",
			factory:         nil,
			events:          nil,
			defaultTimeout:  0,
			defaultInterval: 0,
		},
		{
			name:            "config_with_timeout",
			factory:         nil,
			events:          nil,
			defaultTimeout:  5 * time.Second,
			defaultInterval: 0,
		},
		{
			name:            "config_with_interval",
			factory:         nil,
			events:          nil,
			defaultTimeout:  0,
			defaultInterval: 10 * time.Second,
		},
		{
			name:            "config_with_all_fields",
			factory:         &internalTestCreator{},
			events:          make(chan domain.Event, 1),
			defaultTimeout:  5 * time.Second,
			defaultInterval: 10 * time.Second,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			config := ProbeMonitorConfig{
				Factory:         tt.factory,
				Events:          tt.events,
				DefaultTimeout:  tt.defaultTimeout,
				DefaultInterval: tt.defaultInterval,
			}

			// Verify factory field.
			assert.Equal(t, tt.factory, config.Factory)
			// Verify events field.
			assert.Equal(t, tt.events, config.Events)
			// Verify default timeout field.
			assert.Equal(t, tt.defaultTimeout, config.DefaultTimeout)
			// Verify default interval field.
			assert.Equal(t, tt.defaultInterval, config.DefaultInterval)
		})
	}
}
