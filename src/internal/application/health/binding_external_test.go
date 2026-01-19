package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/application/health"
)

// TestProbeType tests the ProbeType constants.
func TestProbeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		probe    health.ProbeType
		expected string
	}{
		{
			name:     "tcp_probe_type",
			probe:    health.ProbeTCP,
			expected: "tcp",
		},
		{
			name:     "udp_probe_type",
			probe:    health.ProbeUDP,
			expected: "udp",
		},
		{
			name:     "http_probe_type",
			probe:    health.ProbeHTTP,
			expected: "http",
		},
		{
			name:     "grpc_probe_type",
			probe:    health.ProbeGRPC,
			expected: "grpc",
		},
		{
			name:     "exec_probe_type",
			probe:    health.ProbeExec,
			expected: "exec",
		},
		{
			name:     "icmp_probe_type",
			probe:    health.ProbeICMP,
			expected: "icmp",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify probe type string value.
			assert.Equal(t, tt.expected, string(tt.probe))
		})
	}
}
