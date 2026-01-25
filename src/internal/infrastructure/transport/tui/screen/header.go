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

// HeaderRenderer renders the header section.
type HeaderRenderer struct {
	theme ansi.Theme
	width int
}

// NewHeaderRenderer creates a header renderer.
func NewHeaderRenderer(width int) *HeaderRenderer {
	return &HeaderRenderer{
		theme: ansi.DefaultTheme(),
		width: width,
	}
}

// SetWidth updates the renderer width.
func (h *HeaderRenderer) SetWidth(width int) {
	h.width = width
}

// Render returns the header for the given snapshot.
func (h *HeaderRenderer) Render(snap *model.Snapshot, showTime bool) string {
	layout := terminal.GetLayout(terminal.Size{Cols: h.width, Rows: 24})

	switch layout {
	case terminal.LayoutCompact:
		return h.renderCompact(snap, showTime)
	case terminal.LayoutNormal:
		return h.renderNormal(snap, showTime)
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		return h.renderWide(snap, showTime)
	}
	return h.renderNormal(snap, showTime)
}

// renderCompact renders a minimal header for small terminals.
func (h *HeaderRenderer) renderCompact(snap *model.Snapshot, showTime bool) string {
	// Logo.
	logo := h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset

	// Version.
	version := h.theme.Accent + "v" + snap.Context.Version + ansi.Reset

	// Time (optional).
	var timeStr string
	if showTime {
		timeStr = snap.Timestamp.Format("15:04:05")
	}

	// Build lines.
	line1 := fmt.Sprintf("%s %s", logo, version)
	if timeStr != "" {
		line1 = fmt.Sprintf("%s %s", logo, timeStr)
	}

	line2 := fmt.Sprintf("%s │ Up %s",
		snap.Context.Hostname,
		widget.FormatDurationShort(snap.Context.Uptime))

	// Box.
	box := widget.NewBox(h.width).
		SetStyle(widget.RoundedBox).
		AddLine("  " + line1).
		AddLine("  " + line2)

	return box.Render()
}

// renderNormal renders a standard header.
func (h *HeaderRenderer) renderNormal(snap *model.Snapshot, showTime bool) string {
	// Logo.
	logo := h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset +
		" " + h.theme.Accent + "v" + snap.Context.Version + ansi.Reset

	// Time.
	var timeStr string
	if showTime {
		timeStr = snap.Timestamp.Format("15:04:05")
	}

	// Context line.
	mode := snap.Context.Mode.String()
	if snap.Context.ContainerRuntime != "" {
		mode += " (" + snap.Context.ContainerRuntime + ")"
	}

	// Build content.
	line1Parts := []string{logo}
	if timeStr != "" {
		padding := h.width - 4 - widget.VisibleLen(logo) - len(timeStr)
		if padding > 0 {
			line1Parts = append(line1Parts, strings.Repeat(" ", padding)+timeStr)
		}
	}
	line1 := strings.Join(line1Parts, "")

	line2 := fmt.Sprintf("%s │ %s │ %s",
		snap.Context.Hostname,
		snap.Context.PrimaryIP,
		mode)

	line3 := fmt.Sprintf("%s %s │ Up %s",
		snap.Context.OS,
		snap.Context.Arch,
		widget.FormatDuration(snap.Context.Uptime))

	box := widget.NewBox(h.width).
		SetStyle(widget.RoundedBox).
		AddLine("  " + line1).
		AddLine("  " + h.theme.Muted + "PID1 Process Supervisor" + ansi.Reset).
		AddLine("  " + line2).
		AddLine("  " + line3)

	return box.Render()
}

// renderWide renders an expanded header for wide terminals.
func (h *HeaderRenderer) renderWide(snap *model.Snapshot, showTime bool) string {
	// Logo with more spacing.
	logo := h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset

	version := h.theme.Accent + "v" + snap.Context.Version + ansi.Reset
	subtitle := h.theme.Muted + "PID1 Process Supervisor" + ansi.Reset

	// Right side info.
	rightParts := []string{
		snap.Context.Hostname,
		snap.Context.PrimaryIP,
	}

	if snap.Context.Mode == model.ModeContainer && snap.Context.ContainerRuntime != "" {
		rightParts = append(rightParts, snap.Context.ContainerRuntime)
	}

	rightInfo := strings.Join(rightParts, " │ ")

	// Time.
	var timeStr string
	if showTime {
		timeStr = snap.Timestamp.Format("15:04:05")
	}

	// System info.
	sysInfo := fmt.Sprintf("%s %s %s │ Up %s",
		snap.Context.OS,
		snap.Context.Kernel,
		snap.Context.Arch,
		widget.FormatDuration(snap.Context.Uptime))

	// Build lines.
	line1 := fmt.Sprintf("%s %s", logo, version)
	line1Right := rightInfo
	if timeStr != "" {
		line1Right = timeStr + "  " + rightInfo
	}

	// Pad line1.
	padding := h.width - 4 - widget.VisibleLen(line1) - widget.VisibleLen(line1Right)
	if padding > 0 {
		line1 = line1 + strings.Repeat(" ", padding) + line1Right
	}

	box := widget.NewBox(h.width).
		SetStyle(widget.RoundedBox).
		AddLine("  " + line1).
		AddLine("  " + subtitle + strings.Repeat(" ", max(0, h.width-4-24-len(sysInfo))) + sysInfo)

	return box.Render()
}

// RenderBrandOnly returns just the brand logo.
func (h *HeaderRenderer) RenderBrandOnly() string {
	return h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset
}

// Helper for max.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
