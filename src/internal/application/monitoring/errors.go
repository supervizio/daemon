// Package monitoring provides the application service for external target monitoring.
package monitoring

import "errors"

// Sentinel errors for the monitoring package.
var (
	// ErrProberFactoryMissing indicates the prober factory was not configured.
	ErrProberFactoryMissing error = errors.New("prober factory not configured")

	// ErrEmptyProbeType indicates the target has no probe type configured.
	ErrEmptyProbeType error = errors.New("target has no probe type configured")

	// ErrTargetNotFound indicates the target was not found in the registry.
	ErrTargetNotFound error = errors.New("target not found")

	// ErrTargetExists indicates a target with the same ID already exists.
	ErrTargetExists error = errors.New("target already exists")

	// ErrMonitorNotRunning indicates an operation was attempted on a stopped monitor.
	ErrMonitorNotRunning error = errors.New("monitor is not running")

	// ErrMonitorAlreadyRunning indicates Start was called on an already running monitor.
	ErrMonitorAlreadyRunning error = errors.New("monitor is already running")
)
