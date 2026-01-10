// Package probe provides internal tests for TCP prober.
package probe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTCPProber_internalFields tests internal struct fields.
func TestTCPProber_internalFields(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "timeout_is_stored",
			timeout:         5 * time.Second,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "zero_timeout_is_stored",
			timeout:         0,
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TCP prober.
			prober := NewTCPProber(tt.timeout)

			// Verify internal timeout field.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
		})
	}
}

// TestProberTypeTCP_constant tests the constant value.
func TestProberTypeTCP_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeTCP)
		})
	}
}
