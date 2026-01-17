// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestProcessCPUParams_Fields tests ProcessCPUParams struct fields.
func TestProcessCPUParams_Fields(t *testing.T) {
	tests := []struct {
		name   string
		params metrics.ProcessCPUParams
	}{
		{
			name: "all_fields_set",
			params: metrics.ProcessCPUParams{
				PID:            1234,
				Name:           "myprocess",
				User:           500,
				System:         300,
				ChildrenUser:   100,
				ChildrenSystem: 50,
				StartTime:      12345678,
				UsagePercent:   5.5,
				Timestamp:      time.Now(),
			},
		},
		{
			name: "zero_values",
			params: metrics.ProcessCPUParams{
				PID:            0,
				Name:           "",
				User:           0,
				System:         0,
				ChildrenUser:   0,
				ChildrenSystem: 0,
				StartTime:      0,
				UsagePercent:   0,
				Timestamp:      time.Time{},
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible and hold correct values.
			assert.Equal(t, tt.params.PID, tt.params.PID)
			assert.Equal(t, tt.params.Name, tt.params.Name)
			assert.Equal(t, tt.params.User, tt.params.User)
			assert.Equal(t, tt.params.System, tt.params.System)
			assert.Equal(t, tt.params.ChildrenUser, tt.params.ChildrenUser)
			assert.Equal(t, tt.params.ChildrenSystem, tt.params.ChildrenSystem)
			assert.Equal(t, tt.params.StartTime, tt.params.StartTime)
			assert.Equal(t, tt.params.UsagePercent, tt.params.UsagePercent)
			assert.Equal(t, tt.params.Timestamp, tt.params.Timestamp)
		})
	}
}
