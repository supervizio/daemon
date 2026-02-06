// Package config provides domain value objects for service configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestNewMonitoringConfig tests the NewMonitoringConfig constructor.
//
// Params:
//   - t: testing context
func TestNewMonitoringConfig(t *testing.T) {
	// testCase defines a test case for NewMonitoringConfig
	type testCase struct {
		name              string
		wantInterval      shared.Duration
		wantTimeout       shared.Duration
		wantSuccessThresh int
		wantFailureThresh int
	}

	// tests defines all test cases for NewMonitoringConfig
	tests := []testCase{
		{
			name:              "creates monitoring config with defaults",
			wantInterval:      shared.Seconds(30),
			wantTimeout:       shared.Seconds(5),
			wantSuccessThresh: 1,
			wantFailureThresh: 3,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create new monitoring config
			mc := config.NewMonitoringConfig()

			// verify defaults are set
			assert.Equal(t, tt.wantInterval, mc.Defaults.Interval, "default interval mismatch")
			assert.Equal(t, tt.wantTimeout, mc.Defaults.Timeout, "default timeout mismatch")
			assert.Equal(t, tt.wantSuccessThresh, mc.Defaults.SuccessThreshold, "default success threshold mismatch")
			assert.Equal(t, tt.wantFailureThresh, mc.Defaults.FailureThreshold, "default failure threshold mismatch")
			assert.Nil(t, mc.Targets, "targets should be nil")
		})
	}
}

// TestDefaultMonitoringDefaults tests the DefaultMonitoringDefaults function.
//
// Params:
//   - t: testing context
func TestDefaultMonitoringDefaults(t *testing.T) {
	// testCase defines a test case for DefaultMonitoringDefaults
	type testCase struct {
		name              string
		wantInterval      shared.Duration
		wantTimeout       shared.Duration
		wantSuccessThresh int
		wantFailureThresh int
	}

	// tests defines all test cases for DefaultMonitoringDefaults
	tests := []testCase{
		{
			name:              "returns correct default values",
			wantInterval:      shared.Seconds(30),
			wantTimeout:       shared.Seconds(5),
			wantSuccessThresh: 1,
			wantFailureThresh: 3,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// get default monitoring values
			defaults := config.DefaultMonitoringDefaults()

			// verify all fields have expected values
			assert.Equal(t, tt.wantInterval, defaults.Interval, "interval mismatch")
			assert.Equal(t, tt.wantTimeout, defaults.Timeout, "timeout mismatch")
			assert.Equal(t, tt.wantSuccessThresh, defaults.SuccessThreshold, "success threshold mismatch")
			assert.Equal(t, tt.wantFailureThresh, defaults.FailureThreshold, "failure threshold mismatch")
		})
	}
}

