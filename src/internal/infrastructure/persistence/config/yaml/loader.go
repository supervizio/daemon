// Package yaml provides YAML configuration loading infrastructure.
package yaml

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/kodflow/daemon/internal/domain/config"
)

// Default configuration values.
const (
	// defaultVersion is the default configuration schema version.
	defaultVersion string = "1"
	// defaultBaseDir is the default base directory for log files.
	defaultBaseDir string = "/var/log/daemon"
	// defaultTimestampFormat is the default timestamp format for logs.
	defaultTimestampFormat string = "iso8601"
	// defaultMaxSize is the default maximum log file size.
	defaultMaxSize string = "100MB"
	// defaultMaxFiles is the default maximum number of rotated log files.
	defaultMaxFiles int = 10
	// defaultMaxRetries is the default maximum restart retries.
	defaultMaxRetries int = 3
	// defaultRestartDelay is the default delay between restart attempts.
	defaultRestartDelay string = "5s"
	// defaultHTTPMethod is the default HTTP method for health checks.
	defaultHTTPMethod string = "GET"
	// defaultHTTPStatus is the default expected HTTP status code.
	defaultHTTPStatus int = 200
	// defaultHealthRetries is the default number of health check retries.
	defaultHealthRetries int = 3
)

// ErrNoConfigurationLoaded is returned when Reload is called without a prior Load.
var ErrNoConfigurationLoaded error = errors.New("no configuration loaded")

// Loader loads configuration from YAML files.
// It maintains state about the last loaded configuration path
// to support configuration reloading.
type Loader struct {
	lastPath string
}

// New creates a new YAML configuration loader.
//
// Returns:
//   - *Loader: a new loader instance ready to load configurations
func New() *Loader {
	// Initialize and return a new loader with default state.
	return &Loader{}
}

