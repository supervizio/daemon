package yaml_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/persistence/config/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMonitoringConfigDTO_ToDomain tests yaml.MonitoringConfigDTO to domain conversion.
// It verifies that monitoring configuration fields are correctly mapped.
//
// Params:
//   - t: testing context
func TestMonitoringConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		dto                      *yaml.MonitoringConfigDTO
		expectedInterval         time.Duration
		expectedTimeout          time.Duration
		expectedSuccessThreshold int
		expectedFailureThreshold int
		expectedTargetCount      int
	}{
		{
			name: "full monitoring config converts correctly",
			dto: &yaml.MonitoringConfigDTO{
				Defaults: yaml.MonitoringDefaultsDTO{
					Interval:         yaml.Duration(30 * time.Second),
					Timeout:          yaml.Duration(5 * time.Second),
					SuccessThreshold: 1,
					FailureThreshold: 3,
				},
				Discovery: yaml.DiscoveryConfigDTO{
					Systemd: &yaml.SystemdDiscoveryDTO{
						Enabled:  true,
						Patterns: []string{"nginx.service"},
					},
				},
				Targets: []yaml.TargetConfigDTO{
					{
						Name:    "test-target",
						Address: "localhost:8080",
						Probe: yaml.ProbeDTO{
							Type: "http",
							Path: "/health",
						},
					},
				},
			},
			expectedInterval:         30 * time.Second,
			expectedTimeout:          5 * time.Second,
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
			expectedTargetCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			// Verify defaults
			assert.Equal(t, tt.expectedInterval, result.Defaults.Interval.Duration())
			assert.Equal(t, tt.expectedTimeout, result.Defaults.Timeout.Duration())
			assert.Equal(t, tt.expectedSuccessThreshold, result.Defaults.SuccessThreshold)
			assert.Equal(t, tt.expectedFailureThreshold, result.Defaults.FailureThreshold)

			// Verify targets
			assert.Len(t, result.Targets, tt.expectedTargetCount)
		})
	}
}

// TestDiscoveryConfigDTO_ToDomain tests yaml.DiscoveryConfigDTO to domain conversion.
// It verifies that discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestDiscoveryConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		dto                    *yaml.DiscoveryConfigDTO
		expectedSystemdEnabled bool
		expectedDockerEnabled  bool
	}{
		{
			name: "systemd discovery enabled",
			dto: &yaml.DiscoveryConfigDTO{
				Systemd: &yaml.SystemdDiscoveryDTO{
					Enabled:  true,
					Patterns: []string{"*.service"},
				},
			},
			expectedSystemdEnabled: true,
			expectedDockerEnabled:  false,
		},
		{
			name: "docker discovery enabled",
			dto: &yaml.DiscoveryConfigDTO{
				Docker: &yaml.DockerDiscoveryDTO{
					Enabled:    true,
					SocketPath: "/var/run/docker.sock",
				},
			},
			expectedSystemdEnabled: false,
			expectedDockerEnabled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			if tt.expectedSystemdEnabled {
				require.NotNil(t, result.Systemd)
				assert.True(t, result.Systemd.Enabled)
			}

			if tt.expectedDockerEnabled {
				require.NotNil(t, result.Docker)
				assert.True(t, result.Docker.Enabled)
			}
		})
	}
}

// TestTargetConfigDTO_ToDomain tests yaml.TargetConfigDTO to domain conversion.
// It verifies that target configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestTargetConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dto             *yaml.TargetConfigDTO
		expectedName    string
		expectedAddress string
		expectedType    string
	}{
		{
			name: "http target",
			dto: &yaml.TargetConfigDTO{
				Name:    "api",
				Type:    "remote",
				Address: "api.example.com:443",
				Probe: yaml.ProbeDTO{
					Type: "http",
					Path: "/health",
				},
				Interval: yaml.Duration(60 * time.Second),
				Timeout:  yaml.Duration(10 * time.Second),
			},
			expectedName:    "api",
			expectedAddress: "api.example.com:443",
			expectedType:    "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedAddress, result.Address)
			assert.Equal(t, tt.expectedType, result.Type)
		})
	}
}

