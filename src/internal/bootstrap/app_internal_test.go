package bootstrap

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// mockSignalHandler is a test double for signalHandler interface.
type mockSignalHandler struct {
	reloadCalled atomic.Bool
	stopCalled   atomic.Bool
	reloadErr    error
	stopErr      error
}

// Reload records that Reload was called and returns configured error.
func (m *mockSignalHandler) Reload() error {
	// Mark reload as called.
	m.reloadCalled.Store(true)
	// Return configured error or nil.
	return m.reloadErr
}

// Stop records that Stop was called and returns configured error.
func (m *mockSignalHandler) Stop() error {
	// Mark stop as called.
	m.stopCalled.Store(true)
	// Return configured error or nil.
	return m.stopErr
}

// TestWaitForSignals_Internal verifies WaitForSignals behavior with various signals.
// Goroutine lifecycle:
//   - Spawns a goroutine that runs WaitForSignals until the signal is processed.
//   - The goroutine exits when Stop is called via signal handling or context cancellation.
//   - Cleanup occurs when the function returns and the done channel receives the result.
func TestWaitForSignals_Internal(t *testing.T) {
	t.Parallel()

	// Define test cases for WaitForSignals.
	tests := []struct {
		name           string
		triggerType    string
		wantStopCalled bool
		wantReload     bool
		followUpSignal bool
	}{
		{
			name:           "SIGINT triggers stop",
			triggerType:    "SIGINT",
			wantStopCalled: true,
			wantReload:     false,
			followUpSignal: false,
		},
		{
			name:           "SIGHUP triggers reload then SIGINT stops",
			triggerType:    "SIGHUP",
			wantStopCalled: true,
			wantReload:     true,
			followUpSignal: true,
		},
		{
			name:           "context done triggers stop",
			triggerType:    "context",
			wantStopCalled: true,
			wantReload:     false,
			followUpSignal: false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create context and signal channel for test.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			mock := &mockSignalHandler{}

			done := make(chan error, 1)
			go func() {
				done <- WaitForSignals(ctx, cancel, sigCh, mock)
			}()

			// Trigger the appropriate signal or action.
			switch tt.triggerType {
			case "SIGINT":
				sigCh <- syscall.SIGINT
			case "SIGHUP":
				sigCh <- syscall.SIGHUP
				// Wait briefly for reload to be called.
				time.Sleep(100 * time.Millisecond)
				// Verify reload was called.
				if !mock.reloadCalled.Load() {
					t.Error("Reload should have been called")
				}
				// Send SIGINT to stop the loop if follow-up is needed.
				if tt.followUpSignal {
					sigCh <- syscall.SIGINT
				}
			case "context":
				cancel()
			}

			// Wait for function to complete with timeout.
			select {
			case err := <-done:
				// Verify no error occurred.
				if err != nil {
					t.Errorf("WaitForSignals returned error: %v", err)
				}
				// Verify stop was called if expected.
				if tt.wantStopCalled && !mock.stopCalled.Load() {
					t.Error("Stop should have been called")
				}
				// Verify reload was called if expected.
				if tt.wantReload && !mock.reloadCalled.Load() {
					t.Error("Reload should have been called")
				}
			case <-time.After(time.Second):
				t.Fatal("WaitForSignals did not complete in time")
			}
		})
	}
}
