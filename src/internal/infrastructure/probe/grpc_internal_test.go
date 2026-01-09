// Package probe provides internal tests for gRPC prober.
package probe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGRPCProber_internalFields tests internal struct fields.
func TestGRPCProber_internalFields(t *testing.T) {
	tests := []struct {
		name             string
		timeout          time.Duration
		expectedTimeout  time.Duration
		expectedInsecure bool
		useSecure        bool
	}{
		{
			name:             "insecure_prober",
			timeout:          5 * time.Second,
			expectedTimeout:  5 * time.Second,
			expectedInsecure: true,
			useSecure:        false,
		},
		{
			name:             "secure_prober",
			timeout:          5 * time.Second,
			expectedTimeout:  5 * time.Second,
			expectedInsecure: false,
			useSecure:        true,
		},
		{
			name:             "zero_timeout_insecure",
			timeout:          0,
			expectedTimeout:  0,
			expectedInsecure: true,
			useSecure:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prober *GRPCProber
			if tt.useSecure {
				// Create secure gRPC prober.
				prober = NewGRPCProberSecure(tt.timeout)
			} else {
				// Create insecure gRPC prober.
				prober = NewGRPCProber(tt.timeout)
			}

			// Verify internal fields.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
			assert.Equal(t, tt.expectedInsecure, prober.insecure)
		})
	}
}

// TestProberTypeGRPC_constant tests the constant value.
func TestProberTypeGRPC_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeGRPC)
		})
	}
}
