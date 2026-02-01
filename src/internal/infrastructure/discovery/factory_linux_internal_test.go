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
			wantNil: false, // Nomad discoverer is implemented on Linux
		},
		{
			name: "enabled nomad with address",
			config: &config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{
					Enabled: true,
					Address: "http://nomad.example.com:4646",
				},
			},
			wantNil: false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createNomadDiscoverer()
			// Verify discoverer matches expectation.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createNomadDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactory_createOpenRCDiscoverer tests createOpenRCDiscoverer method.
func TestFactory_createOpenRCDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil openrc config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled openrc",
			config: &config.DiscoveryConfig{
				OpenRC: &config.OpenRCDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: true, // Returns nil on Linux (not implemented)
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createOpenRCDiscoverer()
			// Verify discoverer is nil on Linux.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createOpenRCDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactory_createBSDRCDiscoverer tests createBSDRCDiscoverer method.
func TestFactory_createBSDRCDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil bsdrc config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled bsdrc",
			config: &config.DiscoveryConfig{
				BSDRC: &config.BSDRCDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: true, // Returns nil on Linux (not available)
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createBSDRCDiscoverer()
			// Verify discoverer is nil on Linux.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createBSDRCDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactory_createPodmanDiscoverer tests createPodmanDiscoverer method.
func TestFactory_createPodmanDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil podman config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled podman with socket",
			config: &config.DiscoveryConfig{
				Podman: &config.PodmanDiscoveryConfig{
					Enabled:    true,
					SocketPath: "/run/podman/podman.sock",
				},
			},
			wantNil: false,
		},
		{
			name: "enabled podman with default socket",
			config: &config.DiscoveryConfig{
				Podman: &config.PodmanDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: false,
		},
		{
			name: "enabled podman with labels",
			config: &config.DiscoveryConfig{
				Podman: &config.PodmanDiscoveryConfig{
					Enabled: true,
					Labels:  map[string]string{"app": "test"},
				},
			},
			wantNil: false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createPodmanDiscoverer()
			// Verify discoverer matches expectation.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createPodmanDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}

// TestFactory_createPortScanDiscoverer tests createPortScanDiscoverer method.
func TestFactory_createPortScanDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil portscan config",
			config:  &config.DiscoveryConfig{},
			wantNil: true,
		},
		{
			name: "enabled portscan",
			config: &config.DiscoveryConfig{
				PortScan: &config.PortScanDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: false,
		},
		{
			name: "enabled portscan with interfaces",
			config: &config.DiscoveryConfig{
				PortScan: &config.PortScanDiscoveryConfig{
					Enabled:    true,
					Interfaces: []string{"lo", "eth0"},
				},
			},
			wantNil: false,
		},
		{
			name: "enabled portscan with port filters",
			config: &config.DiscoveryConfig{
				PortScan: &config.PortScanDiscoveryConfig{
					Enabled:      true,
					ExcludePorts: []int{22, 25},
					IncludePorts: []int{80, 443},
				},
			},
			wantNil: false,
		},
	}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Factory{config: tc.config}
			discoverer := f.createPortScanDiscoverer()
			// Verify discoverer matches expectation.
			if (discoverer == nil) != tc.wantNil {
				t.Errorf("createPortScanDiscoverer() = %v, wantNil = %v", discoverer, tc.wantNil)
			}
		})
	}
}
