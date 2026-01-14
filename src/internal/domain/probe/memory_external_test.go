package probe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/probe"
)

func TestSystemMemory_SwapUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		mem      probe.SystemMemory
		expected float64
	}{
		{
			name: "50% swap usage",
			mem: probe.SystemMemory{
				SwapTotal: 1000,
				SwapUsed:  500,
			},
			expected: 50.0,
		},
		{
			name: "no swap",
			mem: probe.SystemMemory{
				SwapTotal: 0,
				SwapUsed:  0,
			},
			expected: 0.0,
		},
		{
			name: "100% swap usage",
			mem: probe.SystemMemory{
				SwapTotal: 1000,
				SwapUsed:  1000,
			},
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mem.SwapUsagePercent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSystemMemory_Fields(t *testing.T) {
	now := time.Now()
	mem := probe.SystemMemory{
		Total:        16 * 1024 * 1024 * 1024, // 16 GB
		Available:    8 * 1024 * 1024 * 1024,  // 8 GB
		Used:         8 * 1024 * 1024 * 1024,  // 8 GB
		Free:         2 * 1024 * 1024 * 1024,  // 2 GB
		Cached:       4 * 1024 * 1024 * 1024,  // 4 GB
		Buffers:      1 * 1024 * 1024 * 1024,  // 1 GB
		SwapTotal:    4 * 1024 * 1024 * 1024,  // 4 GB
		SwapUsed:     1 * 1024 * 1024 * 1024,  // 1 GB
		SwapFree:     3 * 1024 * 1024 * 1024,  // 3 GB
		Shared:       512 * 1024 * 1024,       // 512 MB
		UsagePercent: 50.0,
		Timestamp:    now,
	}

	assert.Equal(t, uint64(16*1024*1024*1024), mem.Total)
	assert.Equal(t, uint64(8*1024*1024*1024), mem.Available)
	assert.Equal(t, uint64(8*1024*1024*1024), mem.Used)
	assert.Equal(t, 50.0, mem.UsagePercent)
	assert.Equal(t, now, mem.Timestamp)
}

func TestProcessMemory_TotalResident(t *testing.T) {
	proc := probe.ProcessMemory{
		PID:  1234,
		Name: "test",
		RSS:  100 * 1024 * 1024, // 100 MB
		Swap: 10 * 1024 * 1024,  // 10 MB
	}

	total := proc.TotalResident()
	assert.Equal(t, uint64(110*1024*1024), total)
}

func TestProcessMemory_Fields(t *testing.T) {
	now := time.Now()
	proc := probe.ProcessMemory{
		PID:          1234,
		Name:         "myprocess",
		RSS:          100 * 1024 * 1024,
		VMS:          500 * 1024 * 1024,
		Shared:       20 * 1024 * 1024,
		Swap:         5 * 1024 * 1024,
		Data:         50 * 1024 * 1024,
		Stack:        8 * 1024 * 1024,
		UsagePercent: 2.5,
		Timestamp:    now,
	}

	assert.Equal(t, 1234, proc.PID)
	assert.Equal(t, "myprocess", proc.Name)
	assert.Equal(t, uint64(100*1024*1024), proc.RSS)
	assert.Equal(t, uint64(500*1024*1024), proc.VMS)
	assert.Equal(t, 2.5, proc.UsagePercent)
	assert.Equal(t, now, proc.Timestamp)
}
