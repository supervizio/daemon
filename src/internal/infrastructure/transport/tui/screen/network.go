// Package screen provides complete screen renderers.
package screen

import (
	"fmt"
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

// Render returns the network section.
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
		return n.renderNormal(interfaces)
	}
	return n.renderNormal(interfaces)
}

// filterInterfaces removes uninteresting interfaces.
func (n *NetworkRenderer) filterInterfaces(ifaces []model.NetworkInterface) []model.NetworkInterface {
	result := make([]model.NetworkInterface, 0, len(ifaces))

	for _, iface := range ifaces {
		// Skip down interfaces (except loopback).
		if !iface.IsUp && !iface.IsLoopback {
			continue
		}

		// Skip interfaces without IP.
		if iface.IP == "" && !iface.IsLoopback {
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

		// Compact: just name and IP.
		line := fmt.Sprintf("  %-6s %s", iface.Name, ip)
		lines = append(lines, line)
	}

	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
				AddLines(lines)

	return box.Render()
}

// renderNormal renders a standard network table.
func (n *NetworkRenderer) renderNormal(ifaces []model.NetworkInterface) string {
	table := widget.NewTable(n.width-4).
		AddColumn("IFACE", 8, widget.AlignLeft).
		AddFlexColumn("IP", 15, widget.AlignLeft).
		AddColumn("RX/s", 10, widget.AlignRight).
		AddColumn("TX/s", 10, widget.AlignRight).
		AddColumn("SPEED", 8, widget.AlignRight)

	for _, iface := range ifaces {
		ip := iface.IP
		if ip == "" {
			ip = "-"
		}

		rx := "-"
		tx := "-"
		if iface.RxBytesPerSec > 0 || iface.TxBytesPerSec > 0 {
			rx = "↓" + widget.FormatBytesPerSec(iface.RxBytesPerSec)
			tx = "↑" + widget.FormatBytesPerSec(iface.TxBytesPerSec)
		}

		speed := "-"
		if iface.Speed > 0 {
			speed = widget.FormatSpeedShort(iface.Speed)
		}

		table.AddRow(iface.Name, ip, rx, tx, speed)
	}

	lines := strings.Split(table.Render(), "\n")

	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
				AddLines(prefixLines(lines, "  "))

	return box.Render()
}

// RenderInline returns a single-line summary.
func (n *NetworkRenderer) RenderInline(snap *model.Snapshot) string {
	// Find primary interface (non-loopback, has IP).
	for _, iface := range snap.Network {
		if !iface.IsLoopback && iface.IP != "" {
			return fmt.Sprintf("%s %s ↓%s ↑%s",
				iface.Name,
				iface.IP,
				widget.FormatBytesPerSec(iface.RxBytesPerSec),
				widget.FormatBytesPerSec(iface.TxBytesPerSec))
		}
	}
	return "No network"
}
