// Package tui provides terminal user interface for superviz.io.
package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// Standard terminal width (80 cols baseline).
const standardWidth = 80

// RawRenderer renders a static MOTD snapshot.
type RawRenderer struct {
	width  int
	height int
	out    io.Writer
	theme  ansi.Theme
}

// NewRawRenderer creates a raw renderer.
func NewRawRenderer(out io.Writer) *RawRenderer {
	size := terminal.GetSize()
	return &RawRenderer{
		width:  size.Cols,
		height: size.Rows,
		out:    out,
		theme:  ansi.DefaultTheme(),
	}
}

// SetSize updates the renderer dimensions.
func (r *RawRenderer) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// Render outputs a complete MOTD snapshot for raw mode.
// Shows only startup-time static information, no dynamic data.
func (r *RawRenderer) Render(snap *model.Snapshot) error {
	var sb strings.Builder

	// Header (simplified - version, timestamp, host, config).
	sb.WriteString(r.renderHeader(snap))
	sb.WriteString("\n")

	// For wide terminals (>= 2x standard width), use side-by-side layout.
	if r.width >= standardWidth*2 {
		sb.WriteString(r.renderSystemAndSandboxesSideBySide(snap))
	} else {
		// Stack vertically for narrower terminals.
		sb.WriteString(r.renderSystemSection(snap))
		sb.WriteString("\n")
		if r.hasAnySandboxes(snap) {
			sb.WriteString(r.renderSandboxes(snap))
			sb.WriteString("\n")
		}
	}

	// Services (names only in columns).
	services := screen.NewServicesRenderer(r.width)
	sb.WriteString(services.RenderNamesOnly(snap))
	sb.WriteString("\n")

	// Output.
	_, err := fmt.Fprint(r.out, sb.String())
	return err
}

// RenderCompact outputs a condensed MOTD for small terminals.
func (r *RawRenderer) RenderCompact(snap *model.Snapshot) error {
	var sb strings.Builder

	// Minimal header.
	sb.WriteString(r.renderHeaderCompact(snap))
	sb.WriteString("\n")

	// Services (names only).
	services := screen.NewServicesRenderer(r.width)
	sb.WriteString(services.RenderNamesOnly(snap))
	sb.WriteString("\n")

	// Output.
	_, err := fmt.Fprint(r.out, sb.String())
	return err
}

// renderHeader renders the simplified header for raw mode.
func (r *RawRenderer) renderHeader(snap *model.Snapshot) string {
	ctx := snap.Context

	// Version string (remove leading 'v' if present to avoid "vv").
	version := ctx.Version
	if len(version) > 0 && version[0] != 'v' {
		version = "v" + version
	}

	// Title line: "   superviz.io ─────────────────────────── v0.2.0   "
	logo := r.theme.Primary + "superviz" + ansi.Reset + r.theme.Accent + ".io" + ansi.Reset
	versionStr := r.theme.Accent + version + ansi.Reset

	// Calculate dimensions for symmetric layout.
	innerWidth := r.width - 2 // Box inner width (excluding borders)
	pad := 3                  // Padding on each side
	logoLen := 11             // "superviz.io" = 8 + 3 = 11 visible chars
	versionLen := len(version)

	// Calculate separator to fill space between logo and version.
	// We want: [pad] logo [sp] separator [sp] version [pad]
	// Content visible length should be < innerWidth to let box pad the right side.
	// Content = pad + logoLen + 1 + sepLen + 1 + versionLen (no right pad, box adds it)
	// We want box to add exactly 'pad' spaces on the right.
	// So: content visible = innerWidth - pad
	// sepLen = innerWidth - pad - pad - logoLen - 1 - 1 - versionLen
	separatorLen := innerWidth - (2 * pad) - logoLen - 2 - versionLen
	if separatorLen < 3 {
		separatorLen = 3
	}
	separator := r.theme.Muted + strings.Repeat("─", separatorLen) + ansi.Reset

	// Build title line (box will auto-pad to add right padding).
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

	// Bullet point style.
	bullet := r.theme.Accent + "▸" + ansi.Reset

	// Build content lines with visual hierarchy.
	box := widget.NewBox(r.width).
		AddLine("").
		AddLine(titleLine).
		AddLine("").
		AddLine("   " + bullet + " " + r.theme.Muted + "Host" + ansi.Reset + "       " + ctx.Hostname).
		AddLine("   " + bullet + " " + r.theme.Muted + "Platform" + ansi.Reset + "   " + platform).
		AddLine("   " + bullet + " " + r.theme.Muted + "Runtime" + ansi.Reset + "    " + runtime).
		AddLine("   " + bullet + " " + r.theme.Muted + "Config" + ansi.Reset + "     " + configPath).
		AddLine("   " + bullet + " " + r.theme.Muted + "Started" + ansi.Reset + "    " + ctx.StartTime.Format("2006-01-02T15:04:05Z")).
		AddLine("")

	return box.Render()
}

