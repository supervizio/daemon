// Package kernel provides domain interfaces for OS abstraction.
// These ports define the contract between the application layer and
// platform-specific implementations in the infrastructure layer.
package kernel

// ZombieReaper handles zombie process reaping for PID1 scenarios.
// When running as PID1 in a container, orphaned child processes become
// zombies if not reaped. This interface abstracts the reaping behavior.
type ZombieReaper interface {
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
