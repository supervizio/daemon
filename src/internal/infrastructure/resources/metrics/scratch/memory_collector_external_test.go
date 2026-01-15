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

// TestMemoryCollector_CollectSystem tests system memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectSystem(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewMemoryCollector()
			require.NotNil(t, collector)

			_, err := collector.CollectSystem(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestMemoryCollector_CollectProcess tests process memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectProcess(t *testing.T) {
	tests := []struct {
		name string
		pid  int
	}{
		{name: "pid 1", pid: 1},
		{name: "pid 0", pid: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewMemoryCollector()

			_, err := collector.CollectProcess(context.Background(), tt.pid)

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestMemoryCollector_CollectAllProcesses tests all processes memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectAllProcesses(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewMemoryCollector()

			_, err := collector.CollectAllProcesses(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestMemoryCollector_CollectPressure tests memory pressure collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewMemoryCollector()

			_, err := collector.CollectPressure(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}