// renderHeaderCompact renders a minimal header.
func (r *RawRenderer) renderHeaderCompact(snap *model.Snapshot) string {
	ctx := snap.Context

	logo := r.theme.Primary + "superviz" + ansi.Reset +
		r.theme.Accent + ".io" + ansi.Reset +
		" " + r.theme.Accent + "v" + ctx.Version + ansi.Reset

	mode := ctx.Mode.String()
	if ctx.ContainerRuntime != "" {
		mode = ctx.ContainerRuntime
	}

	line := fmt.Sprintf("%s │ %s │ %s", ctx.Hostname, mode, ctx.StartTime.Format("15:04:05"))

	box := widget.NewBox(r.width).
		SetStyle(widget.RoundedBox).
		AddLine("  " + logo).
		AddLine("  " + line)

	return box.Render()
}

// renderSystemSection renders the System (at start) section.
func (r *RawRenderer) renderSystemSection(snap *model.Snapshot) string {
	system := screen.NewSystemRenderer(r.width)
	return system.RenderForRaw(snap)
}

// renderSystemAndSandboxesSideBySide renders System and Sandboxes side by side.
func (r *RawRenderer) renderSystemAndSandboxesSideBySide(snap *model.Snapshot) string {
	// Calculate widths: each panel gets half minus 1 for gap.
	halfWidth := (r.width - 1) / 2

	// Build content for both boxes to determine heights.
	systemLines := r.buildSystemContentLines(snap, halfWidth)
	sandboxLines := r.buildSandboxContentLines(snap, halfWidth)

	// Equalize heights by padding shorter content.
	maxContent := len(systemLines)
	if len(sandboxLines) > maxContent {
		maxContent = len(sandboxLines)
	}
	for len(systemLines) < maxContent {
		systemLines = append(systemLines, "")
	}
	for len(sandboxLines) < maxContent {
		sandboxLines = append(sandboxLines, "")
	}

	// Render both boxes with equalized content.
	systemBox := widget.NewBox(halfWidth).
		SetTitle("System (at start)").
		SetTitleColor(r.theme.Header).
		AddLines(systemLines)

	sandboxBox := widget.NewBox(halfWidth).
		SetTitle("Sandboxes").
		SetTitleColor(r.theme.Header).
		AddLines(sandboxLines)

	// Merge side by side.
	return mergeSideBySide(systemBox.Render(), sandboxBox.Render(), " ")
}

// hasAnySandboxes checks if there are any sandboxes to display.
func (r *RawRenderer) hasAnySandboxes(snap *model.Snapshot) bool {
	return len(snap.Sandboxes) > 0
}

// buildSystemContentLines builds the content lines for system section.
func (r *RawRenderer) buildSystemContentLines(snap *model.Snapshot, width int) []string {
	sys := snap.System
	limits := snap.Limits

	barWidth := width - 40
	if barWidth < 10 {
		barWidth = 10
	}

	// CPU bar.
	cpuBar := widget.NewProgressBar(barWidth, sys.CPUPercent).
		SetLabel("CPU ").
		SetColorByPercent()

	// RAM bar.
	ramBar := widget.NewProgressBar(barWidth, sys.MemoryPercent).
		SetLabel("RAM ").
		SetColorByPercent()

	// Swap bar.
	swapBar := widget.NewProgressBar(barWidth, sys.SwapPercent).
		SetLabel("Swap").
		SetColorByPercent()

	// Disk bar.
	diskBar := widget.NewProgressBar(barWidth, sys.DiskPercent).
		SetLabel("Disk").
		SetColorByPercent()

	// Build load average string without fmt.Sprintf.
	loadAvgStr := "  Load: " + strconv.FormatFloat(sys.LoadAvg1, 'f', 2, 64) + " " + strconv.FormatFloat(sys.LoadAvg5, 'f', 2, 64)

	lines := []string{
		"  " + cpuBar.Render() + loadAvgStr,
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render() + "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
		"  " + diskBar.Render() + "  " + widget.FormatBytes(sys.DiskUsed) + " / " + widget.FormatBytes(sys.DiskTotal),
	}

	// Limits if present.
	if limits.HasLimits {
		// Pre-allocate for max 3 limit parts.
		limitParts := make([]string, 0, 3)
		if limits.CPUSet != "" {
			limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
		}
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, "CPU: "+strconv.FormatFloat(limits.CPUQuota, 'f', 1, 64)+" cores")
		}
		if limits.MemoryMax > 0 {
			limitParts = append(limitParts, "MEM: "+widget.FormatBytes(limits.MemoryMax))
		}
		if len(limitParts) > 0 {
			lines = append(lines, "  "+r.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
		}
	}

	return lines
}

