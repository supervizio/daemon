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

// NewLoader creates a new YAML configuration loader.
//
// Returns:
//   - *Loader: a new loader instance ready to load configurations
func NewLoader() *Loader {
	// return initialized loader.
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
	// file read failed.
	if err != nil {
		// return wrapped error with context.
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg, err := l.Parse(data)
	// parsing or validation failed.
	if err != nil {
		// return parse error to caller.
		return nil, err
	}

	cfg.ConfigPath = path
	l.lastPath = path

	// return successfully loaded config.
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

	// unmarshal YAML bytes into DTO.
	if err := yaml.Unmarshal(data, &dto); err != nil {
		// return YAML parsing error.
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	applyDefaults(&dto)

	cfg := dto.ToDomain("")

	// validate domain configuration.
	if err := config.Validate(cfg); err != nil {
		// return validation error.
		return nil, fmt.Errorf("validating config: %w", err)
	}

	// return validated configuration.
	return cfg, nil
}

// Reload reloads configuration from the last loaded path.
//
// Returns:
//   - *config.Config: reloaded and validated configuration
//   - error: error if no configuration was previously loaded or reload fails
func (l *Loader) Reload() (*config.Config, error) {
	// no previous configuration loaded.
	if l.lastPath == "" {
		// return error indicating no prior load.
		return nil, fmt.Errorf("%w", ErrNoConfigurationLoaded)
	}

	// reload from last known path.
	return l.Load(l.lastPath)
}

// applyDefaults sets default values for unset configuration options.
//
// Params:
//   - cfg: configuration DTO to apply defaults to
func applyDefaults(cfg *ConfigDTO) {
	// set default version if not specified.
	if cfg.Version == "" {
		cfg.Version = defaultVersion
	}

	// set default logging base directory.
	if cfg.Logging.BaseDir == "" {
		cfg.Logging.BaseDir = defaultBaseDir
	}

	// set default timestamp format.
	if cfg.Logging.Defaults.TimestampFormat == "" {
		cfg.Logging.Defaults.TimestampFormat = defaultTimestampFormat
	}

	// set default max log file size.
	if cfg.Logging.Defaults.Rotation.MaxSize == "" {
		cfg.Logging.Defaults.Rotation.MaxSize = defaultMaxSize
	}

	// set default max rotated files count.
	if cfg.Logging.Defaults.Rotation.MaxFiles == 0 {
		cfg.Logging.Defaults.Rotation.MaxFiles = defaultMaxFiles
	}

	// apply service-specific defaults.
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
	applyRestartDefaults(&svc.Restart)

	// Stdout logging defaults - inherit from global config.
	// set default stdout log file name.
	if svc.Logging.Stdout.File == "" {
		svc.Logging.Stdout.File = svc.Name + ".out.log"
	}
	// inherit timestamp format from global config.
	if svc.Logging.Stdout.TimestampFormat == "" {
		svc.Logging.Stdout.TimestampFormat = logging.Defaults.TimestampFormat
	}
	// inherit rotation config from global defaults.
	if svc.Logging.Stdout.Rotation.MaxSize == "" {
		svc.Logging.Stdout.Rotation = logging.Defaults.Rotation
	}

	// Stderr logging defaults - inherit from global config.
	// set default stderr log file name.
	if svc.Logging.Stderr.File == "" {
		svc.Logging.Stderr.File = svc.Name + ".err.log"
	}
	// inherit timestamp format from global config.
	if svc.Logging.Stderr.TimestampFormat == "" {
		svc.Logging.Stderr.TimestampFormat = logging.Defaults.TimestampFormat
	}
	// inherit rotation config from global defaults.
	if svc.Logging.Stderr.Rotation.MaxSize == "" {
		svc.Logging.Stderr.Rotation = logging.Defaults.Rotation
	}

	// apply health check defaults.
	for j := range svc.HealthChecks {
		applyHealthCheckDefaults(&svc.HealthChecks[j])
	}
}

// applyRestartDefaults applies default values to restart configuration.
//
// Params:
//   - restart: restart configuration DTO to apply defaults to
func applyRestartDefaults(restart *RestartConfigDTO) {
	// set default restart policy.
	if restart.Policy == "" {
		restart.Policy = string(config.RestartOnFailure)
	}

	// set default max restart retries.
	if restart.MaxRetries == 0 {
		restart.MaxRetries = defaultMaxRetries
	}

	// set default restart delay.
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
	// set default health check retry count.
	if hc.Retries == 0 {
		hc.Retries = defaultHealthRetries
	}

	// HTTP health checks need method and status code defaults.
	// set default HTTP method for HTTP checks.
	if hc.Method == "" && hc.Type == string(config.HealthCheckHTTP) {
		hc.Method = defaultHTTPMethod
	}
	// set default expected status code for HTTP checks.
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

	// Reuse UnmarshalYAML logic for consistency with YAML parsing.
	err := duration.UnmarshalYAML(func(v any) error {
		// assign string to pointer if type assertion succeeds
		if sp, ok := v.(*string); ok {
			*sp = s
		}

		// callback must return error interface
		return nil
	})

	// return parsed duration and any error.
	return duration, err
}
