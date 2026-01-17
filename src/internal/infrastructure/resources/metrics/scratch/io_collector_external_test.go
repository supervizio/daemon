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
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewIOCollector()
			require.NotNil(t, collector)

			_, err := collector.CollectStats(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestIOCollector_CollectPressure tests I/O pressure collection.
//
// Params:
//   - t: the testing context
func TestIOCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewIOCollector()

			_, err := collector.CollectPressure(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// Test_NewIOCollector verifies NewIOCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewIOCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantNotNil  bool
		description string
	}{
		{
			name:        "returns_valid_collector",
			wantNotNil:  true,
			description: "NewIOCollector should return a non-nil collector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewIOCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector, tt.description)
			}
		})
	}
}
