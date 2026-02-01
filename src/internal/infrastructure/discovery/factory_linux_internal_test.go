//go:build linux

// Package discovery provides internal tests for Linux factory methods.
package discovery

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
)

// TestFactoryCreateSystemdDiscoverer tests createSystemdDiscoverer method.
func TestFactory_createSystemdDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil systemd config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled systemd with patterns",
			config: &config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled:  true,
					Patterns: []string{"nginx.service"},
				},
			},
			wantNil: false,
		},
		{
			name: "enabled systemd without patterns",
			config: &config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createSystemdDiscoverer()
			// Verify discoverer matches expectation.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createSystemdDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactoryCreateDockerDiscoverer tests createDockerDiscoverer method.
func TestFactory_createDockerDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil docker config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled docker with socket",
			config: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled:    true,
					SocketPath: "/var/run/docker.sock",
				},
			},
			wantNil: false,
		},
		{
			name: "enabled docker with default socket",
			config: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createDockerDiscoverer()
			// Verify discoverer matches expectation.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createDockerDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactoryCreateKubernetesDiscoverer tests createKubernetesDiscoverer method.
func TestFactory_createKubernetesDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil kubernetes config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled kubernetes",
			config: &config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: true,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createKubernetesDiscoverer()
			// Verify discoverer is nil (not implemented).
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createKubernetesDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactoryCreateNomadDiscoverer tests createNomadDiscoverer method.
func TestFactory_createNomadDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil nomad config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled nomad",
			config: &config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: true,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createNomadDiscoverer()
			// Verify discoverer is nil (not implemented).
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createNomadDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}
