// Package service provides domain value objects for service configuration.
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateService validates the internal validateService function.
//
// Params:
//   - t: testing context for assertions
func TestValidateService(t *testing.T) {
	tests := []struct {
		name      string
		svc       *ServiceConfig
		wantErr   bool
		errTarget error
	}{
		{
			name:      "empty name",
			svc:       &ServiceConfig{Name: "", Command: "/bin/app"},
			wantErr:   true,
			errTarget: ErrEmptyServiceName,
		},
		{
			name:      "empty command",
			svc:       &ServiceConfig{Name: "app", Command: ""},
			wantErr:   true,
			errTarget: ErrEmptyCommand,
		},
		{
			name:    "valid service",
			svc:     &ServiceConfig{Name: "app", Command: "/bin/app"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateService(tt.svc)

			if tt.wantErr {
				assert.ErrorIs(t, err, tt.errTarget)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateHealthCheck validates the internal validateHealthCheck function.
//
// Params:
//   - t: testing context for assertions
func TestValidateHealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		hc        *HealthCheckConfig
		wantErr   bool
		errTarget error
	}{
		{
			name:    "valid HTTP check",
			hc:      &HealthCheckConfig{Type: HealthCheckHTTP, Endpoint: "http://localhost/health"},
			wantErr: false,
		},
		{
			name:      "HTTP missing endpoint",
			hc:        &HealthCheckConfig{Type: HealthCheckHTTP, Endpoint: ""},
			wantErr:   true,
			errTarget: ErrMissingHTTPEndpoint,
		},
		{
			name:    "valid TCP check",
			hc:      &HealthCheckConfig{Type: HealthCheckTCP, Host: "localhost", Port: 8080},
			wantErr: false,
		},
		{
			name:      "TCP missing host",
			hc:        &HealthCheckConfig{Type: HealthCheckTCP, Host: "", Port: 8080},
			wantErr:   true,
			errTarget: ErrMissingTCPHost,
		},
		{
			name:      "TCP missing port",
			hc:        &HealthCheckConfig{Type: HealthCheckTCP, Host: "localhost", Port: 0},
			wantErr:   true,
			errTarget: ErrMissingTCPPort,
		},
		{
			name:    "valid command check",
			hc:      &HealthCheckConfig{Type: HealthCheckCommand, Command: "/bin/check"},
			wantErr: false,
		},
		{
			name:      "command missing command",
			hc:        &HealthCheckConfig{Type: HealthCheckCommand, Command: ""},
			wantErr:   true,
			errTarget: ErrMissingHealthCommand,
		},
		{
			name:    "invalid type",
			hc:      &HealthCheckConfig{Type: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHealthCheck(tt.hc)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
