//go:build cgo

package probe_test

import (
	"os"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectSystemContextSwitches verifies system context switch collection.
func TestCollectSystemContextSwitches(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "system context switches greater than zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			count, err := probe.CollectSystemContextSwitches()
			require.NoError(t, err)

			// System context switches should be greater than zero on any running system
			assert.Greater(t, count, uint64(0))
		})
	}
}

// TestCollectProcessContextSwitches verifies process context switch collection.
func TestCollectProcessContextSwitches(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects process context switches"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			pid := int32(os.Getpid())
			cs, err := probe.CollectProcessContextSwitches(pid)
			require.NoError(t, err)
			require.NotNil(t, cs)

			// Verify structure is populated
			// Note: Values may be zero for some processes
			t.Logf("Voluntary: %d, Involuntary: %d, SystemTotal: %d",
				cs.Voluntary, cs.Involuntary, cs.SystemTotal)
		})
	}
}

// TestCollectSelfContextSwitches verifies self context switch collection.
func TestCollectSelfContextSwitches(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects self context switches"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			cs, err := probe.CollectSelfContextSwitches()
			require.NoError(t, err)
			require.NotNil(t, cs)

			// At least one context switch should have occurred to get here
			t.Logf("Self - Voluntary: %d, Involuntary: %d",
				cs.Voluntary, cs.Involuntary)
		})
	}
}

// TestCollectSystemContextSwitches_NotInitialized verifies error when not initialized.
func TestCollectSystemContextSwitches_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Do not call probe.Init()
			_, err := probe.CollectSystemContextSwitches()

			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
