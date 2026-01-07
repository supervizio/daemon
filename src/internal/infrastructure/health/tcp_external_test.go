// Package health_test provides black-box tests for the health infrastructure package.
package health_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/health"
)

// TestTCPChecker_Name tests the Name method returns the expected checker name.
func TestTCPChecker_Name(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedName string
	}{
		{
			name: "returns_custom_name",
			config: &service.HealthCheckConfig{
				Name:    "custom-tcp-check",
				Host:    "localhost",
				Port:    8080,
				Timeout: shared.Seconds(5),
			},
			expectedName: "custom-tcp-check",
		},
		{
			name: "returns_generated_name_from_host_and_port",
			config: &service.HealthCheckConfig{
				Host:    "localhost",
				Port:    9090,
				Timeout: shared.Seconds(3),
			},
			expectedName: "tcp-localhost:9090",
		},
		{
			name: "returns_generated_name_with_ip",
			config: &service.HealthCheckConfig{
				Host:    "192.168.1.1",
				Port:    5000,
				Timeout: shared.Seconds(1),
			},
			expectedName: "tcp-192.168.1.1:5000",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewTCPChecker(tt.config)

			// Verify the checker name matches expected value.
			assert.Equal(t, tt.expectedName, checker.Name())
		})
	}
}

// TestTCPChecker_Type tests the Type method returns the expected checker type.
func TestTCPChecker_Type(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedType string
	}{
		{
			name: "returns_tcp_type",
			config: &service.HealthCheckConfig{
				Name:    "test-checker",
				Host:    "localhost",
				Port:    8080,
				Timeout: shared.Seconds(5),
			},
			expectedType: "tcp",
		},
		{
			name: "returns_tcp_type_with_ip_address",
			config: &service.HealthCheckConfig{
				Host:    "127.0.0.1",
				Port:    3000,
				Timeout: shared.Seconds(1),
			},
			expectedType: "tcp",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewTCPChecker(tt.config)

			// Verify the checker type is always tcp.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestNewTCPChecker tests the NewTCPChecker constructor with various configurations.
func TestNewTCPChecker(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedName string
		expectedType string
	}{
		{
			name: "with_custom_name",
			config: &service.HealthCheckConfig{
				Name:    "custom-tcp-check",
				Host:    "localhost",
				Port:    8080,
				Timeout: shared.Seconds(5),
			},
			expectedName: "custom-tcp-check",
			expectedType: "tcp",
		},
		{
			name: "without_name_generates_default",
			config: &service.HealthCheckConfig{
				Host:    "localhost",
				Port:    9090,
				Timeout: shared.Seconds(3),
			},
			expectedName: "tcp-localhost:9090",
			expectedType: "tcp",
		},
		{
			name: "with_ip_address",
			config: &service.HealthCheckConfig{
				Name:    "test-checker",
				Host:    "127.0.0.1",
				Port:    3000,
				Timeout: shared.Seconds(1),
			},
			expectedName: "test-checker",
			expectedType: "tcp",
		},
		{
			name: "without_name_with_ip",
			config: &service.HealthCheckConfig{
				Host:    "192.168.1.1",
				Port:    5000,
				Timeout: shared.Seconds(1),
			},
			expectedName: "tcp-192.168.1.1:5000",
			expectedType: "tcp",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewTCPChecker(tt.config)

			// Verify the checker name matches expected value.
			assert.Equal(t, tt.expectedName, checker.Name())
			// Verify the checker type is always TCP.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestTCPChecker_Check tests the Check method with various scenarios.
func TestTCPChecker_Check(t *testing.T) {
	// Create a listener for healthy check tests.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close() //nolint:errcheck // Close listener when test completes
	addr := listener.Addr().(*net.TCPAddr)

	tests := []struct {
		name           string
		config         *service.HealthCheckConfig
		setupContext   func() context.Context
		expectedStatus domain.Status
		messageContain string
		expectError    bool
	}{
		{
			name: "healthy_connection_successful",
			config: &service.HealthCheckConfig{
				Name:    "test-healthy",
				Host:    "127.0.0.1",
				Port:    addr.Port,
				Timeout: shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusHealthy,
			messageContain: "connected to",
			expectError:    false,
		},
		{
			name: "unhealthy_connection_refused",
			config: &service.HealthCheckConfig{
				Name:    "test-unhealthy",
				Host:    "127.0.0.1",
				Port:    59999,
				Timeout: shared.Seconds(1),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "connection failed",
			expectError:    true,
		},
		{
			name: "unhealthy_context_canceled",
			config: &service.HealthCheckConfig{
				Name:    "test-canceled",
				Host:    "10.255.255.1",
				Port:    80,
				Timeout: shared.Seconds(30),
			},
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				// Cancel context immediately to simulate cancellation.
				cancel()
				// Return the canceled context.
				return ctx
			},
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "",
			expectError:    true,
		},
		{
			name: "unhealthy_timeout",
			config: &service.HealthCheckConfig{
				Name:    "test-timeout",
				Host:    "10.255.255.1",
				Port:    80,
				Timeout: shared.FromTimeDuration(100 * time.Millisecond),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "",
			expectError:    true,
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewTCPChecker(tt.config)
			ctx := tt.setupContext()

			result := checker.Check(ctx)

			// Verify the status matches expected value.
			assert.Equal(t, tt.expectedStatus, result.Status)
			// Verify the duration is positive.
			assert.Greater(t, result.Duration, time.Duration(0))

			// Verify message contains expected substring if specified.
			if tt.messageContain != "" {
				assert.Contains(t, result.Message, tt.messageContain)
			}

			// Verify error state matches expectation.
			if tt.expectError {
				assert.NotNil(t, result.Error)
			} else {
				assert.Nil(t, result.Error)
			}
		})
	}
}
