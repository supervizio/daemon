// Package health_test provides black-box tests for the health package.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewTarget tests generic target creation.
func TestNewTarget(t *testing.T) {
	tests := []struct {
		name            string
		network         string
		address         string
		expectedNetwork string
		expectedAddress string
	}{
		{
			name:            "tcp_localhost",
			network:         "tcp",
			address:         "localhost:8080",
			expectedNetwork: "tcp",
			expectedAddress: "localhost:8080",
		},
		{
			name:            "tcp4_ip",
			network:         "tcp4",
			address:         "192.168.1.1:9090",
			expectedNetwork: "tcp4",
			expectedAddress: "192.168.1.1:9090",
		},
		{
			name:            "tcp6_ip",
			network:         "tcp6",
			address:         "[::1]:8080",
			expectedNetwork: "tcp6",
			expectedAddress: "[::1]:8080",
		},
		{
			name:            "udp_localhost",
			network:         "udp",
			address:         "localhost:5353",
			expectedNetwork: "udp",
			expectedAddress: "localhost:5353",
		},
		{
			name:            "icmp_ip",
			network:         "icmp",
			address:         "192.168.1.1",
			expectedNetwork: "icmp",
			expectedAddress: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewTarget(tt.network, tt.address)

			// Verify fields.
			assert.Equal(t, tt.expectedNetwork, target.Network)
			assert.Equal(t, tt.expectedAddress, target.Address)
		})
	}
}

// TestNewTCPTarget tests TCP target creation.
func TestNewTCPTarget(t *testing.T) {
	tests := []struct {
		name            string
		address         string
		expectedNetwork string
		expectedAddress string
	}{
		{
			name:            "localhost",
			address:         "localhost:8080",
			expectedNetwork: "tcp",
			expectedAddress: "localhost:8080",
		},
		{
			name:            "ip_address",
			address:         "192.168.1.1:9090",
			expectedNetwork: "tcp",
			expectedAddress: "192.168.1.1:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewTCPTarget(tt.address)

			// Verify fields.
			assert.Equal(t, tt.expectedNetwork, target.Network)
			assert.Equal(t, tt.expectedAddress, target.Address)
		})
	}
}

// TestNewUDPTarget tests UDP target creation.
func TestNewUDPTarget(t *testing.T) {
	tests := []struct {
		name            string
		address         string
		expectedNetwork string
		expectedAddress string
	}{
		{
			name:            "localhost",
			address:         "localhost:5353",
			expectedNetwork: "udp",
			expectedAddress: "localhost:5353",
		},
		{
			name:            "ip_address",
			address:         "8.8.8.8:53",
			expectedNetwork: "udp",
			expectedAddress: "8.8.8.8:53",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewUDPTarget(tt.address)

			// Verify fields.
			assert.Equal(t, tt.expectedNetwork, target.Network)
			assert.Equal(t, tt.expectedAddress, target.Address)
		})
	}
}

// TestNewHTTPTarget tests HTTP target creation.
func TestNewHTTPTarget(t *testing.T) {
	tests := []struct {
		name               string
		address            string
		method             string
		statusCode         int
		expectedMethod     string
		expectedStatusCode int
	}{
		{
			name:               "get_200",
			address:            "http://localhost:8080/health",
			method:             "GET",
			statusCode:         200,
			expectedMethod:     "GET",
			expectedStatusCode: 200,
		},
		{
			name:               "head_204",
			address:            "http://localhost:9090/ready",
			method:             "HEAD",
			statusCode:         204,
			expectedMethod:     "HEAD",
			expectedStatusCode: 204,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewHTTPTarget(tt.address, tt.method, tt.statusCode)

			// Verify fields.
			assert.Equal(t, tt.address, target.Address)
			assert.Equal(t, tt.expectedMethod, target.Method)
			assert.Equal(t, tt.expectedStatusCode, target.StatusCode)
		})
	}
}

// TestNewGRPCTarget tests gRPC target creation.
func TestNewGRPCTarget(t *testing.T) {
	tests := []struct {
		name            string
		address         string
		service         string
		expectedService string
	}{
		{
			name:            "with_service",
			address:         "localhost:50051",
			service:         "myapp.v1.UserService",
			expectedService: "myapp.v1.UserService",
		},
		{
			name:            "without_service",
			address:         "localhost:50051",
			service:         "",
			expectedService: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewGRPCTarget(tt.address, tt.service)

			// Verify fields.
			assert.Equal(t, tt.address, target.Address)
			assert.Equal(t, tt.expectedService, target.Service)
		})
	}
}

// TestNewExecTarget tests exec target creation.
func TestNewExecTarget(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		args            []string
		expectedCommand string
		expectedArgs    []string
	}{
		{
			name:            "simple_command",
			command:         "/app/health.sh",
			args:            nil,
			expectedCommand: "/app/health.sh",
			expectedArgs:    nil,
		},
		{
			name:            "command_with_args",
			command:         "/bin/sh",
			args:            []string{"-c", "echo ok"},
			expectedCommand: "/bin/sh",
			expectedArgs:    []string{"-c", "echo ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewExecTarget(tt.command, tt.args...)

			// Verify fields.
			assert.Equal(t, tt.expectedCommand, target.Command)
			assert.Equal(t, tt.expectedArgs, target.Args)
		})
	}
}

// TestNewICMPTarget tests ICMP target creation.
func TestNewICMPTarget(t *testing.T) {
	tests := []struct {
		name            string
		address         string
		expectedNetwork string
		expectedAddress string
	}{
		{
			name:            "ip_address",
			address:         "192.168.1.1",
			expectedNetwork: "icmp",
			expectedAddress: "192.168.1.1",
		},
		{
			name:            "hostname",
			address:         "google.com",
			expectedNetwork: "icmp",
			expectedAddress: "google.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create target.
			target := health.NewICMPTarget(tt.address)

			// Verify fields.
			assert.Equal(t, tt.expectedNetwork, target.Network)
			assert.Equal(t, tt.expectedAddress, target.Address)
		})
	}
}