// buildSandboxContentLines builds the content lines for sandboxes section.
// Pre-allocates slice capacity for efficiency.
func (r *RawRenderer) buildSandboxContentLines(snap *model.Snapshot, width int) []string {
	status := widget.NewStatusIndicator()
	// Pre-allocate for all sandboxes.
	lines := make([]string, 0, len(snap.Sandboxes))

	maxEndpoint := width - 20
	detectedIcon := status.Detected(true)
	notDetectedIcon := status.Detected(false)

	// Show detected sandboxes first.
	for _, sb := range snap.Sandboxes {
		if sb.Detected {
			endpoint := sb.Endpoint
			if len(endpoint) > maxEndpoint && maxEndpoint > 3 {
				endpoint = endpoint[:maxEndpoint-3] + "..."
			}
			lines = append(lines, r.formatSandboxLine(detectedIcon, sb.Name, endpoint))
		}
	}

	// Then show not detected ones.
	notDetectedText := r.theme.Muted + "not detected" + ansi.Reset
	for _, sb := range snap.Sandboxes {
		if !sb.Detected {
			lines = append(lines, r.formatSandboxLine(notDetectedIcon, sb.Name, notDetectedText))
		}
	}

	return lines
}

// formatSandboxLine formats a sandbox line with strings.Builder.
func (r *RawRenderer) formatSandboxLine(icon, name, endpoint string) string {
	var sb strings.Builder
	sb.Grow(32 + len(endpoint))
	sb.WriteString("  ")
	sb.WriteString(icon)
	sb.WriteByte(' ')
	sb.WriteString(name)
	// Pad name to 12 chars.
	for i := len(name); i < 12; i++ {
		sb.WriteByte(' ')
	}
	sb.WriteByte(' ')
	sb.WriteString(endpoint)
	return sb.String()
}

// renderSandboxes renders the sandboxes section.
func (r *RawRenderer) renderSandboxes(snap *model.Snapshot) string {
	return r.renderSandboxesWithWidth(snap, r.width)
}

// renderSandboxesWithWidth renders sandboxes with a specific width.
func (r *RawRenderer) renderSandboxesWithWidth(snap *model.Snapshot, width int) string {
	// Reuse buildSandboxContentLines for consistency and efficiency.
	lines := r.buildSandboxContentLines(snap, width)

	box := widget.NewBox(width).
		SetTitle("Sandboxes").
		SetTitleColor(r.theme.Header).
		AddLines(lines)

	return box.Render()
}

// mergeSideBySide merges two rendered boxes horizontally.
func mergeSideBySide(left, right, separator string) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	// Determine the width of the left panel (from visible characters).
	leftWidth := 0
	for _, line := range leftLines {
		w := widget.VisibleLen(line)
		if w > leftWidth {
			leftWidth = w
		}
	}

	// Remove trailing empty lines.
	for len(leftLines) > 0 && strings.TrimSpace(leftLines[len(leftLines)-1]) == "" {
		leftLines = leftLines[:len(leftLines)-1]
	}
	for len(rightLines) > 0 && strings.TrimSpace(rightLines[len(rightLines)-1]) == "" {
		rightLines = rightLines[:len(rightLines)-1]
	}

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var result strings.Builder
	for i := range maxLines {
		// Get left line, pad to width.
		leftLine := ""
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}

		// Pad left line to consistent width.
		visLen := widget.VisibleLen(leftLine)
		if visLen < leftWidth {
			leftLine += strings.Repeat(" ", leftWidth-visLen)
		}

		// Get right line.
		rightLine := ""
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}

		result.WriteString(leftLine)
		result.WriteString(separator)
		result.WriteString(rightLine)
		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
