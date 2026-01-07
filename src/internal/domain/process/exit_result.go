// Package process provides domain entities and value objects for process lifecycle management.
package process

// ExitResult contains the result of a process exit.
//
// ExitResult captures the exit code returned by the process and any
// error that occurred during execution or termination.
type ExitResult struct {
	// Code is the process exit code (0 indicates success).
	Code int
	// Error is any error that occurred during process execution.
	Error error
}
