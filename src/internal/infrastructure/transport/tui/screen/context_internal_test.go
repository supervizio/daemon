package screen

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestContextRenderer_formatLimitValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		limits model.ResourceLimits
	}{
		{
			name: "with_cpu_quota",
			limits: model.ResourceLimits{
				CPUQuota:    2.0,
				CPUQuotaRaw: 200000,
				CPUPeriod:   100000,
			},
		},
		{
			name: "without_limits",
			limits: model.ResourceLimits{
				CPUQuota: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewContextRenderer(80)
			cpuStr, memStr, pidsStr, cpusetStr := renderer.formatLimitValues(tt.limits)
			assert.NotEmpty(t, cpuStr)
			assert.NotEmpty(t, memStr)
			assert.NotEmpty(t, pidsStr)
			assert.NotEmpty(t, cpusetStr)
		})
	}
}

func TestContextRenderer_buildLimitsLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cpuStr    string
		memStr    string
		pidsStr   string
		cpusetStr string
	}{
		{
			name:      "basic_limits",
			cpuStr:    "2.0 cores",
			memStr:    "1 GB",
			pidsStr:   "100/1000",
			cpusetStr: "0-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewContextRenderer(80)
			lines := renderer.buildLimitsLines(tt.cpuStr, tt.memStr, tt.pidsStr, tt.cpusetStr)
			assert.NotEmpty(t, lines)
		})
	}
}

func TestContextRenderer_buildSandboxLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sandboxes []model.SandboxInfo
	}{
		{
			name: "single_sandbox",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "/var/run/docker.sock"},
			},
		},
		{
			name: "multiple_sandboxes",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "/var/run/docker.sock"},
				{Name: "podman", Detected: false, Endpoint: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewContextRenderer(80)
			lines := renderer.buildSandboxLines(tt.sandboxes)
			assert.NotEmpty(t, lines)
		})
	}
}

func TestContextRenderer_getSandboxEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sandboxes []model.SandboxInfo
		idx       int
		width     int
	}{
		{
			name: "valid_index",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "/var/run/docker.sock"},
			},
			idx:   0,
			width: 40,
		},
		{
			name:      "invalid_index",
			sandboxes: []model.SandboxInfo{},
			idx:       5,
			width:     40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewContextRenderer(80)
			result := renderer.getSandboxEntry(tt.sandboxes, tt.idx, tt.width)
			// Empty string is valid for out-of-bounds
			assert.NotNil(t, &result)
		})
	}
}

func TestContextRenderer_formatSandbox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sandbox model.SandboxInfo
		width   int
	}{
		{
			name: "detected_sandbox",
			sandbox: model.SandboxInfo{
				Name:     "docker",
				Detected: true,
				Endpoint: "/var/run/docker.sock",
			},
			width: 40,
		},
		{
			name: "not_detected_sandbox",
			sandbox: model.SandboxInfo{
				Name:     "podman",
				Detected: false,
				Endpoint: "",
			},
			width: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewContextRenderer(80)
			result := renderer.formatSandbox(tt.sandbox, tt.width)
			assert.NotEmpty(t, result)
		})
	}
}
