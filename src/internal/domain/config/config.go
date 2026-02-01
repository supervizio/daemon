// Package config provides domain value objects for service configuration.
package config

const (
	// defaultMaxLogFiles is the default number of rotated log files to keep.
	defaultMaxLogFiles int = 10
)

// Config represents the root configuration structure.
// It contains global settings, logging configuration, and service definitions.
type Config struct {
	// Version specifies the configuration schema version for compatibility.
	Version string
	// Logging defines global logging defaults applied to all services.
	Logging LoggingConfig
	// Monitoring defines external target monitoring configuration.
	Monitoring MonitoringConfig
	// Services contains the list of service configurations to manage.
	Services []ServiceConfig
	// ConfigPath stores the path from which this configuration was loaded.
	ConfigPath string
}

// FindService returns a service configuration by name.
//
// Params:
//   - name: service name to find
//
// Returns:
//   - *ServiceConfig: service configuration or nil if not found
func (c *Config) FindService(name string) *ServiceConfig {
	// search services by name
	for i := range c.Services {
		// check if service name matches
		if c.Services[i].Name == name {
			// return matching service
			return &c.Services[i]
		}
	}
	// no match found
	return nil
}

// Validate validates the configuration.
//
// Returns:
//   - error: validation error if any
func (c *Config) Validate() error {
	// delegate to validation function
	return Validate(c)
}

// GetServiceLogPath returns the full path for a service log file.
//
// Params:
//   - serviceName: name of the service
//   - logFile: name of the log file
//
// Returns:
//   - string: full path to the service log file
func (c *Config) GetServiceLogPath(serviceName, logFile string) string {
	// Construct path by joining base directory, service name, and log filename
	// construct path from base directory, service name, and log file
	return c.Logging.BaseDir + "/" + serviceName + "/" + logFile
}

// NewConfig creates a new Config with the provided services.
//
// Params:
//   - services: list of service configurations to manage.
//
// Returns:
//   - *Config: configuration with the provided services and default logging settings.
func NewConfig(services []ServiceConfig) *Config {
	// create config with version 1 and defaults
	return &Config{
		Version:    "1",
		Logging:    DefaultLoggingConfig(),
		Monitoring: NewMonitoringConfig(),
		Services:   services,
	}
}

// DefaultConfig returns a new Config with default values.
//
// Returns:
//   - *Config: configuration with sensible defaults for logging and rotation
func DefaultConfig() *Config {
	// return config with default values
	return &Config{
		Version: "1",
		Logging: LoggingConfig{
			BaseDir: "/var/log/daemon",
			Defaults: LogDefaults{
				TimestampFormat: "iso8601",
				Rotation: RotationConfig{
					MaxSize:  "100MB",
					MaxFiles: defaultMaxLogFiles,
				},
			},
		},
		Monitoring: NewMonitoringConfig(),
	}
}
