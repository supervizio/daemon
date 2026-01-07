package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a configuration file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return nil, err
	}

	cfg.ConfigPath = path
	return cfg, nil
}

// Parse parses configuration from YAML bytes.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	// Validate
	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// applyDefaults sets default values for unset configuration options.
func applyDefaults(cfg *Config) {
	// Version defaults
	if cfg.Version == "" {
		cfg.Version = "1"
	}

	// Logging defaults
	if cfg.Logging.BaseDir == "" {
		cfg.Logging.BaseDir = "/var/log/daemon"
	}
	if cfg.Logging.Defaults.TimestampFormat == "" {
		cfg.Logging.Defaults.TimestampFormat = "iso8601"
	}
	if cfg.Logging.Defaults.Rotation.MaxSize == "" {
		cfg.Logging.Defaults.Rotation.MaxSize = "100MB"
	}
	if cfg.Logging.Defaults.Rotation.MaxFiles == 0 {
		cfg.Logging.Defaults.Rotation.MaxFiles = 10
	}

	// Service defaults
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		applyServiceDefaults(svc, &cfg.Logging)
	}
}

// applyServiceDefaults applies default values to a service configuration.
func applyServiceDefaults(svc *ServiceConfig, logging *LoggingConfig) {
	// Restart policy defaults
	if svc.Restart.Policy == "" {
		svc.Restart.Policy = RestartOnFailure
	}
	if svc.Restart.MaxRetries == 0 {
		svc.Restart.MaxRetries = 3
	}
	if svc.Restart.Delay == 0 {
		svc.Restart.Delay = Duration(5e9) // 5s
	}

	// Logging defaults - inherit from global if not set
	if svc.Logging.Stdout.File == "" {
		svc.Logging.Stdout.File = svc.Name + ".out.log"
	}
	if svc.Logging.Stderr.File == "" {
		svc.Logging.Stderr.File = svc.Name + ".err.log"
	}
	if svc.Logging.Stdout.TimestampFormat == "" {
		svc.Logging.Stdout.TimestampFormat = logging.Defaults.TimestampFormat
	}
	if svc.Logging.Stderr.TimestampFormat == "" {
		svc.Logging.Stderr.TimestampFormat = logging.Defaults.TimestampFormat
	}

	// Copy rotation defaults if not set
	if svc.Logging.Stdout.Rotation.MaxSize == "" {
		svc.Logging.Stdout.Rotation = logging.Defaults.Rotation
	}
	if svc.Logging.Stderr.Rotation.MaxSize == "" {
		svc.Logging.Stderr.Rotation = logging.Defaults.Rotation
	}

	// Health check defaults
	for j := range svc.HealthChecks {
		hc := &svc.HealthChecks[j]
		if hc.Retries == 0 {
			hc.Retries = 3
		}
		if hc.Method == "" && hc.Type == HealthCheckHTTP {
			hc.Method = "GET"
		}
		if hc.StatusCode == 0 && hc.Type == HealthCheckHTTP {
			hc.StatusCode = 200
		}
	}
}

// GetServiceLogPath returns the full path for a service log file.
func (c *Config) GetServiceLogPath(serviceName, logFile string) string {
	return filepath.Join(c.Logging.BaseDir, serviceName, logFile)
}

// FindService returns a service configuration by name.
func (c *Config) FindService(name string) *ServiceConfig {
	for i := range c.Services {
		if c.Services[i].Name == name {
			return &c.Services[i]
		}
	}
	return nil
}

// ParseSize parses a size string like "100MB" into bytes.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Check suffixes from longest to shortest to avoid matching "B" before "KB"
	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"G", 1024 * 1024 * 1024},
		{"M", 1024 * 1024},
		{"K", 1024},
		{"B", 1},
	}

	for _, sf := range suffixes {
		if strings.HasSuffix(s, sf.suffix) {
			numStr := strings.TrimSuffix(s, sf.suffix)
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size: %s", s)
			}
			return num * sf.mult, nil
		}
	}

	// Try parsing as plain number (bytes)
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %s", s)
	}
	return num, nil
}
