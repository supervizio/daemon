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
// Note: showTime parameter was removed as it was reserved for future use but unused.
//
// Params:
//   - snap: system snapshot
//
// Returns:
//   - string: rendered header section
func (h *HeaderRenderer) Render(snap *model.Snapshot) string {
	// defaultRows is the default number of terminal rows for layout calculations.
	const defaultRows int = 24

	layout := terminal.GetLayout(terminal.Size{Cols: h.width, Rows: defaultRows})

	switch layout {
	case terminal.LayoutCompact:
		return h.renderCompact(snap)
	case terminal.LayoutNormal:
		return h.renderNormal(snap)
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		return h.renderWide(snap)
	}
	return h.renderNormal(snap)
}

// renderCompact renders a minimal header for small terminals.
//
// Params:
//   - snap: system snapshot
//
// Returns:
//   - string: rendered compact header
func (h *HeaderRenderer) renderCompact(snap *model.Snapshot) string {
	ctx := snap.Context

	line1 := h.buildCompactLogoLine(ctx.Version)
	line2 := h.buildCompactInfoLine(ctx)

	box := widget.NewBox(h.width).
		SetStyle(widget.RoundedBox).
		AddLine(line1).
		AddLine(line2)

	return box.Render()
}

// buildCompactLogoLine builds the logo line for compact header.
//
// Params:
//   - version: application version string.
//
// Returns:
//   - string: formatted logo line.
func (h *HeaderRenderer) buildCompactLogoLine(version string) string {
	// compactBuilderSize is the initial capacity for the string builder.
	const compactBuilderSize int = 64

	var sb strings.Builder
	sb.Grow(compactBuilderSize)
	sb.WriteString("  ")
	sb.WriteString(h.theme.Primary)
	sb.WriteString("superviz")
	sb.WriteString(ansi.Reset)
	sb.WriteString(h.theme.Accent)
	sb.WriteString(".io")
	sb.WriteString(ansi.Reset)
	sb.WriteByte(' ')
	sb.WriteString(h.theme.Accent)
	// Add 'v' prefix if version doesn't have it.
	if len(version) > 0 && version[0] != 'v' {
		sb.WriteByte('v')
	}
	sb.WriteString(version)
	sb.WriteString(ansi.Reset)

	return sb.String()
}

// buildCompactInfoLine builds the info line for compact header.
//
// Params:
//   - ctx: context information.
//
// Returns:
//   - string: formatted info line.
func (h *HeaderRenderer) buildCompactInfoLine(ctx model.RuntimeContext) string {
	// compactBuilderSize is the initial capacity for the string builder.
	const compactBuilderSize int = 64

	runtime := ctx.Mode.String()
	if ctx.ContainerRuntime != "" {
		runtime = ctx.ContainerRuntime
	}

	var sb strings.Builder
	sb.Grow(compactBuilderSize)
	sb.WriteString("  ")
	sb.WriteString(ctx.Hostname)
	sb.WriteString(" │ ")
	sb.WriteString(runtime)
	sb.WriteString(" │ Up ")
	sb.WriteString(widget.FormatDurationShort(ctx.Uptime))

	return sb.String()
}

// renderNormal renders a standard header matching raw mode format.
//
// Params:
//   - snap: system snapshot
//
// Returns:
//   - string: rendered header
func (h *HeaderRenderer) renderNormal(snap *model.Snapshot) string {
	ctx := snap.Context

	titleLine := h.buildNormalTitleLine(ctx.Version)
	contentLines := h.buildNormalContentLines(ctx)

	box := widget.NewBox(h.width).
		AddLine("").
		AddLine(titleLine).
		AddLine("").
		AddLines(contentLines).
		AddLine("")

	return box.Render()
}

// buildNormalTitleLine builds the title line for normal header.
//
// Params:
//   - version: application version string.
//
// Returns:
//   - string: formatted title line.
func (h *HeaderRenderer) buildNormalTitleLine(version string) string {
	const (
		// boxBorder is the width of box borders (left + right).
		boxBorder int = 2
		// padding is the number of spaces for content padding.
		padding int = 3
		// logoLength is the visible character length of the logo.
		logoLength int = 11
		// minSeparator is the minimum separator line length.
		minSeparator int = 3
		// paddingSides is the multiplier for calculating side padding.
		paddingSides int = 2
	)

	if len(version) > 0 && version[0] != 'v' {
		version = "v" + version
	}

	logo := h.theme.Primary + "superviz" + ansi.Reset + h.theme.Accent + ".io" + ansi.Reset
	versionStr := h.theme.Accent + version + ansi.Reset

	innerWidth := h.width - boxBorder
	separatorLen := max(innerWidth-(paddingSides*padding)-logoLength-boxBorder-len(version), minSeparator)
	separator := h.theme.Muted + strings.Repeat("─", separatorLen) + ansi.Reset

	return strings.Repeat(" ", padding) + logo + " " + separator + " " + versionStr
}

// buildNormalContentLines builds the content lines for normal header.
//
// Params:
//   - ctx: context information.
//
// Returns:
//   - []string: content lines for header.
func (h *HeaderRenderer) buildNormalContentLines(ctx model.RuntimeContext) []string {
	runtime := ctx.Mode.String()
	if ctx.ContainerRuntime != "" {
		runtime = ctx.Mode.String() + " (" + ctx.ContainerRuntime + ")"
	}

	platform := ctx.OS + "/" + ctx.Arch
	configPath := ctx.ConfigPath
	if configPath == "" {
		configPath = "/etc/supervizio/config.yaml"
	}

	bullet := h.theme.Accent + "▸" + ansi.Reset

	return []string{
		"   " + bullet + " " + h.theme.Muted + "Host" + ansi.Reset + "       " + ctx.Hostname,
		"   " + bullet + " " + h.theme.Muted + "Platform" + ansi.Reset + "   " + platform,
		"   " + bullet + " " + h.theme.Muted + "Runtime" + ansi.Reset + "    " + runtime,
		"   " + bullet + " " + h.theme.Muted + "Config" + ansi.Reset + "     " + configPath,
		"   " + bullet + " " + h.theme.Muted + "Uptime" + ansi.Reset + "     " + widget.FormatDuration(ctx.Uptime),
	}
}

// renderWide renders an expanded header for wide terminals (same as normal).
//
// Params:
//   - snap: system snapshot
//
// Returns:
//   - string: rendered header
func (h *HeaderRenderer) renderWide(snap *model.Snapshot) string {
	// Wide mode uses the same format as normal for consistency.
	return h.renderNormal(snap)
}

// RenderBrandOnly returns just the brand logo.
//
// Returns:
//   - string: brand logo with colors
func (h *HeaderRenderer) RenderBrandOnly() string {
	return h.theme.Primary + "superviz" + ansi.Reset +
		h.theme.Accent + ".io" + ansi.Reset
}

// NOTE: Tests needed - create header_external_test.go and header_internal_test.go
