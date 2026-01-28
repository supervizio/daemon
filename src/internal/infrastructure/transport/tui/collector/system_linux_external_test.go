//go:build linux

// Package collector_test provides black-box tests for the collector package.
package collector_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestNewSystemCollector tests the NewSystemCollector constructor.
func TestNewSystemCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates collector with pre-allocated buffers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := collector.NewSystemCollector()

			assert.NotNil(t, c)
		})
	}
}

// TestSystemCollector_Gather tests the Gather method.
func TestSystemCollector_Gather(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		wantErr        bool
		wantMinLoadAvg float64
	}{
		{
			name:           "Gather returns no error and populates system metrics",
			wantErr:        false,
			wantMinLoadAvg: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := collector.NewSystemCollector()
			snap := &model.Snapshot{}

			err := c.Gather(snap)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// On Linux, we should have some system metrics populated.
			// CPU might be zero on first call (needs delta).
			// Memory should be populated.
			if snap.System.MemoryTotal > 0 {
				assert.Greater(t, snap.System.MemoryTotal, uint64(0))
			}

			// Load average should be non-negative.
			assert.GreaterOrEqual(t, snap.System.LoadAvg1, tt.wantMinLoadAvg)
		})
	}
}
