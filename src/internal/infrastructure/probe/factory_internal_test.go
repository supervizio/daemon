// Package probe provides internal tests for the Factory.
package probe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domainprobe "github.com/kodflow/daemon/internal/domain/probe"
)

// TestFactory_internalFields tests internal struct fields.
func TestFactory_internalFields(t *testing.T) {
	tests := []struct {
		name                   string
		defaultTimeout         time.Duration
		expectedDefaultTimeout time.Duration
	}{
		{
			name:                   "standard_timeout",
			defaultTimeout:         5 * time.Second,
			expectedDefaultTimeout: 5 * time.Second,
		},
		{
			name:                   "short_timeout",
			defaultTimeout:         100 * time.Millisecond,
			expectedDefaultTimeout: 100 * time.Millisecond,
		},
		{
			name:                   "zero_timeout",
			defaultTimeout:         0,
			expectedDefaultTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory.
			factory := NewFactory(tt.defaultTimeout)

			// Verify internal field.
			assert.Equal(t, tt.expectedDefaultTimeout, factory.defaultTimeout)
		})
	}
}

// TestProberTypeConstants tests all prober type constants.
func TestProberTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "tcp_constant",
			constant: proberTypeTCP,
			expected: "tcp",
		},
		{
			name:     "udp_constant",
			constant: proberTypeUDP,
			expected: "udp",
		},
		{
			name:     "http_constant",
			constant: proberTypeHTTP,
			expected: "http",
		},
		{
			name:     "grpc_constant",
			constant: proberTypeGRPC,
			expected: "grpc",
		},
		{
			name:     "exec_constant",
			constant: proberTypeExec,
			expected: "exec",
		},
		{
			name:     "icmp_constant",
			constant: proberTypeICMP,
			expected: "icmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant value.
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// TestFactory_Create_usesDefaultTimeout tests that default timeout is used.
func TestFactory_Create_usesDefaultTimeout(t *testing.T) {
	tests := []struct {
		name           string
		defaultTimeout time.Duration
		proberType     string
	}{
		{
			name:           "tcp_uses_default",
			defaultTimeout: 10 * time.Second,
			proberType:     proberTypeTCP,
		},
		{
			name:           "udp_uses_default",
			defaultTimeout: 10 * time.Second,
			proberType:     proberTypeUDP,
		},
		{
			name:           "http_uses_default",
			defaultTimeout: 10 * time.Second,
			proberType:     proberTypeHTTP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory with specific default.
			factory := NewFactory(tt.defaultTimeout)

			// Create prober with zero timeout.
			prober, err := factory.Create(tt.proberType, 0)

			// Verify prober created without error.
			assert.NoError(t, err)
			assert.NotNil(t, prober)
		})
	}
}

// TestErrUnknownProberType tests the error constant.
func TestErrUnknownProberType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "error_is_defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is not nil.
			assert.NotNil(t, ErrUnknownProberType)
			assert.Contains(t, ErrUnknownProberType.Error(), "unknown")
		})
	}
}

// Test_Factory_normalizeTimeout tests the internal normalizeTimeout method.
func Test_Factory_normalizeTimeout(t *testing.T) {
	tests := []struct {
		name            string
		factoryTimeout  time.Duration
		inputTimeout    time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "positive_input_preserved",
			factoryTimeout:  5 * time.Second,
			inputTimeout:    10 * time.Second,
			expectedTimeout: 10 * time.Second,
		},
		{
			name:            "zero_input_uses_factory_default",
			factoryTimeout:  5 * time.Second,
			inputTimeout:    0,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "negative_input_uses_factory_default",
			factoryTimeout:  5 * time.Second,
			inputTimeout:    -1 * time.Second,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "zero_factory_uses_probe_default",
			factoryTimeout:  0,
			inputTimeout:    0,
			expectedTimeout: domainprobe.DefaultTimeout,
		},
		{
			name:            "negative_factory_uses_probe_default",
			factoryTimeout:  -1 * time.Second,
			inputTimeout:    0,
			expectedTimeout: domainprobe.DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory with specified timeout.
			factory := NewFactory(tt.factoryTimeout)

			// Call normalizeTimeout.
			result := factory.normalizeTimeout(tt.inputTimeout)

			// Verify result.
			assert.Equal(t, tt.expectedTimeout, result)
		})
	}
}
