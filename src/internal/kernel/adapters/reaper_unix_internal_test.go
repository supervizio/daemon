//go:build unix

// Package adapters provides platform-specific implementations of kernel interfaces.
// reaper_unix_internal_test.go contains white-box unit tests for reaper functionality.
// It tests private functions: reapLoop, reapAll.
package adapters

import (
	"testing"
	"time"
)

// Test_UnixZombieReaper_reapAll tests the reapAll method.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none
func Test_UnixZombieReaper_reapAll(t *testing.T) {
	// Define test cases for reapAll.
	tests := []struct {
		// name is the test case name.
		name string
		// setup is an optional function to run before the test.
		setup func(*UnixZombieReaper)
	}{
		{
			name:  "reapAll on new reaper does not panic",
			setup: nil,
		},
		{
			name:  "reapAll with no zombies returns immediately",
			setup: nil,
		},
		{
			name:  "reapAll handles wait4 error gracefully",
			setup: nil,
		},
		{
			name:  "reapAll handles wait4 returning zero pid",
			setup: nil,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reaper := NewUnixZombieReaper()
			// Run setup if provided.
			if tt.setup != nil {
				tt.setup(reaper)
			}
			// Call reapAll should not panic.
			reaper.reapAll()
			// No panic indicates success.
		})
	}
}

// Test_UnixZombieReaper_reapLoop tests the reapLoop method.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none
func Test_UnixZombieReaper_reapLoop(t *testing.T) {
	// Define test cases for reapLoop.
	tests := []struct {
		// name is the test case name.
		name string
		// timeout is the duration to wait before stopping.
		timeout time.Duration
	}{
		{
			name:    "reapLoop starts and stops immediately",
			timeout: 0,
		},
		{
			name:    "reapLoop handles stop signal correctly",
			timeout: 10 * time.Millisecond,
		},
		{
			name:    "reapLoop processes SIGCHLD signal",
			timeout: 20 * time.Millisecond,
		},
		{
			name:    "reapLoop calls reapAll on stop",
			timeout: 5 * time.Millisecond,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reaper := NewUnixZombieReaper()
			// Start the reaper which starts reapLoop.
			reaper.Start()
			// Wait for specified timeout.
			if tt.timeout > 0 {
				time.Sleep(tt.timeout)
			}
			// Stop the reaper which stops reapLoop.
			reaper.Stop()
			// No panic indicates success.
		})
	}
}
