// Package health provides white-box tests for private functions.
package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// Test_HTTPChecker_getStatusCode tests the private getStatusCode method.
func Test_HTTPChecker_getStatusCode(t *testing.T) {
	// Create test server for status code tests.
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return 200 OK status.
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	createdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return 201 Created status.
		w.WriteHeader(http.StatusCreated)
	}))
	defer createdServer.Close()

	tests := []struct {
		name               string
		config             *service.HealthCheckConfig
		setupContext       func() context.Context
		expectedStatusCode int
		expectError        bool
	}{
		{
			name: "returns_200_ok",
			config: &service.HealthCheckConfig{
				Name:     "test-200",
				Endpoint: okServer.URL,
				Method:   "GET",
				Timeout:  shared.Seconds(5),
			},
			setupContext:       context.Background,
			expectedStatusCode: http.StatusOK,
			expectError:        false,
		},
		{
			name: "returns_201_created",
			config: &service.HealthCheckConfig{
				Name:     "test-201",
				Endpoint: createdServer.URL,
				Method:   "GET",
				Timeout:  shared.Seconds(5),
			},
			setupContext:       context.Background,
			expectedStatusCode: http.StatusCreated,
			expectError:        false,
		},
		{
			name: "error_on_invalid_url",
			config: &service.HealthCheckConfig{
				Name:     "test-invalid-url",
				Endpoint: "://invalid-url",
				Method:   "GET",
				Timeout:  shared.Seconds(1),
			},
			setupContext:       context.Background,
			expectedStatusCode: 0,
			expectError:        true,
		},
		{
			name: "error_on_connection_refused",
			config: &service.HealthCheckConfig{
				Name:     "test-connection-refused",
				Endpoint: "http://127.0.0.1:59999/health",
				Method:   "GET",
				Timeout:  shared.Seconds(1),
			},
			setupContext:       context.Background,
			expectedStatusCode: 0,
			expectError:        true,
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewHTTPChecker(tt.config)
			ctx := tt.setupContext()

			statusCode, err := checker.getStatusCode(ctx)

			// Verify error state matches expectation.
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, statusCode)
				// Return early for error cases.
				return
			}

			require.NoError(t, err)
			// Verify the status code matches expected value.
			assert.Equal(t, tt.expectedStatusCode, statusCode)
		})
	}
}
