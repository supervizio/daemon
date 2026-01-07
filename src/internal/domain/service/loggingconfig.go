// Package service provides domain value objects for service configuration.
package service

// LoggingConfig defines global logging defaults.
// It specifies base directory and default settings inherited by all services.
type LoggingConfig struct {
	// Defaults specifies default logging settings inherited by services.
	Defaults LogDefaults
	// BaseDir specifies the base directory for all log files.
	BaseDir string
}

// DefaultLoggingConfig returns a LoggingConfig with sensible defaults.
//
// Returns:
//   - LoggingConfig: a configuration with base directory and default settings.
func DefaultLoggingConfig() LoggingConfig {
	// Return default logging configuration with standard base directory.
	return LoggingConfig{
		BaseDir: "/var/log/daemon",
		Defaults: LogDefaults{
			TimestampFormat: "iso8601",
			Rotation:        DefaultRotationConfig(),
		},
	}
}
