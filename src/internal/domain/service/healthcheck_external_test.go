// Package service provides domain value objects for service configuration.
package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/service"
)

// TestHealthCheckType_String verifies string representation of health check types.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that each HealthCheckType constant returns its expected
// string representation for HTTP, TCP, and command check types.
func TestHealthCheckType_String(t *testing.T) {
	tests := []struct {
		name     string
		hcType   service.HealthCheckType
		expected string
	}{
		{"http", service.HealthCheckHTTP, "http"},
		{"tcp", service.HealthCheckTCP, "tcp"},
		{"command", service.HealthCheckCommand, "command"},
	}

	// Iterate through all test cases to verify string conversion
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.hcType.String())
		})
	}
}

// TestHealthCheckType_IsHTTP verifies the IsHTTP method for all health check types.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that IsHTTP returns true only for HealthCheckHTTP type
// and false for all other health check types.
func TestHealthCheckType_IsHTTP(t *testing.T) {
	tests := []struct {
		name     string
		hcType   service.HealthCheckType
		expected bool
	}{
		{"http returns true", service.HealthCheckHTTP, true},
		{"tcp returns false", service.HealthCheckTCP, false},
		{"command returns false", service.HealthCheckCommand, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.hcType.IsHTTP())
		})
	}
}

// TestHealthCheckType_IsTCP verifies the IsTCP method for all health check types.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that IsTCP returns true only for HealthCheckTCP type
// and false for all other health check types.
func TestHealthCheckType_IsTCP(t *testing.T) {
	tests := []struct {
		name     string
		hcType   service.HealthCheckType
		expected bool
	}{
		{"http returns false", service.HealthCheckHTTP, false},
		{"tcp returns true", service.HealthCheckTCP, true},
		{"command returns false", service.HealthCheckCommand, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.hcType.IsTCP())
		})
	}
}

// TestHealthCheckType_IsCommand verifies the IsCommand method for all health check types.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that IsCommand returns true only for HealthCheckCommand type
// and false for all other health check types.
func TestHealthCheckType_IsCommand(t *testing.T) {
	tests := []struct {
		name     string
		hcType   service.HealthCheckType
		expected bool
	}{
		{"http returns false", service.HealthCheckHTTP, false},
		{"tcp returns false", service.HealthCheckTCP, false},
		{"command returns true", service.HealthCheckCommand, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.hcType.IsCommand())
		})
	}
}

// TestDefaultHealthCheckConfig verifies default configuration values for each health check type.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that DefaultHealthCheckConfig returns proper default values
// for HTTP, TCP, and command health check configurations including retries,
// HTTP method, and expected status code.
func TestDefaultHealthCheckConfig(t *testing.T) {
	// defaultRetries is the expected number of retry attempts.
	const defaultRetries int = 3
	// httpStatusOK is the expected HTTP status code for successful checks.
	const httpStatusOK int = 200

	tests := []struct {
		name           string
		checkType      service.HealthCheckType
		expectedType   service.HealthCheckType
		checkMethod    bool
		expectedMethod string
		checkStatus    bool
		expectedStatus int
	}{
		{
			name:           "HTTP defaults",
			checkType:      service.HealthCheckHTTP,
			expectedType:   service.HealthCheckHTTP,
			checkMethod:    true,
			expectedMethod: "GET",
			checkStatus:    true,
			expectedStatus: httpStatusOK,
		},
		{
			name:         "TCP defaults",
			checkType:    service.HealthCheckTCP,
			expectedType: service.HealthCheckTCP,
		},
		{
			name:         "Command defaults",
			checkType:    service.HealthCheckCommand,
			expectedType: service.HealthCheckCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.DefaultHealthCheckConfig(tt.checkType)
			assert.Equal(t, tt.expectedType, cfg.Type)
			assert.Equal(t, defaultRetries, cfg.Retries)

			if tt.checkMethod {
				assert.Equal(t, tt.expectedMethod, cfg.Method)
			}

			if tt.checkStatus {
				assert.Equal(t, tt.expectedStatus, cfg.StatusCode)
			}
		})
	}
}