// TestMonitoringDefaultsDTO_ToDomain tests yaml.MonitoringDefaultsDTO to domain conversion.
// It verifies that monitoring default values are correctly mapped and defaults applied.
//
// Params:
//   - t: testing context
func TestMonitoringDefaultsDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		dto                      *yaml.MonitoringDefaultsDTO
		expectedInterval         time.Duration
		expectedTimeout          time.Duration
		expectedSuccessThreshold int
		expectedFailureThreshold int
	}{
		{
			name: "all values specified",
			dto: &yaml.MonitoringDefaultsDTO{
				Interval:         yaml.Duration(30 * time.Second),
				Timeout:          yaml.Duration(5 * time.Second),
				SuccessThreshold: 2,
				FailureThreshold: 5,
			},
			expectedInterval:         30 * time.Second,
			expectedTimeout:          5 * time.Second,
			expectedSuccessThreshold: 2,
			expectedFailureThreshold: 5,
		},
		{
			name: "defaults applied when zero values",
			dto: &yaml.MonitoringDefaultsDTO{
				Interval:         0,
				Timeout:          0,
				SuccessThreshold: 0,
				FailureThreshold: 0,
			},
			expectedInterval:         30 * time.Second,
			expectedTimeout:          5 * time.Second,
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
		},
		{
			name: "partial values with defaults",
			dto: &yaml.MonitoringDefaultsDTO{
				Interval:         yaml.Duration(60 * time.Second),
				Timeout:          0,
				SuccessThreshold: 3,
				FailureThreshold: 0,
			},
			expectedInterval:         60 * time.Second,
			expectedTimeout:          5 * time.Second,
			expectedSuccessThreshold: 3,
			expectedFailureThreshold: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedInterval, result.Interval.Duration())
			assert.Equal(t, tt.expectedTimeout, result.Timeout.Duration())
			assert.Equal(t, tt.expectedSuccessThreshold, result.SuccessThreshold)
			assert.Equal(t, tt.expectedFailureThreshold, result.FailureThreshold)
		})
	}
}

// TestSystemdDiscoveryDTO_ToDomain tests yaml.SystemdDiscoveryDTO to domain conversion.
// It verifies that systemd discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestSystemdDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dto              *yaml.SystemdDiscoveryDTO
		expectedEnabled  bool
		expectedPatterns []string
	}{
		{
			name: "enabled with patterns",
			dto: &yaml.SystemdDiscoveryDTO{
				Enabled:  true,
				Patterns: []string{"nginx.service", "postgres.service"},
			},
			expectedEnabled:  true,
			expectedPatterns: []string{"nginx.service", "postgres.service"},
		},
		{
			name: "disabled without patterns",
			dto: &yaml.SystemdDiscoveryDTO{
				Enabled:  false,
				Patterns: nil,
			},
			expectedEnabled:  false,
			expectedPatterns: nil,
		},
		{
			name: "enabled with wildcard pattern",
			dto: &yaml.SystemdDiscoveryDTO{
				Enabled:  true,
				Patterns: []string{"*.service"},
			},
			expectedEnabled:  true,
			expectedPatterns: []string{"*.service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedPatterns, result.Patterns)
		})
	}
}

// TestOpenRCDiscoveryDTO_ToDomain tests yaml.OpenRCDiscoveryDTO to domain conversion.
// It verifies that OpenRC discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestOpenRCDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dto              *yaml.OpenRCDiscoveryDTO
		expectedEnabled  bool
		expectedPatterns []string
	}{
		{
			name: "enabled with patterns",
			dto: &yaml.OpenRCDiscoveryDTO{
				Enabled:  true,
				Patterns: []string{"nginx", "postgresql"},
			},
			expectedEnabled:  true,
			expectedPatterns: []string{"nginx", "postgresql"},
		},
		{
			name: "disabled without patterns",
			dto: &yaml.OpenRCDiscoveryDTO{
				Enabled:  false,
				Patterns: nil,
			},
			expectedEnabled:  false,
			expectedPatterns: nil,
		},
		{
			name: "enabled with wildcard pattern",
			dto: &yaml.OpenRCDiscoveryDTO{
				Enabled:  true,
				Patterns: []string{"*"},
			},
			expectedEnabled:  true,
			expectedPatterns: []string{"*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedPatterns, result.Patterns)
		})
	}
}

// TestBSDRCDiscoveryDTO_ToDomain tests yaml.BSDRCDiscoveryDTO to domain conversion.
// It verifies that BSD rc.d discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestBSDRCDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dto              *yaml.BSDRCDiscoveryDTO
		expectedEnabled  bool
		expectedPatterns []string
	}{
		{
			name: "enabled with patterns",
			dto: &yaml.BSDRCDiscoveryDTO{
				Enabled:  true,
				Patterns: []string{"nginx", "postgresql"},
			},
			expectedEnabled:  true,
			expectedPatterns: []string{"nginx", "postgresql"},
		},
		{
			name: "disabled without patterns",
			dto: &yaml.BSDRCDiscoveryDTO{
				Enabled:  false,
				Patterns: nil,
			},
			expectedEnabled:  false,
			expectedPatterns: nil,
		},
		{
			name: "enabled with wildcard pattern",
			dto: &yaml.BSDRCDiscoveryDTO{
				Enabled:  true,
				Patterns: []string{"*"},
			},
			expectedEnabled:  true,
			expectedPatterns: []string{"*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedPatterns, result.Patterns)
		})
	}
}

