// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// proberTypeHTTP is the type identifier for HTTP probers.
const proberTypeHTTP string = "http"

// defaultHTTPMethod is the default HTTP method for probes.
const defaultHTTPMethod string = http.MethodGet

// defaultHTTPStatusCode is the default expected status code.
const defaultHTTPStatusCode int = http.StatusOK

// ErrHTTPStatusMismatch indicates the status code didn't match.
var ErrHTTPStatusMismatch error = errors.New("status code mismatch")

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
	// use default timeout if not specified
	if timeout <= 0 {
		timeout = health.DefaultTimeout
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
	}

	// configure client with timeout and transport
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
	// identify this prober as http type
	return proberTypeHTTP
}

// Probe performs an HTTP endpoint healthcheck.
// It sends an HTTP request and validates the response status code.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to healthcheck.
//
// Returns:
//   - health.CheckResult: the probe result with latency and status code.
func (p *HTTPProber) Probe(ctx context.Context, target health.Target) health.CheckResult {
	start := time.Now()

	method := target.Method
	// use default HTTP method if not specified
	if method == "" {
		method = defaultHTTPMethod
	}

	expectedStatus := target.StatusCode
	// use default status code if not specified
	if expectedStatus == 0 {
		expectedStatus = defaultHTTPStatusCode
	}

	statusCode, err := p.getStatusCode(ctx, method, target.Address, target.Path)
	latency := time.Since(start)

	// handle request failure
	if err != nil {
		// request errors indicate network or server issues
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("request failed: %v", err),
			err,
		)
	}

	// validate status code matches expectation
	if statusCode != expectedStatus {
		// status mismatch indicates service is responding incorrectly
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("unexpected status code: %d (expected %d)", statusCode, expectedStatus),
			ErrHTTPStatusMismatch,
		)
	}

	// return success with status code
	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("HTTP %d", statusCode),
	)
}

// getStatusCode performs the HTTP request and returns the status code.
//
// Params:
//   - ctx: context for cancellation.
//   - method: the HTTP method to use.
//   - address: the base URL to request.
//   - path: optional path to append to the URL.
//
// Returns:
//   - int: the HTTP status code from the response.
//   - error: any error that occurred during the request.
func (p *HTTPProber) getStatusCode(ctx context.Context, method, address, path string) (int, error) {
	targetURL, err := url.Parse(address)
	// handle malformed URL
	if err != nil {
		// malformed urls indicate configuration error
		return 0, fmt.Errorf("failed to parse url: %w", err)
	}

	// Go's url.Parse treats "host:port" as "scheme:opaque", so prepend http:// if needed.
	// handle missing scheme
	if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
		targetURL, err = url.Parse("http://" + address)
		// handle malformed URL after adding scheme
		if err != nil {
			// second parse failure means address is truly invalid
			return 0, fmt.Errorf("failed to parse url: %w", err)
		}
	}

	// append path to URL if provided
	if path != "" {
		targetURL.Path = strings.TrimRight(targetURL.Path, "/") + "/" + strings.TrimLeft(path, "/")
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL.String(), http.NoBody)
	// handle request creation failure
	if err != nil {
		// request creation errors indicate programming error
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	// handle request execution failure
	if err != nil {
		// network errors or timeouts during request execution
		return 0, err
	}
	// ensure response body is closed
	defer func() { _ = resp.Body.Close() }()

	// return status code from successful response
	return resp.StatusCode, nil
}
