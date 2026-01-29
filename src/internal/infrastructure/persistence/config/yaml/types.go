// Package yaml provides YAML configuration loading infrastructure.
// It handles parsing and conversion of YAML configuration files to domain objects.
package yaml

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
)

const (
	// defaultFailureThreshold defines how many consecutive failures before marking unhealthy.
	// Three failures provides a balance between quick detection and avoiding false positives
	// from transient network issues or temporary service hiccups.
	defaultFailureThreshold int = 3

	// defaultHTTPStatusCode is the expected HTTP response code for healthy endpoints.
	// HTTP 200 OK is the standard success response indicating the request was successful.
	defaultHTTPStatusCode int = 200

	// defaultProbeInterval is the default interval between probe executions.
	// 10 seconds is a reasonable balance between responsiveness and resource usage.
	defaultProbeInterval time.Duration = 10 * time.Second

	// defaultProbeTimeout is the default timeout for a single probe execution.
	// 5 seconds allows most network operations to complete without false timeouts.
	defaultProbeTimeout time.Duration = 5 * time.Second
)

// Duration is a wrapper around time.Duration for YAML serialization.
// It enables parsing of human-readable duration strings like "30s" or "5m" from YAML files.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
// It parses a string duration value from YAML into a Duration type.
//
// Params:
//   - unmarshal: callback function to unmarshal the YAML value
//
// Returns:
//   - error: parsing error if the duration string is invalid
func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var s string

	// unmarshal string from YAML.
	if err := unmarshal(&s); err != nil {
		// return unmarshal error.
		return err
	}

	// parse duration string.
	parsed, err := time.ParseDuration(s)
	// parsing failed.
	if err != nil {
		// return parse error.
		return err
	}

	*d = Duration(parsed)

	// duration successfully parsed.
	return nil
}

// MarshalText implements encoding.TextMarshaler for Duration.
// It converts a Duration back to a byte slice for serialization.
// This approach is used instead of yaml.Marshaler to avoid returning interface{}.
//
// Returns:
//   - []byte: the duration as a formatted string in bytes
//   - error: always nil for this implementation
func (d *Duration) MarshalText() ([]byte, error) {
	// convert duration to string and return as bytes.
	return []byte(time.Duration(*d).String()), nil
}

// ConfigDTO is the YAML representation of the root configuration.
// It serves as the data transfer object for parsing the main configuration file.
type ConfigDTO struct {
	Version  string             `yaml:"version"`
	Logging  LoggingConfigDTO   `yaml:"logging"`
	Services []ServiceConfigDTO `yaml:"services"`
}

// ServiceConfigDTO is the YAML representation of a service configuration.
// It contains all settings needed to define and manage a supervised config.
type ServiceConfigDTO struct {
	Name             string            `yaml:"name"`
	Command          string            `yaml:"command"`
	Args             []string          `yaml:"args,omitempty"`
	User             string            `yaml:"user,omitempty"`
	Group            string            `yaml:"group,omitempty"`
	WorkingDirectory string            `yaml:"working_dir,omitempty"`
	Environment      map[string]string `yaml:"environment,omitempty"`
	Restart          RestartConfigDTO  `yaml:"restart"`
	HealthChecks     []HealthCheckDTO  `yaml:"health_checks,omitempty"`
	Listeners        []ListenerDTO     `yaml:"listeners,omitempty"`
	Logging          ServiceLoggingDTO `yaml:"logging,omitempty"`
	DependsOn        []string          `yaml:"depends_on,omitempty"`
	Oneshot          bool              `yaml:"oneshot,omitempty"`
}

// ListenerDTO is the YAML representation of a network listener.
// It defines a port with optional health probe configuration.
type ListenerDTO struct {
	Name     string   `yaml:"name"`
	Port     int      `yaml:"port"`
	Protocol string   `yaml:"protocol,omitempty"`
	Address  string   `yaml:"address,omitempty"`
	Exposed  bool     `yaml:"exposed,omitempty"`
	Probe    ProbeDTO `yaml:"probe,omitempty"`
}

