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

// TestNetworkCollector_ListInterfaces tests interface listing.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_ListInterfaces(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()
		require.NotNil(t, collector)

		_, err := collector.ListInterfaces(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.ListInterfaces(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestNetworkCollector_CollectStats tests interface stats collection.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_CollectStats(t *testing.T) {
	t.Parallel()

	t.Run("eth0 interface returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()

		_, err := collector.CollectStats(context.Background(), "eth0")

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("lo interface returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()

		_, err := collector.CollectStats(context.Background(), "lo")

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectStats(ctx, "eth0")

		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("empty interface returns ErrEmptyInterface", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()

		_, err := collector.CollectStats(context.Background(), "")

		assert.True(t, errors.Is(err, scratch.ErrEmptyInterface))
	})
}

// TestNetworkCollector_CollectAllStats tests all interface stats collection.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_CollectAllStats(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()

		_, err := collector.CollectAllStats(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewNetworkCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectAllStats(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// Test_NewNetworkCollector verifies NewNetworkCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewNetworkCollector(t *testing.T) {
	t.Parallel()

	t.Run("returns valid collector", func(t *testing.T) {
		t.Parallel()

		collector := scratch.NewNetworkCollector()

		assert.NotNil(t, collector, "NewNetworkCollector should return a non-nil collector")
	})
}
