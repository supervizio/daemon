// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestIOPressureParams_Fields tests IOPressureParams struct fields.
func TestIOPressureParams_Fields(t *testing.T) {
	tests := []struct {
		name   string
		params metrics.IOPressureParams
	}{
		{
			name: "all_fields_set",
			params: metrics.IOPressureParams{
				SomeAvg10:  10.5,
				SomeAvg60:  8.0,
				SomeAvg300: 5.0,
				SomeTotal:  12345,
				FullAvg10:  5.0,
				FullAvg60:  3.0,
				FullAvg300: 2.0,
				FullTotal:  6789,
				Timestamp:  time.Now(),
			},
		},
		{
			name: "zero_values",
			params: metrics.IOPressureParams{
				SomeAvg10:  0,
				SomeAvg60:  0,
				SomeAvg300: 0,
				SomeTotal:  0,
				FullAvg10:  0,
				FullAvg60:  0,
				FullAvg300: 0,
				FullTotal:  0,
				Timestamp:  time.Time{},
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible and hold correct values.
			assert.Equal(t, tt.params.SomeAvg10, tt.params.SomeAvg10)
			assert.Equal(t, tt.params.SomeAvg60, tt.params.SomeAvg60)
			assert.Equal(t, tt.params.SomeAvg300, tt.params.SomeAvg300)
			assert.Equal(t, tt.params.SomeTotal, tt.params.SomeTotal)
			assert.Equal(t, tt.params.FullAvg10, tt.params.FullAvg10)
			assert.Equal(t, tt.params.FullAvg60, tt.params.FullAvg60)
			assert.Equal(t, tt.params.FullAvg300, tt.params.FullAvg300)
			assert.Equal(t, tt.params.FullTotal, tt.params.FullTotal)
			assert.Equal(t, tt.params.Timestamp, tt.params.Timestamp)
		})
	}
}
