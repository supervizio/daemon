// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestIOStats tests IOStats value object methods.
func TestIOStats(t *testing.T) {
	tests := []struct {
		name           string
		stats          metrics.IOStats
		wantTotalBytes uint64
		wantTotalOps   uint64
	}{
		{
			name: "system_io_stats",
			stats: metrics.IOStats{
				ReadBytesTotal:  1024 * 1024 * 100, // 100MB read
				WriteBytesTotal: 1024 * 1024 * 50,  // 50MB written
				ReadOpsTotal:    10000,
				WriteOpsTotal:   5000,
				Timestamp:       time.Now(),
			},
			wantTotalBytes: 1024 * 1024 * 150,
			wantTotalOps:   15000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalBytes, tt.stats.TotalBytes())
			assert.Equal(t, tt.wantTotalOps, tt.stats.TotalOps())
		})
	}
}

// TestIOPressure tests IOPressure value object methods.
func TestIOPressure(t *testing.T) {
	tests := []struct {
		name              string
		pressure          metrics.IOPressure
		wantUnderPressure bool
	}{
		{
			name: "low_pressure",
			pressure: metrics.IOPressure{
				SomeAvg10:  5.0,
				SomeAvg60:  3.0,
				SomeAvg300: 2.0,
				FullAvg10:  1.0,
				FullAvg60:  0.5,
				FullAvg300: 0.2,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: false,
		},
		{
			name: "high_some_pressure",
			pressure: metrics.IOPressure{
				SomeAvg10:  15.0, // Above threshold
				SomeAvg60:  10.0,
				SomeAvg300: 5.0,
				FullAvg10:  5.0,
				FullAvg60:  3.0,
				FullAvg300: 1.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
		{
			name: "high_full_pressure",
			pressure: metrics.IOPressure{
				SomeAvg10:  8.0,
				SomeAvg60:  5.0,
				SomeAvg300: 3.0,
				FullAvg10:  12.0, // Above threshold
				FullAvg60:  8.0,
				FullAvg300: 4.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantUnderPressure, tt.pressure.IsUnderPressure())
		})
	}
}

// TestMemoryPressure tests MemoryPressure value object methods.
func TestMemoryPressure(t *testing.T) {
	tests := []struct {
		name              string
		pressure          metrics.MemoryPressure
		wantUnderPressure bool
	}{
		{
			name: "no_memory_pressure",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  2.0,
				SomeAvg60:  1.0,
				SomeAvg300: 0.5,
				FullAvg10:  0.5,
				FullAvg60:  0.2,
				FullAvg300: 0.1,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: false,
		},
		{
			name: "memory_pressure",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  25.0, // High pressure
				SomeAvg60:  15.0,
				SomeAvg300: 10.0,
				FullAvg10:  8.0,
				FullAvg60:  5.0,
				FullAvg300: 3.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantUnderPressure, tt.pressure.IsUnderPressure())
		})
	}
}

// TestCPUPressure tests CPUPressure value object methods.
func TestCPUPressure(t *testing.T) {
	tests := []struct {
		name              string
		pressure          metrics.CPUPressure
		wantUnderPressure bool
	}{
		{
			name: "no_cpu_pressure",
			pressure: metrics.CPUPressure{
				SomeAvg10:  5.0,
				SomeAvg60:  3.0,
				SomeAvg300: 2.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: false,
		},
		{
			name: "cpu_pressure",
			pressure: metrics.CPUPressure{
				SomeAvg10:  50.0, // High CPU contention
				SomeAvg60:  30.0,
				SomeAvg300: 20.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantUnderPressure, tt.pressure.IsUnderPressure())
		})
	}
}

// TestLoadAverage tests LoadAverage value object methods.
func TestLoadAverage(t *testing.T) {
	tests := []struct {
		name           string
		load           metrics.LoadAverage
		numCPU         int
		wantOverloaded bool
	}{
		{
			name: "normal_load_4_cpu",
			load: metrics.LoadAverage{
				Load1:            2.0,
				Load5:            1.5,
				Load15:           1.0,
				RunningProcesses: 5,
				TotalProcesses:   200,
				LastPID:          12345,
				Timestamp:        time.Now(),
			},
			numCPU:         4,
			wantOverloaded: false,
		},
		{
			name: "overloaded_4_cpu",
			load: metrics.LoadAverage{
				Load1:            8.0, // 2x the CPU count
				Load5:            6.0,
				Load15:           4.0,
				RunningProcesses: 20,
				TotalProcesses:   300,
				LastPID:          12346,
				Timestamp:        time.Now(),
			},
			numCPU:         4,
			wantOverloaded: true,
		},
		{
			name: "exactly_at_cpu_count",
			load: metrics.LoadAverage{
				Load1:            4.0,
				Load5:            3.0,
				Load15:           2.0,
				RunningProcesses: 8,
				TotalProcesses:   250,
				LastPID:          12347,
				Timestamp:        time.Now(),
			},
			numCPU:         4,
			wantOverloaded: false,
		},
		{
			name: "just_over_cpu_count",
			load: metrics.LoadAverage{
				Load1:            4.1,
				Load5:            3.0,
				Load15:           2.0,
				RunningProcesses: 8,
				TotalProcesses:   250,
				LastPID:          12348,
				Timestamp:        time.Now(),
			},
			numCPU:         4,
			wantOverloaded: true,
		},
		{
			name: "zero_cpu_defaults_to_one",
			load: metrics.LoadAverage{
				Load1:     2.0,
				Load5:     1.5,
				Load15:    1.0,
				Timestamp: time.Now(),
			},
			numCPU:         0,
			wantOverloaded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantOverloaded, tt.load.IsOverloaded(tt.numCPU))
		})
	}
}