// ProbeDTO is the YAML representation of a probe configuration.
// It defines how to probe a listener for health checking.
type ProbeDTO struct {
	Type             string   `yaml:"type"`
	Interval         Duration `yaml:"interval,omitempty"`
	Timeout          Duration `yaml:"timeout,omitempty"`
	SuccessThreshold int      `yaml:"success_threshold,omitempty"`
	FailureThreshold int      `yaml:"failure_threshold,omitempty"`
	Path             string   `yaml:"path,omitempty"`
	Method           string   `yaml:"method,omitempty"`
	StatusCode       int      `yaml:"status_code,omitempty"`
	Service          string   `yaml:"service,omitempty"`
	Command          string   `yaml:"command,omitempty"`
	Args             []string `yaml:"args,omitempty"`
}

// RestartConfigDTO is the YAML representation of restart configuration.
// It defines the restart policy and timing parameters for service recovery.
type RestartConfigDTO struct {
	Policy          string   `yaml:"policy"`
	MaxRetries      int      `yaml:"max_retries,omitempty"`
	Delay           Duration `yaml:"delay,omitempty"`
	DelayMax        Duration `yaml:"delay_max,omitempty"`
	StabilityWindow Duration `yaml:"stability_window,omitempty"`
}

// HealthCheckDTO is the YAML representation of a health check.
// It defines how to verify that a service is running correctly.
type HealthCheckDTO struct {
	Name       string   `yaml:"name,omitempty"`
	Type       string   `yaml:"type"`
	Interval   Duration `yaml:"interval"`
	Timeout    Duration `yaml:"timeout"`
	Retries    int      `yaml:"retries"`
	Endpoint   string   `yaml:"endpoint,omitempty"`
	Method     string   `yaml:"method,omitempty"`
	StatusCode int      `yaml:"status_code,omitempty"`
	Host       string   `yaml:"host,omitempty"`
	Port       int      `yaml:"port,omitempty"`
	Command    string   `yaml:"command,omitempty"`
}

// LoggingConfigDTO is the YAML representation of logging configuration.
// It contains global logging settings including defaults and base directory.
type LoggingConfigDTO struct {
	Defaults LogDefaultsDTO   `yaml:"defaults"`
	BaseDir  string           `yaml:"base_dir"`
	Daemon   DaemonLoggingDTO `yaml:"daemon,omitempty"`
}

// DaemonLoggingDTO is the YAML representation of daemon-level logging.
// It defines writers for daemon event logging.
type DaemonLoggingDTO struct {
	Writers []WriterConfigDTO `yaml:"writers,omitempty"`
}

// WriterConfigDTO is the YAML representation of a log writer configuration.
// It defines the type, level, and specific writer settings for file or JSON output.
type WriterConfigDTO struct {
	Type  string              `yaml:"type"`
	Level string              `yaml:"level,omitempty"`
	File  FileWriterConfigDTO `yaml:"file,omitempty"`
	JSON  JSONWriterConfigDTO `yaml:"json,omitempty"`
}

// FileWriterConfigDTO is the YAML representation of file writer configuration.
// It specifies the output file path and rotation policy for file-based logging.
type FileWriterConfigDTO struct {
	Path     string            `yaml:"path,omitempty"`
	Rotation RotationConfigDTO `yaml:"rotation,omitempty"`
}

// JSONWriterConfigDTO is the YAML representation of JSON writer configuration.
// It specifies the output file path and rotation policy for JSON-formatted logging.
type JSONWriterConfigDTO struct {
	Path     string            `yaml:"path,omitempty"`
	Rotation RotationConfigDTO `yaml:"rotation,omitempty"`
}

// LogDefaultsDTO is the YAML representation of logging defaults.
// It defines default timestamp format and rotation settings for all log streams.
type LogDefaultsDTO struct {
	TimestampFormat string            `yaml:"timestamp_format"`
	Rotation        RotationConfigDTO `yaml:"rotation"`
}

// RotationConfigDTO is the YAML representation of rotation configuration.
// It specifies log file rotation parameters like size limits and retention.
type RotationConfigDTO struct {
	MaxSize  string `yaml:"max_size"`
	MaxAge   string `yaml:"max_age"`
	MaxFiles int    `yaml:"max_files"`
	Compress bool   `yaml:"compress"`
}

