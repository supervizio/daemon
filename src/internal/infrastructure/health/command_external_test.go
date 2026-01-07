// Package health_test provides black-box tests for the health infrastructure package.
package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/health"
)

// TestCommandChecker_Name tests the Name method returns the expected checker name.
func TestCommandChecker_Name(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedName string
	}{
		{
			name: "returns_custom_name",
			config: &service.HealthCheckConfig{
				Name:    "custom-command-check",
				Command: "echo hello",
				Timeout: shared.Seconds(5),
			},
			expectedName: "custom-command-check",
		},
		{
			name: "returns_generated_name_from_command",
			config: &service.HealthCheckConfig{
				Command: "curl http://localhost",
				Timeout: shared.Seconds(3),
			},
			expectedName: "cmd-curl",
		},
		{
			name: "returns_default_name_for_empty_command",
			config: &service.HealthCheckConfig{
				Command: "",
				Timeout: shared.Seconds(1),
			},
			expectedName: "cmd-empty",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewCommandChecker(tt.config)

			// Verify the checker name matches expected value.
			assert.Equal(t, tt.expectedName, checker.Name())
		})
	}
}

// TestCommandChecker_Type tests the Type method returns the expected checker type.
func TestCommandChecker_Type(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedType string
	}{
		{
			name: "returns_command_type",
			config: &service.HealthCheckConfig{
				Name:    "test-checker",
				Command: "echo hello",
				Timeout: shared.Seconds(5),
			},
			expectedType: "command",
		},
		{
			name: "returns_command_type_for_empty_command",
			config: &service.HealthCheckConfig{
				Command: "",
				Timeout: shared.Seconds(1),
			},
			expectedType: "command",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewCommandChecker(tt.config)

			// Verify the checker type is always command.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestNewCommandChecker tests the NewCommandChecker constructor with various configurations.
func TestNewCommandChecker(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedName string
		expectedType string
	}{
		{
			name: "with_custom_name",
			config: &service.HealthCheckConfig{
				Name:    "custom-command-check",
				Command: "echo hello",
				Timeout: shared.Seconds(5),
			},
			expectedName: "custom-command-check",
			expectedType: "command",
		},
		{
			name: "without_name_generates_from_command",
			config: &service.HealthCheckConfig{
				Command: "curl http://localhost",
				Timeout: shared.Seconds(3),
			},
			expectedName: "cmd-curl",
			expectedType: "command",
		},
		{
			name: "empty_command_uses_default_name",
			config: &service.HealthCheckConfig{
				Command: "",
				Timeout: shared.Seconds(1),
			},
			expectedName: "cmd-empty",
			expectedType: "command",
		},
		{
			name: "command_with_arguments",
			config: &service.HealthCheckConfig{
				Command: "test -f /etc/passwd",
				Timeout: shared.Seconds(1),
			},
			expectedName: "cmd-test",
			expectedType: "command",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewCommandChecker(tt.config)

			// Verify the checker name matches expected value.
			assert.Equal(t, tt.expectedName, checker.Name())
			// Verify the checker type is always command.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestCommandChecker_Check tests the Check method with various scenarios.
func TestCommandChecker_Check(t *testing.T) {
	tests := []struct {
		name           string
		config         *service.HealthCheckConfig
		setupContext   func() context.Context
		expectedStatus domain.Status
		messageContain string
		expectError    bool
	}{
		{
			name: "healthy_successful_command",
			config: &service.HealthCheckConfig{
				Name:    "test-echo",
				Command: "echo hello",
				Timeout: shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusHealthy,
			messageContain: "hello",
			expectError:    false,
		},
		{
			name: "healthy_true_command",
			config: &service.HealthCheckConfig{
				Name:    "test-true",
				Command: "true",
				Timeout: shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusHealthy,
			messageContain: "",
			expectError:    false,
		},
		{
			name: "unhealthy_failed_command",
			config: &service.HealthCheckConfig{
				Name:    "test-false",
				Command: "false",
				Timeout: shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "command failed",
			expectError:    true,
		},
		{
			name: "unhealthy_empty_command",
			config: &service.HealthCheckConfig{
				Name:    "test-empty",
				Command: "",
				Timeout: shared.Seconds(1),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "empty command",
			expectError:    true,
		},
		{
			name: "unhealthy_nonexistent_command",
			config: &service.HealthCheckConfig{
				Name:    "test-nonexistent",
				Command: "nonexistent_command_12345",
				Timeout: shared.Seconds(1),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "command failed",
			expectError:    true,
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewCommandChecker(tt.config)
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
