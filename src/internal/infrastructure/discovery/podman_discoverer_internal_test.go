//go:build unix

package discovery

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

// TestPodmanDiscoverer_matchesLabels verifies label filtering logic.
func TestPodmanDiscoverer_matchesLabels(t *testing.T) {
	tests := []struct {
		name        string
		filter      map[string]string
		container   dockerContainer
		shouldMatch bool
	}{
		{
			name:   "no filter accepts all",
			filter: nil,
			container: dockerContainer{
				Labels: map[string]string{"app": "web"},
			},
			shouldMatch: true,
		},
		{
			name:   "empty filter accepts all",
			filter: map[string]string{},
			container: dockerContainer{
				Labels: map[string]string{"app": "web"},
			},
			shouldMatch: true,
		},
		{
			name:   "matching label",
			filter: map[string]string{"app": "web"},
			container: dockerContainer{
				Labels: map[string]string{"app": "web"},
			},
			shouldMatch: true,
		},
		{
			name:   "non-matching value",
			filter: map[string]string{"app": "web"},
			container: dockerContainer{
				Labels: map[string]string{"app": "db"},
			},
			shouldMatch: false,
		},
		{
			name:   "missing label",
			filter: map[string]string{"app": "web"},
			container: dockerContainer{
				Labels: map[string]string{},
			},
			shouldMatch: false,
		},
		{
			name:   "multiple labels all match",
			filter: map[string]string{"app": "web", "env": "prod"},
			container: dockerContainer{
				Labels: map[string]string{"app": "web", "env": "prod"},
			},
			shouldMatch: true,
		},
		{
			name:   "multiple labels partial match",
			filter: map[string]string{"app": "web", "env": "prod"},
			container: dockerContainer{
				Labels: map[string]string{"app": "web", "env": "dev"},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PodmanDiscoverer{
				labelFilter: tt.filter,
			}

			result := d.matchesLabels(tt.container)

			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestPodmanDiscoverer_containerToTarget verifies container conversion.
func TestPodmanDiscoverer_containerToTarget(t *testing.T) {
	tests := []struct {
		name      string
		container dockerContainer
		wantID    string
		wantName  string
		wantType  target.Type
	}{
		{
			name: "container with name",
			container: dockerContainer{
				ID:     "abcdef123456789",
				Names:  []string{"/my-container"},
				State:  "running",
				Status: "Up 2 hours",
				Labels: map[string]string{"app": "web"},
				Ports:  []dockerPort{{Type: "tcp", PublicPort: 8080, PrivatePort: 80}},
			},
			wantID:   "podman:abcdef123456",
			wantName: "my-container",
			wantType: target.TypePodman,
		},
		{
			name: "container without name uses ID",
			container: dockerContainer{
				ID:     "abcdef123456789",
				Names:  []string{},
				State:  "running",
				Status: "Up 2 hours",
			},
			wantID:   "podman:abcdef123456",
			wantName: "abcdef123456",
			wantType: target.TypePodman,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PodmanDiscoverer{}

			target := d.containerToTarget(tt.container)

			assert.Equal(t, tt.wantID, target.ID)
			assert.Equal(t, tt.wantName, target.Name)
			assert.Equal(t, tt.wantType, target.Type)
		})
	}
}

// TestPodmanDiscoverer_configureProbe verifies probe configuration.
func TestPodmanDiscoverer_configureProbe(t *testing.T) {
	tests := []struct {
		name      string
		container dockerContainer
		wantProbe bool
	}{
		{
			name: "uses public port",
			container: dockerContainer{
				Ports: []dockerPort{{Type: "tcp", PublicPort: 8080, PrivatePort: 80}},
			},
			wantProbe: true,
		},
		{
			name: "falls back to private port",
			container: dockerContainer{
				Ports: []dockerPort{{Type: "tcp", PrivatePort: 80}},
			},
			wantProbe: true,
		},
		{
			name:      "no ports",
			container: dockerContainer{},
			wantProbe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgt := target.ExternalTarget{}

			configureContainerProbe(&tgt, tt.container, podmanProbeTypeTCP)

			// Check if probe was configured based on ProbeType.
			if tt.wantProbe {
				assert.NotEmpty(t, tgt.ProbeType)
			} else {
				assert.Empty(t, tgt.ProbeType)
			}
		})
	}
}
