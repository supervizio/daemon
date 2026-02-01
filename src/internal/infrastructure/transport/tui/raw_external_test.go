// Package tui_test provides external tests for the TUI package.
package tui_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestSnapshot creates a comprehensive snapshot for testing.
func createTestSnapshot() *model.Snapshot {
	snap := model.NewSnapshot()
	snap.Context = model.RuntimeContext{
		Hostname:         "production-server",
		Version:          "2.5.0",
		OS:               "linux",
		Arch:             "amd64",
		Kernel:           "6.5.0",
		Mode:             model.ModeContainer,
		ContainerRuntime: "docker",
		DaemonPID:        1,
		StartTime:        time.Date(2024, 1, 20, 14, 30, 0, 0, time.UTC),
		Uptime:           5 * time.Hour,
		PrimaryIP:        "172.17.0.2",
		ConfigPath:       "/etc/supervizio/config.yaml",
	}

	snap.System = model.SystemMetrics{
		CPUPercent:    65.5,
		LoadAvg1:      3.2,
		LoadAvg5:      2.8,
		MemoryTotal:   32 * 1024 * 1024 * 1024,
		MemoryUsed:    20 * 1024 * 1024 * 1024,
		MemoryPercent: 62.5,
		SwapTotal:     8 * 1024 * 1024 * 1024,
		SwapUsed:      2 * 1024 * 1024 * 1024,
		SwapPercent:   25.0,
		DiskTotal:     1024 * 1024 * 1024 * 1024,
		DiskUsed:      512 * 1024 * 1024 * 1024,
		DiskPercent:   50.0,
		DiskPath:      "/",
	}

	snap.Limits = model.ResourceLimits{
		HasLimits:   true,
		CPUQuota:    4.0,
		CPUSet:      "0-7",
		MemoryMax:   16 * 1024 * 1024 * 1024,
		CPUQuotaRaw: 400000,
		CPUPeriod:   100000,
	}

	snap.Services = []model.ServiceSnapshot{
		{
			Name:            "web-api",
			State:           process.StateRunning,
			PID:             1234,
			Uptime:          2 * time.Hour,
			RestartCount:    0,
			Health:          health.StatusHealthy,
			HasHealthChecks: true,
			HealthLatency:   15 * time.Millisecond,
			Ports:           []int{8080, 8443},
			CPUPercent:      25.5,
			MemoryRSS:       512 * 1024 * 1024,
			MemoryPercent:   1.6,
		},
		{
			Name:            "worker",
			State:           process.StateRunning,
			PID:             1235,
			Uptime:          2 * time.Hour,
			RestartCount:    1,
			Health:          health.StatusHealthy,
			HasHealthChecks: false,
			CPUPercent:      10.0,
			MemoryRSS:       256 * 1024 * 1024,
			MemoryPercent:   0.8,
		},
		{
			Name:          "failed-service",
			State:         process.StateFailed,
			PID:           0,
			RestartCount:  3,
			LastExitCode:  1,
			LastError:     "connection refused",
			Health:        health.StatusUnhealthy,
			CPUPercent:    0,
			MemoryRSS:     0,
			MemoryPercent: 0,
		},
	}

	snap.Sandboxes = []model.SandboxInfo{
		{
			Name:     "docker",
			Detected: true,
			Endpoint: "unix:///var/run/docker.sock",
			Version:  "24.0.7",
		},
		{
			Name:     "podman",
			Detected: false,
			Endpoint: "",
			Version:  "",
		},
		{
			Name:     "kubernetes",
			Detected: true,
			Endpoint: "https://kubernetes.default.svc:443",
			Version:  "v1.28.0",
		},
	}

	snap.Network = []model.NetworkInterface{
		{
			Name:          "eth0",
			IP:            "172.17.0.2",
			RxBytesPerSec: 1024 * 1024,
			TxBytesPerSec: 512 * 1024,
			Speed:         1000000000,
			IsUp:          true,
			IsLoopback:    false,
		},
		{
			Name:       "lo",
			IP:         "127.0.0.1",
			IsUp:       true,
			IsLoopback: true,
		},
	}

	return snap
}

// TestNewRawRenderer tests the RawRenderer constructor.
func TestNewRawRenderer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates_renderer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			r := tui.NewRawRenderer(&buf)

			require.NotNil(t, r)
		})
	}
}

// TestRawRenderer_SetSize tests terminal size configuration.
func TestRawRenderer_SetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "standard_size",
			width:  80,
			height: 24,
		},
		{
			name:   "wide_terminal",
			width:  160,
			height: 40,
		},
		{
			name:   "narrow_terminal",
			width:  60,
			height: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			r := tui.NewRawRenderer(&buf)
			r.SetSize(tt.width, tt.height)

			// Test that SetSize doesn't panic or error.
			snap := createTestSnapshot()
			err := r.Render(snap)

			assert.NoError(t, err)
		})
	}
}

