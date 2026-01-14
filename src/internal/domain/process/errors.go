// Package process provides domain entities and value objects for process lifecycle management.
package process

import "errors"

// Process domain sentinel errors.
var (
	// ErrAlreadyRunning indicates an attempt to start a process that is already running.
	ErrAlreadyRunning error = errors.New("process already running")
	// ErrNotRunning indicates an attempt to operate on a non-running process.
	ErrNotRunning error = errors.New("process not running")
	// ErrMaxRetriesExceeded indicates the maximum restart retries have been exceeded.
	ErrMaxRetriesExceeded error = errors.New("max retries exceeded")
	// ErrInvalidTransition indicates an invalid state transition was attempted.
	ErrInvalidTransition error = errors.New("invalid state transition")
	// ErrProcessFailed indicates the process exited with a non-zero exit code.
	ErrProcessFailed error = errors.New("process failed")
)
