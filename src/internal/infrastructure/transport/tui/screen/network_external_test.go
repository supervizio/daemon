package screen_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/stretchr/testify/assert"
)

func TestNewNetworkRenderer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard_width",
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(tt.width)
			assert.NotNil(t, renderer)
		})
	}
}

func TestNetworkRenderer_SetWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		initial  int
		newWidth int
	}{
		{
			name:     "update_width",
			initial:  80,
			newWidth: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(tt.initial)
			renderer.SetWidth(tt.newWidth)
			assert.NotNil(t, renderer)
		})
	}
}

func TestNetworkRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_render",
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(80)
			result := renderer.Render(tt.snap)
			assert.NotNil(t, &result)
		})
	}
}

func TestNetworkRenderer_RenderInline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
		want string
	}{
		{
			name: "with_interface",
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{
					{
						Name:          "eth0",
						IsLoopback:    false,
						IsUp:          true,
						IP:            "192.168.1.1",
						RxBytesPerSec: 1024,
						TxBytesPerSec: 512,
					},
				},
			},
			want: "eth0",
		},
		{
			name: "only_loopback",
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{
					{
						Name:       "lo",
						IsLoopback: true,
						IsUp:       true,
						IP:         "127.0.0.1",
					},
				},
			},
			want: "No network",
		},
		{
			name: "no_interfaces",
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{},
			},
			want: "No network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(80)
			result := renderer.RenderInline(tt.snap)
			assert.Contains(t, result, tt.want)
		})
	}
}

func TestNetworkRenderer_Render_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "empty_network",
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(80)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Network")
		})
	}
}

func TestNetworkRenderer_Render_Compact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "compact_layout",
			width: 60,
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{
					{
						Name:          "eth0",
						IsLoopback:    false,
						IsUp:          true,
						IP:            "192.168.1.1",
						RxBytesPerSec: 1024,
						TxBytesPerSec: 512,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(tt.width)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestNetworkRenderer_Render_WithBars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "normal_layout_with_speed",
			width: 100,
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{
					{
						Name:          "eth0",
						IsLoopback:    false,
						IsUp:          true,
						IP:            "192.168.1.1",
						Speed:         1000000000,
						RxBytesPerSec: 1024000,
						TxBytesPerSec: 512000,
					},
				},
			},
		},
		{
			name:  "wide_layout_without_speed",
			width: 180,
			snap: &model.Snapshot{
				Network: []model.NetworkInterface{
					{
						Name:          "lo",
						IsLoopback:    true,
						IsUp:          true,
						IP:            "127.0.0.1",
						Speed:         0,
						RxBytesPerSec: 100,
						TxBytesPerSec: 100,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewNetworkRenderer(tt.width)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}
