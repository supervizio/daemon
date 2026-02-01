// Package config_test provides external tests for portscan discovery configuration.
package config_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
)

// TestNewPortScanDiscoveryConfig tests the NewPortScanDiscoveryConfig constructor.
func TestNewPortScanDiscoveryConfig(t *testing.T) {
	tests := []struct {
		name        string
		wantNil     bool
		wantEnabled bool
		wantSSH     bool
	}{
		{
			name:        "creates non-nil config",
			wantNil:     false,
			wantEnabled: false,
			wantSSH:     true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewPortScanDiscoveryConfig()
			// Verify config is not nil.
			if (cfg == nil) != tt.wantNil {
				t.Errorf("NewPortScanDiscoveryConfig() = %v, wantNil = %v", cfg, tt.wantNil)
			}
			// Verify disabled by default.
			if cfg != nil && cfg.Enabled != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.wantEnabled)
			}
			// Verify SSH excluded by default.
			if cfg != nil && tt.wantSSH {
				foundSSH := false
				for _, port := range cfg.ExcludePorts {
					if port == 22 {
						foundSSH = true
						break
					}
				}
				if !foundSSH {
					t.Error("Expected SSH port 22 in ExcludePorts by default")
				}
			}
		})
	}
}

// TestPortScanDiscoveryConfig_Defaults tests default values.
func TestPortScanDiscoveryConfig_Defaults(t *testing.T) {
	tests := []struct {
		name               string
		wantEnabledDefault bool
		wantSSHExcluded    bool
	}{
		{
			name:               "default disabled with SSH excluded",
			wantEnabledDefault: false,
			wantSSHExcluded:    true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewPortScanDiscoveryConfig()
			// Verify Enabled default.
			if cfg.Enabled != tt.wantEnabledDefault {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.wantEnabledDefault)
			}
			// Verify ExcludePorts contains SSH.
			if tt.wantSSHExcluded {
				foundSSH := false
				for _, port := range cfg.ExcludePorts {
					if port == 22 {
						foundSSH = true
						break
					}
				}
				if !foundSSH {
					t.Error("Expected SSH (22) in ExcludePorts by default")
				}
			}
			// Verify empty slices are initialized.
			if cfg.Interfaces == nil {
				t.Error("Interfaces should be initialized, not nil")
			}
			if cfg.IncludePorts == nil {
				t.Error("IncludePorts should be initialized, not nil")
			}
		})
	}
}

// TestPortScanDiscoveryConfig_Fields tests field assignments.
func TestPortScanDiscoveryConfig_Fields(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		interfaces   []string
		excludePorts []int
		includePorts []int
	}{
		{
			name:         "empty config",
			enabled:      false,
			interfaces:   []string{},
			excludePorts: []int{},
			includePorts: []int{},
		},
		{
			name:         "with interfaces",
			enabled:      true,
			interfaces:   []string{"eth0", "lo"},
			excludePorts: []int{22, 25},
			includePorts: []int{},
		},
		{
			name:         "with include ports",
			enabled:      true,
			interfaces:   []string{},
			excludePorts: []int{},
			includePorts: []int{80, 443},
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.PortScanDiscoveryConfig{
				Enabled:      tt.enabled,
				Interfaces:   tt.interfaces,
				ExcludePorts: tt.excludePorts,
				IncludePorts: tt.includePorts,
			}
			// Verify all fields match.
			if cfg.Enabled != tt.enabled {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.enabled)
			}
			if len(cfg.Interfaces) != len(tt.interfaces) {
				t.Errorf("len(Interfaces) = %v, want %v", len(cfg.Interfaces), len(tt.interfaces))
			}
			if len(cfg.ExcludePorts) != len(tt.excludePorts) {
				t.Errorf("len(ExcludePorts) = %v, want %v", len(cfg.ExcludePorts), len(tt.excludePorts))
			}
			if len(cfg.IncludePorts) != len(tt.includePorts) {
				t.Errorf("len(IncludePorts) = %v, want %v", len(cfg.IncludePorts), len(tt.includePorts))
			}
		})
	}
}
