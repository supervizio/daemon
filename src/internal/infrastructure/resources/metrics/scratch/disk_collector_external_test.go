//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

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
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewDiskCollector()
			require.NotNil(t, collector)

			_, err := collector.ListPartitions(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestDiskCollector_CollectUsage tests disk usage collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectUsage(t *testing.T) {
	tests := []struct {
		name       string
		mountpoint string
	}{
		{name: "root mount", mountpoint: "/"},
		{name: "home mount", mountpoint: "/home"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewDiskCollector()

			_, err := collector.CollectUsage(context.Background(), tt.mountpoint)

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestDiskCollector_CollectAllUsage tests all disk usage collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectAllUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewDiskCollector()

			_, err := collector.CollectAllUsage(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestDiskCollector_CollectIO tests I/O stats collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectIO(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewDiskCollector()

			_, err := collector.CollectIO(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestDiskCollector_CollectDeviceIO tests device I/O stats collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectDeviceIO(t *testing.T) {
	tests := []struct {
		name   string
		device string
	}{
		{name: "sda device", device: "sda"},
		{name: "nvme device", device: "nvme0n1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewDiskCollector()

			_, err := collector.CollectDeviceIO(context.Background(), tt.device)

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}
