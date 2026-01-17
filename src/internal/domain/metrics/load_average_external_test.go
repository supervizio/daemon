// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewLoadAverage tests the NewLoadAverage constructor.
func TestNewLoadAverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params metrics.LoadAverageParams
	}{
		{
			name: "all_fields_populated",
			params: metrics.LoadAverageParams{
				Load1:            2.5,
				Load5:            1.5,
				Load15:           1.0,
				RunningProcesses: 10,
				TotalProcesses:   200,
				LastPID:          12345,
				Timestamp:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "zero_values",
			params: metrics.LoadAverageParams{
				Load1:            0.0,
				Load5:            0.0,
				Load15:           0.0,
				RunningProcesses: 0,
				TotalProcesses:   0,
				LastPID:          0,
				Timestamp:        time.Time{},
			},
		},
		{
			name: "high_load_values",
			params: metrics.LoadAverageParams{
				Load1:            16.0,
				Load5:            12.0,
				Load15:           8.0,
				RunningProcesses: 50,
				TotalProcesses:   500,
				LastPID:          99999,
				Timestamp:        time.Now(),
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create LoadAverage using constructor.
			load := metrics.NewLoadAverage(tt.params)

			// Verify all fields are correctly set.
			assert.Equal(t, tt.params.Load1, load.Load1)
			assert.Equal(t, tt.params.Load5, load.Load5)
			assert.Equal(t, tt.params.Load15, load.Load15)
			assert.Equal(t, tt.params.RunningProcesses, load.RunningProcesses)
			assert.Equal(t, tt.params.TotalProcesses, load.TotalProcesses)
			assert.Equal(t, tt.params.LastPID, load.LastPID)
			assert.Equal(t, tt.params.Timestamp, load.Timestamp)
		})
	}
}

// TestLoadAverage_IsOverloaded tests the IsOverloaded method on LoadAverage.
func TestLoadAverage_IsOverloaded(t *testing.T) {
	t.Parallel()

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
				Load1:            8.0,
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
		{
			name: "negative_cpu_defaults_to_one",
			load: metrics.LoadAverage{
				Load1:     2.0,
				Load5:     1.5,
				Load15:    1.0,
				Timestamp: time.Now(),
			},
			numCPU:         -1,
			wantOverloaded: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Check IsOverloaded result.
			assert.Equal(t, tt.wantOverloaded, tt.load.IsOverloaded(tt.numCPU))
		})
	}
}
