// Package config provides domain value objects for service configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
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
		hcType   config.HealthCheckType
		expected string
	}{
		{"http", config.HealthCheckHTTP, "http"},
		{"tcp", config.HealthCheckTCP, "tcp"},
		{"command", config.HealthCheckCommand, "command"},
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
		hcType   config.HealthCheckType
		expected bool
	}{
		{"http returns true", config.HealthCheckHTTP, true},
		{"tcp returns false", config.HealthCheckTCP, false},
		{"command returns false", config.HealthCheckCommand, false},
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
		hcType   config.HealthCheckType
		expected bool
	}{
		{"http returns false", config.HealthCheckHTTP, false},
		{"tcp returns true", config.HealthCheckTCP, true},
		{"command returns false", config.HealthCheckCommand, false},
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
		hcType   config.HealthCheckType
		expected bool
	}{
		{"http returns false", config.HealthCheckHTTP, false},
		{"tcp returns false", config.HealthCheckTCP, false},
		{"command returns true", config.HealthCheckCommand, true},
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
		checkType      config.HealthCheckType
		expectedType   config.HealthCheckType
		checkMethod    bool
		expectedMethod string
		checkStatus    bool
		expectedStatus int
	}{
		{
			name:           "HTTP defaults",
			checkType:      config.HealthCheckHTTP,
			expectedType:   config.HealthCheckHTTP,
			checkMethod:    true,
			expectedMethod: "GET",
			checkStatus:    true,
			expectedStatus: httpStatusOK,
		},
		{
			name:         "TCP defaults",
			checkType:    config.HealthCheckTCP,
			expectedType: config.HealthCheckTCP,
		},
		{
			name:         "Command defaults",
			checkType:    config.HealthCheckCommand,
			expectedType: config.HealthCheckCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultHealthCheckConfig(tt.checkType)
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
