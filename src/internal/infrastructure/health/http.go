// Package health provides infrastructure adapters for health checking.
// It implements the health check interfaces for HTTP, TCP, and command-based checks.
package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
)

// checkerTypeHTTP is the type identifier for HTTP health checkers.
// This constant is used to distinguish HTTP checkers from other types.
const checkerTypeHTTP string = "http"

// defaultHTTPStatusCode is the default expected status code for HTTP health checks.
// When no status code is specified, HTTP 200 OK is expected.
const defaultHTTPStatusCode int = http.StatusOK

// ErrHTTPStatusMismatch indicates the HTTP response status code did not match expected.
// This error is returned when the actual status code differs from the configured expected code.
var ErrHTTPStatusMismatch error = errors.New("status code mismatch")

// HTTPChecker performs HTTP health checks against a specified endpoint.
// It supports configurable HTTP methods, expected status codes, and timeouts
// for flexible health check implementations.
type HTTPChecker struct {
	name       string
	endpoint   string
	method     string
	statusCode int
	client     *http.Client
}

// NewHTTPChecker creates a new HTTP health checker.
// It initializes the checker with the provided configuration, applying defaults
// for any unspecified values.
//
// Params:
//   - cfg: The health check configuration containing endpoint, method, and timeout settings.
//
// Returns:
//   - *HTTPChecker: A configured HTTP health checker ready for use.
func NewHTTPChecker(cfg *service.HealthCheckConfig) *HTTPChecker {
	name := cfg.Name
	// Use endpoint as name if not specified.
	if name == "" {
		name = fmt.Sprintf("http-%s", cfg.Endpoint)
	}

	method := cfg.Method
	// Default to GET method if not specified.
	if method == "" {
		method = http.MethodGet
	}

	statusCode := cfg.StatusCode
	// Default to 200 OK if status code not specified.
	if statusCode == 0 {
		statusCode = defaultHTTPStatusCode
	}

	timeout := cfg.Timeout.Duration()
	// Configure transport with response header timeout.
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
	}

	// Return the configured HTTP checker.
	// Note: CheckRedirect is not needed because we use transport.RoundTrip()
	// directly, which does not follow redirects automatically.
	return &HTTPChecker{
		name:       name,
		endpoint:   cfg.Endpoint,
		method:     method,
		statusCode: statusCode,
		client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// Name returns the checker name.
// This identifier is used for logging and status reporting.
//
// Returns:
//   - string: The name of this health checker.
func (c *HTTPChecker) Name() string {
	// Return the configured checker name.
	return c.name
}

// Type returns the checker type.
// This is used to identify the kind of health check being performed.
//
// Returns:
//   - string: The type identifier "http" for HTTP checkers.
func (c *HTTPChecker) Type() string {
	// Return the HTTP checker type constant.
	return checkerTypeHTTP
}

// Check performs an HTTP health check.
// It sends an HTTP request to the configured endpoint and validates the response
// status code against the expected value.
//
// Params:
//   - ctx: The context for request cancellation and timeout.
//
// Returns:
//   - domain.Result: The health check result indicating healthy or unhealthy status.
func (c *HTTPChecker) Check(ctx context.Context) domain.Result {
	start := time.Now()

	statusCode, err := c.getStatusCode(ctx)
	// Handle request errors.
	if err != nil {
		// Return unhealthy result with error details.
		return domain.NewUnhealthyResult(
			fmt.Sprintf("request failed: %v", err),
			time.Since(start),
			err,
		)
	}

	// Check if status code matches expected.
	if statusCode != c.statusCode {
		// Return unhealthy result for status mismatch.
		return domain.NewUnhealthyResult(
			fmt.Sprintf("unexpected status code: %d (expected %d)", statusCode, c.statusCode),
			time.Since(start),
			ErrHTTPStatusMismatch,
		)
	}

	// Return healthy result with status code.
	return domain.NewHealthyResult(
		fmt.Sprintf("HTTP %d", statusCode),
		time.Since(start),
	)
}

// getStatusCode performs the HTTP request and returns the status code.
// It handles request creation and transport-level operations.
//
// Params:
//   - ctx: The context for request cancellation and timeout.
//
// Returns:
//   - int: The HTTP status code from the response.
//   - error: Any error that occurred during the request.
func (c *HTTPChecker) getStatusCode(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, c.method, c.endpoint, http.NoBody)
	// Handle request creation errors.
	if err != nil {
		// Return wrapped error with context.
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	transport := c.client.Transport
	// Use default transport if none configured.
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(req)
	// Handle transport errors.
	if err != nil {
		// Return the transport error.
		return 0, err
	}
	// Ensure response body is closed.
	defer func() {
		// Ignore close error for response body.
		_ = resp.Body.Close()
	}()

	// Return the response status code.
	return resp.StatusCode, nil
}
