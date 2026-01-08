// Package probe provides domain abstractions for service probing.
package probe

import "errors"

// ErrInvalidTimeout indicates the timeout value is invalid.
var ErrInvalidTimeout = errors.New("timeout must be positive")

// ErrInvalidInterval indicates the interval value is invalid.
var ErrInvalidInterval = errors.New("interval must be positive")

// ErrInvalidSuccessThreshold indicates the success threshold is invalid.
var ErrInvalidSuccessThreshold = errors.New("success threshold must be positive")

// ErrInvalidFailureThreshold indicates the failure threshold is invalid.
var ErrInvalidFailureThreshold = errors.New("failure threshold must be positive")

// ErrProbeTimeout indicates the probe timed out.
var ErrProbeTimeout = errors.New("probe timeout")

// ErrConnectionRefused indicates the connection was refused.
var ErrConnectionRefused = errors.New("connection refused")

// ErrEmptyCommand indicates the command is empty.
var ErrEmptyCommand = errors.New("empty command")
