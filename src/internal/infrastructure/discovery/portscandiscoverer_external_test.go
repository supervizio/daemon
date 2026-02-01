//go:build linux

package discovery_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPortScanDiscoverer verifies that a new port scan discoverer can be created.
func TestNewPortScanDiscoverer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		interfaces   []string
		excludePorts []int
		includePorts []int
		wantNotNil   bool
	}{
		{
			name:         "creates discoverer with full config",
			interfaces:   []string{"eth0", "lo"},
			excludePorts: []int{22, 23},
			includePorts: []int{80, 443},
			wantNotNil:   true,
		},
		{
			name:         "creates discoverer with empty config",
			interfaces:   []string{},
			excludePorts: []int{},
			includePorts: []int{},
			wantNotNil:   true,
		},
		{
			name:         "creates discoverer with only interfaces",
			interfaces:   []string{"lo"},
			excludePorts: nil,
			includePorts: nil,
			wantNotNil:   true,
		},
		{
			name:         "creates discoverer with only exclude ports",
			interfaces:   nil,
			excludePorts: []int{22},
			includePorts: nil,
			wantNotNil:   true,
		},
		{
			name:         "creates discoverer with only include ports",
			interfaces:   nil,
			excludePorts: nil,
			includePorts: []int{80, 443, 8080},
			wantNotNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{
				Interfaces:   tt.interfaces,
				ExcludePorts: tt.excludePorts,
				IncludePorts: tt.includePorts,
			}

			discoverer := discovery.NewPortScanDiscoverer(cfg)

			if tt.wantNotNil {
				require.NotNil(t, discoverer)
			}
		})
	}
}

// TestPortScanDiscoverer_Type verifies the discoverer returns the correct type.
func TestPortScanDiscoverer_Type(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      *config.PortScanDiscoveryConfig
		wantType target.Type
	}{
		{
			name:     "returns TypeCustom with empty config",
			cfg:      &config.PortScanDiscoveryConfig{},
			wantType: target.TypeCustom,
		},
		{
			name: "returns TypeCustom with configured ports",
			cfg: &config.PortScanDiscoveryConfig{
				IncludePorts: []int{80},
			},
			wantType: target.TypeCustom,
		},
		{
			name: "returns TypeCustom with interfaces",
			cfg: &config.PortScanDiscoveryConfig{
				Interfaces: []string{"lo"},
			},
			wantType: target.TypeCustom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			discoverer := discovery.NewPortScanDiscoverer(tt.cfg)

			targetType := discoverer.Type()

			assert.Equal(t, tt.wantType, targetType)
		})
	}
}

// TestPortScanDiscoverer_Discover verifies discovery can be called.
func TestPortScanDiscoverer_Discover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		excludePorts []int
		wantErr      bool
	}{
		{
			name:         "discovers with exclude ports",
			excludePorts: []int{22},
			wantErr:      false,
		},
		{
			name:         "discovers with no exclude ports",
			excludePorts: nil,
			wantErr:      false,
		},
		{
			name:         "discovers with empty exclude ports",
			excludePorts: []int{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{
				ExcludePorts: tt.excludePorts,
			}
			discoverer := discovery.NewPortScanDiscoverer(cfg)

			ctx := context.Background()
			targets, err := discoverer.Discover(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Discovery may return empty or populated depending on system state.
				// We just verify it doesn't error and returns a valid slice.
				require.NoError(t, err)
				assert.NotNil(t, targets)
			}
		})
	}
}

// TestPortScanDiscoverer_DiscoverWithCancelledContext verifies context cancellation.
func TestPortScanDiscoverer_DiscoverWithCancelledContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.PortScanDiscoveryConfig
		wantErr bool
	}{
		{
			name:    "returns error with cancelled context empty config",
			cfg:     &config.PortScanDiscoveryConfig{},
			wantErr: true,
		},
		{
			name: "returns error with cancelled context with exclude ports",
			cfg: &config.PortScanDiscoveryConfig{
				ExcludePorts: []int{22},
			},
			wantErr: true,
		},
		{
			name: "returns error with cancelled context with include ports",
			cfg: &config.PortScanDiscoveryConfig{
				IncludePorts: []int{80, 443},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			discoverer := discovery.NewPortScanDiscoverer(tt.cfg)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := discoverer.Discover(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPortScanDiscoverer_WithEmptyConfig verifies discoverer works with empty config.
func TestPortScanDiscoverer_WithEmptyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        *config.PortScanDiscoveryConfig
		wantNotNil bool
		wantType   target.Type
	}{
		{
			name:       "works with empty config struct",
			cfg:        &config.PortScanDiscoveryConfig{},
			wantNotNil: true,
			wantType:   target.TypeCustom,
		},
		{
			name: "works with zeroed fields",
			cfg: &config.PortScanDiscoveryConfig{
				Interfaces:   nil,
				ExcludePorts: nil,
				IncludePorts: nil,
			},
			wantNotNil: true,
			wantType:   target.TypeCustom,
		},
		{
			name: "works with empty slices",
			cfg: &config.PortScanDiscoveryConfig{
				Interfaces:   []string{},
				ExcludePorts: []int{},
				IncludePorts: []int{},
			},
			wantNotNil: true,
			wantType:   target.TypeCustom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			discoverer := discovery.NewPortScanDiscoverer(tt.cfg)

			if tt.wantNotNil {
				require.NotNil(t, discoverer)
			}
			assert.Equal(t, tt.wantType, discoverer.Type())
		})
	}
}
