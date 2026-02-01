package discovery_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
)

// TestFactoryCreateDiscoverers tests CreateDiscoverers method.
func TestFactory_CreateDiscoverers(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.DiscoveryConfig
		wantCount int
	}{
		{
			name:      "nil config returns nil",
			cfg:       nil,
			wantCount: 0,
		},
		{
			name:      "empty config returns empty slice",
			cfg:       &config.DiscoveryConfig{},
			wantCount: 0,
		},
		{
			name: "docker only",
			cfg: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: true,
				},
			},
			wantCount: 1,
		},
		{
			name: "multiple discoverers",
			cfg: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: true,
				},
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: true,
				},
			},
			wantCount: 2,
		},
		{
			name: "disabled discoverers not created",
			cfg: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: false,
				},
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: false,
				},
			},
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := discovery.NewFactory(tc.cfg)
			discoverers := factory.CreateDiscoverers()

			if tc.cfg == nil {
				if discoverers != nil {
					t.Errorf("CreateDiscoverers() with nil config = %v, want nil", discoverers)
				}
				return
			}

			if len(discoverers) != tc.wantCount {
				t.Errorf("CreateDiscoverers() count = %d, want %d", len(discoverers), tc.wantCount)
			}
		})
	}
}

// TestFactoryAddSystemdDiscoverer tests addSystemdDiscoverer method.
func TestFactory_addSystemdDiscoverer(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.DiscoveryConfig
		wantAdded bool
	}{
		{
			name:      "nil config no add",
			cfg:       &config.DiscoveryConfig{},
			wantAdded: false,
		},
		{
			name: "disabled no add",
			cfg: &config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: false,
				},
			},
			wantAdded: false,
		},
		{
			name: "enabled adds discoverer",
			cfg: &config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: true,
				},
			},
			wantAdded: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := discovery.NewFactory(tc.cfg)
			discoverers := factory.CreateDiscoverers()

			added := false
			for _, d := range discoverers {
				if d.Type() == "systemd" {
					added = true
					break
				}
			}

			if added != tc.wantAdded {
				t.Errorf("addSystemdDiscoverer() added = %v, want %v", added, tc.wantAdded)
			}
		})
	}
}

// TestFactoryAddDockerDiscoverer tests addDockerDiscoverer method.
func TestFactory_addDockerDiscoverer(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.DiscoveryConfig
		wantAdded bool
	}{
		{
			name:      "nil config no add",
			cfg:       &config.DiscoveryConfig{},
			wantAdded: false,
		},
		{
			name: "disabled no add",
			cfg: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: false,
				},
			},
			wantAdded: false,
		},
		{
			name: "enabled adds discoverer",
			cfg: &config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{
					Enabled: true,
				},
			},
			wantAdded: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := discovery.NewFactory(tc.cfg)
			discoverers := factory.CreateDiscoverers()

			added := false
			for _, d := range discoverers {
				if d.Type() == "docker" {
					added = true
					break
				}
			}

			if added != tc.wantAdded {
				t.Errorf("addDockerDiscoverer() added = %v, want %v", added, tc.wantAdded)
			}
		})
	}
}

// TestFactoryAddKubernetesDiscoverer tests addKubernetesDiscoverer method.
func TestFactory_addKubernetesDiscoverer(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.DiscoveryConfig
		wantAdded bool
	}{
		{
			name:      "nil config no add",
			cfg:       &config.DiscoveryConfig{},
			wantAdded: false,
		},
		{
			name: "disabled no add",
			cfg: &config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{
					Enabled: false,
				},
			},
			wantAdded: false,
		},
		{
			name: "enabled returns nil (not implemented)",
			cfg: &config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{
					Enabled: true,
				},
			},
			wantAdded: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := discovery.NewFactory(tc.cfg)
			discoverers := factory.CreateDiscoverers()

			added := false
			for _, d := range discoverers {
				if d.Type() == "kubernetes" {
					added = true
					break
				}
			}

			if added != tc.wantAdded {
				t.Errorf("addKubernetesDiscoverer() added = %v, want %v", added, tc.wantAdded)
			}
		})
	}
}

// TestFactoryAddNomadDiscoverer tests addNomadDiscoverer method.
func TestFactory_addNomadDiscoverer(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.DiscoveryConfig
		wantAdded bool
	}{
		{
			name:      "nil config no add",
			cfg:       &config.DiscoveryConfig{},
			wantAdded: false,
		},
		{
			name: "disabled no add",
			cfg: &config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{
					Enabled: false,
				},
			},
			wantAdded: false,
		},
		{
			name: "enabled returns nil (not implemented)",
			cfg: &config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{
					Enabled: true,
				},
			},
			wantAdded: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := discovery.NewFactory(tc.cfg)
			discoverers := factory.CreateDiscoverers()

			added := false
			for _, d := range discoverers {
				if d.Type() == "nomad" {
					added = true
					break
				}
			}

			if added != tc.wantAdded {
				t.Errorf("addNomadDiscoverer() added = %v, want %v", added, tc.wantAdded)
			}
		})
	}
}

// TestNewFactory tests NewFactory constructor.
func TestNewFactory(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.DiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantNil: false,
		},
		{
			name:    "empty config",
			cfg:     &config.DiscoveryConfig{},
			wantNil: false,
		},
		{
			name: "with systemd config",
			cfg: &config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{
					Enabled: true,
				},
			},
			wantNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			factory := discovery.NewFactory(tc.cfg)

			if (factory == nil) != tc.wantNil {
				t.Errorf("NewFactory() = %v, wantNil %v", factory, tc.wantNil)
			}
		})
	}
}
