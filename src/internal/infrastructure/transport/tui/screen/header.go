// Package screen provides complete screen renderers.
package screen

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// HeaderRenderer renders the header section.
// Displays application branding, version, and system information.
type HeaderRenderer struct {
	theme ansi.Theme
	width int
}

// NewHeaderRenderer creates a header renderer.
//
// Params:
//   - width: terminal width for rendering
//
// Returns:
//   - *HeaderRenderer: configured renderer instance
func NewHeaderRenderer(width int) *HeaderRenderer {
	// Initialize with default theme.
	return &HeaderRenderer{
		theme: ansi.DefaultTheme(),
		width: width,
	}
}

// SetWidth updates the renderer width.
//
// Params:
//   - width: new terminal width
func (h *HeaderRenderer) SetWidth(width int) {
	h.width = width
}

// Render returns the header for the given snapshot.
//
// Params:
//   - snap: system snapshot
//   - showTime: whether to show timestamp
//
// Returns:
//   - string: rendered header section
func (h *HeaderRenderer) Render(snap *model.Snapshot, showTime bool) string {
	// defaultRows is the default number of terminal rows for layout calculations.
	const defaultRows int = 24

	layout := terminal.GetLayout(terminal.Size{Cols: h.width, Rows: defaultRows})

	// Select rendering mode based on terminal layout.
	switch layout {
	// Compact mode for small terminals.
	case terminal.LayoutCompact:
		// Return minimal header for small screens.
		return h.renderCompact(snap, showTime)
	// Standard mode for normal terminals.
	case terminal.LayoutNormal:
		// Return standard header layout.
		return h.renderNormal(snap, showTime)
	// Wide mode for large terminals.
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		// Return expanded header for wide screens.
		return h.renderWide(snap, showTime)
	}
	// Default to normal rendering.
	return h.renderNormal(snap, showTime)
}

// renderCompact renders a minimal header for small terminals.
//
// Params:
//   - snap: system snapshot
//   - _: showTime (unused, kept for consistency)
//
// Returns:
//   - string: rendered compact header
func (h *HeaderRenderer) renderCompact(snap *model.Snapshot, _ bool) string {
	// compactBuilderSize is the pre-allocated buffer size for compact header rendering.
	const compactBuilderSize int = 64

	ctx := snap.Context

	// Build logo line with version using strings.Builder for efficiency.
	var sb1 strings.Builder
	sb1.Grow(compactBuilderSize)
	sb1.WriteString("  ")
	sb1.WriteString(h.theme.Primary)
	sb1.WriteString("superviz")
	sb1.WriteString(ansi.Reset)
	sb1.WriteString(h.theme.Accent)
	sb1.WriteString(".io")
	sb1.WriteString(ansi.Reset)
	sb1.WriteByte(' ')
	sb1.WriteString(h.theme.Accent)
	// Add 'v' prefix if version doesn't have it.
	if len(ctx.Version) > 0 && ctx.Version[0] != 'v' {
		sb1.WriteByte('v')
	}
	sb1.WriteString(ctx.Version)
	sb1.WriteString(ansi.Reset)
	line1 := sb1.String()

	// Determine runtime display.
	runtime := ctx.Mode.String()
	// Use container runtime name when available.
	if ctx.ContainerRuntime != "" {
		runtime = ctx.ContainerRuntime
	}

	// Build info line with hostname, runtime, and uptime.
	var sb2 strings.Builder
	sb2.Grow(compactBuilderSize)
	sb2.WriteString("  ")
	sb2.WriteString(ctx.Hostname)
	sb2.WriteString(" │ ")
	sb2.WriteString(runtime)
	sb2.WriteString(" │ Up ")
	sb2.WriteString(widget.FormatDurationShort(ctx.Uptime))
	line2 := sb2.String()

	// Render as rounded box.
	box := widget.NewBox(h.width).
		SetStyle(widget.RoundedBox).
		AddLine(line1).
		AddLine(line2)

	// Return rendered compact header.
	return box.Render()
}

// renderNormal renders a standard header matching raw mode format.
//
// Params:
//   - snap: system snapshot
//   - _: showTime (unused, kept for consistency)
//
// Returns:
//   - string: rendered header
func (h *HeaderRenderer) renderNormal(snap *model.Snapshot, _ bool) string {
	const (
		// boxBorder is the total width consumed by box borders.
		boxBorder int = 2
		// padding is the horizontal padding for header content.
		padding int = 3
		// logoLength is the visible character length of the logo text.
		logoLength int = 11
		// minSeparator is the minimum width for the separator line.
		minSeparator int = 3
		// paddingSides is the number of sides with padding.
		paddingSides int = 2
	)

	ctx := snap.Context

	// Add 'v' prefix to version if not present.
	version := ctx.Version
	// Prepend 'v' for consistent version display.
	if len(version) > 0 && version[0] != 'v' {
		version = "v" + version
	}

	// Format logo and version with colors.
	logo := h.theme.Primary + "superviz" + ansi.Reset + h.theme.Accent + ".io" + ansi.Reset
	versionStr := h.theme.Accent + version + ansi.Reset

	// Calculate separator width for title line.
	innerWidth := h.width - boxBorder
	versionLen := len(version)
	separatorLen := innerWidth - (paddingSides * padding) - logoLength - boxBorder - versionLen
	// Ensure minimum separator width for readability.
	if separatorLen < minSeparator {
		separatorLen = minSeparator
	}
	separator := h.theme.Muted + strings.Repeat("─", separatorLen) + ansi.Reset

	titleLine := strings.Repeat(" ", padding) + logo + " " + separator + " " + versionStr

	// Format runtime mode with optional container runtime.
	runtime := ctx.Mode.String()
	// Append container runtime info when available.
	if ctx.ContainerRuntime != "" {
		runtime = ctx.Mode.String() + " (" + ctx.ContainerRuntime + ")"
	}

	// Determine platform string.
	platform := ctx.OS + "/" + ctx.Arch

	// Use config path or default.
	configPath := ctx.ConfigPath
	// Use default config path when not specified.
	if configPath == "" {
		configPath = "/etc/supervizio/config.yaml"
	}

	// Format uptime for interactive mode (dynamic).
	uptime := widget.FormatDuration(ctx.Uptime)

	// Bullet point for list items.
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

	// Return rendered normal header.
	return box.Render()
}

// renderWide renders an expanded header for wide terminals (same as normal).
//
// Params:
//   - snap: system snapshot
//   - _: showTime (unused, kept for consistency)
//
// Returns:
//   - string: rendered header
func (h *HeaderRenderer) renderWide(snap *model.Snapshot, _ bool) string {
	// Use same format as normal for consistency.
	// Return normal header as wide uses same format.
	return h.renderNormal(snap, false)
}

// RenderBrandOnly returns just the brand logo.
//
// Returns:
//   - string: brand logo with colors
func (h *HeaderRenderer) RenderBrandOnly() string {
	// Return brand logo with colors.
	return h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset
}

// NOTE: Tests needed - create header_external_test.go and header_internal_test.go
