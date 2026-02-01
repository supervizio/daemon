// Package discovery provides internal tests for the discovery factory.
package discovery

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
)

// TestFactory_addSystemdDiscoverer tests addSystemdDiscoverer method.
func TestFactory_addSystemdDiscoverer(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DiscoveryConfig
		wantLen  int
		inputLen int
	}{
		{
			name:     "nil systemd config",
			config:   &config.DiscoveryConfig{},
			wantLen:  0,
			inputLen: 0,
		},
		{
			name: "disabled systemd",
			config: &config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: false,
				},
			},
			wantLen:  0,
			inputLen: 0,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			input := make([]target.Discoverer, tc.inputLen)
			result := f.addSystemdDiscoverer(input)
			// Verify result length.
			if len(result) != tc.wantLen {
				t.Errorf("addSystemdDiscoverer() len = %d, want %d", len(result), tc.wantLen)
			}
		})
	}
}

// TestFactory_addDockerDiscoverer tests addDockerDiscoverer method.
func TestFactory_addDockerDiscoverer(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DiscoveryConfig
		wantLen  int
		inputLen int
	}{
		{
			name:     "nil docker config",
			config:   &config.DiscoveryConfig{},
			wantLen:  0,
			inputLen: 0,
		},
		{
			name: "disabled docker",
			config: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: false,
				},
			},
			wantLen:  0,
			inputLen: 0,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			input := make([]target.Discoverer, tc.inputLen)
			result := f.addDockerDiscoverer(input)
			// Verify result length.
			if len(result) != tc.wantLen {
				t.Errorf("addDockerDiscoverer() len = %d, want %d", len(result), tc.wantLen)
			}
		})
	}
}

// TestFactory_addKubernetesDiscoverer tests addKubernetesDiscoverer method.
func TestFactory_addKubernetesDiscoverer(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DiscoveryConfig
		wantLen  int
		inputLen int
	}{
		{
			name:     "nil kubernetes config",
			config:   &config.DiscoveryConfig{},
			wantLen:  0,
			inputLen: 0,
		},
		{
			name: "disabled kubernetes",
			config: &config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{
					Enabled: false,
				},
			},
			wantLen:  0,
			inputLen: 0,
		},
		{
			name: "enabled kubernetes",
			config: &config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{
					Enabled: true,
				},
			},
			wantLen:  0, // Still 0 because createKubernetesDiscoverer returns nil
			inputLen: 0,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			input := make([]target.Discoverer, tc.inputLen)
			result := f.addKubernetesDiscoverer(input)
			// Verify result length.
			if len(result) != tc.wantLen {
				t.Errorf("addKubernetesDiscoverer() len = %d, want %d", len(result), tc.wantLen)
			}
		})
	}
}

// TestFactory_addNomadDiscoverer tests addNomadDiscoverer method.
func TestFactory_addNomadDiscoverer(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DiscoveryConfig
		wantLen  int
		inputLen int
	}{
		{
			name:     "nil nomad config",
			config:   &config.DiscoveryConfig{},
			wantLen:  0,
			inputLen: 0,
		},
		{
			name: "disabled nomad",
			config: &config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{
					Enabled: false,
				},
			},
			wantLen:  0,
			inputLen: 0,
		},
		{
			name: "enabled nomad",
			config: &config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{
					Enabled: true,
				},
			},
			wantLen:  0, // Still 0 because createNomadDiscoverer returns nil
			inputLen: 0,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			input := make([]target.Discoverer, tc.inputLen)
			result := f.addNomadDiscoverer(input)
			// Verify result length.
			if len(result) != tc.wantLen {
				t.Errorf("addNomadDiscoverer() len = %d, want %d", len(result), tc.wantLen)
			}
		})
	}
}
