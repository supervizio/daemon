// Package config provides configuration types and parsing for daemon.
package config

import "time"

// Config represents the root configuration structure.
type Config struct {
	Version    string          `yaml:"version"`
	Logging    LoggingConfig   `yaml:"logging"`
	Services   []ServiceConfig `yaml:"services"`
	ConfigPath string          `yaml:"-"` // Path to the config file (not serialized)
}

// LoggingConfig defines global logging defaults.
type LoggingConfig struct {
	Defaults LogDefaults `yaml:"defaults"`
	BaseDir  string      `yaml:"base_dir"`
}

// LogDefaults defines default logging settings.
type LogDefaults struct {
	TimestampFormat string         `yaml:"timestamp_format"`
	Rotation        RotationConfig `yaml:"rotation"`
}

// RotationConfig defines log rotation settings.
type RotationConfig struct {
	MaxSize   string `yaml:"max_size"`
	MaxAge    string `yaml:"max_age"`
	MaxFiles  int    `yaml:"max_files"`
	Compress  bool   `yaml:"compress"`
}

// ServiceConfig defines a single service configuration.
type ServiceConfig struct {
	Name             string              `yaml:"name"`
	Command          string              `yaml:"command"`
	Args             []string            `yaml:"args,omitempty"`
	User             string              `yaml:"user,omitempty"`
	Group            string              `yaml:"group,omitempty"`
	WorkingDirectory string              `yaml:"working_dir,omitempty"`
	Environment      map[string]string   `yaml:"environment,omitempty"`
	Restart          RestartConfig       `yaml:"restart"`
	HealthChecks     []HealthCheckConfig `yaml:"health_checks,omitempty"`
	Logging          ServiceLogging      `yaml:"logging,omitempty"`
	DependsOn        []string            `yaml:"depends_on,omitempty"`
	Oneshot          bool                `yaml:"oneshot,omitempty"`
}

// RestartConfig defines service restart behavior.
type RestartConfig struct {
	Policy     RestartPolicy `yaml:"policy"`
	MaxRetries int           `yaml:"max_retries,omitempty"`
	Delay      Duration      `yaml:"delay,omitempty"`
	DelayMax   Duration      `yaml:"delay_max,omitempty"`
}

// RestartPolicy defines when to restart a service.
type RestartPolicy string

const (
	RestartAlways    RestartPolicy = "always"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartNever     RestartPolicy = "never"
	RestartUnless    RestartPolicy = "unless-stopped"
)

// HealthCheckConfig defines a health check for a service.
type HealthCheckConfig struct {
	Name     string          `yaml:"name,omitempty"`
	Type     HealthCheckType `yaml:"type"`
	Interval Duration        `yaml:"interval"`
	Timeout  Duration        `yaml:"timeout"`
	Retries  int             `yaml:"retries"`
	// HTTP check fields
	Endpoint   string `yaml:"endpoint,omitempty"`
	Method     string `yaml:"method,omitempty"`
	StatusCode int    `yaml:"status_code,omitempty"`
	// TCP check fields
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
	// Command check fields
	Command string `yaml:"command,omitempty"`
}

// HealthCheckType defines the type of health check.
type HealthCheckType string

const (
	HealthCheckHTTP    HealthCheckType = "http"
	HealthCheckTCP     HealthCheckType = "tcp"
	HealthCheckCommand HealthCheckType = "command"
)

// ServiceLogging defines per-service logging configuration.
type ServiceLogging struct {
	Stdout LogStreamConfig `yaml:"stdout,omitempty"`
	Stderr LogStreamConfig `yaml:"stderr,omitempty"`
}

// LogStreamConfig defines configuration for a log stream.
type LogStreamConfig struct {
	File            string         `yaml:"file,omitempty"`
	TimestampFormat string         `yaml:"timestamp_format,omitempty"`
	Rotation        RotationConfig `yaml:"rotation,omitempty"`
}

// Duration is a wrapper around time.Duration that supports YAML unmarshaling.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// MarshalYAML implements yaml.Marshaler for Duration.
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
