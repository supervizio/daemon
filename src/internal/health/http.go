package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// HTTPChecker performs HTTP health checks.
type HTTPChecker struct {
	name       string
	endpoint   string
	method     string
	statusCode int
	client     *http.Client
}

// NewHTTPChecker creates a new HTTP health checker.
func NewHTTPChecker(cfg *config.HealthCheckConfig) *HTTPChecker {
	name := cfg.Name
	if name == "" {
		name = fmt.Sprintf("http-%s", cfg.Endpoint)
	}

	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}

	statusCode := cfg.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return &HTTPChecker{
		name:       name,
		endpoint:   cfg.Endpoint,
		method:     method,
		statusCode: statusCode,
		client: &http.Client{
			Timeout: cfg.Timeout.Duration(),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects for health checks
				return http.ErrUseLastResponse
			},
		},
	}
}

// Name returns the checker name.
func (c *HTTPChecker) Name() string {
	return c.name
}

// Type returns the checker type.
func (c *HTTPChecker) Type() string {
	return "http"
}

// Check performs an HTTP health check.
func (c *HTTPChecker) Check(ctx context.Context) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, c.method, c.endpoint, nil)
	if err != nil {
		return Result{
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("failed to create request: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     err,
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("request failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != c.statusCode {
		return Result{
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("unexpected status code: %d (expected %d)", resp.StatusCode, c.statusCode),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     fmt.Errorf("status code mismatch"),
		}
	}

	return Result{
		Status:    StatusHealthy,
		Message:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}
