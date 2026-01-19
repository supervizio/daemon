// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewCPUPressure tests the NewCPUPressure constructor.
func TestNewCPUPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		someAvg10  float64
		someAvg60  float64
		someAvg300 float64
		someTotal  uint64
		timestamp  time.Time
	}{
		{
			name:       "all_fields_populated",
			someAvg10:  5.0,
			someAvg60:  3.0,
			someAvg300: 2.0,
			someTotal:  1000000,
			timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:       "zero_values",
			someAvg10:  0.0,
			someAvg60:  0.0,
			someAvg300: 0.0,
			someTotal:  0,
			timestamp:  time.Time{},
		},
		{
			name:       "high_pressure_values",
			someAvg10:  50.0,
			someAvg60:  40.0,
			someAvg300: 30.0,
			someTotal:  5000000,
			timestamp:  time.Now(),
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create CPUPressure using constructor.
			pressure := metrics.NewCPUPressure(tt.someAvg10, tt.someAvg60, tt.someAvg300, tt.someTotal, tt.timestamp)

			// Verify all fields are correctly set.
			assert.Equal(t, tt.someAvg10, pressure.SomeAvg10)
			assert.Equal(t, tt.someAvg60, pressure.SomeAvg60)
			assert.Equal(t, tt.someAvg300, pressure.SomeAvg300)
			assert.Equal(t, tt.someTotal, pressure.SomeTotal)
			assert.Equal(t, tt.timestamp, pressure.Timestamp)
		})
	}
}

// TestCPUPressure_IsUnderPressure tests the IsUnderPressure method on CPUPressure.
func TestCPUPressure_IsUnderPressure(t *testing.T) {
	t.Parallel()

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
			name: "cpu_pressure_above_threshold",
			pressure: metrics.CPUPressure{
				SomeAvg10:  50.0,
				SomeAvg60:  30.0,
				SomeAvg300: 20.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
		{
			name: "cpu_pressure_exactly_at_threshold",
			pressure: metrics.CPUPressure{
				SomeAvg10:  10.0,
				SomeAvg60:  5.0,
				SomeAvg300: 3.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: false,
		},
		{
			name: "cpu_pressure_just_above_threshold",
			pressure: metrics.CPUPressure{
				SomeAvg10:  10.1,
				SomeAvg60:  5.0,
				SomeAvg300: 3.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Check IsUnderPressure result.
			assert.Equal(t, tt.wantUnderPressure, tt.pressure.IsUnderPressure())
		})
	}
}
