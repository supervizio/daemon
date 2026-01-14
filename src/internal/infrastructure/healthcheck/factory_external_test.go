// Package healthcheck_test provides black-box tests for the probe package.
package healthcheck_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/healthcheck"
)

// TestNewFactory tests factory creation.
func TestNewFactory(t *testing.T) {
	tests := []struct {
		name           string
		defaultTimeout time.Duration
	}{
		{
			name:           "default_timeout",
			defaultTimeout: 5 * time.Second,
		},
		{
			name:           "custom_timeout",
			defaultTimeout: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory.
			f := healthcheck.NewFactory(tt.defaultTimeout)

			// Verify factory is not nil.
			require.NotNil(t, f)
		})
	}
}

// TestFactory_Create tests prober creation.
func TestFactory_Create(t *testing.T) {
	tests := []struct {
		name        string
		proberType  string
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "tcp_prober",
			proberType:  "tcp",
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "udp_prober",
			proberType:  "udp",
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "http_prober",
			proberType:  "http",
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "grpc_prober",
			proberType:  "grpc",
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "exec_prober",
			proberType:  "exec",
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "icmp_prober",
			proberType:  "icmp",
			timeout:     time.Second,
			expectError: false,
		},
		{
			name:        "unknown_prober",
			proberType:  "unknown",
			timeout:     time.Second,
			expectError: true,
		},
		{
			name:        "default_timeout",
			proberType:  "tcp",
			timeout:     0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory.
			f := healthcheck.NewFactory(5 * time.Second)

			// Create prober.
			prober, err := f.Create(tt.proberType, tt.timeout)

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, prober)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, prober)
				assert.Equal(t, tt.proberType, prober.Type())
			}
		})
	}
}

// TestFactory_CreateTCP tests TCP prober creation.
func TestFactory_CreateTCP(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "with_timeout",
			timeout: time.Second,
		},
		{
			name:    "default_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory and TCP prober.
			f := healthcheck.NewFactory(5 * time.Second)
			prober := f.CreateTCP(tt.timeout)

			// Verify prober.
			require.NotNil(t, prober)
			assert.Equal(t, "tcp", prober.Type())
		})
	}
}

// TestFactory_CreateUDP tests UDP prober creation.
func TestFactory_CreateUDP(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "with_timeout",
			timeout: time.Second,
		},
		{
			name:    "default_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory and UDP prober.
			f := healthcheck.NewFactory(5 * time.Second)
			prober := f.CreateUDP(tt.timeout)

			// Verify prober.
			require.NotNil(t, prober)
			assert.Equal(t, "udp", prober.Type())
		})
	}
}

// TestFactory_CreateHTTP tests HTTP prober creation.
func TestFactory_CreateHTTP(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "with_timeout",
			timeout: time.Second,
		},
		{
			name:    "default_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory and HTTP prober.
			f := healthcheck.NewFactory(5 * time.Second)
			prober := f.CreateHTTP(tt.timeout)

			// Verify prober.
			require.NotNil(t, prober)
			assert.Equal(t, "http", prober.Type())
		})
	}
}

// TestFactory_CreateGRPC tests gRPC prober creation.
func TestFactory_CreateGRPC(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "with_timeout",
			timeout: time.Second,
		},
		{
			name:    "default_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory and gRPC prober.
			f := healthcheck.NewFactory(5 * time.Second)
			prober := f.CreateGRPC(tt.timeout)

			// Verify prober.
			require.NotNil(t, prober)
			assert.Equal(t, "grpc", prober.Type())
		})
	}
}

// TestFactory_CreateExec tests Exec prober creation.
func TestFactory_CreateExec(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "with_timeout",
			timeout: time.Second,
		},
		{
			name:    "default_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory and Exec prober.
			f := healthcheck.NewFactory(5 * time.Second)
			prober := f.CreateExec(tt.timeout)

			// Verify prober.
			require.NotNil(t, prober)
			assert.Equal(t, "exec", prober.Type())
		})
	}
}

// TestFactory_CreateICMP tests ICMP prober creation.
func TestFactory_CreateICMP(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "with_timeout",
			timeout: time.Second,
		},
		{
			name:    "default_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory and ICMP prober.
			f := healthcheck.NewFactory(5 * time.Second)
			prober := f.CreateICMP(tt.timeout)

			// Verify prober.
			require.NotNil(t, prober)
			assert.Equal(t, "icmp", prober.Type())
		})
	}
}
