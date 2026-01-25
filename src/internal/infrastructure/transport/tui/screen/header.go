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
func (h *HeaderRenderer) renderCompact(snap *model.Snapshot, _ bool) string {
	ctx := snap.Context

	// Logo with version.
	logo := h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset

	version := ctx.Version
	if len(version) > 0 && version[0] != 'v' {
		version = "v" + version
	}
	versionStr := h.theme.Accent + version + ansi.Reset

	// Runtime.
	runtime := ctx.Mode.String()
	if ctx.ContainerRuntime != "" {
		runtime = ctx.ContainerRuntime
	}

	line1 := fmt.Sprintf("  %s %s", logo, versionStr)
	line2 := fmt.Sprintf("  %s │ %s │ Up %s",
		ctx.Hostname,
		runtime,
		widget.FormatDurationShort(ctx.Uptime))

	// Box.
	box := widget.NewBox(h.width).
		SetStyle(widget.RoundedBox).
		AddLine(line1).
		AddLine(line2)

	return box.Render()
}

// renderNormal renders a standard header matching raw mode format.
func (h *HeaderRenderer) renderNormal(snap *model.Snapshot, _ bool) string {
	ctx := snap.Context

	// Version string (add 'v' prefix if not present).
	version := ctx.Version
	if len(version) > 0 && version[0] != 'v' {
		version = "v" + version
	}

	// Title line: "   superviz.io ─────────────────────── v0.2.0   "
	logo := h.theme.Primary + "superviz" + ansi.Reset + h.theme.Accent + ".io" + ansi.Reset
	versionStr := h.theme.Accent + version + ansi.Reset

	// Calculate separator.
	innerWidth := h.width - 2
	pad := 3
	logoLen := 11 // "superviz.io" visible
	versionLen := len(version)
	separatorLen := innerWidth - (2 * pad) - logoLen - 2 - versionLen
	if separatorLen < 3 {
		separatorLen = 3
	}
	separator := h.theme.Muted + strings.Repeat("─", separatorLen) + ansi.Reset

	titleLine := strings.Repeat(" ", pad) + logo + " " + separator + " " + versionStr

	// Runtime mode.
	runtime := ctx.Mode.String()
	if ctx.ContainerRuntime != "" {
		runtime = ctx.Mode.String() + " (" + ctx.ContainerRuntime + ")"
	}

	// Platform.
	platform := ctx.OS + "/" + ctx.Arch

	// Config path.
	configPath := ctx.ConfigPath
	if configPath == "" {
		configPath = "/etc/supervizio/config.yaml"
	}

	// Uptime for interactive mode (dynamic).
	uptime := widget.FormatDuration(ctx.Uptime)

	// Bullet point style.
	bullet := h.theme.Accent + "▸" + ansi.Reset

	// Build content lines with visual hierarchy.
	box := widget.NewBox(h.width).
		AddLine("").
		AddLine(titleLine).
		AddLine("").
		AddLine("   " + bullet + " " + h.theme.Muted + "Host" + ansi.Reset + "       " + ctx.Hostname).
		AddLine("   " + bullet + " " + h.theme.Muted + "Platform" + ansi.Reset + "   " + platform).
		AddLine("   " + bullet + " " + h.theme.Muted + "Runtime" + ansi.Reset + "    " + runtime).
		AddLine("   " + bullet + " " + h.theme.Muted + "Config" + ansi.Reset + "     " + configPath).
		AddLine("   " + bullet + " " + h.theme.Muted + "Uptime" + ansi.Reset + "     " + uptime).
		AddLine("")

	return box.Render()
}

// renderWide renders an expanded header for wide terminals (same as normal).
func (h *HeaderRenderer) renderWide(snap *model.Snapshot, _ bool) string {
	// Use same format as normal for consistency.
	return h.renderNormal(snap, false)
}

// RenderBrandOnly returns just the brand logo.
func (h *HeaderRenderer) RenderBrandOnly() string {
	return h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset
}
