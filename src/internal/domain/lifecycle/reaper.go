// Package lifecycle provides domain types for daemon lifecycle management.
// This file contains the port interface for zombie process cleanup.
// Implementation is in infrastructure/process/reaper.
package lifecycle

// Reaper handles zombie process reaping for PID1 scenarios.
// When running as PID1 in a container, orphaned child processes become
// zombies if not reaped. This interface abstracts the reaping behavior.
//
// This is a DOMAIN PORT: it defines what the application layer needs,
// and the infrastructure layer provides the implementation.
type Reaper interface {
	// Start begins the background reaping loop.
	// This should be called when the supervisor starts.
	Start()

	// Stop stops the reaping loop.
	// This should be called when the supervisor stops.
	Stop()

	// ReapOnce performs a single reap cycle and returns the count of reaped processes.
	// This is useful for testing or manual reaping.
	ReapOnce() int

	// IsPID1 returns true if the current process is running as PID 1.
	// When running as PID1, zombie reaping is mandatory.
	IsPID1() bool
}