// Load reads and parses a configuration file from the given path.
//
// Params:
//   - path: absolute or relative path to the YAML configuration file
//
// Returns:
//   - *config.Config: parsed and validated configuration
//   - error: any error during reading, parsing, or validation
func (l *Loader) Load(path string) (*config.Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 - config path is trusted input
	// Check if file reading failed.
	if err != nil {
		// Return wrapped error for context.
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Parse the YAML data into domain configuration.
	cfg, err := l.Parse(data)
	// Check if parsing failed.
	if err != nil {
		// Return the parse error as-is.
		return nil, err
	}

	// Store the config path in the configuration and loader state.
	cfg.ConfigPath = path
	l.lastPath = path

	// Return the successfully parsed configuration.
	return cfg, nil
}

// Parse parses configuration from YAML bytes.
//
// Params:
//   - data: raw YAML configuration bytes
//
// Returns:
//   - *config.Config: parsed and validated configuration
//   - error: any error during parsing or validation
func (l *Loader) Parse(data []byte) (*config.Config, error) {
	var dto ConfigDTO

	// Unmarshal YAML data into the DTO structure.
	if err := yaml.Unmarshal(data, &dto); err != nil {
		// Return wrapped error for context.
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	// Apply default values to unset fields.
	applyDefaults(&dto)

	// Convert DTO to domain model.
	cfg := dto.ToDomain("")

	// Validate the configuration against domain rules.
	if err := config.Validate(cfg); err != nil {
		// Return wrapped validation error.
		return nil, fmt.Errorf("validating config: %w", err)
	}

	// Return the validated configuration.
	return cfg, nil
}

// Reload reloads configuration from the last loaded path.
//
// Returns:
//   - *config.Config: reloaded and validated configuration
//   - error: error if no configuration was previously loaded or reload fails
func (l *Loader) Reload() (*config.Config, error) {
	// Check if a configuration was previously loaded.
	if l.lastPath == "" {
		// Return error when no previous load exists.
		return nil, fmt.Errorf("%w", ErrNoConfigurationLoaded)
	}
	// Reload from the stored path.
	return l.Load(l.lastPath)
}

// applyDefaults sets default values for unset configuration options.
//
// Params:
//   - cfg: configuration DTO to apply defaults to
func applyDefaults(cfg *ConfigDTO) {
	// Set default version if not specified.
	if cfg.Version == "" {
		cfg.Version = defaultVersion
	}

	// Set default logging base directory if not specified.
	if cfg.Logging.BaseDir == "" {
		cfg.Logging.BaseDir = defaultBaseDir
	}

	// Set default timestamp format if not specified.
	if cfg.Logging.Defaults.TimestampFormat == "" {
		cfg.Logging.Defaults.TimestampFormat = defaultTimestampFormat
	}

	// Set default maximum log file size if not specified.
	if cfg.Logging.Defaults.Rotation.MaxSize == "" {
		cfg.Logging.Defaults.Rotation.MaxSize = defaultMaxSize
	}

	// Set default maximum rotated files if not specified.
	if cfg.Logging.Defaults.Rotation.MaxFiles == 0 {
		cfg.Logging.Defaults.Rotation.MaxFiles = defaultMaxFiles
	}

	// Apply defaults to each service configuration.
	for i := range cfg.Services {
		applyServiceDefaults(&cfg.Services[i], &cfg.Logging)
	}
}

// applyServiceDefaults applies default values to a service configuration.
//
// Params:
//   - svc: service configuration DTO to apply defaults to
//   - logging: global logging configuration for inheriting defaults
func applyServiceDefaults(svc *ServiceConfigDTO, logging *LoggingConfigDTO) {
	// Apply restart configuration defaults.
	applyRestartDefaults(&svc.Restart)

	// Set default stdout log file name if not specified.
	if svc.Logging.Stdout.File == "" {
		svc.Logging.Stdout.File = svc.Name + ".out.log"
	}
	// Inherit stdout timestamp format from global defaults if not specified.
	if svc.Logging.Stdout.TimestampFormat == "" {
		svc.Logging.Stdout.TimestampFormat = logging.Defaults.TimestampFormat
	}
	// Inherit stdout rotation config from global defaults if not specified.
	if svc.Logging.Stdout.Rotation.MaxSize == "" {
		svc.Logging.Stdout.Rotation = logging.Defaults.Rotation
	}

	// Set default stderr log file name if not specified.
	if svc.Logging.Stderr.File == "" {
		svc.Logging.Stderr.File = svc.Name + ".err.log"
	}
	// Inherit stderr timestamp format from global defaults if not specified.
	if svc.Logging.Stderr.TimestampFormat == "" {
		svc.Logging.Stderr.TimestampFormat = logging.Defaults.TimestampFormat
	}
	// Inherit stderr rotation config from global defaults if not specified.
	if svc.Logging.Stderr.Rotation.MaxSize == "" {
		svc.Logging.Stderr.Rotation = logging.Defaults.Rotation
	}

	// Apply defaults to each health check configuration.
	for j := range svc.HealthChecks {
		applyHealthCheckDefaults(&svc.HealthChecks[j])
	}
}

// applyRestartDefaults applies default values to restart configuration.
//
// Params:
//   - restart: restart configuration DTO to apply defaults to
func applyRestartDefaults(restart *RestartConfigDTO) {
	// Set default restart policy if not specified.
	if restart.Policy == "" {
		restart.Policy = string(config.RestartOnFailure)
	}

	// Set default maximum retries if not specified.
	if restart.MaxRetries == 0 {
		restart.MaxRetries = defaultMaxRetries
	}

	// Set default restart delay if not specified.
	if restart.Delay == 0 {
		parsed, _ := parseDuration(defaultRestartDelay)
		restart.Delay = parsed
	}
}

// applyHealthCheckDefaults applies default values to a health check.
//
// Params:
//   - hc: health check DTO to apply defaults to
func applyHealthCheckDefaults(hc *HealthCheckDTO) {
	// Set default retries if not specified.
	if hc.Retries == 0 {
		hc.Retries = defaultHealthRetries
	}

	// Set default HTTP method for HTTP health checks if not specified.
	if hc.Method == "" && hc.Type == string(config.HealthCheckHTTP) {
		hc.Method = defaultHTTPMethod
	}

	// Set default expected status code for HTTP health checks if not specified.
	if hc.StatusCode == 0 && hc.Type == string(config.HealthCheckHTTP) {
		hc.StatusCode = defaultHTTPStatus
	}
}

// parseDuration parses a duration string.
//
// Params:
//   - s: duration string in Go duration format (e.g., "5s", "1m30s")
//
// Returns:
//   - Duration: parsed duration value
//   - error: any error during parsing
func parseDuration(s string) (Duration, error) {
	var duration Duration
	// Use UnmarshalYAML to parse the duration string.
	err := duration.UnmarshalYAML(func(v any) error {
		// Set the string value for parsing.
		if sp, ok := v.(*string); ok {
			*sp = s
		}
		// Return nil to indicate successful unmarshaling.
		return nil
	})
	// Return the parsed duration and any error.
	return duration, err
}
