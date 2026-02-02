//go:build unix

// Package discovery provides internal tests for container helper functions.
package discovery

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

// TestContainerToExternalTarget tests the containerToExternalTarget function.
func TestContainerToExternalTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   containerToTargetParams
		wantID   string
		wantType target.Type
	}{
		{
			name: "docker container with name",
			params: containerToTargetParams{
				Container: dockerContainer{
					ID:     "abc123def456",
					Names:  []string{"/nginx"},
					State:  "running",
					Status: "Up 5 minutes",
				},
				RuntimePrefix:  "docker",
				TargetType:     target.TypeDocker,
				MetadataLabels: 2,
				ProbeType:      "tcp",
			},
			wantID:   "docker:abc123def456",
			wantType: target.TypeDocker,
		},
		{
			name: "podman container without name",
			params: containerToTargetParams{
				Container: dockerContainer{
					ID:     "xyz789abc123",
					Names:  []string{},
					State:  "running",
					Status: "Up 10 minutes",
				},
				RuntimePrefix:  "podman",
				TargetType:     target.TypePodman,
				MetadataLabels: 2,
				ProbeType:      "tcp",
			},
			wantID:   "podman:xyz789abc123",
			wantType: target.TypePodman,
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := containerToExternalTarget(tc.params)
			assert.Equal(t, tc.wantID, result.ID)
			assert.Equal(t, tc.wantType, result.Type)
			assert.Equal(t, target.SourceDiscovered, result.Source)
		})
	}
}

// TestConfigureContainerProbe tests the configureContainerProbe function.
func TestConfigureContainerProbe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		container      dockerContainer
		probeType      string
		expectProbe    bool
		expectedTarget string
	}{
		{
			name: "container with public port",
			container: dockerContainer{
				Ports: []dockerPort{
					{PrivatePort: 80, PublicPort: 8080, Type: "tcp"},
				},
			},
			probeType:      "tcp",
			expectProbe:    true,
			expectedTarget: "127.0.0.1:8080",
		},
		{
			name: "container with only private port",
			container: dockerContainer{
				Ports: []dockerPort{
					{PrivatePort: 80, PublicPort: 0, Type: "tcp"},
				},
			},
			probeType:      "tcp",
			expectProbe:    true,
			expectedTarget: "127.0.0.1:80",
		},
		{
			name: "container without ports",
			container: dockerContainer{
				Ports: []dockerPort{},
			},
			probeType:   "tcp",
			expectProbe: false,
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tgt := &target.ExternalTarget{}
			configureContainerProbe(tgt, tc.container, tc.probeType)

			if tc.expectProbe {
				assert.Equal(t, tc.probeType, tgt.ProbeType)
				assert.NotNil(t, tgt.ProbeTarget)
			} else {
				assert.Empty(t, tgt.ProbeType)
			}
		})
	}
}

// mockDoer implements Doer for testing.
type mockDoer struct {
	response *http.Response
	err      error
}

// Do implements Doer.
func (m *mockDoer) Do(_ *http.Request) (*http.Response, error) {
	// Return configured response and error.
	return m.response, m.err
}

// TestFetchContainers tests the fetchContainers function.
func TestFetchContainers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		response    *http.Response
		err         error
		wantErr     bool
		wantLen     int
		runtimeName string
	}{
		{
			name: "successful fetch",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`[{"Id":"abc123","Names":["/test"]}]`)),
			},
			err:         nil,
			wantErr:     false,
			wantLen:     1,
			runtimeName: "docker",
		},
		{
			name: "empty response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`[]`)),
			},
			err:         nil,
			wantErr:     false,
			wantLen:     0,
			runtimeName: "docker",
		},
		{
			name:        "request error",
			response:    nil,
			err:         errors.New("connection refused"),
			wantErr:     true,
			runtimeName: "podman",
		},
		{
			name: "non-OK status",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			},
			err:         nil,
			wantErr:     true,
			runtimeName: "docker",
		},
		{
			name: "invalid JSON",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`invalid`)),
			},
			err:         nil,
			wantErr:     true,
			runtimeName: "docker",
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := &mockDoer{response: tc.response, err: tc.err}
			result, err := fetchContainers(context.Background(), client, "http://test/containers/json", tc.runtimeName)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tc.wantLen)
			}
		})
	}
}