// ServiceLoggingDTO is the YAML representation of service logging.
// It defines separate configurations for stdout and stderr log streams.
type ServiceLoggingDTO struct {
	Stdout LogStreamConfigDTO `yaml:"stdout,omitempty"`
	Stderr LogStreamConfigDTO `yaml:"stderr,omitempty"`
}

// LogStreamConfigDTO is the YAML representation of a log stream.
// It configures file path, format, and rotation for a single log stream.
type LogStreamConfigDTO struct {
	File            string            `yaml:"file,omitempty"`
	TimestampFormat string            `yaml:"timestamp_format,omitempty"`
	Rotation        RotationConfigDTO `yaml:"rotation,omitempty"`
}

// ToDomain converts ConfigDTO to domain Config.
// It transforms the YAML data transfer object into the domain model.
//
// Params:
//   - configPath: the filesystem path of the loaded configuration file
//
// Returns:
//   - *config.Config: the converted domain configuration object
func (c *ConfigDTO) ToDomain(configPath string) *config.Config {
	services := make([]config.ServiceConfig, 0, len(c.Services))

	// convert each service to domain model.
	for i := range c.Services {
		services = append(services, c.Services[i].ToDomain())
	}

	// return assembled domain configuration.
	return &config.Config{
		Version:    c.Version,
		ConfigPath: configPath,
		Logging:    c.Logging.ToDomain(),
		Services:   services,
	}
}

// ToDomain converts ServiceConfigDTO to domain ServiceConfig.
// It maps all service settings from YAML format to the domain model.
//
// Returns:
//   - config.ServiceConfig: the converted domain service configuration
func (s *ServiceConfigDTO) ToDomain() config.ServiceConfig {
	healthChecks := make([]config.HealthCheckConfig, 0, len(s.HealthChecks))

	// convert each health check to domain model.
	for i := range s.HealthChecks {
		healthChecks = append(healthChecks, s.HealthChecks[i].ToDomain())
	}

	listeners := make([]config.ListenerConfig, 0, len(s.Listeners))

	// convert each listener to domain model.
	for i := range s.Listeners {
		listeners = append(listeners, s.Listeners[i].ToDomain())
	}

	// return assembled domain service config.
	return config.ServiceConfig{
		Name:             s.Name,
		Command:          s.Command,
		Args:             s.Args,
		User:             s.User,
		Group:            s.Group,
		WorkingDirectory: s.WorkingDirectory,
		Environment:      s.Environment,
		Restart:          s.Restart.ToDomain(),
		DependsOn:        s.DependsOn,
		Oneshot:          s.Oneshot,
		Logging:          s.Logging.ToDomain(),
		HealthChecks:     healthChecks,
		Listeners:        listeners,
	}
}

// ToDomain converts ListenerDTO to domain ListenerConfig.
// It maps listener settings from YAML format to the domain model.
//
// Returns:
//   - config.ListenerConfig: the converted domain listener configuration
func (l *ListenerDTO) ToDomain() config.ListenerConfig {
	// Default to TCP when no protocol is specified.
	protocol := l.Protocol
	// apply default TCP protocol.
	if protocol == "" {
		protocol = "tcp"
	}

	listener := config.ListenerConfig{
		Name:     l.Name,
		Port:     l.Port,
		Protocol: protocol,
		Address:  l.Address,
		Exposed:  l.Exposed,
	}

	// add probe configuration if present.
	if l.Probe.Type != "" {
		probe := l.Probe.ToDomain()
		listener.Probe = &probe
	}

	// return assembled listener config.
	return listener
}

// ToDomain converts ProbeDTO to domain ProbeConfig.
// It maps probe settings from YAML format to the domain model.
//
// Returns:
//   - config.ProbeConfig: the converted domain probe configuration
func (p *ProbeDTO) ToDomain() config.ProbeConfig {
	successThreshold, failureThreshold := p.getThresholdDefaults()
	interval, timeout := p.getTimingDefaults()
	method, statusCode := p.getHTTPDefaults()

	// return assembled probe config with defaults applied.
	return config.ProbeConfig{
		Type:             p.Type,
		Interval:         shared.FromTimeDuration(interval),
		Timeout:          shared.FromTimeDuration(timeout),
		SuccessThreshold: successThreshold,
		FailureThreshold: failureThreshold,
		Path:             p.Path,
		Method:           method,
		StatusCode:       statusCode,
		Service:          p.Service,
		Command:          p.Command,
		Args:             p.Args,
	}
}

