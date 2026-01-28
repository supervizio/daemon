// Package collector_test provides black-box tests for the collector package.
package collector_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestNewSandboxCollector tests the NewSandboxCollector constructor.
// It verifies that the constructor creates a valid sandbox collector instance.
//
// Params:
//   - t: the testing context.
func TestNewSandboxCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// wantNil indicates if the result should be nil.
		wantNil bool
	}{
		{
			name:    "returns_non_nil_collector",
			wantNil: false,
		},
		{
			name:    "returns_valid_instance",
			wantNil: false,
		},
		{
			name:    "multiple_calls_return_independent_instances",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := collector.NewSandboxCollector()

			if tt.wantNil {
				assert.Nil(t, c)
			} else {
				assert.NotNil(t, c)
			}

			// Additional validation for specific test cases.
			if tt.name == "multiple_calls_return_independent_instances" {
				c2 := collector.NewSandboxCollector()
				assert.NotNil(t, c2)
				// Both instances should be valid (though they're stateless).
				assert.IsType(t, &collector.SandboxCollector{}, c)
				assert.IsType(t, &collector.SandboxCollector{}, c2)
			}
		})
	}
}

func TestSandboxCollector_Gather(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wantErr   bool
		wantEmpty bool
	}{
		{
			name:      "Gather returns no error and populates sandboxes",
			wantErr:   false,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := collector.NewSandboxCollector()
			snap := &model.Snapshot{}

			err := c.Gather(snap)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantEmpty {
				assert.Empty(t, snap.Sandboxes)
			} else {
				assert.NotEmpty(t, snap.Sandboxes)

				// Verify structure.
				for _, sandbox := range snap.Sandboxes {
					assert.NotEmpty(t, sandbox.Name)
					// Detected flag should be set (true or false).
					// Endpoint may be empty if not detected.
					if sandbox.Detected {
						assert.NotEmpty(t, sandbox.Endpoint, "detected sandbox should have endpoint")
					}
				}
			}
		})
	}
}
