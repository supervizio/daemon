//go:build unix

package discovery_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
)

// TestNomadDiscoverer_Type verifies the discoverer returns correct type.
func TestNomadDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name string
		want target.Type
	}{
		{
			name: "returns TypeNomad",
			want: target.TypeNomad,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.NomadDiscoveryConfig{
				Address: "http://localhost:4646",
			}
			d := discovery.NewNomadDiscoverer(cfg)

			got := d.Type()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNomadDiscoverer_Discover verifies error handling for invalid address.
func TestNomadDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "invalid address returns error",
			address: "http://invalid-nomad-server:9999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.NomadDiscoveryConfig{
				Address: tt.address,
			}
			d := discovery.NewNomadDiscoverer(cfg)

			_, err := d.Discover(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewNomadDiscoverer verifies constructor behavior.
func TestNewNomadDiscoverer(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.NomadDiscoveryConfig
		wantPanic bool
	}{
		{
			name: "creates with default address",
			cfg: &config.NomadDiscoveryConfig{
				Address: "",
			},
			wantPanic: false,
		},
		{
			name: "creates with custom address",
			cfg: &config.NomadDiscoveryConfig{
				Address: "http://nomad.example.com:4646",
			},
			wantPanic: false,
		},
		{
			name: "creates with namespace filter",
			cfg: &config.NomadDiscoveryConfig{
				Address:   "http://localhost:4646",
				Namespace: "production",
			},
			wantPanic: false,
		},
		{
			name: "creates with job filter",
			cfg: &config.NomadDiscoveryConfig{
				Address:   "http://localhost:4646",
				JobFilter: "web-",
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewNomadDiscoverer(tt.cfg)

			assert.NotNil(t, d)
		})
	}
}
