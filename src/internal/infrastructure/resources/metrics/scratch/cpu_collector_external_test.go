// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCPUCollector_CollectSystem tests system CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()
		require.NotNil(t, collector)

		_, err := collector.CollectSystem(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectSystem(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestCPUCollector_CollectProcess tests process CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectProcess(t *testing.T) {
	t.Parallel()

	t.Run("pid 1 returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()

		_, err := collector.CollectProcess(context.Background(), 1)

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("pid 0 returns ErrInvalidPID", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()

		_, err := collector.CollectProcess(context.Background(), 0)

		assert.True(t, errors.Is(err, scratch.ErrInvalidPID))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectProcess(ctx, 1)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestCPUCollector_CollectAllProcesses tests all processes CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()

		_, err := collector.CollectAllProcesses(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectAllProcesses(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestCPUCollector_CollectLoadAverage tests load average collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectLoadAverage(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()

		_, err := collector.CollectLoadAverage(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectLoadAverage(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestCPUCollector_CollectPressure tests CPU pressure collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()

		_, err := collector.CollectPressure(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectPressure(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// Test_NewCPUCollector verifies NewCPUCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewCPUCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantNotNil  bool
		description string
	}{
		{
			name:        "returns_valid_collector",
			wantNotNil:  true,
			description: "NewCPUCollector should return a non-nil collector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewCPUCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector, tt.description)
			}
		})
	}
}