// TestMonitoringConfig_HasDiscoveryEnabled tests the HasDiscoveryEnabled method.
//
// Params:
//   - t: testing context
func TestMonitoringConfig_HasDiscoveryEnabled(t *testing.T) {
	// testCase defines a test case for HasDiscoveryEnabled
	type testCase struct {
		name      string
		setupFunc func() config.MonitoringConfig
		want      bool
	}

	// tests defines all test cases for HasDiscoveryEnabled
	tests := []testCase{
		{
			name: "no discovery enabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with no discovery
				return config.MonitoringConfig{}
			},
			want: false,
		},
		{
			name: "systemd discovery enabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with systemd discovery
				return config.MonitoringConfig{
					Discovery: config.DiscoveryConfig{
						Systemd: &config.SystemdDiscoveryConfig{
							Enabled: true,
						},
					},
				}
			},
			want: true,
		},
		{
			name: "systemd discovery disabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with systemd discovery disabled
				return config.MonitoringConfig{
					Discovery: config.DiscoveryConfig{
						Systemd: &config.SystemdDiscoveryConfig{
							Enabled: false,
						},
					},
				}
			},
			want: false,
		},
		{
			name: "docker discovery enabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with docker discovery
				return config.MonitoringConfig{
					Discovery: config.DiscoveryConfig{
						Docker: &config.DockerDiscoveryConfig{
							Enabled: true,
						},
					},
				}
			},
			want: true,
		},
		{
			name: "kubernetes discovery enabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with kubernetes discovery
				return config.MonitoringConfig{
					Discovery: config.DiscoveryConfig{
						Kubernetes: &config.KubernetesDiscoveryConfig{
							Enabled: true,
						},
					},
				}
			},
			want: true,
		},
		{
			name: "nomad discovery enabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with nomad discovery
				return config.MonitoringConfig{
					Discovery: config.DiscoveryConfig{
						Nomad: &config.NomadDiscoveryConfig{
							Enabled: true,
						},
					},
				}
			},
			want: true,
		},
		{
			name: "multiple discovery enabled",
			setupFunc: func() config.MonitoringConfig {
				// create config with multiple discovery types
				return config.MonitoringConfig{
					Discovery: config.DiscoveryConfig{
						Systemd: &config.SystemdDiscoveryConfig{
							Enabled: true,
						},
						Docker: &config.DockerDiscoveryConfig{
							Enabled: true,
						},
					},
				}
			},
			want: true,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup config
			mc := tt.setupFunc()

			// check discovery status
			got := mc.HasDiscoveryEnabled()

			// verify result
			assert.Equal(t, tt.want, got, "HasDiscoveryEnabled() result mismatch")
		})
	}
}

// TestMonitoringConfig_HasStaticTargets tests the HasStaticTargets method.
//
// Params:
//   - t: testing context
func TestMonitoringConfig_HasStaticTargets(t *testing.T) {
	// testCase defines a test case for HasStaticTargets
	type testCase struct {
		name string
		mc   config.MonitoringConfig
		want bool
	}

	// tests defines all test cases for HasStaticTargets
	tests := []testCase{
		{
			name: "no targets",
			mc: config.MonitoringConfig{
				Targets: nil,
			},
			want: false,
		},
		{
			name: "empty targets slice",
			mc: config.MonitoringConfig{
				Targets: []config.TargetConfig{},
			},
			want: false,
		},
		{
			name: "one target",
			mc: config.MonitoringConfig{
				Targets: []config.TargetConfig{
					{Name: "target1"},
				},
			},
			want: true,
		},
		{
			name: "multiple targets",
			mc: config.MonitoringConfig{
				Targets: []config.TargetConfig{
					{Name: "target1"},
					{Name: "target2"},
				},
			},
			want: true,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// check static targets
			got := tt.mc.HasStaticTargets()

			// verify result
			assert.Equal(t, tt.want, got, "HasStaticTargets() result mismatch")
		})
	}
}

// TestMonitoringConfig_IsEmpty tests the IsEmpty method.
//
// Params:
//   - t: testing context
func TestMonitoringConfig_IsEmpty(t *testing.T) {
	// testCase defines a test case for IsEmpty
	type testCase struct {
		name string
		mc   config.MonitoringConfig
		want bool
	}

	// tests defines all test cases for IsEmpty
	tests := []testCase{
		{
			name: "completely empty",
			mc:   config.MonitoringConfig{},
			want: true,
		},
		{
			name: "has static targets",
			mc: config.MonitoringConfig{
				Targets: []config.TargetConfig{
					{Name: "target1"},
				},
			},
			want: false,
		},
		{
			name: "has discovery enabled",
			mc: config.MonitoringConfig{
				Discovery: config.DiscoveryConfig{
					Systemd: &config.SystemdDiscoveryConfig{
						Enabled: true,
					},
				},
			},
			want: false,
		},
		{
			name: "has both targets and discovery",
			mc: config.MonitoringConfig{
				Targets: []config.TargetConfig{
					{Name: "target1"},
				},
				Discovery: config.DiscoveryConfig{
					Docker: &config.DockerDiscoveryConfig{
						Enabled: true,
					},
				},
			},
			want: false,
		},
		{
			name: "discovery config present but disabled",
			mc: config.MonitoringConfig{
				Discovery: config.DiscoveryConfig{
					Systemd: &config.SystemdDiscoveryConfig{
						Enabled: false,
					},
				},
			},
			want: true,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// check if empty
			got := tt.mc.IsEmpty()

			// verify result
			assert.Equal(t, tt.want, got, "IsEmpty() result mismatch")
		})
	}
}

