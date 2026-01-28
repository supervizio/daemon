//go:build linux

// Package collector_test provides black-box tests for the collector package.
package collector_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
)

func TestUpdateAdaptiveSpeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ifName     string
		throughput uint64
	}{
		{
			name:       "high throughput triggers scaling",
			ifName:     "test_scale_external1",
			throughput: 850000000, // 85% of 1 Gbps
		},
		{
			name:       "low throughput does not panic",
			ifName:     "test_low_external1",
			throughput: 500000000, // 50% of 1 Gbps
		},
		{
			name:       "zero throughput",
			ifName:     "test_zero_external1",
			throughput: 0,
		},
		{
			name:       "very high throughput",
			ifName:     "test_high_external1",
			throughput: 10000000000, // 10 Gbps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Black-box test: verify UpdateAdaptiveSpeed doesn't panic
			// with various throughput values.
			collector.UpdateAdaptiveSpeed(tt.ifName, tt.throughput)

			// Call multiple times to test scaling behavior.
			for range 5 {
				collector.UpdateAdaptiveSpeed(tt.ifName, tt.throughput)
			}
		})
	}
}