// TestDockerDiscoveryDTO_ToDomain tests yaml.DockerDiscoveryDTO to domain conversion.
// It verifies that Docker discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestDockerDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		dto                *yaml.DockerDiscoveryDTO
		expectedEnabled    bool
		expectedSocketPath string
		expectedLabels     map[string]string
	}{
		{
			name: "enabled with socket and labels",
			dto: &yaml.DockerDiscoveryDTO{
				Enabled:    true,
				SocketPath: "/var/run/docker.sock",
				Labels:     map[string]string{"app": "web", "env": "prod"},
			},
			expectedEnabled:    true,
			expectedSocketPath: "/var/run/docker.sock",
			expectedLabels:     map[string]string{"app": "web", "env": "prod"},
		},
		{
			name: "disabled without socket",
			dto: &yaml.DockerDiscoveryDTO{
				Enabled:    false,
				SocketPath: "",
				Labels:     nil,
			},
			expectedEnabled:    false,
			expectedSocketPath: "",
			expectedLabels:     nil,
		},
		{
			name: "enabled with custom socket path",
			dto: &yaml.DockerDiscoveryDTO{
				Enabled:    true,
				SocketPath: "/custom/docker.sock",
				Labels:     nil,
			},
			expectedEnabled:    true,
			expectedSocketPath: "/custom/docker.sock",
			expectedLabels:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedSocketPath, result.SocketPath)
			assert.Equal(t, tt.expectedLabels, result.Labels)
		})
	}
}

// TestPodmanDiscoveryDTO_ToDomain tests yaml.PodmanDiscoveryDTO to domain conversion.
// It verifies that Podman discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestPodmanDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		dto                *yaml.PodmanDiscoveryDTO
		expectedEnabled    bool
		expectedSocketPath string
		expectedLabels     map[string]string
	}{
		{
			name: "enabled with socket and labels",
			dto: &yaml.PodmanDiscoveryDTO{
				Enabled:    true,
				SocketPath: "/run/podman/podman.sock",
				Labels:     map[string]string{"app": "api", "tier": "backend"},
			},
			expectedEnabled:    true,
			expectedSocketPath: "/run/podman/podman.sock",
			expectedLabels:     map[string]string{"app": "api", "tier": "backend"},
		},
		{
			name: "disabled without socket",
			dto: &yaml.PodmanDiscoveryDTO{
				Enabled:    false,
				SocketPath: "",
				Labels:     nil,
			},
			expectedEnabled:    false,
			expectedSocketPath: "",
			expectedLabels:     nil,
		},
		{
			name: "enabled with custom socket path",
			dto: &yaml.PodmanDiscoveryDTO{
				Enabled:    true,
				SocketPath: "/custom/podman.sock",
				Labels:     nil,
			},
			expectedEnabled:    true,
			expectedSocketPath: "/custom/podman.sock",
			expectedLabels:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedSocketPath, result.SocketPath)
			assert.Equal(t, tt.expectedLabels, result.Labels)
		})
	}
}

// TestKubernetesDiscoveryDTO_ToDomain tests yaml.KubernetesDiscoveryDTO to domain conversion.
// It verifies that Kubernetes discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestKubernetesDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		dto                    *yaml.KubernetesDiscoveryDTO
		expectedEnabled        bool
		expectedKubeconfigPath string
		expectedNamespaces     []string
		expectedLabelSelector  string
	}{
		{
			name: "enabled with full config",
			dto: &yaml.KubernetesDiscoveryDTO{
				Enabled:        true,
				KubeconfigPath: "/home/user/.kube/config",
				Namespaces:     []string{"default", "production"},
				LabelSelector:  "app=web",
			},
			expectedEnabled:        true,
			expectedKubeconfigPath: "/home/user/.kube/config",
			expectedNamespaces:     []string{"default", "production"},
			expectedLabelSelector:  "app=web",
		},
		{
			name: "disabled without config",
			dto: &yaml.KubernetesDiscoveryDTO{
				Enabled:        false,
				KubeconfigPath: "",
				Namespaces:     nil,
				LabelSelector:  "",
			},
			expectedEnabled:        false,
			expectedKubeconfigPath: "",
			expectedNamespaces:     nil,
			expectedLabelSelector:  "",
		},
		{
			name: "enabled with in-cluster config",
			dto: &yaml.KubernetesDiscoveryDTO{
				Enabled:        true,
				KubeconfigPath: "",
				Namespaces:     []string{"monitoring"},
				LabelSelector:  "tier=backend",
			},
			expectedEnabled:        true,
			expectedKubeconfigPath: "",
			expectedNamespaces:     []string{"monitoring"},
			expectedLabelSelector:  "tier=backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedKubeconfigPath, result.KubeconfigPath)
			assert.Equal(t, tt.expectedNamespaces, result.Namespaces)
			assert.Equal(t, tt.expectedLabelSelector, result.LabelSelector)
		})
	}
}

