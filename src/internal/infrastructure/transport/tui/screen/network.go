// Package screen provides complete screen renderers.
package screen

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// NetworkRenderer renders network interface information.
type NetworkRenderer struct {
	theme ansi.Theme
	width int
}

// NewNetworkRenderer creates a network renderer.
func NewNetworkRenderer(width int) *NetworkRenderer {
	return &NetworkRenderer{
		theme: ansi.DefaultTheme(),
		width: width,
	}
}

// SetWidth updates the renderer width.
func (n *NetworkRenderer) SetWidth(width int) {
	n.width = width
}

// Render returns the network section with progress bars.
func (n *NetworkRenderer) Render(snap *model.Snapshot) string {
	// Filter to interesting interfaces.
	interfaces := n.filterInterfaces(snap.Network)

	if len(interfaces) == 0 {
		return n.renderEmpty()
	}

	layout := terminal.GetLayout(terminal.Size{Cols: n.width, Rows: 24})

	switch layout {
	case terminal.LayoutCompact:
		return n.renderCompact(interfaces)
	case terminal.LayoutNormal, terminal.LayoutWide, terminal.LayoutUltraWide:
		return n.renderWithBars(interfaces)
	}
	return n.renderWithBars(interfaces)
}

// filterInterfaces removes uninteresting interfaces.
func (n *NetworkRenderer) filterInterfaces(ifaces []model.NetworkInterface) []model.NetworkInterface {
	result := make([]model.NetworkInterface, 0, len(ifaces))

	for _, iface := range ifaces {
		// Skip down interfaces.
		if !iface.IsUp {
			continue
		}

		// Include loopback even without explicit IP.
		if iface.IsLoopback {
			if iface.IP == "" {
				iface.IP = "127.0.0.1"
			}
			result = append(result, iface)
			continue
		}

		// Skip non-loopback interfaces without IP.
		if iface.IP == "" {
			continue
		}

		result = append(result, iface)
	}

	return result
}

// renderEmpty renders when there are no interfaces.
func (n *NetworkRenderer) renderEmpty() string {
	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
		AddLine("  " + n.theme.Muted + "No network interfaces" + ansi.Reset)

	return box.Render()
}

// renderCompact renders a minimal network view.
func (n *NetworkRenderer) renderCompact(ifaces []model.NetworkInterface) string {
	lines := make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		ip := iface.IP
		if ip == "" {
			ip = "-"
		}

		// Compact: name, IP, and rates. Use strings.Builder to avoid fmt.Sprintf.
		rx := widget.FormatBytesPerSec(iface.RxBytesPerSec)
		tx := widget.FormatBytesPerSec(iface.TxBytesPerSec)
		var sb strings.Builder
		sb.Grow(50)
		sb.WriteString("  ")
		sb.WriteString(widget.PadRight(iface.Name, 6))
		sb.WriteByte(' ')
		sb.WriteString(widget.PadRight(ip, 15))
		sb.WriteString(" ↓")
		sb.WriteString(rx)
		sb.WriteString(" ↑")
		sb.WriteString(tx)
		lines = append(lines, sb.String())
	}

	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
		AddLines(lines)

	return box.Render()
}

// renderWithBars renders network interfaces with progress bars like CPU/RAM.
// For interfaces without speed info (loopback, virtual), shows throughput only.
func (n *NetworkRenderer) renderWithBars(ifaces []model.NetworkInterface) string {
	// Calculate bar width.
	// Format: "  eth0   ↓[bar] 1.2 MB/s  ↑[bar] 256 KB/s  1 Gbps"
	barWidth := (n.width - 55) / 2
	if barWidth < 8 {
		barWidth = 8
	}

	lines := make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		// Format rates.
		rxRate := widget.FormatBytesPerSec(iface.RxBytesPerSec)
		txRate := widget.FormatBytesPerSec(iface.TxBytesPerSec)

		var line string

		if iface.Speed > 0 {
			// Has speed - show progress bars with percentage.
			rxBps := iface.RxBytesPerSec * 8
			txBps := iface.TxBytesPerSec * 8
			rxPercent := float64(rxBps) / float64(iface.Speed) * 100
			txPercent := float64(txBps) / float64(iface.Speed) * 100
			if rxPercent > 100 {
				rxPercent = 100
			}
			if txPercent > 100 {
				txPercent = 100
			}

			// Create progress bars.
			rxBar := widget.NewProgressBar(barWidth, rxPercent).
				SetLabel("").
				SetColorByPercent()
			rxBar.ShowValue = false

			txBar := widget.NewProgressBar(barWidth, txPercent).
				SetLabel("").
				SetColorByPercent()
			txBar.ShowValue = false

			// Format speed.
			speed := widget.FormatSpeedShort(iface.Speed)

			// Format: "  eth0   ↓[bar] 1.2 MB/s  ↑[bar] 256 KB/s  1 Gbps"
			// Use strings.Builder to avoid fmt.Sprintf allocation.
			var sb strings.Builder
			sb.Grow(100)
			sb.WriteString("  ")
			sb.WriteString(widget.PadRight(iface.Name, 6))
			sb.WriteString(" ↓")
			sb.WriteString(rxBar.Render())
			sb.WriteByte(' ')
			sb.WriteString(widget.PadLeft(rxRate, 8))
			sb.WriteString("  ↑")
			sb.WriteString(txBar.Render())
			sb.WriteByte(' ')
			sb.WriteString(widget.PadLeft(txRate, 8))
			sb.WriteString("  ")
			sb.WriteString(speed)
			line = sb.String()
		} else {
			// No speed info (non-Linux platforms) - show throughput only.
			// Format: "  lo     ↓ 1.2 MB/s  ↑ 256 KB/s  (no limit)"
			// Use strings.Builder to avoid fmt.Sprintf allocation.
			var sb strings.Builder
			sb.Grow(60)
			sb.WriteString("  ")
			sb.WriteString(widget.PadRight(iface.Name, 6))
			sb.WriteString(" ↓ ")
			sb.WriteString(widget.PadLeft(rxRate, 8))
			sb.WriteString("  ↑ ")
			sb.WriteString(widget.PadLeft(txRate, 8))
			sb.WriteString("  ")
			sb.WriteString(n.theme.Muted)
			sb.WriteString("(no limit)")
			sb.WriteString(ansi.Reset)
			line = sb.String()
		}

		lines = append(lines, line)
	}

	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
		AddLines(lines)

	return box.Render()
}

// RenderInline returns a single-line summary.
// Uses string concatenation to avoid fmt.Sprintf allocation.
func (n *NetworkRenderer) RenderInline(snap *model.Snapshot) string {
	// Find primary interface (non-loopback, has IP).
	for _, iface := range snap.Network {
		if !iface.IsLoopback && iface.IP != "" {
			return iface.Name + " " + iface.IP +
				" ↓" + widget.FormatBytesPerSec(iface.RxBytesPerSec) +
				" ↑" + widget.FormatBytesPerSec(iface.TxBytesPerSec)
		}
	}
	return "No network"
}
