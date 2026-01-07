// Package health_test provides black-box tests for the health infrastructure package.
package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/health"
)

// TestHTTPChecker_Name tests the Name method returns the expected checker name.
func TestHTTPChecker_Name(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedName string
	}{
		{
			name: "returns_custom_name",
			config: &service.HealthCheckConfig{
				Name:       "custom-http-check",
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			expectedName: "custom-http-check",
		},
		{
			name: "returns_generated_name_from_endpoint",
			config: &service.HealthCheckConfig{
				Endpoint:   "http://localhost:9090/api/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(3),
			},
			expectedName: "http-http://localhost:9090/api/health",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewHTTPChecker(tt.config)

			// Verify the checker name matches expected value.
			assert.Equal(t, tt.expectedName, checker.Name())
		})
	}
}

// TestHTTPChecker_Type tests the Type method returns the expected checker type.
func TestHTTPChecker_Type(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedType string
	}{
		{
			name: "returns_http_type",
			config: &service.HealthCheckConfig{
				Name:       "test-checker",
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			expectedType: "http",
		},
		{
			name: "returns_http_type_with_different_method",
			config: &service.HealthCheckConfig{
				Endpoint: "http://localhost:8080/health",
				Method:   "POST",
				Timeout:  shared.Seconds(1),
			},
			expectedType: "http",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewHTTPChecker(tt.config)

			// Verify the checker type is always http.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestNewHTTPChecker tests the NewHTTPChecker constructor with various configurations.
func TestNewHTTPChecker(t *testing.T) {
	tests := []struct {
		name         string
		config       *service.HealthCheckConfig
		expectedName string
		expectedType string
	}{
		{
			name: "with_custom_name",
			config: &service.HealthCheckConfig{
				Name:       "custom-http-check",
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			expectedName: "custom-http-check",
			expectedType: "http",
		},
		{
			name: "without_name_generates_from_endpoint",
			config: &service.HealthCheckConfig{
				Endpoint:   "http://localhost:9090/api/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(3),
			},
			expectedName: "http-http://localhost:9090/api/health",
			expectedType: "http",
		},
		{
			name: "default_method_is_get",
			config: &service.HealthCheckConfig{
				Name:       "test-default-method",
				Endpoint:   "http://localhost:8080/health",
				StatusCode: 200,
				Timeout:    shared.Seconds(1),
			},
			expectedName: "test-default-method",
			expectedType: "http",
		},
		{
			name: "default_status_code_is_200",
			config: &service.HealthCheckConfig{
				Name:     "test-default-status",
				Endpoint: "http://localhost:8080/health",
				Method:   "GET",
				Timeout:  shared.Seconds(1),
			},
			expectedName: "test-default-status",
			expectedType: "http",
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewHTTPChecker(tt.config)

			// Verify the checker name matches expected value.
			assert.Equal(t, tt.expectedName, checker.Name())
			// Verify the checker type is always http.
			assert.Equal(t, tt.expectedType, checker.Type())
		})
	}
}

// TestNewHTTPChecker_DefaultMethod_UsedInRequest verifies that when Method is empty,
// the default GET method is actually used when making requests.
func TestNewHTTPChecker_DefaultMethod_UsedInRequest(t *testing.T) {
	var receivedMethod string

	// Create test server that records the HTTP method received.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &service.HealthCheckConfig{
		Name:       "test-default-method-used",
		Endpoint:   server.URL,
		Method:     "", // Empty method should default to GET.
		StatusCode: 200,
		Timeout:    shared.Seconds(5),
	}

	checker := health.NewHTTPChecker(cfg)
	result := checker.Check(context.Background())

	// Verify the request was healthy.
	assert.Equal(t, domain.StatusHealthy, result.Status)
	// Verify GET method was used.
	assert.Equal(t, http.MethodGet, receivedMethod)
}

// TestNewHTTPChecker_DefaultStatusCode_UsedInCheck verifies that when StatusCode is 0,
// the default 200 status code is expected.
func TestNewHTTPChecker_DefaultStatusCode_UsedInCheck(t *testing.T) {
	// Create test server that returns 200 OK.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &service.HealthCheckConfig{
		Name:       "test-default-status-used",
		Endpoint:   server.URL,
		Method:     "GET",
		StatusCode: 0, // Zero status code should default to 200.
		Timeout:    shared.Seconds(5),
	}

	checker := health.NewHTTPChecker(cfg)
	result := checker.Check(context.Background())

	// Verify the request was healthy (200 expected and received).
	assert.Equal(t, domain.StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "HTTP 200")
}

// TestHTTPChecker_Check_RedirectNotFollowed verifies that redirects are not followed
// and the redirect status code is returned instead.
func TestHTTPChecker_Check_RedirectNotFollowed(t *testing.T) {
	// Create test server that returns a redirect.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/redirected")
		w.WriteHeader(http.StatusMovedPermanently)
	}))
	defer server.Close()

	cfg := &service.HealthCheckConfig{
		Name:       "test-redirect",
		Endpoint:   server.URL,
		Method:     "GET",
		StatusCode: http.StatusMovedPermanently, // Expect the redirect status code.
		Timeout:    shared.Seconds(5),
	}

	checker := health.NewHTTPChecker(cfg)
	result := checker.Check(context.Background())

	// Verify the redirect status is received without following.
	assert.Equal(t, domain.StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "HTTP 301")
}

// TestHTTPChecker_Check tests the Check method with various scenarios.
func TestHTTPChecker_Check(t *testing.T) {
	// Create test servers for different scenarios.
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return 200 OK for healthy status.
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()

	notFoundServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return 404 Not Found for unhealthy status.
		w.WriteHeader(http.StatusNotFound)
	}))
	defer notFoundServer.Close()

	internalErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return 500 Internal Server Error.
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer internalErrorServer.Close()

	tests := []struct {
		name           string
		config         *service.HealthCheckConfig
		setupContext   func() context.Context
		expectedStatus domain.Status
		messageContain string
		expectError    bool
	}{
		{
			name: "healthy_status_200_ok",
			config: &service.HealthCheckConfig{
				Name:       "test-healthy",
				Endpoint:   healthyServer.URL,
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusHealthy,
			messageContain: "HTTP 200",
			expectError:    false,
		},
		{
			name: "unhealthy_status_mismatch_404",
			config: &service.HealthCheckConfig{
				Name:       "test-404",
				Endpoint:   notFoundServer.URL,
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "unexpected status code",
			expectError:    true,
		},
		{
			name: "unhealthy_status_mismatch_500",
			config: &service.HealthCheckConfig{
				Name:       "test-500",
				Endpoint:   internalErrorServer.URL,
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "unexpected status code",
			expectError:    true,
		},
		{
			name: "unhealthy_connection_refused",
			config: &service.HealthCheckConfig{
				Name:       "test-connection-refused",
				Endpoint:   "http://127.0.0.1:59999/health",
				Method:     "GET",
				StatusCode: 200,
				Timeout:    shared.Seconds(1),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusUnhealthy,
			messageContain: "request failed",
			expectError:    true,
		},
		{
			name: "healthy_accepts_404_when_expected",
			config: &service.HealthCheckConfig{
				Name:       "test-expected-404",
				Endpoint:   notFoundServer.URL,
				Method:     "GET",
				StatusCode: 404,
				Timeout:    shared.Seconds(5),
			},
			setupContext:   context.Background,
			expectedStatus: domain.StatusHealthy,
			messageContain: "HTTP 404",
			expectError:    false,
		},
	}

	// Iterate over test cases using table-driven pattern.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := health.NewHTTPChecker(tt.config)
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
