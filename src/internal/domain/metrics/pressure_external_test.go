// Package metrics_test provides external tests for the metrics domain package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/stretchr/testify/assert"
)

// TestPressure_IsUnderPressure tests the IsUnderPressure method.
func TestPressure_IsUnderPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pressure metrics.Pressure
		expected bool
	}{
		{
			name:     "no pressure when below threshold",
			pressure: metrics.Pressure{SomeAvg10: 5.0, FullAvg10: 5.0},
			expected: false,
		},
		{
			name:     "under pressure when SomeAvg10 exceeds threshold",
			pressure: metrics.Pressure{SomeAvg10: 15.0, FullAvg10: 5.0},
			expected: true,
		},
		{
			name:     "under pressure when FullAvg10 exceeds threshold",
			pressure: metrics.Pressure{SomeAvg10: 5.0, FullAvg10: 15.0},
			expected: true,
		},
		{
			name:     "under pressure when both exceed threshold",
			pressure: metrics.Pressure{SomeAvg10: 15.0, FullAvg10: 15.0},
			expected: true,
		},
		{
			name:     "no pressure at exact threshold",
			pressure: metrics.Pressure{SomeAvg10: 10.0, FullAvg10: 10.0},
			expected: false,
		},
		{
			name:     "zero values not under pressure",
			pressure: metrics.Pressure{},
			expected: false,
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.pressure.IsUnderPressure()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestPressureParams_ToPressure tests the ToPressure conversion method.
func TestPressureParams_ToPressure(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name   string
		params metrics.PressureParams
	}{
		{
			name: "converts all fields correctly",
			params: metrics.PressureParams{
				SomeAvg10:  1.5,
				SomeAvg60:  2.5,
				SomeAvg300: 3.5,
				SomeTotal:  1000,
				FullAvg10:  4.5,
				FullAvg60:  5.5,
				FullAvg300: 6.5,
				FullTotal:  2000,
				Timestamp:  now,
			},
		},
		{
			name:   "handles zero values",
			params: metrics.PressureParams{},
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pressure := tc.params.ToPressure()

			// Verify all fields are copied correctly.
			assert.Equal(t, tc.params.SomeAvg10, pressure.SomeAvg10)
			assert.Equal(t, tc.params.SomeAvg60, pressure.SomeAvg60)
			assert.Equal(t, tc.params.SomeAvg300, pressure.SomeAvg300)
			assert.Equal(t, tc.params.SomeTotal, pressure.SomeTotal)
			assert.Equal(t, tc.params.FullAvg10, pressure.FullAvg10)
			assert.Equal(t, tc.params.FullAvg60, pressure.FullAvg60)
			assert.Equal(t, tc.params.FullAvg300, pressure.FullAvg300)
			assert.Equal(t, tc.params.FullTotal, pressure.FullTotal)
			assert.Equal(t, tc.params.Timestamp, pressure.Timestamp)
		})
	}
}

// TestNewPressure tests the NewPressure constructor.
func TestNewPressure(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name   string
		params *metrics.PressureParams
	}{
		{
			name: "creates pressure from full params",
			params: &metrics.PressureParams{
				SomeAvg10:  11.0,
				SomeAvg60:  12.0,
				SomeAvg300: 13.0,
				SomeTotal:  3000,
				FullAvg10:  14.0,
				FullAvg60:  15.0,
				FullAvg300: 16.0,
				FullTotal:  4000,
				Timestamp:  now,
			},
		},
		{
			name:   "creates pressure from zero params",
			params: &metrics.PressureParams{},
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pressure := metrics.NewPressure(tc.params)

			// Verify constructor creates correct instance.
			assert.NotNil(t, pressure)
			assert.Equal(t, tc.params.SomeAvg10, pressure.SomeAvg10)
			assert.Equal(t, tc.params.FullAvg10, pressure.FullAvg10)
			assert.Equal(t, tc.params.Timestamp, pressure.Timestamp)
		})
	}
}

// TestNewPressureParams tests the NewPressureParams constructor.
func TestNewPressureParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates zero-value params",
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			params := metrics.NewPressureParams()

			// Verify constructor returns non-nil with zero values.
			assert.NotNil(t, params)
			assert.Equal(t, float64(0), params.SomeAvg10)
			assert.Equal(t, float64(0), params.FullAvg10)
		})
	}
}
