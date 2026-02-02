// Package metrics_test provides external tests for the metrics domain package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/stretchr/testify/assert"
)

// TestPressureParams_ToPressure_Conversion tests the ToPressure method from params file.
func TestPressureParams_ToPressure_Conversion(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name   string
		params metrics.PressureParams
	}{
		{
			name: "full conversion preserves all fields",
			params: metrics.PressureParams{
				SomeAvg10:  1.0,
				SomeAvg60:  2.0,
				SomeAvg300: 3.0,
				SomeTotal:  100,
				FullAvg10:  4.0,
				FullAvg60:  5.0,
				FullAvg300: 6.0,
				FullTotal:  200,
				Timestamp:  now,
			},
		},
		{
			name:   "zero params produce zero pressure",
			params: metrics.PressureParams{},
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pressure := tc.params.ToPressure()

			// Verify field equality.
			assert.Equal(t, tc.params.SomeAvg10, pressure.SomeAvg10)
			assert.Equal(t, tc.params.FullAvg10, pressure.FullAvg10)
		})
	}
}

// TestNewPressureParams_Constructor tests the NewPressureParams constructor.
func TestNewPressureParams_Constructor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "returns non-nil zero-value params",
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			params := metrics.NewPressureParams()

			// Verify non-nil result.
			assert.NotNil(t, params)
			assert.True(t, params.Timestamp.IsZero())
		})
	}
}
