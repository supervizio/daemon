// Package executor provides infrastructure adapters for OS process execution.
// It implements the domain process interfaces using Unix system calls.
package executor

import (
	"os"
)

// osProcessWrapper wraps os.Process to implement the Process interface.
// This adapter allows testing with mock implementations.
type osProcessWrapper struct {
	proc *os.Process
}

// Signal delegates to os.Process.Signal.
//
// Params:
//   - sig: signal to send to the process
//
// Returns:
//   - error: if signal delivery fails (process not found, permission denied, etc)
func (w *osProcessWrapper) Signal(sig os.Signal) error { return w.proc.Signal(sig) }

// Kill delegates to os.Process.Kill.
//
// Returns:
//   - error: if SIGKILL delivery fails
func (w *osProcessWrapper) Kill() error { return w.proc.Kill() }

// Wait delegates to os.Process.Wait.
//
// Returns:
//   - *os.ProcessState: exit status and resource usage
//   - error: if wait fails (process not a child, already reaped, etc)
func (w *osProcessWrapper) Wait() (*os.ProcessState, error) { return w.proc.Wait() }

// defaultFindProcess creates a handle; existence check deferred to Signal/Kill.
// On Unix, FindProcess always succeeds; actual existence verified on signal.
//
// Params:
//   - pid: process ID to create handle for
//
// Returns:
//   - Process: wrapped process handle
//   - error: always nil (Unix FindProcess never fails)
func defaultFindProcess(pid int) (Process, error) {
	proc, _ := os.FindProcess(pid)
	return &osProcessWrapper{proc: proc}, nil
}
