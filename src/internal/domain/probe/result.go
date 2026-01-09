// Package probe provides domain abstractions for service probing.
package probe

import "time"

// Result represents the outcome of a probe execution.
// It contains the probe status, latency measurement, output, and any error.
type Result struct {
	// Success indicates whether the probe succeeded.
	Success bool

	// Latency records how long the probe took to complete.
	// This is useful for measuring network latency and service response times.
	Latency time.Duration

	// Output contains any output from the probe.
	// For exec probes: stdout content.
	// For HTTP probes: response body summary.
	// For other probes: connection details.
	Output string

	// Error holds any error that occurred during probing.
	// When Success is false, this should contain the failure reason.
	Error error
}

// NewResult creates a new probe result with the specified parameters.
//
// Params:
//   - success: whether the probe succeeded.
//   - latency: how long the probe took to complete.
//   - output: any output from the probe.
//   - err: any error that occurred (nil for success).
//
// Returns:
//   - Result: a probe result with the specified values.
func NewResult(success bool, latency time.Duration, output string, err error) Result {
	// Return result with all fields set.
	return Result{
		Success: success,
		Latency: latency,
		Output:  output,
		Error:   err,
	}
}

// NewSuccessResult creates a successful probe result.
//
// Params:
//   - latency: how long the probe took to complete.
//   - output: any output from the probe.
//
// Returns:
//   - Result: a successful probe result.
func NewSuccessResult(latency time.Duration, output string) Result {
	// Return successful result.
	return Result{
		Success: true,
		Latency: latency,
		Output:  output,
	}
}

// NewFailureResult creates a failed probe result.
//
// Params:
//   - latency: how long the probe took before failing.
//   - output: any output from the probe.
//   - err: the error that caused the failure.
//
// Returns:
//   - Result: a failed probe result with error.
func NewFailureResult(latency time.Duration, output string, err error) Result {
	// Return failure result with error.
	return Result{
		Success: false,
		Latency: latency,
		Output:  output,
		Error:   err,
	}
}

// IsSuccess returns true if the probe succeeded.
//
// Returns:
//   - bool: true if probe succeeded, false otherwise.
func (r Result) IsSuccess() bool {
	// Return the success status.
	return r.Success
}

// IsFailure returns true if the probe failed.
//
// Returns:
//   - bool: true if probe failed, false otherwise.
func (r Result) IsFailure() bool {
	// Return the inverse of success.
	return !r.Success
}
