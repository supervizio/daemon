package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/application/health"
)

// TestNewProbeBinding tests the NewProbeBinding constructor with various configurations.
func TestNewProbeBinding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		listenerName string
		probeType    health.ProbeType
		target       health.ProbeTarget
		wantInterval time.Duration
		wantTimeout  time.Duration
	}{
		{
			name:         "http_probe_with_path",
			listenerName: "http-listener",
			probeType:    health.ProbeHTTP,
			target: health.ProbeTarget{
				Address: "localhost:8080",
				Path:    "/health",
				Method:  "GET",
			},
			wantInterval: 10 * time.Second,
			wantTimeout:  5 * time.Second,
		},
		{
			name:         "tcp_probe",
			listenerName: "tcp-listener",
			probeType:    health.ProbeTCP,
			target: health.ProbeTarget{
				Address: "localhost:3000",
			},
			wantInterval: 10 * time.Second,
			wantTimeout:  5 * time.Second,
		},
		{
			name:         "grpc_probe_with_service",
			listenerName: "grpc-listener",
			probeType:    health.ProbeGRPC,
			target: health.ProbeTarget{
				Address: "localhost:50051",
				Service: "grpc.health.v1.Health",
			},
			wantInterval: 10 * time.Second,
			wantTimeout:  5 * time.Second,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			binding := health.NewProbeBinding(tt.listenerName, tt.probeType, tt.target)

			// Verify binding properties.
			assert.Equal(t, tt.listenerName, binding.ListenerName)
			assert.Equal(t, tt.probeType, binding.Type)
			assert.Equal(t, tt.target, binding.Target)
			// Verify default config is applied.
			assert.Equal(t, tt.wantInterval, binding.Config.Interval)
			assert.Equal(t, tt.wantTimeout, binding.Config.Timeout)
		})
	}
}

// TestProbeBinding_WithConfig tests the WithConfig method for custom configurations.
func TestProbeBinding_WithConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		listenerName string
		probeType    health.ProbeType
		target       health.ProbeTarget
		customConfig health.ProbeConfig
	}{
		{
			name:         "custom_intervals_and_thresholds",
			listenerName: "tcp-listener",
			probeType:    health.ProbeTCP,
			target: health.ProbeTarget{
				Address: "localhost:3000",
			},
			customConfig: health.ProbeConfig{
				Interval:         30 * time.Second,
				Timeout:          10 * time.Second,
				SuccessThreshold: 2,
				FailureThreshold: 5,
			},
		},
		{
			name:         "fast_healthcheck",
			listenerName: "http-listener",
			probeType:    health.ProbeHTTP,
			target: health.ProbeTarget{
				Address: "localhost:8080",
				Path:    "/health",
			},
			customConfig: health.ProbeConfig{
				Interval:         1 * time.Second,
				Timeout:          500 * time.Millisecond,
				SuccessThreshold: 1,
				FailureThreshold: 2,
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create binding with default config.
			binding := health.NewProbeBinding(tt.listenerName, tt.probeType, tt.target)

			// Apply custom config.
			result := binding.WithConfig(tt.customConfig)

			// Verify custom config is applied.
			assert.Equal(t, tt.customConfig, binding.Config)
			// Verify method returns self for chaining.
			assert.Equal(t, binding, result)
		})
	}
}
