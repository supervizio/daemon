//go:build freebsd || openbsd || netbsd

package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBSDRCDiscoverer_Type tests the Type method returns TypeBSDRC.
func TestBSDRCDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     target.Type
	}{
		{
			name:     "returns bsd-rc type with nil patterns",
			patterns: nil,
			want:     target.TypeBSDRC,
		},
		{
			name:     "returns bsd-rc type with empty patterns",
			patterns: []string{},
			want:     target.TypeBSDRC,
		},
		{
			name:     "returns bsd-rc type with patterns",
			patterns: []string{"nginx", "sshd"},
			want:     target.TypeBSDRC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewBSDRCDiscoverer(tt.patterns)

			got := d.Type()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBSDRCDiscoverer_Discover tests the Discover method.
func TestBSDRCDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantErr     bool
		minTargets  int
		validateFn  func(*testing.T, []target.ExternalTarget)
	}{
		{
			name:       "discovers all services without patterns",
			patterns:   nil,
			wantErr:    false,
			minTargets: 0,
			validateFn: func(t *testing.T, targets []target.ExternalTarget) {
				// All targets should be BSD RC type.
				for _, tgt := range targets {
					assert.Equal(t, target.TypeBSDRC, tgt.Type)
					assert.Equal(t, target.SourceDiscovered, tgt.Source)
					assert.Equal(t, "exec", tgt.ProbeType)
					assert.NotEmpty(t, tgt.ID)
					assert.NotEmpty(t, tgt.Name)
					assert.Greater(t, tgt.Interval, time.Duration(0))
					assert.Greater(t, tgt.Timeout, time.Duration(0))
					assert.NotNil(t, tgt.Labels)
					assert.Contains(t, tgt.Labels, "bsdrc.service")
					assert.Contains(t, tgt.Labels, "bsdrc.os")
				}
			},
		},
		{
			name:       "discovers with empty patterns",
			patterns:   []string{},
			wantErr:    false,
			minTargets: 0,
			validateFn: func(t *testing.T, targets []target.ExternalTarget) {
				// Same validation as nil patterns.
				for _, tgt := range targets {
					assert.Equal(t, target.TypeBSDRC, tgt.Type)
					assert.Equal(t, "exec", tgt.ProbeType)
				}
			},
		},
		{
			name:       "filters by pattern",
			patterns:   []string{"nonexistent_service_*"},
			wantErr:    false,
			minTargets: 0,
			validateFn: func(t *testing.T, targets []target.ExternalTarget) {
				// Should return empty or very few results.
				assert.LessOrEqual(t, len(targets), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewBSDRCDiscoverer(tt.patterns)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			targets, err := d.Discover(ctx)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(targets), tt.minTargets)

			if tt.validateFn != nil {
				tt.validateFn(t, targets)
			}
		})
	}
}

// TestBSDRCDiscoverer_DiscoverCancellation tests context cancellation.
func TestBSDRCDiscoverer_DiscoverCancellation(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		cancelFn func(context.Context, context.CancelFunc)
	}{
		{
			name:     "respects context cancellation",
			patterns: []string{"*"},
			cancelFn: func(ctx context.Context, cancel context.CancelFunc) {
				// Cancel immediately.
				cancel()
			},
		},
		{
			name:     "respects context timeout",
			patterns: []string{"*"},
			cancelFn: func(ctx context.Context, cancel context.CancelFunc) {
				// Wait for timeout to trigger.
				<-ctx.Done()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewBSDRCDiscoverer(tt.patterns)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			if tt.cancelFn != nil {
				tt.cancelFn(ctx, cancel)
			}

			// Discovery may succeed or fail depending on timing.
			// We just verify it doesn't panic and respects cancellation.
			_, _ = d.Discover(ctx)
		})
	}
}

// TestBSDRCDiscoverer_TargetFormat tests target ID and name format.
func TestBSDRCDiscoverer_TargetFormat(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
	}{
		{
			name:     "target IDs have bsd-rc prefix",
			patterns: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewBSDRCDiscoverer(tt.patterns)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			targets, err := d.Discover(ctx)
			require.NoError(t, err)

			for _, tgt := range targets {
				// ID should have "bsd-rc:" prefix.
				assert.Contains(t, tgt.ID, "bsd-rc:")
				// Name should not have prefix.
				assert.NotContains(t, tgt.Name, "bsd-rc:")
				// Labels should contain service name.
				assert.Equal(t, tgt.Name, tgt.Labels["bsdrc.service"])
			}
		})
	}
}

// TestBSDRCDiscoverer_ProbeConfiguration tests probe is correctly configured.
func TestBSDRCDiscoverer_ProbeConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
	}{
		{
			name:     "probe configuration is valid",
			patterns: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewBSDRCDiscoverer(tt.patterns)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			targets, err := d.Discover(ctx)
			require.NoError(t, err)

			for _, tgt := range targets {
				// All targets should use exec probe.
				assert.Equal(t, "exec", tgt.ProbeType)
				// Probe target should have command configured.
				assert.NotEmpty(t, tgt.ProbeTarget.Command)
				// Timing should be positive.
				assert.Greater(t, tgt.Interval, time.Duration(0))
				assert.Greater(t, tgt.Timeout, time.Duration(0))
				// Thresholds should be positive.
				assert.Greater(t, tgt.SuccessThreshold, 0)
				assert.Greater(t, tgt.FailureThreshold, 0)
			}
		})
	}
}
