// Package config provides domain value objects for service configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
)

// TestNewListenerConfig verifies creation of a new listener configuration.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that NewListenerConfig creates a listener with the
// specified name, port, and defaults to TCP protocol.
func TestNewListenerConfig(t *testing.T) {
	tests := []struct {
		name          string
		listenerName  string
		port          int
		expectedProto string
	}{
		{
			name:          "HTTP listener on port 8080",
			listenerName:  "http",
			port:          8080,
			expectedProto: "tcp",
		},
		{
			name:          "Admin listener on port 9090",
			listenerName:  "admin",
			port:          9090,
			expectedProto: "tcp",
		},
		{
			name:          "gRPC listener on port 50051",
			listenerName:  "grpc",
			port:          50051,
			expectedProto: "tcp",
		},
	}

	// Iterate through all test cases to verify listener creation.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener := config.NewListenerConfig(tt.listenerName, tt.port)

			assert.Equal(t, tt.listenerName, listener.Name)
			assert.Equal(t, tt.port, listener.Port)
			assert.Equal(t, tt.expectedProto, listener.Protocol)
			assert.Nil(t, listener.Probe)
		})
	}
}

// TestListenerConfig_WithProbe verifies adding probe configuration to a listener.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that WithProbe correctly attaches a probe configuration
// to the listener and returns a new listener with the probe.
func TestListenerConfig_WithProbe(t *testing.T) {
	tests := []struct {
		name         string
		listenerName string
		port         int
		probeType    string
		expectedType string
	}{
		{
			name:         "TCP probe on HTTP listener",
			listenerName: "http",
			port:         8080,
			probeType:    "tcp",
			expectedType: "tcp",
		},
		{
			name:         "HTTP probe on API listener",
			listenerName: "api",
			port:         3000,
			probeType:    "http",
			expectedType: "http",
		},
		{
			name:         "gRPC probe on gRPC listener",
			listenerName: "grpc",
			port:         50051,
			probeType:    "grpc",
			expectedType: "grpc",
		},
	}

	// Iterate through all test cases to verify probe attachment.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a base listener to test probe attachment.
			listener := config.NewListenerConfig(tt.listenerName, tt.port)

			// Create a probe configuration to attach.
			probe := config.NewProbeConfig(tt.probeType)

			// Attach probe to listener.
			result := listener.WithProbe(&probe)

			// Verify probe was attached correctly.
			assert.NotNil(t, result.Probe)
			assert.Equal(t, tt.expectedType, result.Probe.Type)
		})
	}
}

// TestListenerConfig_WithTCPProbe verifies adding TCP probe to a listener.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that WithTCPProbe correctly creates and attaches
// a TCP probe configuration to the listener.
func TestListenerConfig_WithTCPProbe(t *testing.T) {
	tests := []struct {
		name         string
		listenerName string
		port         int
		expectedType string
	}{
		{
			name:         "HTTP listener with TCP probe",
			listenerName: "http",
			port:         8080,
			expectedType: "tcp",
		},
		{
			name:         "Admin listener with TCP probe",
			listenerName: "admin",
			port:         9090,
			expectedType: "tcp",
		},
		{
			name:         "Database listener with TCP probe",
			listenerName: "db",
			port:         5432,
			expectedType: "tcp",
		},
	}

	// Iterate through all test cases to verify TCP probe creation.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a listener and add TCP probe.
			listener := config.NewListenerConfig(tt.listenerName, tt.port).WithTCPProbe()

			// Verify TCP probe was attached correctly.
			assert.NotNil(t, listener.Probe)
			assert.Equal(t, tt.expectedType, listener.Probe.Type)
		})
	}
}

// TestListenerConfig_WithHTTPProbe verifies adding HTTP probe to a listener.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that WithHTTPProbe correctly creates and attaches
// an HTTP probe configuration with the specified path.
func TestListenerConfig_WithHTTPProbe(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedPath string
	}{
		{
			name:         "Health endpoint",
			path:         "/health",
			expectedPath: "/health",
		},
		{
			name:         "Ready endpoint",
			path:         "/ready",
			expectedPath: "/ready",
		},
		{
			name:         "Root path",
			path:         "/",
			expectedPath: "/",
		},
	}

	// Iterate through all test cases to verify HTTP probe creation.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a listener and add HTTP probe with path.
			listener := config.NewListenerConfig("http", 8080).WithHTTPProbe(tt.path)

			// Verify HTTP probe was attached correctly.
			assert.NotNil(t, listener.Probe)
			assert.Equal(t, "http", listener.Probe.Type)
			assert.Equal(t, tt.expectedPath, listener.Probe.Path)
		})
	}
}

// TestListenerConfig_WithGRPCProbe verifies adding gRPC probe to a listener.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that WithGRPCProbe correctly creates and attaches
// a gRPC probe configuration with the specified service name.
func TestListenerConfig_WithGRPCProbe(t *testing.T) {
	tests := []struct {
		name            string
		serviceName     string
		expectedService string
	}{
		{
			name:            "Health service",
			serviceName:     "grpc.health.v1.Health",
			expectedService: "grpc.health.v1.Health",
		},
		{
			name:            "Empty service checks overall health",
			serviceName:     "",
			expectedService: "",
		},
		{
			name:            "Custom service",
			serviceName:     "myapp.v1.MyService",
			expectedService: "myapp.v1.MyService",
		},
	}

	// Iterate through all test cases to verify gRPC probe creation.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a listener and add gRPC probe with config.
			listener := config.NewListenerConfig("grpc", 50051).WithGRPCProbe(tt.serviceName)

			// Verify gRPC probe was attached correctly.
			assert.NotNil(t, listener.Probe)
			assert.Equal(t, "grpc", listener.Probe.Type)
			assert.Equal(t, tt.expectedService, listener.Probe.Service)
		})
	}
}

// TestListenerConfig_ChainedProbes verifies that probe methods can be chained.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that calling multiple probe methods in sequence
// results in the last probe being attached to the listener.
func TestListenerConfig_ChainedProbes(t *testing.T) {
	tests := []struct {
		name             string
		listenerName     string
		port             int
		chainedProbeType string
		expectedType     string
		expectedPath     string
	}{
		{
			name:             "TCP then HTTP probe",
			listenerName:     "http",
			port:             8080,
			chainedProbeType: "http",
			expectedType:     "http",
			expectedPath:     "/health",
		},
		{
			name:             "HTTP then gRPC probe",
			listenerName:     "grpc",
			port:             50051,
			chainedProbeType: "grpc",
			expectedType:     "grpc",
			expectedPath:     "",
		},
	}

	// Iterate through all test cases to verify probe chaining.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listener config.ListenerConfig

			// Chain probe methods based on expected final type.
			if tt.chainedProbeType == "http" {
				// Chain TCP then HTTP probes.
				listener = config.NewListenerConfig(tt.listenerName, tt.port).
					WithTCPProbe().
					WithHTTPProbe("/health")
			} else {
				// Chain HTTP then gRPC probes.
				listener = config.NewListenerConfig(tt.listenerName, tt.port).
					WithHTTPProbe("/ready").
					WithGRPCProbe("")
			}

			// Verify the last probe was attached.
			assert.NotNil(t, listener.Probe)
			assert.Equal(t, tt.expectedType, listener.Probe.Type)
		})
	}
}
