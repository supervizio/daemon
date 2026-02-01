//go:build unix

package discovery_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
)

// TestDockerDiscoverer_Type verifies the discoverer returns correct type.
func TestDockerDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name string
		want target.Type
	}{
		{
			name: "returns TypeDocker",
			want: target.TypeDocker,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewDockerDiscoverer("/var/run/docker.sock", nil)

			got := d.Type()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDockerDiscoverer_Discover verifies error handling for invalid socket.
func TestDockerDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name       string
		socketPath string
		wantErr    bool
	}{
		{
			name:       "invalid socket returns error",
			socketPath: "/nonexistent/socket",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewDockerDiscoverer(tt.socketPath, nil)

			_, err := d.Discover(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewDockerDiscoverer verifies constructor behavior.
func TestNewDockerDiscoverer(t *testing.T) {
	tests := []struct {
		name   string
		socket string
		labels map[string]string
	}{
		{
			name:   "creates with nil labels",
			socket: "/var/run/docker.sock",
			labels: nil,
		},
		{
			name:   "creates with label filter",
			socket: "/var/run/docker.sock",
			labels: map[string]string{"app": "web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewDockerDiscoverer(tt.socket, tt.labels)

			assert.NotNil(t, d)
		})
	}
}
