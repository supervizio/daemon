//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Doer abstracts HTTP client operations for testing and flexibility.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// errUnexpectedContainerStatus is returned when container runtime API returns non-OK status.
var errUnexpectedContainerStatus error = errors.New("unexpected status code")

// fetchContainers fetches containers from a container runtime API.
// It handles HTTP request creation, execution, and JSON decoding for Docker-compatible APIs.
//
// Params:
//   - ctx: context for cancellation.
//   - client: HTTP client configured for the runtime socket.
//   - apiURL: the API URL (e.g., "http://docker/containers/json").
//   - runtimeName: name of the runtime for error messages (e.g., "docker", "podman").
//
// Returns:
//   - []dockerContainer: the fetched containers.
//   - error: any error during fetch.
func fetchContainers(
	ctx context.Context,
	client Doer,
	apiURL, runtimeName string,
) ([]dockerContainer, error) {
	// Build API request for container list.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	// Check for request creation error.
	if err != nil {
		// Return error with request context.
		return nil, fmt.Errorf("create %s request: %w", runtimeName, err)
	}

	// Execute request against runtime API.
	resp, err := client.Do(req)
	// Check for API communication error.
	if err != nil {
		// Return error with API context.
		return nil, fmt.Errorf("%s api request: %w", runtimeName, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify successful response from runtime API.
	if resp.StatusCode != http.StatusOK {
		// Return error for non-OK status.
		return nil, fmt.Errorf("%s api: %w (status %d)", runtimeName, errUnexpectedContainerStatus, resp.StatusCode)
	}

	// Parse JSON response into container structs.
	var containers []dockerContainer
	// Check for JSON decoding error.
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		// Return error with decode context.
		return nil, fmt.Errorf("decode %s response: %w", runtimeName, err)
	}

	// Return fetched containers.
	return containers, nil
}
