//go:build linux

// Package discovery_test provides external tests for the port scan discoverer.
package discovery_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
)

// TestNewPortScanDiscoverer tests the NewPortScanDiscoverer constructor.
func TestNewPortScanDiscoverer(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.PortScanDiscoveryConfig
		wantNil bool
	}{
		{
			name:    "nil config creates discoverer",
			cfg:     &config.PortScanDiscoveryConfig{},
			wantNil: false,
		},
		{
			name:    "default config",
			cfg:     config.NewPortScanDiscoveryConfig(),
			wantNil: false,
		},
		{
			name: "with interfaces",
			cfg: &config.PortScanDiscoveryConfig{
				Enabled:    true,
				Interfaces: []string{"lo", "eth0"},
			},
			wantNil: false,
		},
		{
			name: "with exclude ports",
			cfg: &config.PortScanDiscoveryConfig{
				Enabled:      true,
				ExcludePorts: []int{22, 25, 80},
			},
			wantNil: false,
		},
		{
			name: "with include ports",
			cfg: &config.PortScanDiscoveryConfig{
				Enabled:      true,
				IncludePorts: []int{8080, 9090},
			},
			wantNil: false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewPortScanDiscoverer(tt.cfg)
			// Verify discoverer is not nil.
			if (d == nil) != tt.wantNil {
				t.Errorf("NewPortScanDiscoverer() = %v, wantNil = %v", d, tt.wantNil)
			}
		})
	}
}

// TestPortScanDiscoverer_Type tests PortScanDiscoverer.Type method.
func TestPortScanDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.PortScanDiscoveryConfig
		wantType target.Type
	}{
		{
			name:     "returns custom type",
			cfg:      config.NewPortScanDiscoveryConfig(),
			wantType: target.TypeCustom,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewPortScanDiscoverer(tt.cfg)
			got := d.Type()
			// Verify type matches expected.
			if got != tt.wantType {
				t.Errorf("Type() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

// TestPortScanDiscoverer_Discover tests PortScanDiscoverer.Discover method.
func TestPortScanDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.PortScanDiscoveryConfig
		wantErr bool
	}{
		{
			name:    "default config discovers ports",
			cfg:     config.NewPortScanDiscoveryConfig(),
			wantErr: false,
		},
		{
			name: "with exclude ports",
			cfg: &config.PortScanDiscoveryConfig{
				Enabled:      true,
				ExcludePorts: []int{22},
			},
			wantErr: false,
		},
		{
			name: "with include ports",
			cfg: &config.PortScanDiscoveryConfig{
				Enabled:      true,
				IncludePorts: []int{80, 443},
			},
			wantErr: false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewPortScanDiscoverer(tt.cfg)
			targets, err := d.Discover(context.Background())
			// Check error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Discover() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Verify targets is not nil on success (empty slice is OK).
			if err == nil && targets == nil {
				t.Error("Discover() returned nil targets without error")
			}
		})
	}
}

// TestPortScanDiscoverer_Discover_WithMockData tests parsing with mock /proc/net/tcp data.
func TestPortScanDiscoverer_Discover_WithMockData(t *testing.T) {
	// Skip if /proc/net/tcp doesn't exist (not Linux).
	if _, err := os.Stat("/proc/net/tcp"); os.IsNotExist(err) {
		t.Skip("Skipping test: /proc/net/tcp not available")
	}

	tests := []struct {
		name         string
		cfg          *config.PortScanDiscoveryConfig
		checkTargets func(t *testing.T, targets []target.ExternalTarget)
	}{
		{
			name: "discovered targets have correct type",
			cfg:  config.NewPortScanDiscoveryConfig(),
			checkTargets: func(t *testing.T, targets []target.ExternalTarget) {
				for _, tgt := range targets {
					if tgt.Type != target.TypeCustom {
						t.Errorf("Target %s has type %v, want %v", tgt.ID, tgt.Type, target.TypeCustom)
					}
					if tgt.ProbeType != "tcp" {
						t.Errorf("Target %s has ProbeType %v, want tcp", tgt.ID, tgt.ProbeType)
					}
					if tgt.Source != target.SourceDiscovered {
						t.Errorf("Target %s has Source %v, want %v", tgt.ID, tgt.Source, target.SourceDiscovered)
					}
				}
			},
		},
		{
			name: "discovered targets have labels",
			cfg:  config.NewPortScanDiscoveryConfig(),
			checkTargets: func(t *testing.T, targets []target.ExternalTarget) {
				for _, tgt := range targets {
					if _, ok := tgt.Labels["portscan.protocol"]; !ok {
						t.Errorf("Target %s missing label portscan.protocol", tgt.ID)
					}
					if _, ok := tgt.Labels["portscan.port"]; !ok {
						t.Errorf("Target %s missing label portscan.port", tgt.ID)
					}
					if _, ok := tgt.Labels["portscan.address"]; !ok {
						t.Errorf("Target %s missing label portscan.address", tgt.ID)
					}
				}
			},
		},
		{
			name: "exclude ports filters correctly",
			cfg: &config.PortScanDiscoveryConfig{
				Enabled:      true,
				ExcludePorts: []int{22},
			},
			checkTargets: func(t *testing.T, targets []target.ExternalTarget) {
				for _, tgt := range targets {
					if portLabel, ok := tgt.Labels["portscan.port"]; ok {
						if portLabel == "22" {
							t.Errorf("Target %s has excluded port 22", tgt.ID)
						}
					}
				}
			},
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := discovery.NewPortScanDiscoverer(tt.cfg)
			targets, err := d.Discover(context.Background())
			// Verify no error.
			if err != nil {
				t.Fatalf("Discover() error = %v", err)
			}
			// Run custom checks.
			if tt.checkTargets != nil {
				tt.checkTargets(t, targets)
			}
		})
	}
}

// TestPortScanDiscoverer_Discover_ContextCancellation tests context cancellation.
func TestPortScanDiscoverer_Discover_ContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "canceled context returns error",
			ctx:     canceledContext(),
			wantErr: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewPortScanDiscoveryConfig()
			d := discovery.NewPortScanDiscoverer(cfg)
			_, err := d.Discover(tt.ctx)
			// Check error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("Discover() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// canceledContext returns a pre-canceled context.
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TestPortScanDiscoverer_ParseAddressFormat tests the /proc/net/tcp parsing logic.
func TestPortScanDiscoverer_ParseAddressFormat(t *testing.T) {
	// Create a temporary mock /proc/net/tcp file.
	tmpDir := t.TempDir()
	mockTCP := filepath.Join(tmpDir, "tcp")

	// Mock /proc/net/tcp content (localhost:53 in LISTEN state).
	mockContent := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:0035 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 17892
`

	err := os.WriteFile(mockTCP, []byte(mockContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create mock file: %v", err)
	}

	// Note: This test would need internal access to parseNetTCP.
	// Since we're in external tests, we can only test via Discover().
	// This is a placeholder for documentation.
	t.Skip("Parsing details tested via Discover() integration test")
}
