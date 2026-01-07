package health

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPChecker(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	cfg := &config.HealthCheckConfig{
		Name:       "test-http",
		Type:       config.HealthCheckHTTP,
		Endpoint:   server.URL,
		Method:     "GET",
		StatusCode: 200,
		Timeout:    config.Duration(5 * time.Second),
	}

	checker := NewHTTPChecker(cfg)

	assert.Equal(t, "test-http", checker.Name())
	assert.Equal(t, "http", checker.Type())

	result := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "200")
}

func TestHTTPCheckerUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.HealthCheckConfig{
		Name:       "test-http",
		Type:       config.HealthCheckHTTP,
		Endpoint:   server.URL,
		StatusCode: 200,
		Timeout:    config.Duration(5 * time.Second),
	}

	checker := NewHTTPChecker(cfg)
	result := checker.Check(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "500")
}

func TestHTTPCheckerConnectionRefused(t *testing.T) {
	cfg := &config.HealthCheckConfig{
		Name:       "test-http",
		Type:       config.HealthCheckHTTP,
		Endpoint:   "http://localhost:59999", // Unlikely to be in use
		StatusCode: 200,
		Timeout:    config.Duration(1 * time.Second),
	}

	checker := NewHTTPChecker(cfg)
	result := checker.Check(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.NotNil(t, result.Error)
}

func TestTCPChecker(t *testing.T) {
	// Create test listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	cfg := &config.HealthCheckConfig{
		Name:    "test-tcp",
		Type:    config.HealthCheckTCP,
		Host:    "127.0.0.1",
		Port:    addr.Port,
		Timeout: config.Duration(5 * time.Second),
	}

	checker := NewTCPChecker(cfg)

	assert.Equal(t, "test-tcp", checker.Name())
	assert.Equal(t, "tcp", checker.Type())

	result := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, result.Status)
}

func TestTCPCheckerUnhealthy(t *testing.T) {
	cfg := &config.HealthCheckConfig{
		Name:    "test-tcp",
		Type:    config.HealthCheckTCP,
		Host:    "127.0.0.1",
		Port:    59998, // Unlikely to be in use
		Timeout: config.Duration(1 * time.Second),
	}

	checker := NewTCPChecker(cfg)
	result := checker.Check(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.NotNil(t, result.Error)
}

func TestCommandChecker(t *testing.T) {
	cfg := &config.HealthCheckConfig{
		Name:    "test-cmd",
		Type:    config.HealthCheckCommand,
		Command: "echo healthy",
		Timeout: config.Duration(5 * time.Second),
	}

	checker := NewCommandChecker(cfg)

	assert.Equal(t, "test-cmd", checker.Name())
	assert.Equal(t, "command", checker.Type())

	result := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "healthy", result.Message)
}

func TestCommandCheckerUnhealthy(t *testing.T) {
	cfg := &config.HealthCheckConfig{
		Name:    "test-cmd",
		Type:    config.HealthCheckCommand,
		Command: "false", // Always returns exit code 1
		Timeout: config.Duration(5 * time.Second),
	}

	checker := NewCommandChecker(cfg)
	result := checker.Check(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.NotNil(t, result.Error)
}

func TestNewChecker(t *testing.T) {
	tests := []struct {
		name        string
		checkType   config.HealthCheckType
		expectType  string
		expectError bool
	}{
		{
			name:       "HTTP checker",
			checkType:  config.HealthCheckHTTP,
			expectType: "http",
		},
		{
			name:       "TCP checker",
			checkType:  config.HealthCheckTCP,
			expectType: "tcp",
		},
		{
			name:       "Command checker",
			checkType:  config.HealthCheckCommand,
			expectType: "command",
		},
		{
			name:        "Invalid type",
			checkType:   "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.HealthCheckConfig{
				Type:     tt.checkType,
				Endpoint: "http://localhost",
				Host:     "localhost",
				Port:     80,
				Command:  "echo test",
				Timeout:  config.Duration(5 * time.Second),
			}

			checker, err := NewChecker(cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, checker)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checker)
				assert.Equal(t, tt.expectType, checker.Type())
			}
		})
	}
}

func TestMonitor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configs := []config.HealthCheckConfig{
		{
			Name:       "http-check",
			Type:       config.HealthCheckHTTP,
			Endpoint:   server.URL,
			StatusCode: 200,
			Interval:   config.Duration(100 * time.Millisecond),
			Timeout:    config.Duration(1 * time.Second),
		},
	}

	events := make(chan Event, 10)
	monitor, err := NewMonitor(configs, events)
	require.NoError(t, err)

	ctx := context.Background()
	monitor.Start(ctx)

	// Wait for at least one check
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, StatusHealthy, monitor.Status())
	assert.True(t, monitor.IsHealthy())

	results := monitor.Results()
	assert.Len(t, results, 1)

	monitor.Stop()
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusUnknown, "unknown"},
		{StatusHealthy, "healthy"},
		{StatusUnhealthy, "unhealthy"},
		{StatusDegraded, "degraded"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.status.String())
	}
}