// TestNomadDiscoveryDTO_ToDomain tests yaml.NomadDiscoveryDTO to domain conversion.
// It verifies that Nomad discovery configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestNomadDiscoveryDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		dto               *yaml.NomadDiscoveryDTO
		expectedEnabled   bool
		expectedAddress   string
		expectedNamespace string
		expectedJobFilter string
	}{
		{
			name: "enabled with full config",
			dto: &yaml.NomadDiscoveryDTO{
				Enabled:   true,
				Address:   "http://nomad.service.consul:4646",
				Namespace: "production",
				JobFilter: "web-*",
			},
			expectedEnabled:   true,
			expectedAddress:   "http://nomad.service.consul:4646",
			expectedNamespace: "production",
			expectedJobFilter: "web-*",
		},
		{
			name: "disabled without config",
			dto: &yaml.NomadDiscoveryDTO{
				Enabled:   false,
				Address:   "",
				Namespace: "",
				JobFilter: "",
			},
			expectedEnabled:   false,
			expectedAddress:   "",
			expectedNamespace: "",
			expectedJobFilter: "",
		},
		{
			name: "enabled with default namespace",
			dto: &yaml.NomadDiscoveryDTO{
				Enabled:   true,
				Address:   "http://localhost:4646",
				Namespace: "default",
				JobFilter: "",
			},
			expectedEnabled:   true,
			expectedAddress:   "http://localhost:4646",
			expectedNamespace: "default",
			expectedJobFilter: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedAddress, result.Address)
			assert.Equal(t, tt.expectedNamespace, result.Namespace)
			assert.Equal(t, tt.expectedJobFilter, result.JobFilter)
		})
	}
}

// TestPortScanConfigDTO_ToDomain tests yaml.PortScanConfigDTO to domain conversion.
// It verifies that port scan configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestPortScanConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		dto                  *yaml.PortScanConfigDTO
		expectedEnabled      bool
		expectedInterfaces   []string
		expectedExcludePorts []int
		expectedIncludePorts []int
	}{
		{
			name: "enabled with full config",
			dto: &yaml.PortScanConfigDTO{
				Enabled:      true,
				Interfaces:   []string{"eth0", "lo"},
				ExcludePorts: []int{22, 3306},
				IncludePorts: []int{80, 443, 8080},
			},
			expectedEnabled:      true,
			expectedInterfaces:   []string{"eth0", "lo"},
			expectedExcludePorts: []int{22, 3306},
			expectedIncludePorts: []int{80, 443, 8080},
		},
		{
			name: "disabled without config",
			dto: &yaml.PortScanConfigDTO{
				Enabled:      false,
				Interfaces:   nil,
				ExcludePorts: nil,
				IncludePorts: nil,
			},
			expectedEnabled:      false,
			expectedInterfaces:   nil,
			expectedExcludePorts: nil,
			expectedIncludePorts: nil,
		},
		{
			name: "enabled with only exclude ports",
			dto: &yaml.PortScanConfigDTO{
				Enabled:      true,
				Interfaces:   []string{"eth0"},
				ExcludePorts: []int{22, 23, 25},
				IncludePorts: nil,
			},
			expectedEnabled:      true,
			expectedInterfaces:   []string{"eth0"},
			expectedExcludePorts: []int{22, 23, 25},
			expectedIncludePorts: nil,
		},
		{
			name: "enabled with only include ports",
			dto: &yaml.PortScanConfigDTO{
				Enabled:      true,
				Interfaces:   nil,
				ExcludePorts: nil,
				IncludePorts: []int{80, 443},
			},
			expectedEnabled:      true,
			expectedInterfaces:   nil,
			expectedExcludePorts: nil,
			expectedIncludePorts: []int{80, 443},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedEnabled, result.Enabled)
			assert.Equal(t, tt.expectedInterfaces, result.Interfaces)
			assert.Equal(t, tt.expectedExcludePorts, result.ExcludePorts)
			assert.Equal(t, tt.expectedIncludePorts, result.IncludePorts)
		})
	}
}

