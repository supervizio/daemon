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

// TestDiskCollector_ListPartitions tests partition listing.
//
// Params:
//   - t: the testing context
func TestDiskCollector_ListPartitions(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()
		require.NotNil(t, collector)

		_, err := collector.ListPartitions(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.ListPartitions(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestDiskCollector_CollectUsage tests disk usage collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectUsage(t *testing.T) {
	t.Parallel()

	t.Run("root mount returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectUsage(context.Background(), "/")

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("home mount returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectUsage(context.Background(), "/home")

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectUsage(ctx, "/")

		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("empty path returns ErrEmptyPath", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectUsage(context.Background(), "")

		assert.True(t, errors.Is(err, scratch.ErrEmptyPath))
	})
}

// TestDiskCollector_CollectAllUsage tests all disk usage collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectAllUsage(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectAllUsage(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectAllUsage(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestDiskCollector_CollectIO tests I/O stats collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectIO(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectIO(context.Background())

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectIO(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestDiskCollector_CollectDeviceIO tests device I/O stats collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectDeviceIO(t *testing.T) {
	t.Parallel()

	t.Run("sda device returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectDeviceIO(context.Background(), "sda")

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("nvme device returns ErrNotSupported", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectDeviceIO(context.Background(), "nvme0n1")

		assert.True(t, errors.Is(err, scratch.ErrNotSupported))
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectDeviceIO(ctx, "sda")

		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("empty device returns ErrEmptyDevice", func(t *testing.T) {
		t.Parallel()
		collector := scratch.NewDiskCollector()

		_, err := collector.CollectDeviceIO(context.Background(), "")

		assert.True(t, errors.Is(err, scratch.ErrEmptyDevice))
	})
}

// Test_NewDiskCollector verifies NewDiskCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewDiskCollector(t *testing.T) {
	t.Parallel()

	t.Run("returns valid collector", func(t *testing.T) {
		t.Parallel()

		collector := scratch.NewDiskCollector()

		assert.NotNil(t, collector, "NewDiskCollector should return a non-nil collector")
	})
}
