//go:build linux

// Package linux provides Linux-specific metric collectors using /proc filesystem.
package linux

import "fmt"

// InvalidPIDError indicates an invalid process ID was provided.
// This error occurs when attempting to collect metrics for non-positive PIDs.
type InvalidPIDError struct {
	PID int
}

// NewInvalidPIDError creates a new InvalidPIDError with the given PID.
//
// Params:
//   - pid: the invalid process ID value
//
// Returns:
//   - *InvalidPIDError: configured error instance
func NewInvalidPIDError(pid int) *InvalidPIDError {
	// Create error with the provided PID value.
	return &InvalidPIDError{PID: pid}
}

// Error implements the error interface.
//
// Returns:
//   - string: formatted error message
func (e *InvalidPIDError) Error() string {
	// Format error with the invalid PID value.
	return fmt.Sprintf("invalid pid: %d", e.PID)
}
