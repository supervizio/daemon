// Package health_test provides black-box tests for the health infrastructure package.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphealth "github.com/kodflow/daemon/internal/application/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/health"
)

// TestNewChecker tests the NewChecker factory function with various health check types.
func TestNewChecker(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedType string
		expectError  bool
	}{
		{
			name: "http_checker",
			config: &service.HealthCheckConfig{
				Type:       service.HealthCheckHTTP,
				Name:       "http-test",
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			expectedType: "http",
			expectError:  false,
		},
		{
			name: "tcp_checker",
			config: &service.HealthCheckConfig{
				Type:    service.HealthCheckTCP,
				Name:    "tcp-test",
				Host:    "localhost",
				Port:    8080,
				Timeout: shared.Seconds(5),
			},
			expectedType: "tcp",
			expectError:  false,
		},
		{
			name: "command_checker",
			config: &service.HealthCheckConfig{
				Type:    service.HealthCheckCommand,
				Name:    "cmd-test",
				Command: "echo hello",
				Timeout: shared.Seconds(5),
			},
			expectedType: "command",
			expectError:  false,
		},
		{
			name: "unknown_type_returns_error",
			config: &service.HealthCheckConfig{
				Type:    service.HealthCheckType("unknown"),
				Name:    "unknown-test",
				Timeout: shared.Seconds(5),
			},
			expectedType: "",
			expectError:  true,
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := health.NewChecker(tt.config)

			// Verify error state matches expectation.
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, checker)
				assert.ErrorIs(t, err, health.ErrUnknownHealthCheckType)
				// Return early for error cases.
				return
			}

			require.NoError(t, err)
			require.NotNil(t, checker)
			// Verify the checker type matches expected value.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestNewFactory tests the NewFactory constructor.
func TestNewFactory(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates_factory_instance",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := health.NewFactory()

			// Verify factory is not nil.
			assert.NotNil(t, factory)
		})
	}
}

// TestFactory_Create tests the Factory.Create method with various configurations.
func TestFactory_Create(t *testing.T) {
	tests := []struct {
		name         string
		config       apphealth.CheckerConfig
		expectedType string
		expectError  bool
	}{
		{
			name: "creates_http_checker",
			config: apphealth.CheckerConfig{
				Type:       "http",
				Name:       "http-factory-test",
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
			},
			expectedType: "http",
			expectError:  false,
		},
		{
			name: "creates_tcp_checker",
			config: apphealth.CheckerConfig{
				Type: "tcp",
				Name: "tcp-factory-test",
				Host: "localhost",
				Port: 8080,
			},
			expectedType: "tcp",
			expectError:  false,
		},
		{
			name: "creates_command_checker",
			config: apphealth.CheckerConfig{
				Type:    "command",
				Name:    "cmd-factory-test",
				Command: "echo hello",
			},
			expectedType: "command",
			expectError:  false,
		},
		{
			name: "unknown_type_returns_error",
			config: apphealth.CheckerConfig{
				Type: "invalid",
				Name: "invalid-test",
			},
			expectedType: "",
			expectError:  true,
		},
	}

	factory := health.NewFactory()

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := factory.Create(tt.config)

			// Verify error state matches expectation.
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, checker)
				// Return early for error cases.
				return
			}

			require.NoError(t, err)
			require.NotNil(t, checker)
			// Verify the checker type matches expected value.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}
