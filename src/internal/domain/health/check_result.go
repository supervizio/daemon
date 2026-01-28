// Package health provides domain abstractions for service probing.
package health

import "time"

// CheckResult represents the outcome of a probe execution.
// It contains the probe status, latency measurement, output, and any error.
//
// Fields are ordered by size for optimal memory alignment:
// error interface (16B), string (16B), Duration (8B), bool (1B).
type CheckResult struct {
	// Error holds any error that occurred during probing.
	// When Success is false, this should contain the failure reason.
	Error error

	// Output contains any output from the probe.
	// For exec probes: stdout content.
	// For HTTP probes: response body summary.
	// For other probes: connection details.
	Output string

	// Latency records how long the probe took to complete.
	// This is useful for measuring network latency and service response times.
	Latency time.Duration

	// Success indicates whether the probe succeeded.
	Success bool
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
	return r.Success
}

// IsFailure returns true if the probe failed.
//
// Returns:
//   - bool: true if probe failed, false otherwise.
func (r CheckResult) IsFailure() bool {
	return !r.Success
}
