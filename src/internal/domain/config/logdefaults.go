// Package config provides domain value objects for service configuration.
package config

// LogDefaults defines default logging settings.
// It includes timestamp format and rotation configuration for log files.
type LogDefaults struct {
	// TimestampFormat specifies the Go time format string for log timestamps.
	TimestampFormat string
	// Rotation defines default log rotation settings.
	Rotation RotationConfig
}
