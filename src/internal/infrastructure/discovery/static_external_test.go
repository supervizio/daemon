package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStaticDiscoverer_Type verifies the discoverer returns correct type.
func TestStaticDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name string
		want target.Type
	}{
		{
			name: "returns TypeRemote",
			want: target.TypeRemote,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewStaticDiscoverer(nil)

			got := d.Type()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestStaticDiscoverer_Discover verifies target conversion.
func TestStaticDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name      string
		targets   []config.TargetConfig
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty config",
			targets:   []config.TargetConfig{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single target",
			targets: []config.TargetConfig{
				{
					Name:     "web-server",
					Type:     "remote",
					Address:  "192.168.1.10",
					Probe:    config.ProbeConfig{Type: "tcp"},
					Interval: shared.FromTimeDuration(30 * time.Second),
					Timeout:  shared.FromTimeDuration(5 * time.Second),
					Labels:   map[string]string{"env": "prod"},
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewStaticDiscoverer(tt.targets)

			targets, err := d.Discover(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, targets, tt.wantCount)
			}
		})
	}
}

// TestNewStaticDiscoverer verifies constructor behavior.
func TestNewStaticDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		targets []config.TargetConfig
	}{
		{
			name:    "creates with nil targets",
			targets: nil,
		},
		{
			name:    "creates with empty targets",
			targets: []config.TargetConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewStaticDiscoverer(tt.targets)

			assert.NotNil(t, d)
		})
	}
}
