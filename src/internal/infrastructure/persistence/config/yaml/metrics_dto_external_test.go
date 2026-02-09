package yaml_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/persistence/config/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goyaml "gopkg.in/yaml.v3"
)

// TestMetricsConfigDTO_ToDomain_MinimalTemplate verifies minimal template.
func TestMetricsConfigDTO_ToDomain_MinimalTemplate(t *testing.T) {
	tests := []struct {
		name     string
		yamlText string
	}{
		{
			name: "minimal template with no overrides",
			yamlText: `
monitoring:
  performance_template: "minimal"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configDTO yaml.ConfigDTO
			err := goyaml.Unmarshal([]byte(tt.yamlText), &configDTO)
			require.NoError(t, err)

			mon := configDTO.Monitoring.ToDomain()

			// Verify minimal template applied
			assert.True(t, mon.Metrics.Enabled)
			assert.True(t, mon.Metrics.CPU.Enabled)
			assert.False(t, mon.Metrics.CPU.Pressure)
			assert.True(t, mon.Metrics.Memory.Enabled)
			assert.False(t, mon.Metrics.Memory.Pressure)
			assert.True(t, mon.Metrics.Load.Enabled)
			assert.False(t, mon.Metrics.Disk.Enabled)
			assert.False(t, mon.Metrics.Network.Enabled)
			assert.False(t, mon.Metrics.Connections.Enabled)
		})
	}
}

// TestMetricsConfigDTO_ToDomain_StandardTemplate verifies standard template.
func TestMetricsConfigDTO_ToDomain_StandardTemplate(t *testing.T) {
	tests := []struct {
		name     string
		yamlText string
	}{
		{
			name: "standard template explicit",
			yamlText: `
monitoring:
  performance_template: "standard"
`,
		},
		{
			name: "no template defaults to standard",
			yamlText: `
monitoring:
  defaults:
    interval: "30s"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configDTO yaml.ConfigDTO
			err := goyaml.Unmarshal([]byte(tt.yamlText), &configDTO)
			require.NoError(t, err)

			mon := configDTO.Monitoring.ToDomain()

			// Verify standard template applied (all enabled)
			assert.True(t, mon.Metrics.Enabled)
			assert.True(t, mon.Metrics.CPU.Enabled)
			assert.True(t, mon.Metrics.Memory.Enabled)
			assert.True(t, mon.Metrics.Disk.Enabled)
			assert.True(t, mon.Metrics.Network.Enabled)
			assert.True(t, mon.Metrics.Connections.Enabled)
		})
	}
}

// TestMetricsConfigDTO_ToDomain_WithOverrides verifies template with overrides.
func TestMetricsConfigDTO_ToDomain_WithOverrides(t *testing.T) {
	tests := []struct {
		name     string
		yamlText string
	}{
		{
			name: "standard template with connections disabled",
			yamlText: `
monitoring:
  performance_template: "standard"
  metrics:
    connections:
      enabled: false
`,
		},
		{
			name: "minimal template with disk enabled",
			yamlText: `
monitoring:
  performance_template: "minimal"
  metrics:
    disk:
      enabled: true
      partitions: true
      usage: true
      io: false
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configDTO yaml.ConfigDTO
			err := goyaml.Unmarshal([]byte(tt.yamlText), &configDTO)
			require.NoError(t, err)

			mon := configDTO.Monitoring.ToDomain()

			if tt.name == "standard template with connections disabled" {
				// Standard base with connections override
				assert.True(t, mon.Metrics.CPU.Enabled)
				assert.True(t, mon.Metrics.Memory.Enabled)
				assert.False(t, mon.Metrics.Connections.Enabled)
			} else {
				// Minimal base with disk override
				assert.True(t, mon.Metrics.CPU.Enabled)
				assert.False(t, mon.Metrics.CPU.Pressure)
				assert.True(t, mon.Metrics.Disk.Enabled)
				assert.True(t, mon.Metrics.Disk.Partitions)
				assert.True(t, mon.Metrics.Disk.Usage)
				assert.False(t, mon.Metrics.Disk.IO)
			}
		})
	}
}

// TestMetricsConfigDTO_ToDomain_GranularControl verifies granular control.
func TestMetricsConfigDTO_ToDomain_GranularControl(t *testing.T) {
	tests := []struct {
		name     string
		yamlText string
	}{
		{
			name: "disable specific connection types",
			yamlText: `
monitoring:
  performance_template: "standard"
  metrics:
    connections:
      tcp_connections: false
      udp_sockets: false
      unix_sockets: true
      listening_ports: true
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configDTO yaml.ConfigDTO
			err := goyaml.Unmarshal([]byte(tt.yamlText), &configDTO)
			require.NoError(t, err)

			mon := configDTO.Monitoring.ToDomain()

			// Verify granular connection settings
			assert.True(t, mon.Metrics.Connections.Enabled)
			assert.False(t, mon.Metrics.Connections.TCPConnections)
			assert.False(t, mon.Metrics.Connections.UDPSockets)
			assert.True(t, mon.Metrics.Connections.UnixSockets)
			assert.True(t, mon.Metrics.Connections.ListeningPorts)
		})
	}
}

// TestMetricsConfigDTO_ToDomain_GlobalDisable verifies global disable.
func TestMetricsConfigDTO_ToDomain_GlobalDisable(t *testing.T) {
	tests := []struct {
		name     string
		yamlText string
	}{
		{
			name: "global metrics disabled",
			yamlText: `
monitoring:
  metrics:
    enabled: false
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configDTO yaml.ConfigDTO
			err := goyaml.Unmarshal([]byte(tt.yamlText), &configDTO)
			require.NoError(t, err)

			mon := configDTO.Monitoring.ToDomain()

			// Verify global disabled
			assert.False(t, mon.Metrics.Enabled)
		})
	}
}

// TestMetricsConfigDTO_ToDomain_BackwardCompatibility verifies backward compatibility.
func TestMetricsConfigDTO_ToDomain_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		yamlText string
	}{
		{
			name: "no metrics section defaults to standard",
			yamlText: `
monitoring:
  defaults:
    interval: "30s"
  targets:
    - name: "test"
      type: "http"
      address: "http://localhost:8080"
      probe:
        type: "http"
        path: "/health"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configDTO yaml.ConfigDTO
			err := goyaml.Unmarshal([]byte(tt.yamlText), &configDTO)
			require.NoError(t, err)

			mon := configDTO.Monitoring.ToDomain()

			// Verify standard template applied by default
			assert.True(t, mon.Metrics.Enabled)
			assert.True(t, mon.Metrics.CPU.Enabled)
			assert.True(t, mon.Metrics.Memory.Enabled)
			assert.True(t, mon.Metrics.Disk.Enabled)
			assert.True(t, mon.Metrics.Network.Enabled)
			assert.True(t, mon.Metrics.Connections.Enabled)

			// Verify other monitoring config preserved
			assert.Len(t, mon.Targets, 1)
			assert.Equal(t, "test", mon.Targets[0].Name)
		})
	}
}
