// Package listener provides white-box tests for internal functions.
package listener

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_formatPort tests the internal formatPort function.
func Test_formatPort(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected string
	}{
		{
			name:     "standard_http_port",
			port:     80,
			expected: "80",
		},
		{
			name:     "standard_https_port",
			port:     443,
			expected: "443",
		},
		{
			name:     "high_port",
			port:     8080,
			expected: "8080",
		},
		{
			name:     "zero_port",
			port:     0,
			expected: "0",
		},
		{
			name:     "max_port",
			port:     65535,
			expected: "65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call formatPort.
			result := formatPort(tt.port)

			// Verify result.
			assert.Equal(t, tt.expected, result)
		})
	}
}