// TestDiscoveryConfig_hasInitSystemDiscovery tests the hasInitSystemDiscovery helper.
//
// Params:
//   - t: testing context
func TestDiscoveryConfig_hasInitSystemDiscovery(t *testing.T) {
	// testCase defines a test case for hasInitSystemDiscovery
	type testCase struct {
		name string
		disc config.DiscoveryConfig
		want bool
	}

	// tests defines all test cases for hasInitSystemDiscovery
	tests := []testCase{
		{
			name: "no init system discovery",
			disc: config.DiscoveryConfig{},
			want: false,
		},
		{
			name: "systemd enabled",
			disc: config.DiscoveryConfig{
				Systemd: &config.SystemdDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "openrc enabled",
			disc: config.DiscoveryConfig{
				OpenRC: &config.OpenRCDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "bsdrc enabled",
			disc: config.DiscoveryConfig{
				BSDRC: &config.BSDRCDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test uses an unexported method, so we test via HasDiscoveryEnabled
			mc := config.MonitoringConfig{Discovery: tt.disc}
			got := mc.HasDiscoveryEnabled()
			assert.Equal(t, tt.want, got, "hasInitSystemDiscovery result mismatch")
		})
	}
}

// TestDiscoveryConfig_hasContainerDiscovery tests the hasContainerDiscovery helper.
//
// Params:
//   - t: testing context
func TestDiscoveryConfig_hasContainerDiscovery(t *testing.T) {
	// testCase defines a test case for hasContainerDiscovery
	type testCase struct {
		name string
		disc config.DiscoveryConfig
		want bool
	}

	// tests defines all test cases for hasContainerDiscovery
	tests := []testCase{
		{
			name: "no container discovery",
			disc: config.DiscoveryConfig{},
			want: false,
		},
		{
			name: "docker enabled",
			disc: config.DiscoveryConfig{
				Docker: &config.DockerDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "podman enabled",
			disc: config.DiscoveryConfig{
				Podman: &config.PodmanDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test uses an unexported method, so we test via HasDiscoveryEnabled
			mc := config.MonitoringConfig{Discovery: tt.disc}
			got := mc.HasDiscoveryEnabled()
			assert.Equal(t, tt.want, got, "hasContainerDiscovery result mismatch")
		})
	}
}

// TestDiscoveryConfig_hasOrchestratorDiscovery tests the hasOrchestratorDiscovery helper.
//
// Params:
//   - t: testing context
func TestDiscoveryConfig_hasOrchestratorDiscovery(t *testing.T) {
	// testCase defines a test case for hasOrchestratorDiscovery
	type testCase struct {
		name string
		disc config.DiscoveryConfig
		want bool
	}

	// tests defines all test cases for hasOrchestratorDiscovery
	tests := []testCase{
		{
			name: "no orchestrator discovery",
			disc: config.DiscoveryConfig{},
			want: false,
		},
		{
			name: "kubernetes enabled",
			disc: config.DiscoveryConfig{
				Kubernetes: &config.KubernetesDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
		{
			name: "nomad enabled",
			disc: config.DiscoveryConfig{
				Nomad: &config.NomadDiscoveryConfig{Enabled: true},
			},
			want: true,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test uses an unexported method, so we test via HasDiscoveryEnabled
			mc := config.MonitoringConfig{Discovery: tt.disc}
			got := mc.HasDiscoveryEnabled()
			assert.Equal(t, tt.want, got, "hasOrchestratorDiscovery result mismatch")
		})
	}
}