// TestDuration_UnmarshalYAML tests yaml.Duration YAML unmarshaling.
// It verifies that duration strings are correctly parsed from YAML.
//
// Params:
//   - t: testing context
func TestDuration_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yamlInput      string
		expectedResult time.Duration
		expectError    bool
	}{
		{
			name:           "parse seconds",
			yamlInput:      "30s",
			expectedResult: 30 * time.Second,
			expectError:    false,
		},
		{
			name:           "parse minutes",
			yamlInput:      "5m",
			expectedResult: 5 * time.Minute,
			expectError:    false,
		},
		{
			name:           "parse complex duration",
			yamlInput:      "1h30m",
			expectedResult: 90 * time.Minute,
			expectError:    false,
		},
		{
			name:           "parse milliseconds",
			yamlInput:      "500ms",
			expectedResult: 500 * time.Millisecond,
			expectError:    false,
		},
		{
			name:        "invalid duration string",
			yamlInput:   "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d yaml.Duration
			unmarshal := func(v any) error {
				ptr := v.(*string)
				*ptr = tt.yamlInput
				return nil
			}

			err := d.UnmarshalYAML(unmarshal)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, time.Duration(d))
			}
		})
	}
}

// TestDuration_MarshalText tests yaml.Duration text marshaling.
// It verifies that durations are correctly converted to text format.
//
// Params:
//   - t: testing context
func TestDuration_MarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		duration       yaml.Duration
		expectedResult string
	}{
		{
			name:           "marshal seconds",
			duration:       yaml.Duration(30 * time.Second),
			expectedResult: "30s",
		},
		{
			name:           "marshal minutes",
			duration:       yaml.Duration(5 * time.Minute),
			expectedResult: "5m0s",
		},
		{
			name:           "marshal hour and minutes",
			duration:       yaml.Duration(90 * time.Minute),
			expectedResult: "1h30m0s",
		},
		{
			name:           "marshal zero duration",
			duration:       yaml.Duration(0),
			expectedResult: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.duration.MarshalText()

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, string(result))
		})
	}
}

// TestConfigDTO_ToDomain tests yaml.ConfigDTO to domain conversion.
// It verifies that root configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dto             *yaml.ConfigDTO
		configPath      string
		expectedVersion string
		expectedPath    string
	}{
		{
			name: "full config converts correctly",
			dto: &yaml.ConfigDTO{
				Version: "1.0",
				Logging: yaml.LoggingConfigDTO{
					BaseDir: "/var/log",
				},
				Services: []yaml.ServiceConfigDTO{
					{
						Name:    "test-service",
						Command: "/usr/bin/test",
					},
				},
			},
			configPath:      "/etc/config.yaml",
			expectedVersion: "1.0",
			expectedPath:    "/etc/config.yaml",
		},
		{
			name: "empty config with defaults",
			dto: &yaml.ConfigDTO{
				Version: "2.0",
			},
			configPath:      "/config/daemon.yaml",
			expectedVersion: "2.0",
			expectedPath:    "/config/daemon.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain(tt.configPath)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedVersion, result.Version)
			assert.Equal(t, tt.expectedPath, result.ConfigPath)
		})
	}
}

// TestServiceConfigDTO_ToDomain tests yaml.ServiceConfigDTO to domain conversion.
// It verifies that service configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestServiceConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dto             *yaml.ServiceConfigDTO
		expectedName    string
		expectedCommand string
		expectedOneshot bool
	}{
		{
			name: "full service config",
			dto: &yaml.ServiceConfigDTO{
				Name:             "nginx",
				Command:          "/usr/sbin/nginx",
				Args:             []string{"-g", "daemon off;"},
				User:             "www-data",
				Group:            "www-data",
				WorkingDirectory: "/var/www",
				Environment:      map[string]string{"PORT": "8080"},
				Oneshot:          false,
			},
			expectedName:    "nginx",
			expectedCommand: "/usr/sbin/nginx",
			expectedOneshot: false,
		},
		{
			name: "oneshot service",
			dto: &yaml.ServiceConfigDTO{
				Name:    "init-script",
				Command: "/bin/init.sh",
				Oneshot: true,
			},
			expectedName:    "init-script",
			expectedCommand: "/bin/init.sh",
			expectedOneshot: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedCommand, result.Command)
			assert.Equal(t, tt.expectedOneshot, result.Oneshot)
		})
	}
}

// TestListenerDTO_ToDomain tests yaml.ListenerDTO to domain conversion.
// It verifies that listener configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestListenerDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dto              *yaml.ListenerDTO
		expectedName     string
		expectedPort     int
		expectedProtocol string
		expectedExposed  bool
	}{
		{
			name: "tcp listener with default protocol",
			dto: &yaml.ListenerDTO{
				Name:    "http",
				Port:    8080,
				Address: "0.0.0.0",
				Exposed: true,
			},
			expectedName:     "http",
			expectedPort:     8080,
			expectedProtocol: "tcp",
			expectedExposed:  true,
		},
		{
			name: "udp listener",
			dto: &yaml.ListenerDTO{
				Name:     "dns",
				Port:     53,
				Protocol: "udp",
				Address:  "127.0.0.1",
				Exposed:  false,
			},
			expectedName:     "dns",
			expectedPort:     53,
			expectedProtocol: "udp",
			expectedExposed:  false,
		},
		{
			name: "listener with probe",
			dto: &yaml.ListenerDTO{
				Name:     "api",
				Port:     3000,
				Protocol: "tcp",
				Probe: yaml.ProbeDTO{
					Type: "http",
					Path: "/health",
				},
			},
			expectedName:     "api",
			expectedPort:     3000,
			expectedProtocol: "tcp",
			expectedExposed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedPort, result.Port)
			assert.Equal(t, tt.expectedProtocol, result.Protocol)
			assert.Equal(t, tt.expectedExposed, result.Exposed)
		})
	}
}