// getThresholdDefaults returns threshold values with defaults applied.
//
// Returns:
//   - int: success threshold (default 1).
//   - int: failure threshold (default 3).
func (p *ProbeDTO) getThresholdDefaults() (successThreshold, failureThreshold int) {
	// Require at least one success to mark healthy.
	successThreshold = p.SuccessThreshold
	// apply default success threshold.
	if successThreshold <= 0 {
		successThreshold = 1
	}

	// Allow three failures before marking unhealthy.
	failureThreshold = p.FailureThreshold
	// apply default failure threshold.
	if failureThreshold <= 0 {
		failureThreshold = defaultFailureThreshold
	}

	// return threshold values with defaults applied.
	return successThreshold, failureThreshold
}

// getTimingDefaults returns timing values with defaults applied.
//
// Returns:
//   - time.Duration: interval (default 10s).
//   - time.Duration: timeout (default 5s).
func (p *ProbeDTO) getTimingDefaults() (interval, timeout time.Duration) {
	interval = time.Duration(p.Interval)
	// apply default probe interval.
	if interval <= 0 {
		interval = defaultProbeInterval
	}

	timeout = time.Duration(p.Timeout)
	// apply default probe timeout.
	if timeout <= 0 {
		timeout = defaultProbeTimeout
	}

	// return timing values with defaults applied.
	return interval, timeout
}

// getHTTPDefaults returns HTTP-specific values with defaults applied.
//
// Returns:
//   - string: HTTP method (default "GET").
//   - int: expected status code (default 200).
func (p *ProbeDTO) getHTTPDefaults() (method string, statusCode int) {
	// GET is the standard HTTP method for health probes.
	method = p.Method
	// apply default HTTP method.
	if method == "" {
		method = "GET"
	}

	// HTTP 200 OK indicates a healthy response.
	statusCode = p.StatusCode
	// apply default status code.
	if statusCode == 0 {
		statusCode = defaultHTTPStatusCode
	}

	// return HTTP values with defaults applied.
	return method, statusCode
}

// ToDomain converts RestartConfigDTO to domain RestartConfig.
// It transforms restart policy settings to the domain model format.
//
// Returns:
//   - config.RestartConfig: the converted domain restart configuration
func (r *RestartConfigDTO) ToDomain() config.RestartConfig {
	// return assembled restart config.
	return config.RestartConfig{
		Policy:          config.RestartPolicy(r.Policy),
		MaxRetries:      r.MaxRetries,
		Delay:           shared.FromTimeDuration(time.Duration(r.Delay)),
		DelayMax:        shared.FromTimeDuration(time.Duration(r.DelayMax)),
		StabilityWindow: shared.FromTimeDuration(time.Duration(r.StabilityWindow)),
	}
}

// ToDomain converts HealthCheckDTO to domain HealthCheckConfig.
// It maps health check parameters from YAML format to the domain model.
//
// Returns:
//   - config.HealthCheckConfig: the converted domain health check configuration
func (h *HealthCheckDTO) ToDomain() config.HealthCheckConfig {
	// return assembled health check config.
	return config.HealthCheckConfig{
		Name:       h.Name,
		Type:       config.HealthCheckType(h.Type),
		Interval:   shared.FromTimeDuration(time.Duration(h.Interval)),
		Timeout:    shared.FromTimeDuration(time.Duration(h.Timeout)),
		Retries:    h.Retries,
		Endpoint:   h.Endpoint,
		Method:     h.Method,
		StatusCode: h.StatusCode,
		Host:       h.Host,
		Port:       h.Port,
		Command:    h.Command,
	}
}

// ToDomain converts LoggingConfigDTO to domain LoggingConfig.
// It transforms global logging settings to the domain model format.
//
// Returns:
//   - config.LoggingConfig: the converted domain logging configuration
func (l *LoggingConfigDTO) ToDomain() config.LoggingConfig {
	// return assembled logging config.
	return config.LoggingConfig{
		BaseDir:  l.BaseDir,
		Defaults: l.Defaults.ToDomain(),
		Daemon:   l.Daemon.ToDomain(),
	}
}

