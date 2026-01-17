// Package health provides domain abstractions for service probing.
package health

import "time"

// CheckResult represents the outcome of a probe execution.
// It contains the probe status, latency measurement, output, and any error.
type CheckResult struct {
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

// NewCheckResult creates a new probe result with the specified parameters.
//
// Params:
//   - success: whether the probe succeeded.
//   - latency: how long the probe took to complete.
//   - output: any output from the probe.
//   - err: any error that occurred (nil for success).
//
// Returns:
//   - CheckResult: a probe result with the specified values.
func NewCheckResult(success bool, latency time.Duration, output string, err error) CheckResult {
	// Return result with all fields set.
	return CheckResult{
		Success: success,
		Latency: latency,
		Output:  output,
		Error:   err,
	}
}

// NewSuccessCheckResult creates a successful probe result.
//
// Params:
//   - latency: how long the probe took to complete.
//   - output: any output from the probe.
//
// Returns:
//   - CheckResult: a successful probe result.
func NewSuccessCheckResult(latency time.Duration, output string) CheckResult {
	// Return successful result.
	return CheckResult{
		Success: true,
		Latency: latency,
		Output:  output,
	}
}

// NewFailureCheckResult creates a failed probe result.
//
// Params:
//   - latency: how long the probe took before failing.
//   - output: any output from the probe.
//   - err: the error that caused the failure.
//
// Returns:
//   - CheckResult: a failed probe result with error.
func NewFailureCheckResult(latency time.Duration, output string, err error) CheckResult {
	// Return failure result with error.
	return CheckResult{
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
func (r CheckResult) IsSuccess() bool {
	// Return the success status.
	return r.Success
}

// IsFailure returns true if the probe failed.
//
// Returns:
//   - bool: true if probe failed, false otherwise.
func (r CheckResult) IsFailure() bool {
	// Return the inverse of success.
	return !r.Success
}