// TestProbeDTO_ToDomain tests yaml.ProbeDTO to domain conversion.
// It verifies that probe configuration is correctly mapped with defaults applied.
//
// Params:
//   - t: testing context
func TestProbeDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		dto                *yaml.ProbeDTO
		expectedType       string
		expectedMethod     string
		expectedStatusCode int
	}{
		{
			name: "http probe with defaults",
			dto: &yaml.ProbeDTO{
				Type: "http",
				Path: "/health",
			},
			expectedType:       "http",
			expectedMethod:     "GET",
			expectedStatusCode: 200,
		},
		{
			name: "http probe with custom values",
			dto: &yaml.ProbeDTO{
				Type:       "http",
				Path:       "/status",
				Method:     "POST",
				StatusCode: 201,
			},
			expectedType:       "http",
			expectedMethod:     "POST",
			expectedStatusCode: 201,
		},
		{
			name: "tcp probe",
			dto: &yaml.ProbeDTO{
				Type: "tcp",
			},
			expectedType:       "tcp",
			expectedMethod:     "GET",
			expectedStatusCode: 200,
		},
		{
			name: "exec probe",
			dto: &yaml.ProbeDTO{
				Type:    "exec",
				Command: "/bin/healthcheck",
				Args:    []string{"--fast"},
			},
			expectedType:       "exec",
			expectedMethod:     "GET",
			expectedStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.expectedMethod, result.Method)
			assert.Equal(t, tt.expectedStatusCode, result.StatusCode)
		})
	}
}

// TestRestartConfigDTO_ToDomain tests yaml.RestartConfigDTO to domain conversion.
// It verifies that restart configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestRestartConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dto              *yaml.RestartConfigDTO
		expectedPolicy   string
		expectedMaxRetry int
		expectedHasDelay bool
	}{
		{
			name: "always restart policy",
			dto: &yaml.RestartConfigDTO{
				Policy:     "always",
				MaxRetries: 0,
				Delay:      yaml.Duration(5 * time.Second),
			},
			expectedPolicy:   "always",
			expectedMaxRetry: 0,
			expectedHasDelay: true,
		},
		{
			name: "on-failure with max retries",
			dto: &yaml.RestartConfigDTO{
				Policy:     "on-failure",
				MaxRetries: 3,
				Delay:      yaml.Duration(1 * time.Second),
				DelayMax:   yaml.Duration(30 * time.Second),
			},
			expectedPolicy:   "on-failure",
			expectedMaxRetry: 3,
			expectedHasDelay: true,
		},
		{
			name: "never restart policy",
			dto: &yaml.RestartConfigDTO{
				Policy: "never",
			},
			expectedPolicy:   "never",
			expectedMaxRetry: 0,
			expectedHasDelay: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedPolicy, string(result.Policy))
			assert.Equal(t, tt.expectedMaxRetry, result.MaxRetries)
			if tt.expectedHasDelay {
				assert.True(t, result.Delay.Duration() > 0)
			}
		})
	}
}

// TestHealthCheckDTO_ToDomain tests yaml.HealthCheckDTO to domain conversion.
// It verifies that health check configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestHealthCheckDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dto             *yaml.HealthCheckDTO
		expectedName    string
		expectedType    string
		expectedRetries int
	}{
		{
			name: "http health check",
			dto: &yaml.HealthCheckDTO{
				Name:       "api-health",
				Type:       "http",
				Interval:   yaml.Duration(30 * time.Second),
				Timeout:    yaml.Duration(5 * time.Second),
				Retries:    3,
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
			},
			expectedName:    "api-health",
			expectedType:    "http",
			expectedRetries: 3,
		},
		{
			name: "tcp health check",
			dto: &yaml.HealthCheckDTO{
				Name:     "db-health",
				Type:     "tcp",
				Interval: yaml.Duration(10 * time.Second),
				Timeout:  yaml.Duration(2 * time.Second),
				Retries:  5,
				Host:     "localhost",
				Port:     5432,
			},
			expectedName:    "db-health",
			expectedType:    "tcp",
			expectedRetries: 5,
		},
		{
			name: "command health check",
			dto: &yaml.HealthCheckDTO{
				Type:     "command",
				Interval: yaml.Duration(60 * time.Second),
				Timeout:  yaml.Duration(10 * time.Second),
				Retries:  1,
				Command:  "/bin/check-health.sh",
			},
			expectedName:    "",
			expectedType:    "command",
			expectedRetries: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedType, string(result.Type))
			assert.Equal(t, tt.expectedRetries, result.Retries)
		})
	}
}

