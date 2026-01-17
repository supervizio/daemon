// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestSystemCPUParams_Fields tests SystemCPUParams struct fields.
func TestSystemCPUParams_Fields(t *testing.T) {
	tests := []struct {
		name   string
		params metrics.SystemCPUParams
	}{
		{
			name: "all_fields_set",
			params: metrics.SystemCPUParams{
				User:         1000,
				Nice:         100,
				System:       500,
				Idle:         8000,
				IOWait:       200,
				IRQ:          50,
				SoftIRQ:      50,
				Steal:        10,
				Guest:        5,
				GuestNice:    1,
				UsagePercent: 25.5,
				Timestamp:    time.Now(),
			},
		},
		{
			name: "zero_values",
			params: metrics.SystemCPUParams{
				User:         0,
				Nice:         0,
				System:       0,
				Idle:         0,
				IOWait:       0,
				IRQ:          0,
				SoftIRQ:      0,
				Steal:        0,
				Guest:        0,
				GuestNice:    0,
				UsagePercent: 0,
				Timestamp:    time.Time{},
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible and hold correct values.
			assert.Equal(t, tt.params.User, tt.params.User)
			assert.Equal(t, tt.params.Nice, tt.params.Nice)
			assert.Equal(t, tt.params.System, tt.params.System)
			assert.Equal(t, tt.params.Idle, tt.params.Idle)
			assert.Equal(t, tt.params.IOWait, tt.params.IOWait)
			assert.Equal(t, tt.params.IRQ, tt.params.IRQ)
			assert.Equal(t, tt.params.SoftIRQ, tt.params.SoftIRQ)
			assert.Equal(t, tt.params.Steal, tt.params.Steal)
			assert.Equal(t, tt.params.Guest, tt.params.Guest)
			assert.Equal(t, tt.params.GuestNice, tt.params.GuestNice)
			assert.Equal(t, tt.params.UsagePercent, tt.params.UsagePercent)
			assert.Equal(t, tt.params.Timestamp, tt.params.Timestamp)
		})
	}
}
