// Package config provides domain value objects for service configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
)

// TestProbeTypeConstants verifies that probe type constants have correct values.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that each ProbeType constant returns its expected
// string value for TCP, UDP, HTTP, gRPC, Exec, and ICMP probe types.
func TestProbeTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ProbeTypeTCP", config.ProbeTypeTCP, "tcp"},
		{"ProbeTypeUDP", config.ProbeTypeUDP, "udp"},
		{"ProbeTypeHTTP", config.ProbeTypeHTTP, "http"},
		{"ProbeTypeGRPC", config.ProbeTypeGRPC, "grpc"},
		{"ProbeTypeExec", config.ProbeTypeExec, "exec"},
		{"ProbeTypeICMP", config.ProbeTypeICMP, "icmp"},
	}

	// Iterate through all test cases to verify constant values.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// TestNewProbeConfig verifies creation of a new probe configuration.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that NewProbeConfig creates a probe with the
// specified type and applies default values for all fields.
func TestNewProbeConfig(t *testing.T) {
	// defaultSuccessThreshold is the expected success threshold.
	const defaultSuccessThreshold int = 1
	// defaultFailureThreshold is the expected failure threshold.
	const defaultFailureThreshold int = 3
	// defaultHTTPStatusOK is the expected HTTP status code.
	const defaultHTTPStatusOK int = 200

	tests := []struct {
		name       string
		probeType  string
		expectType string
	}{
		{
			name:       "TCP probe",
			probeType:  "tcp",
			expectType: "tcp",
		},
		{
			name:       "UDP probe",
			probeType:  "udp",
			expectType: "udp",
		},
		{
			name:       "HTTP probe",
			probeType:  "http",
			expectType: "http",
		},
		{
			name:       "gRPC probe",
			probeType:  "grpc",
			expectType: "grpc",
		},
		{
			name:       "Exec probe",
			probeType:  "exec",
			expectType: "exec",
		},
		{
			name:       "ICMP probe",
			probeType:  "icmp",
			expectType: "icmp",
		},
	}

	// Iterate through all test cases to verify probe creation.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := config.NewProbeConfig(tt.probeType)

			// Verify probe type.
			assert.Equal(t, tt.expectType, probe.Type)
			// Verify default thresholds.
			assert.Equal(t, defaultSuccessThreshold, probe.SuccessThreshold)
			assert.Equal(t, defaultFailureThreshold, probe.FailureThreshold)
			// Verify HTTP defaults.
			assert.Equal(t, "GET", probe.Method)
			assert.Equal(t, defaultHTTPStatusOK, probe.StatusCode)
		})
	}
}

// TestDefaultProbeConfig verifies DefaultProbeConfig delegates to NewProbeConfig.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that DefaultProbeConfig creates the same probe
// configuration as NewProbeConfig with identical defaults.
func TestDefaultProbeConfig(t *testing.T) {
	tests := []struct {
		name      string
		probeType string
	}{
		{
			name:      "TCP probe consistency",
			probeType: "tcp",
		},
		{
			name:      "UDP probe consistency",
			probeType: "udp",
		},
		{
			name:      "HTTP probe consistency",
			probeType: "http",
		},
		{
			name:      "gRPC probe consistency",
			probeType: "grpc",
		},
		{
			name:      "Exec probe consistency",
			probeType: "exec",
		},
		{
			name:      "ICMP probe consistency",
			probeType: "icmp",
		},
	}

	// Iterate through all test cases to verify consistency.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create probes using both constructors.
			newProbe := config.NewProbeConfig(tt.probeType)
			defaultProbe := config.DefaultProbeConfig(tt.probeType)

			// Verify both constructors produce identical results.
			assert.Equal(t, newProbe.Type, defaultProbe.Type)
			assert.Equal(t, newProbe.SuccessThreshold, defaultProbe.SuccessThreshold)
			assert.Equal(t, newProbe.FailureThreshold, defaultProbe.FailureThreshold)
			assert.Equal(t, newProbe.Method, defaultProbe.Method)
			assert.Equal(t, newProbe.StatusCode, defaultProbe.StatusCode)
		})
	}
}

// TestProbeConfig_Fields verifies that ProbeConfig fields can be set correctly.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that all fields of ProbeConfig can be assigned
// and retrieved correctly after creation.
func TestProbeConfig_Fields(t *testing.T) {
	// noContentStatusCode is the HTTP status code for no content.
	const noContentStatusCode int = 204

	tests := []struct {
		name           string
		probeType      string
		path           string
		method         string
		statusCode     int
		serviceName    string
		command        string
		args           []string
		expectedType   string
		expectedPath   string
		expectedMethod string
		expectedStatus int
	}{
		{
			name:           "HTTP probe with custom fields",
			probeType:      "http",
			path:           "/health",
			method:         "HEAD",
			statusCode:     noContentStatusCode,
			serviceName:    "",
			command:        "",
			args:           nil,
			expectedType:   "http",
			expectedPath:   "/health",
			expectedMethod: "HEAD",
			expectedStatus: noContentStatusCode,
		},
		{
			name:           "gRPC probe with service name",
			probeType:      "grpc",
			path:           "",
			method:         "GET",
			statusCode:     200,
			serviceName:    "myservice",
			command:        "",
			args:           nil,
			expectedType:   "grpc",
			expectedPath:   "",
			expectedMethod: "GET",
			expectedStatus: 200,
		},
		{
			name:           "Exec probe with command and args",
			probeType:      "exec",
			path:           "",
			method:         "GET",
			statusCode:     200,
			serviceName:    "",
			command:        "/bin/check",
			args:           []string{"--verbose", "--timeout=5"},
			expectedType:   "exec",
			expectedPath:   "",
			expectedMethod: "GET",
			expectedStatus: 200,
		},
	}

	// Iterate through all test cases to verify field assignment.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a probe with specific field values.
			probe := config.NewProbeConfig(tt.probeType)
			probe.Path = tt.path
			probe.Method = tt.method
			probe.StatusCode = tt.statusCode
			probe.Service = tt.serviceName
			probe.Command = tt.command
			probe.Args = tt.args

			// Verify all field values.
			assert.Equal(t, tt.expectedType, probe.Type)
			assert.Equal(t, tt.expectedPath, probe.Path)
			assert.Equal(t, tt.expectedMethod, probe.Method)
			assert.Equal(t, tt.expectedStatus, probe.StatusCode)
			assert.Equal(t, tt.serviceName, probe.Service)
			assert.Equal(t, tt.command, probe.Command)
			assert.Equal(t, tt.args, probe.Args)
		})
	}
}
