// Package probe_test provides black-box tests for the probe package.
package probe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainprobe "github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/infrastructure/probe"
)

// TestNewHTTPProber tests HTTP prober creation.
func TestNewHTTPProber(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "standard_timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "short_timeout",
			timeout: 100 * time.Millisecond,
		},
		{
			name:    "zero_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober with specified timeout.
			prober := probe.NewHTTPProber(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestHTTPProber_Type tests the Type method.
func TestHTTPProber_Type(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns_http",
			expected: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober.
			prober := probe.NewHTTPProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// TestHTTPProber_Probe tests HTTP probing functionality.
func TestHTTPProber_Probe(t *testing.T) {
	// Create test servers with different responses.
	serverOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer serverOK.Close()

	serverNotFound := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer serverNotFound.Close()

	serverCreated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer serverCreated.Close()

	tests := []struct {
		name          string
		target        domainprobe.Target
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name: "successful_ok_response",
			target: domainprobe.Target{
				Address: serverOK.URL,
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_with_explicit_method",
			target: domainprobe.Target{
				Address: serverOK.URL,
				Method:  http.MethodGet,
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_with_explicit_status",
			target: domainprobe.Target{
				Address:    serverCreated.URL,
				StatusCode: http.StatusCreated,
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "failure_status_mismatch",
			target: domainprobe.Target{
				Address: serverNotFound.URL,
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name: "failure_expected_different_status",
			target: domainprobe.Target{
				Address:    serverOK.URL,
				StatusCode: http.StatusCreated,
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name: "failure_invalid_url",
			target: domainprobe.Target{
				Address: "http://invalid.local:99999",
			},
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
		{
			name: "failure_malformed_url",
			target: domainprobe.Target{
				Address: "://invalid",
			},
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober.
			prober := probe.NewHTTPProber(tt.timeout)
			ctx := context.Background()

			// Perform probe.
			result := prober.Probe(ctx, tt.target)

			// Verify result based on expected outcome.
			if tt.expectSuccess {
				assert.True(t, result.Success)
				assert.NoError(t, result.Error)
				assert.Contains(t, result.Output, "HTTP")
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be measured.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}

// TestHTTPProber_Probe_Methods tests different HTTP methods.
func TestHTTPProber_Probe_Methods(t *testing.T) {
	// Create test server that echoes back the method.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "get_method",
			method: http.MethodGet,
		},
		{
			name:   "head_method",
			method: http.MethodHead,
		},
		{
			name:   "post_method",
			method: http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP prober.
			prober := probe.NewHTTPProber(time.Second)

			target := domainprobe.Target{
				Address: server.URL,
				Method:  tt.method,
			}

			// Perform probe.
			result := prober.Probe(context.Background(), target)

			// Should succeed.
			assert.True(t, result.Success)
		})
	}
}

// TestHTTPProber_Probe_ContextCancellation tests context cancellation.
func TestHTTPProber_Probe_ContextCancellation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "cancelled_context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create prober.
			prober := probe.NewHTTPProber(10 * time.Second)

			// Create already cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			target := domainprobe.Target{
				Address: "http://example.com",
			}

			// Probe should fail due to cancelled context.
			result := prober.Probe(ctx, target)
			assert.False(t, result.Success)
		})
	}
}

// TestErrHTTPStatusMismatch tests the exported error variable.
func TestErrHTTPStatusMismatch(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "error_is_accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is not nil.
			assert.NotNil(t, probe.ErrHTTPStatusMismatch)
			assert.Contains(t, probe.ErrHTTPStatusMismatch.Error(), "status")
		})
	}
}
