package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoveryConfig_hasInitSystemDiscovery(t *testing.T) {
	tests := []struct {
		name string
		d    *DiscoveryConfig
		want bool
	}{
		{
			name: "systemd enabled",
			d: &DiscoveryConfig{
				Systemd: &SystemdDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "openrc enabled",
			d: &DiscoveryConfig{
				OpenRC: &OpenRCDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "bsdrc enabled",
			d: &DiscoveryConfig{
				BSDRC: &BSDRCDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "systemd disabled",
			d: &DiscoveryConfig{
				Systemd: &SystemdDiscoveryConfig{Enabled: false},
			},
			want: false,
		},
		{
			name: "no init systems",
			d:    &DiscoveryConfig{},
			want: false,
		},
		{
			name: "nil systemd",
			d: &DiscoveryConfig{
				Systemd: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.hasInitSystemDiscovery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiscoveryConfig_hasContainerDiscovery(t *testing.T) {
	tests := []struct {
		name string
		d    *DiscoveryConfig
		want bool
	}{
		{
			name: "docker enabled",
			d: &DiscoveryConfig{
				Docker: &DockerDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "podman enabled",
			d: &DiscoveryConfig{
				Podman: &PodmanDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "docker disabled",
			d: &DiscoveryConfig{
				Docker: &DockerDiscoveryConfig{Enabled: false},
			},
			want: false,
		},
		{
			name: "no containers",
			d:    &DiscoveryConfig{},
			want: false,
		},
		{
			name: "nil docker",
			d: &DiscoveryConfig{
				Docker: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.hasContainerDiscovery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiscoveryConfig_hasOrchestratorDiscovery(t *testing.T) {
	tests := []struct {
		name string
		d    *DiscoveryConfig
		want bool
	}{
		{
			name: "kubernetes enabled",
			d: &DiscoveryConfig{
				Kubernetes: &KubernetesDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "nomad enabled",
			d: &DiscoveryConfig{
				Nomad: &NomadDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "kubernetes disabled",
			d: &DiscoveryConfig{
				Kubernetes: &KubernetesDiscoveryConfig{Enabled: false},
			},
			want: false,
		},
		{
			name: "no orchestrators",
			d:    &DiscoveryConfig{},
			want: false,
		},
		{
			name: "nil kubernetes",
			d: &DiscoveryConfig{
				Kubernetes: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.hasOrchestratorDiscovery()
			assert.Equal(t, tt.want, got)
		})
	}
}