// TestRawRenderer_Render tests full snapshot rendering.
func TestRawRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		height     int
		modify     func(*model.Snapshot)
		wantSubstr []string
		wantErr    bool
	}{
		{
			name:   "full_render_standard_width",
			width:  80,
			height: 24,
			modify: nil,
			wantSubstr: []string{
				"superviz",
				".io",
				"production-server",
				"linux/amd64",
				"docker",
				"CPU",
				"RAM",
				"Swap",
				"Disk",
				"Limits",
				"web-api",
				"worker",
				"failed-service",
			},
			wantErr: false,
		},
		{
			name:   "wide_terminal_side_by_side",
			width:  160,
			height: 40,
			modify: nil,
			wantSubstr: []string{
				"superviz",
				"System (at start)",
				"Sandboxes",
				"docker",
				"kubernetes",
			},
			wantErr: false,
		},
		{
			name:   "narrow_terminal",
			width:  60,
			height: 20,
			modify: nil,
			wantSubstr: []string{
				"superviz",
				"web-api",
			},
			wantErr: false,
		},
		{
			name:   "with_limits",
			width:  80,
			height: 24,
			modify: nil,
			wantSubstr: []string{
				"Limits",
				"CPUSet: 0-7",
				"CPU: 4 cores",
			},
			wantErr: false,
		},
		{
			name:   "without_limits",
			width:  80,
			height: 24,
			modify: func(s *model.Snapshot) {
				s.Limits.HasLimits = false
			},
			wantSubstr: []string{
				"superviz",
				"web-api",
			},
			wantErr: false,
		},
		{
			name:   "no_sandboxes",
			width:  80,
			height: 24,
			modify: func(s *model.Snapshot) {
				s.Sandboxes = []model.SandboxInfo{}
			},
			wantSubstr: []string{
				"superviz",
				"web-api",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := createTestSnapshot()
			if tt.modify != nil {
				tt.modify(snap)
			}

			var buf bytes.Buffer
			r := tui.NewRawRenderer(&buf)
			r.SetSize(tt.width, tt.height)

			err := r.Render(snap)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				output := buf.String()
				require.NotEmpty(t, output)

				for _, substr := range tt.wantSubstr {
					assert.Contains(t, output, substr, "expected substring not found in output")
				}
			}
		})
	}
}

// TestRawRenderer_RenderCompact tests compact rendering.
func TestRawRenderer_RenderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		height     int
		wantSubstr []string
		wantErr    bool
	}{
		{
			name:   "compact_render",
			width:  80,
			height: 24,
			wantSubstr: []string{
				"superviz",
				".io",
				"production-server",
				"web-api",
				"worker",
			},
			wantErr: false,
		},
		{
			name:   "narrow_compact",
			width:  60,
			height: 20,
			wantSubstr: []string{
				"superviz",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := createTestSnapshot()

			var buf bytes.Buffer
			r := tui.NewRawRenderer(&buf)
			r.SetSize(tt.width, tt.height)

			err := r.RenderCompact(snap)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				output := buf.String()
				require.NotEmpty(t, output)

				for _, substr := range tt.wantSubstr {
					assert.Contains(t, output, substr)
				}

				// Compact should be shorter than full render.
				var fullBuf bytes.Buffer
				fullRenderer := tui.NewRawRenderer(&fullBuf)
				fullRenderer.SetSize(tt.width, tt.height)
				err := fullRenderer.Render(snap)
				require.NoError(t, err)

				assert.Less(t, len(output), len(fullBuf.String()), "compact output should be shorter than full")
			}
		})
	}
}

// TestRawRenderer_Render_EmptySnapshot tests rendering with minimal data.
func TestRawRenderer_Render_EmptySnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "minimal_snapshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := model.NewSnapshot()
			snap.Context = model.RuntimeContext{
				Hostname: "minimal",
				Version:  "1.0.0",
				OS:       "linux",
				Arch:     "amd64",
			}

			var buf bytes.Buffer
			r := tui.NewRawRenderer(&buf)
			r.SetSize(80, 24)

			err := r.Render(snap)

			require.NoError(t, err)
			output := buf.String()
			require.NotEmpty(t, output)
			assert.Contains(t, output, "superviz")
			assert.Contains(t, output, "minimal")
		})
	}
}

// TestRawRenderer_Render_DifferentWidths tests responsive layout.
func TestRawRenderer_Render_DifferentWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{name: "40_cols", width: 40},
		{name: "60_cols", width: 60},
		{name: "80_cols", width: 80},
		{name: "100_cols", width: 100},
		{name: "120_cols", width: 120},
		{name: "160_cols", width: 160},
		{name: "200_cols", width: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := createTestSnapshot()

			var buf bytes.Buffer
			r := tui.NewRawRenderer(&buf)
			r.SetSize(tt.width, 24)

			err := r.Render(snap)

			require.NoError(t, err)
			output := buf.String()
			require.NotEmpty(t, output)

			// Just verify we got valid output.
			assert.Contains(t, output, "superviz")
		})
	}
}

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	// Simple ANSI stripper for testing.
	result := strings.Builder{}
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
