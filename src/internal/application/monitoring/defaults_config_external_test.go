package monitoring_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/application/monitoring"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultsConfig(t *testing.T) {
	// testCase defines a test case for NewDefaultsConfig.
	type testCase struct {
		name string
	}

	// tests defines all test cases for NewDefaultsConfig.
	tests := []testCase{
		{name: "creates config with default values"},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create new defaults config.
			config := monitoring.NewDefaultsConfig()

			// Verify defaults are set.
			assert.Equal(t, monitoring.DefaultInterval, config.Interval)
			assert.Equal(t, monitoring.DefaultTimeout, config.Timeout)
			assert.Equal(t, monitoring.DefaultSuccessThreshold, config.SuccessThreshold)
			assert.Equal(t, monitoring.DefaultFailureThreshold, config.FailureThreshold)
		})
	}
}
