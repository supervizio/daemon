// Package discovery provides internal tests for the static discoverer.
package discovery

import (
	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"testing"
)

// TestStaticDiscoverer_parseTargetType tests parseTargetType method.
func TestStaticDiscoverer_parseTargetType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		wantType target.Type
	}{
		{"systemd", "systemd", target.TypeSystemd},
		{"docker", "docker", target.TypeDocker},
		{"kubernetes", "kubernetes", target.TypeKubernetes},
		{"k8s alias", "k8s", target.TypeKubernetes},
		{"nomad", "nomad", target.TypeNomad},
		{"remote", "remote", target.TypeRemote},
		{"empty string", "", target.TypeRemote},
		{"custom type", "custom-type", target.TypeCustom},
		{"unknown type", "unknown", target.TypeCustom},
	}

	d := &StaticDiscoverer{}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := d.parseTargetType(tc.typeStr)
			// Verify type matches expected.
			if got != tc.wantType {
				t.Errorf("parseTargetType(%q) = %v, want %v", tc.typeStr, got, tc.wantType)
			}
		})
	}
}

// TestStaticDiscoverer_configToTarget tests configToTarget method.
func TestStaticDiscoverer_configToTarget(t *testing.T) {
	tests := []struct {
		name   string
		cfg    config.TargetConfig
		wantID string
	}{
		{
			name:   "simple target",
			cfg:    config.TargetConfig{Name: "test", Type: "remote"},
			wantID: "remote:test",
		},
		{
			name:   "systemd target",
			cfg:    config.TargetConfig{Name: "nginx", Type: "systemd"},
			wantID: "systemd:nginx",
		},
	}

	d := &StaticDiscoverer{}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := d.configToTarget(tc.cfg)
			// Verify ID matches expected.
			if got.ID != tc.wantID {
				t.Errorf("configToTarget().ID = %q, want %q", got.ID, tc.wantID)
			}
		})
	}
}

// TestStaticDiscoverer_configureProbe tests configureProbe method.
func TestStaticDiscoverer_configureProbe(t *testing.T) {
	tests := []struct {
		name          string
		cfg           config.TargetConfig
		wantProbeType string
	}{
		{
			name: "tcp probe",
			cfg: config.TargetConfig{
				Name:    "test",
				Address: "localhost:8080",
				Probe:   config.ProbeConfig{Type: "tcp"},
			},
			wantProbeType: "tcp",
		},
		{
			name: "udp probe",
			cfg: config.TargetConfig{
				Name:    "test",
				Address: "localhost:8080",
				Probe:   config.ProbeConfig{Type: "udp"},
			},
			wantProbeType: "udp",
		},
		{
			name: "http probe",
			cfg: config.TargetConfig{
				Name:    "test",
				Address: "http://localhost:8080",
				Probe:   config.ProbeConfig{Type: "http"},
			},
			wantProbeType: "http",
		},
		{
			name: "https probe",
			cfg: config.TargetConfig{
				Name:    "test",
				Address: "https://localhost:8080",
				Probe:   config.ProbeConfig{Type: "https"},
			},
			wantProbeType: "https",
		},
		{
			name: "icmp probe",
			cfg: config.TargetConfig{
				Name:    "test",
				Address: "192.168.1.1",
				Probe:   config.ProbeConfig{Type: "icmp"},
			},
			wantProbeType: "icmp",
		},
		{
			name: "ping probe",
			cfg: config.TargetConfig{
				Name:    "test",
				Address: "192.168.1.1",
				Probe:   config.ProbeConfig{Type: "ping"},
			},
			wantProbeType: "ping",
		},
		{
			name: "exec probe with command",
			cfg: config.TargetConfig{
				Name: "test",
				Probe: config.ProbeConfig{
					Type:    "exec",
					Command: "/bin/true",
					Args:    []string{"arg1"},
				},
			},
			wantProbeType: "exec",
		},
		{
			name: "exec probe without command",
			cfg: config.TargetConfig{
				Name: "test",
				Probe: config.ProbeConfig{
					Type: "exec",
				},
			},
			wantProbeType: "exec",
		},
		{
			name: "empty probe",
			cfg: config.TargetConfig{
				Name:  "test",
				Probe: config.ProbeConfig{Type: ""},
			},
			wantProbeType: "",
		},
	}

	d := &StaticDiscoverer{}

	// Iterate over test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tgt target.ExternalTarget
			d.configureProbe(&tgt, tc.cfg)
			// Verify probe type matches expected.
			if tgt.ProbeType != tc.wantProbeType {
				t.Errorf("configureProbe() ProbeType = %q, want %q", tgt.ProbeType, tc.wantProbeType)
			}
		})
	}
}
