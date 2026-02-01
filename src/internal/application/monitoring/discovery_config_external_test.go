package monitoring_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/application/monitoring"
	"github.com/stretchr/testify/assert"
)

func TestNewDiscoveryModeConfig(t *testing.T) {
	// testCase defines a test case for NewDiscoveryModeConfig.
	type testCase struct {
		name string
	}

	// tests defines all test cases for NewDiscoveryModeConfig.
	tests := []testCase{
		{name: "creates config with defaults"},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create new discovery config.
			config := monitoring.NewDiscoveryModeConfig()

			// Verify defaults are set.
			assert.False(t, config.Enabled)
			assert.Equal(t, monitoring.DefaultDiscoveryInterval, config.Interval)
			assert.Nil(t, config.Discoverers)
			assert.Nil(t, config.Watchers)
		})
	}
}
