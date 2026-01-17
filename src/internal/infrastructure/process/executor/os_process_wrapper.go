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

// Signal sends a signal to the process.
//
// Params:
//   - sig: the signal to send.
//
// Returns:
//   - error: any error from sending the signal.
func (w *osProcessWrapper) Signal(sig os.Signal) error {
	// Delegate to underlying os.Process.
	return w.proc.Signal(sig)
}

// Kill kills the process.
//
// Returns:
//   - error: any error from killing the process.
func (w *osProcessWrapper) Kill() error {
	// Delegate to underlying os.Process.
	return w.proc.Kill()
}

// Wait waits for the process to exit.
//
// Returns:
//   - *os.ProcessState: the process state after exit.
//   - error: any error from waiting.
func (w *osProcessWrapper) Wait() (*os.ProcessState, error) {
	// Delegate to underlying os.Process.
	return w.proc.Wait()
}

// defaultFindProcess wraps os.FindProcess to return the Process interface.
//
// Params:
//   - pid: the process ID to find.
//
// Returns:
//   - Process: the process interface wrapper.
//   - error: any error from finding the process.
func defaultFindProcess(pid int) (Process, error) {
	// On Unix, os.FindProcess always succeeds - it only creates a handle.
	// The actual existence check happens when Signal/Kill is called.
	proc, _ := os.FindProcess(pid)
	// Wrap and return the process.
	return &osProcessWrapper{proc: proc}, nil
}
