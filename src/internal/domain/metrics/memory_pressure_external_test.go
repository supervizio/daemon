// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewMemoryPressure tests the NewMemoryPressure constructor.
func TestNewMemoryPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *metrics.MemoryPressureParams
	}{
		{
			name: "all_fields_populated",
			params: &metrics.MemoryPressureParams{
				SomeAvg10:  5.0,
				SomeAvg60:  3.0,
				SomeAvg300: 2.0,
				SomeTotal:  1000000,
				FullAvg10:  1.0,
				FullAvg60:  0.5,
				FullAvg300: 0.2,
				FullTotal:  500000,
				Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "zero_values",
			params: &metrics.MemoryPressureParams{
				SomeAvg10:  0.0,
				SomeAvg60:  0.0,
				SomeAvg300: 0.0,
				SomeTotal:  0,
				FullAvg10:  0.0,
				FullAvg60:  0.0,
				FullAvg300: 0.0,
				FullTotal:  0,
				Timestamp:  time.Time{},
			},
		},
		{
			name: "high_pressure_values",
			params: &metrics.MemoryPressureParams{
				SomeAvg10:  50.0,
				SomeAvg60:  40.0,
				SomeAvg300: 30.0,
				SomeTotal:  5000000,
				FullAvg10:  25.0,
				FullAvg60:  20.0,
				FullAvg300: 15.0,
				FullTotal:  2500000,
				Timestamp:  time.Now(),
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create MemoryPressure using constructor.
			pressure := metrics.NewMemoryPressure(tt.params)

			// Verify all fields are correctly set.
			assert.Equal(t, tt.params.SomeAvg10, pressure.SomeAvg10)
			assert.Equal(t, tt.params.SomeAvg60, pressure.SomeAvg60)
			assert.Equal(t, tt.params.SomeAvg300, pressure.SomeAvg300)
			assert.Equal(t, tt.params.SomeTotal, pressure.SomeTotal)
			assert.Equal(t, tt.params.FullAvg10, pressure.FullAvg10)
			assert.Equal(t, tt.params.FullAvg60, pressure.FullAvg60)
			assert.Equal(t, tt.params.FullAvg300, pressure.FullAvg300)
			assert.Equal(t, tt.params.FullTotal, pressure.FullTotal)
			assert.Equal(t, tt.params.Timestamp, pressure.Timestamp)
		})
	}
}

// TestMemoryPressure_IsUnderPressure tests the IsUnderPressure method on MemoryPressure.
func TestMemoryPressure_IsUnderPressure(t *testing.T) {
	t.Parallel()

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
			name: "high_some_pressure",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  25.0,
				SomeAvg60:  15.0,
				SomeAvg300: 10.0,
				FullAvg10:  8.0,
				FullAvg60:  5.0,
				FullAvg300: 3.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
		{
			name: "high_full_pressure",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  5.0,
				SomeAvg60:  3.0,
				SomeAvg300: 2.0,
				FullAvg10:  15.0,
				FullAvg60:  10.0,
				FullAvg300: 8.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
		{
			name: "exactly_at_threshold",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  10.0,
				SomeAvg60:  5.0,
				SomeAvg300: 3.0,
				FullAvg10:  10.0,
				FullAvg60:  5.0,
				FullAvg300: 3.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: false,
		},
		{
			name: "some_just_above_threshold",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  10.1,
				SomeAvg60:  5.0,
				SomeAvg300: 3.0,
				FullAvg10:  5.0,
				FullAvg60:  3.0,
				FullAvg300: 1.0,
				Timestamp:  time.Now(),
			},
			wantUnderPressure: true,
		},
		{
			name: "full_just_above_threshold",
			pressure: metrics.MemoryPressure{
				SomeAvg10:  5.0,
				SomeAvg60:  3.0,
				SomeAvg300: 2.0,
				FullAvg10:  10.1,
				FullAvg60:  5.0,
				FullAvg300: 3.0,
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
