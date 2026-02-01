//go:build unix

package discovery_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
)

// TestPodmanDiscoverer_Type verifies the discoverer returns correct type.
func TestPodmanDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name string
		want target.Type
	}{
		{
			name: "returns TypePodman",
			want: target.TypePodman,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewPodmanDiscoverer("/run/podman/podman.sock", nil)

			got := d.Type()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestPodmanDiscoverer_Discover verifies error handling for invalid socket.
func TestPodmanDiscoverer_Discover(t *testing.T) {
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
			d := discovery.NewPodmanDiscoverer(tt.socketPath, nil)

			_, err := d.Discover(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewPodmanDiscoverer verifies constructor behavior.
func TestNewPodmanDiscoverer(t *testing.T) {
	tests := []struct {
		name   string
		socket string
		labels map[string]string
	}{
		{
			name:   "creates with nil labels",
			socket: "/run/podman/podman.sock",
			labels: nil,
		},
		{
			name:   "creates with label filter",
			socket: "/run/podman/podman.sock",
			labels: map[string]string{"app": "web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewPodmanDiscoverer(tt.socket, tt.labels)

			assert.NotNil(t, d)
		})
	}
}
