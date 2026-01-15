//go:build linux

// Package cgroup_test provides external tests for the cgroup package.
package cgroup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/cgroup"
)

// TestReader_Interface tests that V2Reader implements the Reader interface.
//
// Params:
//   - t: the testing context.
func TestReader_Interface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "V2Reader_implements_Reader_interface"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory.
			mockCgroup := t.TempDir()

			// Create V2Reader.
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Verify interface compliance using type assertion.
			var _ cgroup.Reader = reader
			assert.NotNil(t, reader)
			assert.Equal(t, mockCgroup, reader.Path())
		})
	}
}