// TestLoggingConfigDTO_ToDomain tests yaml.LoggingConfigDTO to domain conversion.
// It verifies that logging configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestLoggingConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dto             *yaml.LoggingConfigDTO
		expectedBaseDir string
	}{
		{
			name: "full logging config",
			dto: &yaml.LoggingConfigDTO{
				BaseDir: "/var/log/daemon",
				Defaults: yaml.LogDefaultsDTO{
					TimestampFormat: "2006-01-02T15:04:05Z07:00",
					Rotation: yaml.RotationConfigDTO{
						MaxSize:  "100MB",
						MaxAge:   "7d",
						MaxFiles: 5,
						Compress: true,
					},
				},
			},
			expectedBaseDir: "/var/log/daemon",
		},
		{
			name: "minimal logging config",
			dto: &yaml.LoggingConfigDTO{
				BaseDir: "/tmp/logs",
			},
			expectedBaseDir: "/tmp/logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedBaseDir, result.BaseDir)
		})
	}
}

// TestDaemonLoggingDTO_ToDomain tests yaml.DaemonLoggingDTO to domain conversion.
// It verifies that daemon logging configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestDaemonLoggingDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dto           *yaml.DaemonLoggingDTO
		expectedCount int
	}{
		{
			name: "multiple writers",
			dto: &yaml.DaemonLoggingDTO{
				Writers: []yaml.WriterConfigDTO{
					{Type: "file", Level: "info"},
					{Type: "json", Level: "debug"},
				},
			},
			expectedCount: 2,
		},
		{
			name: "single writer",
			dto: &yaml.DaemonLoggingDTO{
				Writers: []yaml.WriterConfigDTO{
					{Type: "file", Level: "error"},
				},
			},
			expectedCount: 1,
		},
		{
			name: "no writers",
			dto: &yaml.DaemonLoggingDTO{
				Writers: nil,
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Len(t, result.Writers, tt.expectedCount)
		})
	}
}

// TestWriterConfigDTO_ToDomain tests yaml.WriterConfigDTO to domain conversion.
// It verifies that writer configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestWriterConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dto           *yaml.WriterConfigDTO
		expectedType  string
		expectedLevel string
	}{
		{
			name: "file writer",
			dto: &yaml.WriterConfigDTO{
				Type:  "file",
				Level: "info",
				File: yaml.FileWriterConfigDTO{
					Path: "/var/log/daemon.log",
				},
			},
			expectedType:  "file",
			expectedLevel: "info",
		},
		{
			name: "json writer",
			dto: &yaml.WriterConfigDTO{
				Type:  "json",
				Level: "debug",
				JSON: yaml.JSONWriterConfigDTO{
					Path: "/var/log/daemon.json",
				},
			},
			expectedType:  "json",
			expectedLevel: "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.expectedLevel, result.Level)
		})
	}
}

// TestFileWriterConfigDTO_ToDomain tests yaml.FileWriterConfigDTO to domain conversion.
// It verifies that file writer configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestFileWriterConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dto          *yaml.FileWriterConfigDTO
		expectedPath string
	}{
		{
			name: "with path and rotation",
			dto: &yaml.FileWriterConfigDTO{
				Path: "/var/log/app.log",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "50MB",
					MaxAge:   "30d",
					MaxFiles: 10,
					Compress: true,
				},
			},
			expectedPath: "/var/log/app.log",
		},
		{
			name: "with path only",
			dto: &yaml.FileWriterConfigDTO{
				Path: "/tmp/debug.log",
			},
			expectedPath: "/tmp/debug.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedPath, result.Path)
		})
	}
}

// TestJSONWriterConfigDTO_ToDomain tests yaml.JSONWriterConfigDTO to domain conversion.
// It verifies that JSON writer configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestJSONWriterConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dto          *yaml.JSONWriterConfigDTO
		expectedPath string
	}{
		{
			name: "with path and rotation",
			dto: &yaml.JSONWriterConfigDTO{
				Path: "/var/log/app.json",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "100MB",
					MaxAge:   "14d",
					MaxFiles: 7,
					Compress: false,
				},
			},
			expectedPath: "/var/log/app.json",
		},
		{
			name: "with path only",
			dto: &yaml.JSONWriterConfigDTO{
				Path: "/tmp/events.json",
			},
			expectedPath: "/tmp/events.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedPath, result.Path)
		})
	}
}

