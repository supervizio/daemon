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

// TestIOCollector_CollectStats tests I/O stats collection.
//
// Params:
//   - t: the testing context
func TestIOCollector_CollectStats(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewIOCollector()
		require.NotNil(t, collector)

		_, err := collector.CollectStats(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewIOCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectStats(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestIOCollector_CollectPressure tests I/O pressure collection.
//
// Params:
//   - t: the testing context
func TestIOCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewIOCollector()

		_, err := collector.CollectPressure(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewIOCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectPressure(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// Test_NewIOCollector verifies NewIOCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewIOCollector(t *testing.T) {
	t.Parallel()

	t.Run("returns valid collector", func(t *testing.T) {
		t.Parallel()

		collector := scratch.NewIOCollector()

		assert.NotNil(t, collector, "NewIOCollector should return a non-nil collector")
	})
}
