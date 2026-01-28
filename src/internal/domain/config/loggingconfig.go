// Package config provides domain value objects for service configuration.
package config

// LoggingConfig defines global logging defaults.
// It specifies base directory and default settings inherited by all services.
type LoggingConfig struct {
	// Defaults specifies default logging settings inherited by services.
	Defaults LogDefaults
	// BaseDir specifies the base directory for all log files.
	BaseDir string
	// Daemon specifies daemon-level event logging configuration.
	Daemon DaemonLogging
}

// DefaultLoggingConfig returns a LoggingConfig with sensible defaults.
//
// Returns:
//   - LoggingConfig: a configuration with base directory and default settings.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		BaseDir: "/var/log/daemon",
		Defaults: LogDefaults{
			TimestampFormat: "iso8601",
			Rotation:        DefaultRotationConfig(),
		},
	}
}
