// Package ports defines the interfaces for OS abstraction.
package ports

// ZombieReaper handles zombie process reaping for PID1 scenarios.
type ZombieReaper interface {
	// Start begins the background reaping loop.
	Start()

	// Stop stops the reaping loop.
	Stop()

	// ReapOnce performs a single reap cycle and returns the count of reaped processes.
	ReapOnce() int

	// IsPID1 returns true if the current process is running as PID 1.
	IsPID1() bool
}
