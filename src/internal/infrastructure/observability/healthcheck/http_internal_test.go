// Package healthcheck provides internal tests for HTTP prober.
package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestHTTPProber_internalFields tests internal struct fields.
func TestHTTPProber_internalFields(t *testing.T) {
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
			name:            "zero_timeout_normalized_to_default",
			timeout:         0,
			expectedTimeout: health.DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober.
			prober := NewHTTPProber(tt.timeout)

			// Verify internal timeout field.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)

			// Verify client is created.
			require.NotNil(t, prober.client)
		})
	}
}

// TestHTTPProber_getStatusCode tests the internal getStatusCode method.
func TestHTTPProber_getStatusCode(t *testing.T) {
	// Create test servers with different responses.
	serverOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer serverOK.Close()

	serverCreated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer serverCreated.Close()

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "get_ok_status",
			method:         http.MethodGet,
			url:            serverOK.URL,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "get_created_status",
			method:         http.MethodGet,
			url:            serverCreated.URL,
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "invalid_url",
			method:         http.MethodGet,
			url:            "http://invalid.local:99999",
			expectedStatus: 0,
			expectError:    true,
		},
		{
			name:           "malformed_url",
			method:         http.MethodGet,
			url:            "://invalid",
			expectedStatus: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober.
			prober := NewHTTPProber(100 * time.Millisecond)
			ctx := context.Background()

			// Call internal method with empty path.
			statusCode, err := prober.getStatusCode(ctx, tt.method, tt.url, "")

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, statusCode)
			}
		})
	}
}

// TestProberTypeHTTP_constant tests the constant value.
func TestProberTypeHTTP_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeHTTP)
		})
	}
}

// TestDefaultHTTPMethod_constant tests the default HTTP method constant.
func TestDefaultHTTPMethod_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "default_is_get",
			expected: http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify default method.
			assert.Equal(t, tt.expected, defaultHTTPMethod)
		})
	}
}

// TestDefaultHTTPStatusCode_constant tests the default status code constant.
func TestDefaultHTTPStatusCode_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "default_is_ok",
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify default status code.
			assert.Equal(t, tt.expected, defaultHTTPStatusCode)
		})
	}
}

// TestHTTPProber_getStatusCode_urlHandling tests URL parsing and path handling.
func TestHTTPProber_getStatusCode_urlHandling(t *testing.T) {
	// Create test server that echoes back the request path.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name        string
		url         string
		path        string
		expectError bool
	}{
		{
			name:        "url_with_path_appended",
			url:         server.URL,
			path:        "/health",
			expectError: false,
		},
		{
			name:        "url_with_trailing_slash_and_path",
			url:         server.URL + "/",
			path:        "/status",
			expectError: false,
		},
		{
			name:        "url_with_path_no_leading_slash",
			url:         server.URL,
			path:        "api/health",
			expectError: false,
		},
		{
			name:        "url_host_without_scheme_fails_parse",
			url:         "127.0.0.1:8080",
			path:        "",
			expectError: true,
		},
		{
			name:        "url_without_scheme_with_path_defaults_to_http",
			url:         "localhost/path",
			path:        "",
			expectError: true, // Connection will fail but parsing succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober with short timeout.
			prober := NewHTTPProber(100 * time.Millisecond)
			ctx := context.Background()

			// Call internal method.
			statusCode, err := prober.getStatusCode(ctx, http.MethodGet, tt.url, tt.path)

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, statusCode)
			}
		})
	}
}

// TestHTTPProber_getStatusCode_contextCancellation tests context cancellation.
func TestHTTPProber_getStatusCode_contextCancellation(t *testing.T) {
	// Create test server that delays response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name        string
		cancelAfter time.Duration
		expectError bool
	}{
		{
			name:        "cancelled_context_returns_error",
			cancelAfter: 10 * time.Millisecond,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober.
			prober := NewHTTPProber(5 * time.Second)

			// Create context that will be cancelled.
			ctx, cancel := context.WithCancel(context.Background())

			// Cancel context after short delay.
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			// Call internal method.
			_, err := prober.getStatusCode(ctx, http.MethodGet, server.URL, "")

			// Verify error due to cancellation.
			if tt.expectError {
				assert.Error(t, err)
			}
		})
	}
}
