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
