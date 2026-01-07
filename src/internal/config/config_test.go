package config

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(*testing.T, *Config)
	}{
		{
			name: "valid minimal config",
			yaml: `
version: "1"
services:
  - name: nginx
    command: /usr/sbin/nginx
    health_checks:
      - type: http
        endpoint: http://localhost/health
        interval: 10s
        timeout: 5s
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Services) != 1 {
					t.Errorf("expected 1 service, got %d", len(cfg.Services))
				}
				if cfg.Services[0].Name != "nginx" {
					t.Errorf("expected service name 'nginx', got '%s'", cfg.Services[0].Name)
				}
			},
		},
		{
			name: "applies defaults",
			yaml: `
services:
  - name: app
    command: /bin/app
    health_checks:
      - type: tcp
        host: localhost
        port: 8080
        interval: 5s
        timeout: 2s
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				svc := &cfg.Services[0]
				if svc.Restart.Policy != RestartOnFailure {
					t.Errorf("expected default restart policy 'on-failure', got '%s'", svc.Restart.Policy)
				}
				if svc.Restart.MaxRetries != 3 {
					t.Errorf("expected default max_retries 3, got %d", svc.Restart.MaxRetries)
				}
				if cfg.Logging.BaseDir != "/var/log/daemon" {
					t.Errorf("expected default base_dir '/var/log/daemon', got '%s'", cfg.Logging.BaseDir)
				}
			},
		},
		{
			name: "invalid - no services",
			yaml: `
version: "1"
services: []
`,
			wantErr: true,
		},
		{
			name: "invalid - missing service name",
			yaml: `
services:
  - command: /bin/app
    health_checks:
      - type: command
        command: pgrep app
        interval: 5s
        timeout: 2s
`,
			wantErr: true,
		},
		{
			name: "invalid - missing command",
			yaml: `
services:
  - name: app
    health_checks:
      - type: command
        command: pgrep app
        interval: 5s
        timeout: 2s
`,
			wantErr: true,
		},
		{
			name: "invalid - invalid health check type",
			yaml: `
services:
  - name: app
    command: /bin/app
    health_checks:
      - type: invalid
        interval: 5s
        timeout: 2s
`,
			wantErr: true,
		},
		{
			name: "invalid - duplicate service names",
			yaml: `
services:
  - name: app
    command: /bin/app
    health_checks:
      - type: command
        command: true
        interval: 5s
        timeout: 2s
  - name: app
    command: /bin/app2
    health_checks:
      - type: command
        command: true
        interval: 5s
        timeout: 2s
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"5s", 5 * time.Second, false},
		{"10m", 10 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"500ms", 500 * time.Millisecond, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var d Duration
			err := d.UnmarshalYAML(func(v interface{}) error {
				*(v.(*string)) = tt.input
				return nil
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && d.Duration() != tt.expected {
				t.Errorf("UnmarshalYAML() = %v, want %v", d.Duration(), tt.expected)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"100", 100, false},
		{"100B", 100, false},
		{"1KB", 1024, false},
		{"1K", 1024, false},
		{"10MB", 10 * 1024 * 1024, false},
		{"10M", 10 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ParseSize() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateHTTPHealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		hc      HealthCheckConfig
		wantErr bool
	}{
		{
			name: "valid HTTP check",
			hc: HealthCheckConfig{
				Type:     HealthCheckHTTP,
				Endpoint: "http://localhost:8080/health",
				Interval: Duration(10 * time.Second),
				Timeout:  Duration(5 * time.Second),
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			hc: HealthCheckConfig{
				Type:     HealthCheckHTTP,
				Interval: Duration(10 * time.Second),
				Timeout:  Duration(5 * time.Second),
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			hc: HealthCheckConfig{
				Type:     HealthCheckHTTP,
				Endpoint: "not-a-url",
				Interval: Duration(10 * time.Second),
				Timeout:  Duration(5 * time.Second),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHealthCheck(&tt.hc, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTCPHealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		hc      HealthCheckConfig
		wantErr bool
	}{
		{
			name: "valid TCP check",
			hc: HealthCheckConfig{
				Type:     HealthCheckTCP,
				Host:     "localhost",
				Port:     8080,
				Interval: Duration(5 * time.Second),
				Timeout:  Duration(2 * time.Second),
			},
			wantErr: false,
		},
		{
			name: "missing host",
			hc: HealthCheckConfig{
				Type:     HealthCheckTCP,
				Port:     8080,
				Interval: Duration(5 * time.Second),
				Timeout:  Duration(2 * time.Second),
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			hc: HealthCheckConfig{
				Type:     HealthCheckTCP,
				Host:     "localhost",
				Port:     0,
				Interval: Duration(5 * time.Second),
				Timeout:  Duration(2 * time.Second),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHealthCheck(&tt.hc, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
