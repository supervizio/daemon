//go:build linux

package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_cpuSample_total(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sample   cpuSample
		expected uint64
	}{
		{"zero_values", cpuSample{}, 0},
		{"all_fields", cpuSample{user: 100, nice: 10, system: 50, idle: 200, iowait: 5, irq: 2, softirq: 3, steal: 1}, 371},
		{"idle_only", cpuSample{idle: 500}, 500},
		{"busy_only", cpuSample{user: 100, system: 50}, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.sample.total())
		})
	}
}

func Test_cpuSample_busy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sample   cpuSample
		expected uint64
	}{
		{"zero_values", cpuSample{}, 0},
		{"all_fields", cpuSample{user: 100, nice: 10, system: 50, idle: 200, iowait: 5, irq: 2, softirq: 3, steal: 1}, 166},
		{"idle_excluded", cpuSample{user: 100, idle: 200}, 100},
		{"iowait_excluded", cpuSample{user: 50, iowait: 30}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.sample.busy())
		})
	}
}
