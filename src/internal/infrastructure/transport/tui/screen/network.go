// Package screen provides complete screen renderers.
package screen

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

const (
	// networkDefaultRows is the default terminal rows for layout calculation.
	networkDefaultRows int = 24

	// compactBufferSize is the pre-allocation size for compact network line.
	compactBufferSize int = 50

	// interfaceNameWidth is the padding width for interface names.
	interfaceNameWidth int = 6

	// ipAddressWidth is the padding width for IP addresses.
	ipAddressWidth int = 15

	// barCalculationOffset is the offset for calculating bar width.
	barCalculationOffset int = 55

	// barDivider is the divider for splitting bar width between rx/tx.
	barDivider int = 2

	// networkMinBarWidth is the minimum width for network progress bars.
	networkMinBarWidth int = 8

	// wideBufferSize is the pre-allocation size for wide network line.
	wideBufferSize int = 100

	// ratePaddingWidth is the padding width for rate display.
	ratePaddingWidth int = 8

	// throughputBufferSize is the pre-allocation size for throughput-only line.
	throughputBufferSize int = 60

	// percentMax is the maximum percentage value.
	percentMax float64 = 100

	// bitsPerByte is the number of bits in a byte.
	bitsPerByte uint64 = 8
)

// NetworkRenderer renders network interface information.
// It provides formatted network interface display with throughput bars and rate information.
type NetworkRenderer struct {
	theme ansi.Theme
	width int
}

// NewNetworkRenderer creates a network renderer.
//
// Params:
//   - width: terminal width in columns
//
// Returns:
//   - *NetworkRenderer: configured renderer instance
func NewNetworkRenderer(width int) *NetworkRenderer {
	// Return configured network renderer with defaults.
	return &NetworkRenderer{
		theme: ansi.DefaultTheme(),
		width: width,
	}
}

// SetWidth updates the renderer width.
//
// Params:
//   - width: new terminal width in columns
func (n *NetworkRenderer) SetWidth(width int) {
	n.width = width
}

// Render returns the network section with progress bars.
//
// Params:
//   - snap: snapshot containing network interface data
//
// Returns:
//   - string: rendered network section
func (n *NetworkRenderer) Render(snap *model.Snapshot) string {
	// Filter to interesting interfaces.
	interfaces := n.filterInterfaces(snap.Network)

	// Show empty state when no active interfaces.
	if len(interfaces) == 0 {
		// Return empty network box.
		return n.renderEmpty()
	}

	layout := terminal.GetLayout(terminal.Size{Cols: n.width, Rows: networkDefaultRows})

	// Select rendering mode based on terminal layout.
	switch layout {
	// Compact mode for small terminals.
	case terminal.LayoutCompact:
		// Return minimal network view.
		return n.renderCompact(interfaces)
	// Normal and wide modes with progress bars.
	case terminal.LayoutNormal, terminal.LayoutWide, terminal.LayoutUltraWide:
		// Return network view with throughput bars.
		return n.renderWithBars(interfaces)
	// handle default case.
	default:
		// Default to full network display.
		return n.renderWithBars(interfaces)
	}
}

// filterInterfaces removes uninteresting interfaces.
//
// Params:
//   - ifaces: list of network interfaces to filter
//
// Returns:
//   - []model.NetworkInterface: filtered list of active interfaces
func (n *NetworkRenderer) filterInterfaces(ifaces []model.NetworkInterface) []model.NetworkInterface {
	result := make([]model.NetworkInterface, 0, len(ifaces))

	// Iterate through all interfaces to filter active ones.
	for _, iface := range ifaces {
		// Skip down interfaces.
		// Exclude interfaces that are not up.
		if !iface.IsUp {
			continue
		}

		// Include loopback even without explicit IP.
		// Handle loopback interface specially.
		if iface.IsLoopback {
			// Assign default loopback IP if missing.
			if iface.IP == "" {
				iface.IP = "127.0.0.1"
			}
			result = append(result, iface)
			continue
		}

		// Skip non-loopback interfaces without IP.
		// Exclude interfaces without assigned IP address.
		if iface.IP == "" {
			continue
		}

		result = append(result, iface)
	}

	// Return filtered interface list.
	return result
}

