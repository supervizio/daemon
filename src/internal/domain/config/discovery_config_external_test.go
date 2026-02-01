package config_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

func TestDiscoveryConfig(t *testing.T) {
	tests := []struct {
		name      string
		discovery config.DiscoveryConfig
		wantNil   bool
	}{
		{
			name: "systemd discovery",
			discovery: config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name: "openrc discovery",
			discovery: config.DiscoveryConfig{
				OpenRC: &config.OpenRCDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name: "bsdrc discovery",
			discovery: config.DiscoveryConfig{
				BSDRC: &config.BSDRCDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name: "docker discovery",
			discovery: config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name: "podman discovery",
			discovery: config.DiscoveryConfig{
				Podman: &config.PodmanDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name: "kubernetes discovery",
			discovery: config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name: "nomad discovery",
			discovery: config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{Enabled: true},
			},
			wantNil: false,
		},
		{
			name:      "zero value",
			discovery: config.DiscoveryConfig{},
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantNil {
				assert.Nil(t, tt.discovery.Systemd)
				assert.Nil(t, tt.discovery.OpenRC)
				assert.Nil(t, tt.discovery.BSDRC)
				assert.Nil(t, tt.discovery.Docker)
				assert.Nil(t, tt.discovery.Podman)
				assert.Nil(t, tt.discovery.Kubernetes)
				assert.Nil(t, tt.discovery.Nomad)
			} else {
				// At least one field should be set
				hasAny := tt.discovery.Systemd != nil ||
					tt.discovery.OpenRC != nil ||
					tt.discovery.BSDRC != nil ||
					tt.discovery.Docker != nil ||
					tt.discovery.Podman != nil ||
					tt.discovery.Kubernetes != nil ||
					tt.discovery.Nomad != nil
				assert.True(t, hasAny)
			}
		})
	}
}
