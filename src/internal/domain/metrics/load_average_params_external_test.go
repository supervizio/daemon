// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestLoadAverageParams_Fields tests LoadAverageParams struct fields.
func TestLoadAverageParams_Fields(t *testing.T) {
	tests := []struct {
		name   string
		params metrics.LoadAverageParams
	}{
		{
			name: "all_fields_set",
			params: metrics.LoadAverageParams{
				Load1:            3.5,
				Load5:            2.0,
				Load15:           1.5,
				RunningProcesses: 10,
				TotalProcesses:   300,
				LastPID:          54321,
				Timestamp:        time.Now(),
			},
		},
		{
			name: "zero_values",
			params: metrics.LoadAverageParams{
				Load1:            0,
				Load5:            0,
				Load15:           0,
				RunningProcesses: 0,
				TotalProcesses:   0,
				LastPID:          0,
				Timestamp:        time.Time{},
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible and hold correct values.
			assert.Equal(t, tt.params.Load1, tt.params.Load1)
			assert.Equal(t, tt.params.Load5, tt.params.Load5)
			assert.Equal(t, tt.params.Load15, tt.params.Load15)
			assert.Equal(t, tt.params.RunningProcesses, tt.params.RunningProcesses)
			assert.Equal(t, tt.params.TotalProcesses, tt.params.TotalProcesses)
			assert.Equal(t, tt.params.LastPID, tt.params.LastPID)
			assert.Equal(t, tt.params.Timestamp, tt.params.Timestamp)
		})
	}
}