// TestLogDefaultsDTO_ToDomain tests yaml.LogDefaultsDTO to domain conversion.
// It verifies that log defaults are correctly mapped.
//
// Params:
//   - t: testing context
func TestLogDefaultsDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		dto                     *yaml.LogDefaultsDTO
		expectedTimestampFormat string
	}{
		{
			name: "with timestamp format and rotation",
			dto: &yaml.LogDefaultsDTO{
				TimestampFormat: "2006-01-02T15:04:05Z07:00",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "100MB",
					MaxAge:   "7d",
					MaxFiles: 5,
					Compress: true,
				},
			},
			expectedTimestampFormat: "2006-01-02T15:04:05Z07:00",
		},
		{
			name: "with simple timestamp",
			dto: &yaml.LogDefaultsDTO{
				TimestampFormat: "15:04:05",
			},
			expectedTimestampFormat: "15:04:05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedTimestampFormat, result.TimestampFormat)
		})
	}
}

// TestRotationConfigDTO_ToDomain tests yaml.RotationConfigDTO to domain conversion.
// It verifies that rotation configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestRotationConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dto              *yaml.RotationConfigDTO
		expectedMaxSize  string
		expectedMaxAge   string
		expectedMaxFiles int
		expectedCompress bool
	}{
		{
			name: "full rotation config",
			dto: &yaml.RotationConfigDTO{
				MaxSize:  "100MB",
				MaxAge:   "30d",
				MaxFiles: 10,
				Compress: true,
			},
			expectedMaxSize:  "100MB",
			expectedMaxAge:   "30d",
			expectedMaxFiles: 10,
			expectedCompress: true,
		},
		{
			name: "minimal rotation config",
			dto: &yaml.RotationConfigDTO{
				MaxSize:  "10MB",
				MaxFiles: 3,
			},
			expectedMaxSize:  "10MB",
			expectedMaxAge:   "",
			expectedMaxFiles: 3,
			expectedCompress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedMaxSize, result.MaxSize)
			assert.Equal(t, tt.expectedMaxAge, result.MaxAge)
			assert.Equal(t, tt.expectedMaxFiles, result.MaxFiles)
			assert.Equal(t, tt.expectedCompress, result.Compress)
		})
	}
}

// TestServiceLoggingDTO_ToDomain tests yaml.ServiceLoggingDTO to domain conversion.
// It verifies that service logging configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestServiceLoggingDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		dto                *yaml.ServiceLoggingDTO
		expectedStdoutFile string
		expectedStderrFile string
	}{
		{
			name: "both stdout and stderr configured",
			dto: &yaml.ServiceLoggingDTO{
				Stdout: yaml.LogStreamConfigDTO{
					File:            "/var/log/app/stdout.log",
					TimestampFormat: "2006-01-02T15:04:05",
				},
				Stderr: yaml.LogStreamConfigDTO{
					File:            "/var/log/app/stderr.log",
					TimestampFormat: "2006-01-02T15:04:05",
				},
			},
			expectedStdoutFile: "/var/log/app/stdout.log",
			expectedStderrFile: "/var/log/app/stderr.log",
		},
		{
			name: "only stdout configured",
			dto: &yaml.ServiceLoggingDTO{
				Stdout: yaml.LogStreamConfigDTO{
					File: "/var/log/app/out.log",
				},
			},
			expectedStdoutFile: "/var/log/app/out.log",
			expectedStderrFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedStdoutFile, result.Stdout.FilePath)
			assert.Equal(t, tt.expectedStderrFile, result.Stderr.FilePath)
		})
	}
}

// TestLogStreamConfigDTO_ToDomain tests yaml.LogStreamConfigDTO to domain conversion.
// It verifies that log stream configuration is correctly mapped.
//
// Params:
//   - t: testing context
func TestLogStreamConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		dto            *yaml.LogStreamConfigDTO
		expectedFile   string
		expectedFormat string
	}{
		{
			name: "full log stream config",
			dto: &yaml.LogStreamConfigDTO{
				File:            "/var/log/service/output.log",
				TimestampFormat: "2006-01-02T15:04:05Z07:00",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "50MB",
					MaxFiles: 5,
				},
			},
			expectedFile:   "/var/log/service/output.log",
			expectedFormat: "2006-01-02T15:04:05Z07:00",
		},
		{
			name: "minimal log stream config",
			dto: &yaml.LogStreamConfigDTO{
				File: "/tmp/output.log",
			},
			expectedFile:   "/tmp/output.log",
			expectedFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.dto.ToDomain()

			assert.Equal(t, tt.expectedFile, result.FilePath)
			assert.Equal(t, tt.expectedFormat, result.Format)
		})
	}
}
