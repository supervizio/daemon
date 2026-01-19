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

// TestMemoryCollector_CollectSystem tests system memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()
		require.NotNil(t, collector)

		_, err := collector.CollectSystem(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectSystem(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestMemoryCollector_CollectProcess tests process memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectProcess(t *testing.T) {
	t.Parallel()

	t.Run("pid 1 returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()

		_, err := collector.CollectProcess(context.Background(), 1)

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("pid 0 returns ErrInvalidPID", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()

		_, err := collector.CollectProcess(context.Background(), 0)

		assert.True(t, errors.Is(err, scratch.ErrInvalidPID))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectProcess(ctx, 1)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestMemoryCollector_CollectAllProcesses tests all processes memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()

		_, err := collector.CollectAllProcesses(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectAllProcesses(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestMemoryCollector_CollectPressure tests memory pressure collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()

		_, err := collector.CollectPressure(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewMemoryCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectPressure(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// Test_NewMemoryCollector verifies NewMemoryCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewMemoryCollector(t *testing.T) {
	t.Parallel()

	t.Run("returns valid collector", func(t *testing.T) {
		t.Parallel()

		collector := scratch.NewMemoryCollector()

		assert.NotNil(t, collector, "NewMemoryCollector should return a non-nil collector")
	})
}
