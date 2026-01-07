// Package service provides domain value objects for service configuration.
package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/service"
)

// TestValidate tests the Validate function for configuration validation.
//
// Params:
//   - t: the testing context.
func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *service.Config
		wantErr   bool
		errTarget error
	}{
		{
			name: "valid config with single service",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "app", Command: "/bin/app"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with multiple services",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "web", Command: "/bin/web"},
					{Name: "api", Command: "/bin/api"},
				},
			},
			wantErr: false,
		},
		{
			name: "error on empty services",
			cfg: &service.Config{
				Services: nil,
			},
			wantErr:   true,
			errTarget: service.ErrNoServices,
		},
		{
			name: "error on empty service name",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "", Command: "/bin/app"},
				},
			},
			wantErr: true,
		},
		{
			name: "error on empty command",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "app", Command: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "error on duplicate service names",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "app", Command: "/bin/app1"},
					{Name: "app", Command: "/bin/app2"},
				},
			},
			wantErr:   true,
			errTarget: service.ErrDuplicateServiceName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_BasicErrors tests validation error cases for basic configuration issues.
//
// Params:
//   - t: the testing context.
func TestValidate_BasicErrors(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *service.Config
		wantErr   bool
		errTarget error
		errMsg    string
	}{
		{
			name: "no services configured",
			cfg: &service.Config{
				Services: nil,
			},
			wantErr:   true,
			errTarget: service.ErrNoServices,
		},
		{
			name: "empty service name",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "", Command: "/bin/app"},
				},
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "empty command",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "test", Command: ""},
				},
			},
			wantErr: true,
			errMsg:  "service command is required",
		},
		{
			name: "duplicate service name",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "app", Command: "/bin/app"},
					{Name: "app", Command: "/bin/other"},
				},
			},
			wantErr:   true,
			errTarget: service.ErrDuplicateServiceName,
		},
		{
			name: "valid config with single service",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "app1", Command: "/bin/app1"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with multiple services",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{Name: "app1", Command: "/bin/app1"},
					{Name: "app2", Command: "/bin/app2"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget))
				}
				if tt.errMsg != "" {
					assert.ErrorContains(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_HTTPHealthCheck tests validation for HTTP health check configurations.
//
// Params:
//   - t: the testing context.
func TestValidate_HTTPHealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *service.Config
		wantErr   bool
		errTarget error
	}{
		{
			name: "valid HTTP health check",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "web",
						Command: "/bin/web",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckHTTP, Endpoint: "http://localhost:8080/health"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "HTTP missing endpoint",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "web",
						Command: "/bin/web",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckHTTP, Endpoint: ""},
						},
					},
				},
			},
			wantErr:   true,
			errTarget: service.ErrMissingHTTPEndpoint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.errTarget))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_TCPHealthCheck tests validation for TCP health check configurations.
//
// Params:
//   - t: the testing context.
func TestValidate_TCPHealthCheck(t *testing.T) {
	// postgresPort is the default PostgreSQL port number.
	const postgresPort int = 5432

	tests := []struct {
		name      string
		cfg       *service.Config
		wantErr   bool
		errTarget error
	}{
		{
			name: "valid TCP health check",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "db",
						Command: "/bin/postgres",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckTCP, Host: "localhost", Port: postgresPort},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TCP missing host",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "db",
						Command: "/bin/postgres",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckTCP, Host: "", Port: postgresPort},
						},
					},
				},
			},
			wantErr:   true,
			errTarget: service.ErrMissingTCPHost,
		},
		{
			name: "TCP missing port",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "db",
						Command: "/bin/postgres",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckTCP, Host: "localhost", Port: 0},
						},
					},
				},
			},
			wantErr:   true,
			errTarget: service.ErrMissingTCPPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.errTarget))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_CommandHealthCheck tests validation for command health check configurations.
//
// Params:
//   - t: the testing context.
func TestValidate_CommandHealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *service.Config
		wantErr   bool
		errTarget error
	}{
		{
			name: "valid command health check",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "worker",
						Command: "/bin/worker",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckCommand, Command: "/bin/check-health.sh"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "command missing command",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "worker",
						Command: "/bin/worker",
						HealthChecks: []service.HealthCheckConfig{
							{Type: service.HealthCheckCommand, Command: ""},
						},
					},
				},
			},
			wantErr:   true,
			errTarget: service.ErrMissingHealthCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.errTarget))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_InvalidHealthCheckType tests validation with an invalid health check type.
//
// Params:
//   - t: the testing context.
func TestValidate_InvalidHealthCheckType(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *service.Config
		errMsg string
	}{
		{
			name: "invalid type",
			cfg: &service.Config{
				Services: []service.ServiceConfig{
					{
						Name:    "app",
						Command: "/bin/app",
						HealthChecks: []service.HealthCheckConfig{
							{Type: "invalid"},
						},
					},
				},
			},
			errMsg: "invalid health check type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.cfg)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.errMsg)
		})
	}
}
