//go:build unix

package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testContainerFieldsJSON is a JSON response for testing container field parsing.
const testContainerFieldsJSON string = `[{
	"ID": "abcdef123456789",
	"Names": ["/my-container", "/alias"],
	"State": "running",
	"Status": "Up 2 hours",
	"Labels": {"app": "web", "env": "prod"},
	"Ports": [{"Type": "tcp", "PublicPort": 8080, "PrivatePort": 80}]
}]`

// mockHTTPDoer is a mock HTTP client for testing fetchContainers.
type mockHTTPDoer struct {
	response *http.Response
	err      error
}

// Do implements the Doer interface.
func (m *mockHTTPDoer) Do(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// TestFetchContainers verifies container fetching from a container runtime API.
func TestFetchContainers(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		clientErr   error
		wantErr     bool
		wantCount   int
		errContains string
	}{
		{
			name: "successful fetch with containers",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`[
					{"ID": "abc123", "Names": ["/test-container"], "State": "running"}
				]`)),
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "successful fetch with empty list",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`[]`)),
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:        "client error",
			clientErr:   errors.New("connection refused"),
			wantErr:     true,
			errContains: "api request",
		},
		{
			name: "non-OK status code",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "internal error"}`)),
			},
			wantErr:     true,
			errContains: "unexpected status code",
		},
		{
			name: "invalid JSON response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{invalid json}`)),
			},
			wantErr:     true,
			errContains: "decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockHTTPDoer{
				response: tt.response,
				err:      tt.clientErr,
			}

			containers, err := fetchContainers(
				context.Background(),
				client,
				"http://docker/containers/json",
				"docker",
			)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, containers, tt.wantCount)
			}
		})
	}
}

// TestFetchContainers_ContextCancellation verifies context cancellation handling.
func TestFetchContainers_ContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "canceled context with client error fails",
			ctx:     canceledCtx(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock client that returns context error.
			client := &mockHTTPDoer{
				err: context.Canceled,
			}

			_, err := fetchContainers(tt.ctx, client, "http://docker/containers/json", "docker")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// canceledCtx returns a pre-canceled context.
func canceledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TestFetchContainers_ParsesContainerFields verifies container fields are parsed correctly.
func TestFetchContainers_ParsesContainerFields(t *testing.T) {
	tests := []struct {
		name          string
		jsonResponse  string
		wantID        string
		wantNames     []string
		wantState     string
		wantLabelsLen int
	}{
		{
			name:          "parses all container fields",
			jsonResponse:  testContainerFieldsJSON,
			wantID:        "abcdef123456789",
			wantNames:     []string{"/my-container", "/alias"},
			wantState:     "running",
			wantLabelsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockHTTPDoer{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(tt.jsonResponse)),
				},
			}

			containers, err := fetchContainers(
				context.Background(),
				client,
				"http://docker/containers/json",
				"docker",
			)

			require.NoError(t, err)
			require.Len(t, containers, 1)

			c := containers[0]
			assert.Equal(t, tt.wantID, c.ID)
			assert.Equal(t, tt.wantNames, c.Names)
			assert.Equal(t, tt.wantState, c.State)
			assert.Len(t, c.Labels, tt.wantLabelsLen)
		})
	}
}

// TestDockerContainer_JSONDecoding verifies JSON unmarshaling of dockerContainer.
func TestDockerContainer_JSONDecoding(t *testing.T) {
	tests := []struct {
		name       string
		jsonInput  string
		wantID     string
		wantState  string
		wantPorts  int
		wantLabels int
	}{
		{
			name:       "empty container",
			jsonInput:  `{"ID": "test"}`,
			wantID:     "test",
			wantState:  "",
			wantPorts:  0,
			wantLabels: 0,
		},
		{
			name:       "container with ports",
			jsonInput:  `{"ID": "abc", "Ports": [{"Type": "tcp", "PublicPort": 80}]}`,
			wantID:     "abc",
			wantPorts:  1,
			wantLabels: 0,
		},
		{
			name:       "container with labels",
			jsonInput:  `{"ID": "xyz", "Labels": {"key": "value"}}`,
			wantID:     "xyz",
			wantPorts:  0,
			wantLabels: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var container dockerContainer
			err := json.Unmarshal([]byte(tt.jsonInput), &container)

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, container.ID)
			assert.Equal(t, tt.wantState, container.State)
			assert.Len(t, container.Ports, tt.wantPorts)
			assert.Len(t, container.Labels, tt.wantLabels)
		})
	}
}
