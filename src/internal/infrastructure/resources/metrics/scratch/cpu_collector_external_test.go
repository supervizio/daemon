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
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewCPUCollector()
			require.NotNil(t, collector)

			_, err := collector.CollectSystem(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestCPUCollector_CollectProcess tests process CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectProcess(t *testing.T) {
	tests := []struct {
		name        string
		pid         int
		expectedErr error
	}{
		{name: "pid 1", pid: 1, expectedErr: scratch.ErrNotSupported},
		{name: "pid 0", pid: 0, expectedErr: scratch.ErrInvalidPID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewCPUCollector()

			_, err := collector.CollectProcess(context.Background(), tt.pid)

			// Verify expected error is returned
			assert.True(t, errors.Is(err, tt.expectedErr))
		})
	}
}

// TestCPUCollector_CollectAllProcesses tests all processes CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectAllProcesses(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewCPUCollector()

			_, err := collector.CollectAllProcesses(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestCPUCollector_CollectLoadAverage tests load average collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectLoadAverage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewCPUCollector()

			_, err := collector.CollectLoadAverage(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestCPUCollector_CollectPressure tests CPU pressure collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewCPUCollector()

			_, err := collector.CollectPressure(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
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
