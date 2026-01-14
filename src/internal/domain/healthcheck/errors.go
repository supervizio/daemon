// Package healthcheck provides domain abstractions for service probing.
package healthcheck

import "errors"

var (
	// ErrInvalidTimeout indicates the timeout value is invalid.
	// Used when Config.Timeout is zero or negative during validation.
	ErrInvalidTimeout error = errors.New("timeout must be positive")

	// ErrInvalidInterval indicates the interval value is invalid.
	// Used when Config.Interval is zero or negative during validation.
	ErrInvalidInterval error = errors.New("interval must be positive")

	// ErrInvalidSuccessThreshold indicates the success threshold is invalid.
	// Used when Config.SuccessThreshold is zero or negative during validation.
	ErrInvalidSuccessThreshold error = errors.New("success threshold must be positive")

	// ErrInvalidFailureThreshold indicates the failure threshold is invalid.
	// Used when Config.FailureThreshold is zero or negative during validation.
	ErrInvalidFailureThreshold error = errors.New("failure threshold must be positive")

	// ErrProbeTimeout indicates the probe timed out.
	// Returned when a probe exceeds its configured timeout duration.
	ErrProbeTimeout error = errors.New("probe timeout")

	// ErrConnectionRefused indicates the connection was refused.
	// Returned when the target actively refuses the connection attempt.
	ErrConnectionRefused error = errors.New("connection refused")
)
