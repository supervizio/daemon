// Package process provides domain entities and value objects for process lifecycle management.
package process

import (
	"context"
	"os"
	"time"
)

// Executor abstracts OS process execution.
// This is the primary port for process infrastructure.
type Executor interface {
	// Start starts a process with the given specification.
	// Returns the PID, a channel that receives the exit result, and any error.
	Start(ctx context.Context, spec Spec) (pid int, wait <-chan ExitResult, err error)

	// Stop gracefully stops the process with the given PID.
	Stop(pid int, timeout time.Duration) error

	// Signal sends a signal to the process.
	Signal(pid int, sig os.Signal) error
}