// renderEmpty renders when there are no interfaces.
//
// Returns:
//   - string: rendered empty network box
func (n *NetworkRenderer) renderEmpty() string {
	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
		AddLine("  " + n.theme.Muted + "No network interfaces" + ansi.Reset)

	// Return rendered empty network box.
	return box.Render()
}

// renderCompact renders a minimal network view.
//
// Params:
//   - ifaces: list of network interfaces to render
//
// Returns:
//   - string: rendered compact network view
func (n *NetworkRenderer) renderCompact(ifaces []model.NetworkInterface) string {
	lines := make([]string, 0, len(ifaces))

	// Iterate through interfaces to format compact view.
	for _, iface := range ifaces {
		ip := iface.IP
		// Show placeholder when IP is not available.
		if ip == "" {
			ip = "-"
		}

		// Compact: name, IP, and rates. Use strings.Builder to avoid fmt.Sprintf.
		rx := widget.FormatBytesPerSec(iface.RxBytesPerSec)
		tx := widget.FormatBytesPerSec(iface.TxBytesPerSec)
		var sb strings.Builder
		sb.Grow(compactBufferSize)
		sb.WriteString("  ")
		sb.WriteString(widget.PadRight(iface.Name, interfaceNameWidth))
		sb.WriteByte(' ')
		sb.WriteString(widget.PadRight(ip, ipAddressWidth))
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

	// Return rendered compact network box.
	return box.Render()
}

// renderWithBars renders network interfaces with progress bars like CPU/RAM.
// For interfaces without speed info (loopback, virtual), shows throughput only.
//
// Params:
//   - ifaces: list of network interfaces to render
//
// Returns:
//   - string: rendered network view with throughput bars
func (n *NetworkRenderer) renderWithBars(ifaces []model.NetworkInterface) string {
	barWidth := n.calculateBarWidth()
	lines := make([]string, 0, len(ifaces))

	// Iterate through interfaces to format with progress bars.
	for _, iface := range ifaces {
		line := n.formatInterfaceLine(iface, barWidth)
		lines = append(lines, line)
	}

	box := widget.NewBox(n.width).
		SetTitle("Network").
		SetTitleColor(n.theme.Header).
		AddLines(lines)

	// Return rendered network box with bars.
	return box.Render()
}

// calculateBarWidth computes the progress bar width based on terminal width.
//
// Returns:
//   - int: width for progress bars
func (n *NetworkRenderer) calculateBarWidth() int {
	barWidth := (n.width - barCalculationOffset) / barDivider
	// Ensure minimum bar width for readability.
	if barWidth < networkMinBarWidth {
		// return computed result.
		return networkMinBarWidth
	}
	// return computed result.
	return barWidth
}

// formatInterfaceLine formats a single interface line with or without bars.
//
// Params:
//   - iface: network interface to format
//   - barWidth: width for progress bars
//
// Returns:
//   - string: formatted interface line
func (n *NetworkRenderer) formatInterfaceLine(iface model.NetworkInterface, barWidth int) string {
	rxRate := widget.FormatBytesPerSec(iface.RxBytesPerSec)
	txRate := widget.FormatBytesPerSec(iface.TxBytesPerSec)

	// Show progress bars when interface has speed info.
	if iface.Speed > 0 {
		// return computed result.
		return n.formatInterfaceWithSpeed(iface, barWidth, rxRate, txRate)
	}
	// return computed result.
	return n.formatInterfaceNoSpeed(iface, rxRate, txRate)
}

// formatInterfaceWithSpeed formats interface line with progress bars.
//
// Params:
//   - iface: network interface with speed info
//   - barWidth: width for progress bars
//   - rxRate: formatted receive rate string
//   - txRate: formatted transmit rate string
//
// Returns:
//   - string: formatted interface line with bars
func (n *NetworkRenderer) formatInterfaceWithSpeed(iface model.NetworkInterface, barWidth int, rxRate, txRate string) string {
	rxPercent, txPercent := n.calculatePercentages(iface)
	rxBar := n.createProgressBar(barWidth, rxPercent)
	txBar := n.createProgressBar(barWidth, txPercent)
	speed := widget.FormatSpeedShort(iface.Speed)

	var sb strings.Builder
	sb.Grow(wideBufferSize)
	sb.WriteString("  ")
	sb.WriteString(widget.PadRight(iface.Name, interfaceNameWidth))
	sb.WriteString(" ↓")
	sb.WriteString(rxBar.Render())
	sb.WriteByte(' ')
	sb.WriteString(widget.PadLeft(rxRate, ratePaddingWidth))
	sb.WriteString("  ↑")
	sb.WriteString(txBar.Render())
	sb.WriteByte(' ')
	sb.WriteString(widget.PadLeft(txRate, ratePaddingWidth))
	sb.WriteString("  ")
	sb.WriteString(speed)
	// return computed result.
	return sb.String()
}

// calculatePercentages computes RX/TX percentages capped at 100%.
//
// Params:
//   - iface: network interface with speed info
//
// Returns:
//   - float64: RX percentage (0-100)
//   - float64: TX percentage (0-100)
func (n *NetworkRenderer) calculatePercentages(iface model.NetworkInterface) (rxPercent, txPercent float64) {
	rxBps := iface.RxBytesPerSec * bitsPerByte
	txBps := iface.TxBytesPerSec * bitsPerByte
	rxPercent = float64(rxBps) / float64(iface.Speed) * percentMax
	txPercent = float64(txBps) / float64(iface.Speed) * percentMax
	// Cap percentages at 100%.
	if rxPercent > percentMax {
		rxPercent = percentMax
	}
	// evaluate condition.
	if txPercent > percentMax {
		txPercent = percentMax
	}
	// return computed result.
	return rxPercent, txPercent
}

// createProgressBar creates a progress bar without label or value display.
//
// Params:
//   - width: bar width
//   - percent: fill percentage
//
// Returns:
//   - *widget.ProgressBar: configured progress bar
func (n *NetworkRenderer) createProgressBar(width int, percent float64) *widget.ProgressBar {
	bar := widget.NewProgressBar(width, percent).
		SetLabel("").
		SetColorByPercent()
	bar.ShowValue = false
	// return computed result.
	return bar
}

// formatInterfaceNoSpeed formats interface line without progress bars.
//
// Params:
//   - iface: network interface without speed info
//   - rxRate: formatted receive rate string
//   - txRate: formatted transmit rate string
//
// Returns:
//   - string: formatted interface line with throughput only
func (n *NetworkRenderer) formatInterfaceNoSpeed(iface model.NetworkInterface, rxRate, txRate string) string {
	var sb strings.Builder
	sb.Grow(throughputBufferSize)
	sb.WriteString("  ")
	sb.WriteString(widget.PadRight(iface.Name, interfaceNameWidth))
	sb.WriteString(" ↓ ")
	sb.WriteString(widget.PadLeft(rxRate, ratePaddingWidth))
	sb.WriteString("  ↑ ")
	sb.WriteString(widget.PadLeft(txRate, ratePaddingWidth))
	sb.WriteString("  ")
	sb.WriteString(n.theme.Muted)
	sb.WriteString("(no limit)")
	sb.WriteString(ansi.Reset)
	// return computed result.
	return sb.String()
}

// RenderInline returns a single-line summary.
// Uses string concatenation to avoid fmt.Sprintf allocation.
//
// Params:
//   - snap: snapshot containing network interface data
//
// Returns:
//   - string: single-line network summary
func (n *NetworkRenderer) RenderInline(snap *model.Snapshot) string {
	// Find primary interface (non-loopback, has IP).
	// Search for first non-loopback interface with IP.
	for _, iface := range snap.Network {
		// Use first valid non-loopback interface.
		if !iface.IsLoopback && iface.IP != "" {
			// Return formatted interface summary.
			return iface.Name + " " + iface.IP +
				" ↓" + widget.FormatBytesPerSec(iface.RxBytesPerSec) +
				" ↑" + widget.FormatBytesPerSec(iface.TxBytesPerSec)
		}
	}
	// Return placeholder when no network available.
	return "No network"
}