// ToDomain converts DaemonLoggingDTO to domain DaemonLogging.
// It transforms daemon-level logging settings to the domain model format.
//
// Returns:
//   - config.DaemonLogging: the converted domain daemon logging configuration
func (d *DaemonLoggingDTO) ToDomain() config.DaemonLogging {
	writers := make([]config.WriterConfig, 0, len(d.Writers))

	// convert each writer to domain model.
	for i := range d.Writers {
		writers = append(writers, d.Writers[i].ToDomain())
	}

	// return assembled daemon logging config.
	return config.DaemonLogging{
		Writers: writers,
	}
}

// ToDomain converts WriterConfigDTO to domain WriterConfig.
// It transforms writer configuration to the domain model format.
//
// Returns:
//   - config.WriterConfig: the converted domain writer configuration
func (w *WriterConfigDTO) ToDomain() config.WriterConfig {
	// return assembled writer config.
	return config.WriterConfig{
		Type:  w.Type,
		Level: w.Level,
		File:  w.File.ToDomain(),
		JSON:  w.JSON.ToDomain(),
	}
}

// ToDomain converts FileWriterConfigDTO to domain FileWriterConfig.
// It transforms file writer configuration to the domain model format.
//
// Returns:
//   - config.FileWriterConfig: the converted domain file writer configuration
func (f *FileWriterConfigDTO) ToDomain() config.FileWriterConfig {
	// return assembled file writer config.
	return config.FileWriterConfig{
		Path:     f.Path,
		Rotation: f.Rotation.ToDomain(),
	}
}

// ToDomain converts JSONWriterConfigDTO to domain JSONWriterConfig.
// It transforms JSON writer configuration to the domain model format.
//
// Returns:
//   - config.JSONWriterConfig: the converted domain JSON writer configuration
func (j *JSONWriterConfigDTO) ToDomain() config.JSONWriterConfig {
	// return assembled JSON writer config.
	return config.JSONWriterConfig{
		Path:     j.Path,
		Rotation: j.Rotation.ToDomain(),
	}
}

// ToDomain converts LogDefaultsDTO to domain LogDefaults.
// It maps default logging parameters to the domain model format.
//
// Returns:
//   - config.LogDefaults: the converted domain log defaults
func (l *LogDefaultsDTO) ToDomain() config.LogDefaults {
	// return assembled log defaults.
	return config.LogDefaults{
		TimestampFormat: l.TimestampFormat,
		Rotation:        l.Rotation.ToDomain(),
	}
}

// ToDomain converts RotationConfigDTO to domain RotationConfig.
// It transforms log rotation settings to the domain model format.
//
// Returns:
//   - config.RotationConfig: the converted domain rotation configuration
func (r *RotationConfigDTO) ToDomain() config.RotationConfig {
	// return assembled rotation config.
	return config.RotationConfig{
		MaxSize:  r.MaxSize,
		MaxAge:   r.MaxAge,
		MaxFiles: r.MaxFiles,
		Compress: r.Compress,
	}
}

// ToDomain converts ServiceLoggingDTO to domain ServiceLogging.
// It maps service-specific logging settings to the domain model format.
//
// Returns:
//   - config.ServiceLogging: the converted domain service logging configuration
func (s *ServiceLoggingDTO) ToDomain() config.ServiceLogging {
	// return assembled service logging config.
	return config.ServiceLogging{
		Stdout: s.Stdout.ToDomain(),
		Stderr: s.Stderr.ToDomain(),
	}
}

// ToDomain converts LogStreamConfigDTO to domain LogStreamConfig.
// It transforms individual log stream settings to the domain model format.
//
// Returns:
//   - config.LogStreamConfig: the converted domain log stream configuration
func (l *LogStreamConfigDTO) ToDomain() config.LogStreamConfig {
	// return assembled log stream config.
	return config.LogStreamConfig{
		FilePath:       l.File,
		Format:         l.TimestampFormat,
		RotationConfig: l.Rotation.ToDomain(),
	}
}
