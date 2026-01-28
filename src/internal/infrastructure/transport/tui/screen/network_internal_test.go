package screen

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestNetworkRenderer_filterInterfaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interfaces []model.NetworkInterface
		wantCount  int
	}{
		{
			name:       "empty_interfaces",
			interfaces: []model.NetworkInterface{},
			wantCount:  0,
		},
		{
			name: "only_loopback",
			interfaces: []model.NetworkInterface{
				{Name: "lo", IsLoopback: true, IsUp: true, IP: ""},
			},
			wantCount: 1,
		},
		{
			name: "interface_with_ip",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", IsLoopback: false, IsUp: true, IP: "192.168.1.1"},
			},
			wantCount: 1,
		},
		{
			name: "interface_down_skipped",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", IsLoopback: false, IsUp: false, IP: "192.168.1.1"},
			},
			wantCount: 0,
		},
		{
			name: "interface_no_ip_skipped",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", IsLoopback: false, IsUp: true, IP: ""},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			result := renderer.filterInterfaces(tt.interfaces)
			assert.Len(t, result, tt.wantCount)
		})
	}
}

func TestNetworkRenderer_renderEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "empty_state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			result := renderer.renderEmpty()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Network")
		})
	}
}

func TestNetworkRenderer_renderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interfaces []model.NetworkInterface
	}{
		{
			name: "single_interface",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", IP: "192.168.1.1", RxBytesPerSec: 1024, TxBytesPerSec: 512},
			},
		},
		{
			name: "multiple_interfaces",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", IP: "192.168.1.1", RxBytesPerSec: 1024, TxBytesPerSec: 512},
				{Name: "lo", IP: "127.0.0.1", RxBytesPerSec: 0, TxBytesPerSec: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			result := renderer.renderCompact(tt.interfaces)
			assert.NotEmpty(t, result)
		})
	}
}

func TestNetworkRenderer_calculateBarWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		want  int
	}{
		{
			name:  "standard_width",
			width: 100,
			want:  22,
		},
		{
			name:  "small_width",
			width: 60,
			want:  networkMinBarWidth,
		},
		{
			name:  "large_width",
			width: 200,
			want:  72,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: tt.width,
			}
			result := renderer.calculateBarWidth()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNetworkRenderer_calculatePercentages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		iface      model.NetworkInterface
		wantRxPct  float64
		wantTxPct  float64
	}{
		{
			name: "zero_speed",
			iface: model.NetworkInterface{
				Speed:          1000000000,
				RxBytesPerSec:  0,
				TxBytesPerSec:  0,
			},
			wantRxPct: 0,
			wantTxPct: 0,
		},
		{
			name: "half_speed",
			iface: model.NetworkInterface{
				Speed:          1000000000,
				RxBytesPerSec:  62500000,
				TxBytesPerSec:  31250000,
			},
			wantRxPct: 50,
			wantTxPct: 25,
		},
		{
			name: "over_capacity_capped",
			iface: model.NetworkInterface{
				Speed:          1000000000,
				RxBytesPerSec:  200000000,
				TxBytesPerSec:  150000000,
			},
			wantRxPct: 100,
			wantTxPct: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			rxPct, txPct := renderer.calculatePercentages(tt.iface)
			assert.InDelta(t, tt.wantRxPct, rxPct, 0.1)
			assert.InDelta(t, tt.wantTxPct, txPct, 0.1)
		})
	}
}

func TestNetworkRenderer_createProgressBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		percent float64
	}{
		{
			name:    "zero_percent",
			width:   20,
			percent: 0,
		},
		{
			name:    "half_percent",
			width:   20,
			percent: 50,
		},
		{
			name:    "full_percent",
			width:   20,
			percent: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			bar := renderer.createProgressBar(tt.width, tt.percent)
			assert.NotNil(t, bar)
			assert.False(t, bar.ShowValue)
		})
	}
}

func TestNetworkRenderer_formatInterfaceWithSpeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		iface  model.NetworkInterface
		rxRate string
		txRate string
	}{
		{
			name: "basic_interface",
			iface: model.NetworkInterface{
				Name:           "eth0",
				Speed:          1000000000,
				RxBytesPerSec:  1024000,
				TxBytesPerSec:  512000,
			},
			rxRate: "1.0 MB/s",
			txRate: "512 KB/s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			result := renderer.formatInterfaceWithSpeed(tt.iface, 20, tt.rxRate, tt.txRate)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, tt.iface.Name)
		})
	}
}

func TestNetworkRenderer_formatInterfaceNoSpeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		iface  model.NetworkInterface
		rxRate string
		txRate string
	}{
		{
			name: "loopback",
			iface: model.NetworkInterface{
				Name:           "lo",
				RxBytesPerSec:  100,
				TxBytesPerSec:  100,
			},
			rxRate: "100 B/s",
			txRate: "100 B/s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			result := renderer.formatInterfaceNoSpeed(tt.iface, tt.rxRate, tt.txRate)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "(no limit)")
		})
	}
}

func TestNetworkRenderer_formatInterfaceLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		iface model.NetworkInterface
	}{
		{
			name: "with_speed",
			iface: model.NetworkInterface{
				Name:           "eth0",
				Speed:          1000000000,
				RxBytesPerSec:  1024,
				TxBytesPerSec:  512,
			},
		},
		{
			name: "without_speed",
			iface: model.NetworkInterface{
				Name:           "lo",
				Speed:          0,
				RxBytesPerSec:  100,
				TxBytesPerSec:  100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			result := renderer.formatInterfaceLine(tt.iface, 20)
			assert.NotEmpty(t, result)
		})
	}
}

func TestNetworkRenderer_renderWithBars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interfaces []model.NetworkInterface
	}{
		{
			name: "single_interface",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", Speed: 1000000000, RxBytesPerSec: 1024, TxBytesPerSec: 512},
			},
		},
		{
			name: "multiple_interfaces",
			interfaces: []model.NetworkInterface{
				{Name: "eth0", Speed: 1000000000, RxBytesPerSec: 1024, TxBytesPerSec: 512},
				{Name: "lo", Speed: 0, RxBytesPerSec: 100, TxBytesPerSec: 100},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &NetworkRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			result := renderer.renderWithBars(tt.interfaces)
			assert.NotEmpty(t, result)
		})
	}
}
