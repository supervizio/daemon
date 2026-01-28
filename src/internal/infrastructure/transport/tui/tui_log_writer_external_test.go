// Package tui_test provides external tests for tui_log_writer.go.
package tui_test

import (
	"testing"

	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/stretchr/testify/assert"
)

// TestNewTUILogWriter tests the NewTUILogWriter constructor.
// It verifies that a new TUILogWriter is properly initialized.
//
// Params:
//   - t: the testing context.
func TestNewTUILogWriter(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// withAdapter indicates if adapter should be provided.
		withAdapter bool
	}{
		{
			name:        "with_adapter",
			withAdapter: true,
		},
		{
			name:        "with_nil_adapter",
			withAdapter: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create adapter if needed.
			var adapter *tui.LogAdapter
			if tt.withAdapter {
				adapter = tui.NewLogAdapter()
			}

			// Call constructor.
			writer := tui.NewTUILogWriter(adapter)

			// Verify result.
			assert.NotNil(t, writer)
		})
	}
}

// TestTUILogWriter_Write tests the Write method.
// It verifies that log events are written to the adapter.
//
// Params:
//   - t: the testing context.
func TestTUILogWriter_Write(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// withAdapter indicates if adapter should be provided.
		withAdapter bool
		// wantError indicates if error is expected.
		wantError bool
	}{
		{
			name:        "writes_to_adapter",
			withAdapter: true,
			wantError:   false,
		},
		{
			name:        "nil_adapter_no_error",
			withAdapter: false,
			wantError:   false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create adapter if needed.
			var adapter *tui.LogAdapter
			if tt.withAdapter {
				adapter = tui.NewLogAdapter()
			}

			// Create writer.
			writer := tui.NewTUILogWriter(adapter)

			// Create event.
			event := domainlogging.LogEvent{
				Service: "test-service",
				Message: "test message",
			}

			// Call Write.
			err := writer.Write(event)

			// Verify result.
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify adapter received event.
			if tt.withAdapter {
				entries := adapter.Buffer().Entries()
				assert.Len(t, entries, 1)
				assert.Equal(t, "test-service", entries[0].Service)
			}
		})
	}
}

// TestTUILogWriter_Close tests the Close method.
// It verifies that Close always returns nil.
//
// Params:
//   - t: the testing context.
func TestTUILogWriter_Close(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "close_returns_nil",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create writer.
			writer := tui.NewTUILogWriter(nil)

			// Call Close.
			err := writer.Close()

			// Verify result.
			assert.NoError(t, err)
		})
	}
}
