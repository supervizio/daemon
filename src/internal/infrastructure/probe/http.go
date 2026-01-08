// Package probe provides infrastructure adapters for service probing.
package probe

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// proberTypeHTTP is the type identifier for HTTP probers.
const proberTypeHTTP string = "http"

// defaultHTTPMethod is the default HTTP method for probes.
const defaultHTTPMethod string = http.MethodGet

// defaultHTTPStatusCode is the default expected status code.
const defaultHTTPStatusCode int = http.StatusOK

// ErrHTTPStatusMismatch indicates the status code didn't match.
var ErrHTTPStatusMismatch = errors.New("status code mismatch")

// HTTPProber performs HTTP endpoint probes.
// It verifies service health by making HTTP requests.
type HTTPProber struct {
	// client is the HTTP client used for requests.
	client *http.Client
	// timeout is the maximum duration for requests.
	timeout time.Duration
}

// NewHTTPProber creates a new HTTP prober.
//
// Params:
//   - timeout: the maximum duration for HTTP requests.
//
// Returns:
//   - *HTTPProber: a configured HTTP prober ready to perform probes.
func NewHTTPProber(timeout time.Duration) *HTTPProber {
	// Configure transport with timeout.
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
	}

	// Return configured HTTP prober.
	return &HTTPProber{
		timeout: timeout,
		client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "http" identifying the prober type.
func (p *HTTPProber) Type() string {
	// Return the HTTP prober type identifier.
	return proberTypeHTTP
}

// Probe performs an HTTP endpoint probe.
// It sends an HTTP request and validates the response status code.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to probe.
//
// Returns:
//   - probe.Result: the probe result with latency and status code.
func (p *HTTPProber) Probe(ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Determine HTTP method.
	method := target.Method
	if method == "" {
		// Default to GET if not specified.
		method = defaultHTTPMethod
	}

	// Determine expected status code.
	expectedStatus := target.StatusCode
	if expectedStatus == 0 {
		// Default to 200 OK if not specified.
		expectedStatus = defaultHTTPStatusCode
	}

	// Get the status code.
	statusCode, err := p.getStatusCode(ctx, method, target.Address)
	latency := time.Since(start)

	// Handle request errors.
	if err != nil {
		// Return failure result with error details.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("request failed: %v", err),
			err,
		)
	}

	// Check if status code matches expected.
	if statusCode != expectedStatus {
		// Return failure result for status mismatch.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("unexpected status code: %d (expected %d)", statusCode, expectedStatus),
			ErrHTTPStatusMismatch,
		)
	}

	// Return success result with status code.
	return probe.NewSuccessResult(
		latency,
		fmt.Sprintf("HTTP %d", statusCode),
	)
}

// getStatusCode performs the HTTP request and returns the status code.
//
// Params:
//   - ctx: context for cancellation.
//   - method: the HTTP method to use.
//   - url: the URL to request.
//
// Returns:
//   - int: the HTTP status code from the response.
//   - error: any error that occurred during the request.
func (p *HTTPProber) getStatusCode(ctx context.Context, method, url string) (int, error) {
	// Create the request.
	req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
	if err != nil {
		// Return wrapped error.
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Get the transport.
	transport := p.client.Transport
	if transport == nil {
		// Use default transport if none configured.
		transport = http.DefaultTransport
	}

	// Execute the request.
	resp, err := transport.RoundTrip(req)
	if err != nil {
		// Return the transport error.
		return 0, err
	}
	// Ensure response body is closed.
	defer func() {
		// Ignore close error.
		_ = resp.Body.Close()
	}()

	// Return the status code.
	return resp.StatusCode, nil
}
