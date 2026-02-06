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

// Layout constants for terminal rendering.
const (
	standardWidth       int = 80 // Standard terminal width (80 cols baseline).
	wideModeMultiplier  int = 2  // Multiplier for side-by-side layout detection.
	borderWidth         int = 2  // Box border width (left + right).
	headerPadding       int = 3  // Padding on each side of header.
	logoVisibleLength   int = 11 // Visible length of "superviz.io" logo.
	minSeparatorLength  int = 3  // Minimum separator length in header.
	minBarWidth         int = 10 // Minimum progress bar width.
	barReservedSpace    int = 40 // Space reserved for labels/values in system section.
	endpointReserved    int = 20 // Space reserved for sandbox endpoint prefix.
	sandboxNameWidth    int = 12 // Fixed width for sandbox name column.
	ellipsisLength      int = 3  // Length of "..." for truncation.
	maxLimitParts       int = 3  // Maximum number of limit parts (CPU, CPUSet, MEM).
	stringBuilderGrowth int = 32 // Initial capacity for strings.Builder.
	floatPrecision      int = 2  // Decimal precision for load average.
	floatBitSize        int = 64 // Bit size for float formatting.
	cpuQuotaPrecision   int = 1  // Decimal precision for CPU quota.
)

// RawRenderer renders a static MOTD snapshot.
// It provides a non-interactive terminal UI for displaying system information at startup.
type RawRenderer struct {
	width  int
	height int
	out    io.Writer
	theme  ansi.Theme
}

// NewRawRenderer creates a raw renderer.
//
// Params:
//   - out: writer for output.
//
// Returns:
//   - *RawRenderer: new raw renderer instance.
func NewRawRenderer(out io.Writer) *RawRenderer {
	size := terminal.GetSize()
	// Return renderer with detected terminal size.
	return &RawRenderer{
		width:  size.Cols,
		height: size.Rows,
		out:    out,
		theme:  ansi.DefaultTheme(),
	}
}

// SetSize updates the renderer dimensions.
//
// Params:
//   - width: new width in columns
//   - height: new height in rows
func (r *RawRenderer) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// Render outputs a complete MOTD snapshot for raw mode.
// Shows only startup-time static information, no dynamic data.
//
// Params:
//   - snap: snapshot containing system and service data.
//
// Returns:
//   - error: write error if output fails.
func (r *RawRenderer) Render(snap *model.Snapshot) error {
	var sb strings.Builder

	// Header (simplified - version, timestamp, host, config).
	sb.WriteString(r.renderHeader(snap))
	sb.WriteString("\n")

	// Check if terminal is wide enough for side-by-side layout.
	if r.width >= standardWidth*wideModeMultiplier {
		sb.WriteString(r.renderSystemAndSandboxesSideBySide(snap))
	} else {
		// Stack vertically for narrower terminals.
		sb.WriteString(r.renderSystemSection(snap))
		sb.WriteString("\n")
		// Only show sandboxes section if any sandboxes exist.
		if r.hasAnySandboxes(snap) {
			sb.WriteString(r.renderSandboxes(snap))
			sb.WriteString("\n")
		}
	}

	// Services (names only in columns).
	services := screen.NewServicesRenderer(r.width)
	sb.WriteString(services.RenderNamesOnly(snap))
	sb.WriteString("\n")

	// Write final output to writer.
	_, err := fmt.Fprint(r.out, sb.String())

	// Return write error if any.
	return err
}

// RenderCompact outputs a condensed MOTD for small terminals.
//
// Params:
//   - snap: snapshot containing system and service data.
//
// Returns:
//   - error: write error if output fails.
func (r *RawRenderer) RenderCompact(snap *model.Snapshot) error {
	var sb strings.Builder

	// Minimal header.
	sb.WriteString(r.renderHeaderCompact(snap))
	sb.WriteString("\n")

	// Services (names only).
	services := screen.NewServicesRenderer(r.width)
	sb.WriteString(services.RenderNamesOnly(snap))
	sb.WriteString("\n")

	// Write final output to writer.
	_, err := fmt.Fprint(r.out, sb.String())

	// Return write error if any.
	return err
}

// renderHeader renders the simplified header for raw mode.
//
// Params:
//   - snap: snapshot containing context data.
//
// Returns:
//   - string: rendered header content.
func (r *RawRenderer) renderHeader(snap *model.Snapshot) string {
	ctx := snap.Context

	// Build title line.
	titleLine := r.buildTitleLine(ctx.Version)

	// Build content lines.
	contentLines := r.buildHeaderContentLines(ctx)

	// Build content lines with visual hierarchy.
	box := widget.NewBox(r.width).
		AddLine("").
		AddLine(titleLine).
		AddLine("").
		AddLines(contentLines).
		AddLine("")

	// Render box to string and return.
	return box.Render()
}

// buildTitleLine builds the header title line with logo and version.
//
// Params:
//   - version: application version string.
//
// Returns:
//   - string: formatted title line.
func (r *RawRenderer) buildTitleLine(version string) string {
	// Ensure version starts with 'v' prefix.
	if version != "" && version[0] != 'v' {
		version = "v" + version
	}

	logo := r.theme.Primary + "superviz" + ansi.Reset + r.theme.Accent + ".io" + ansi.Reset
	versionStr := r.theme.Accent + version + ansi.Reset

	// Calculate separator length.
	innerWidth := r.width - borderWidth
	separatorLen := innerWidth - (wideModeMultiplier * headerPadding) - logoVisibleLength - wideModeMultiplier - len(version)
	separatorLen = max(separatorLen, minSeparatorLength)
	separator := r.theme.Muted + strings.Repeat("─", separatorLen) + ansi.Reset

	// Return formatted title line.
	return strings.Repeat(" ", headerPadding) + logo + " " + separator + " " + versionStr
}

// buildHeaderContentLines builds the header content lines.
//
// Params:
//   - ctx: context data for the header.
//
// Returns:
//   - []string: content lines for the header.
func (r *RawRenderer) buildHeaderContentLines(ctx model.RuntimeContext) []string {
	// Runtime mode.
	runtime := ctx.Mode.String()
	// Include container runtime if available.
	if ctx.ContainerRuntime != "" {
		runtime = ctx.Mode.String() + " (" + ctx.ContainerRuntime + ")"
	}

	// Platform and config.
	platform := ctx.OS + "/" + ctx.Arch
	configPath := ctx.ConfigPath
	// Use default path if not specified.
	if configPath == "" {
		configPath = "/etc/supervizio/config.yaml"
	}

	// Bullet point style.
	bullet := r.theme.Accent + "▸" + ansi.Reset

	// Return formatted content lines.
	return []string{
		"   " + bullet + " " + r.theme.Muted + "Host" + ansi.Reset + "       " + ctx.Hostname,
		"   " + bullet + " " + r.theme.Muted + "Platform" + ansi.Reset + "   " + platform,
		"   " + bullet + " " + r.theme.Muted + "Runtime" + ansi.Reset + "    " + runtime,
		"   " + bullet + " " + r.theme.Muted + "Config" + ansi.Reset + "     " + configPath,
		"   " + bullet + " " + r.theme.Muted + "Started" + ansi.Reset + "    " + ctx.StartTime.Format("2006-01-02T15:04:05Z"),
	}
}

// renderHeaderCompact renders a minimal header.
//
// Params:
//   - snap: snapshot containing context data.
//
// Returns:
//   - string: rendered compact header content.
func (r *RawRenderer) renderHeaderCompact(snap *model.Snapshot) string {
	ctx := snap.Context

	logo := r.theme.Primary + "superviz" + ansi.Reset +
		r.theme.Accent + ".io" + ansi.Reset +
		" " + r.theme.Accent + "v" + ctx.Version + ansi.Reset

	mode := ctx.Mode.String()
	// Use container runtime name if available.
	if ctx.ContainerRuntime != "" {
		mode = ctx.ContainerRuntime
	}

	line := fmt.Sprintf("%s │ %s │ %s", ctx.Hostname, mode, ctx.StartTime.Format("15:04:05"))

	box := widget.NewBox(r.width).
		SetStyle(widget.RoundedBox).
		AddLine("  " + logo).
		AddLine("  " + line)

	// Render box to string and return.
	return box.Render()
}

// renderSystemSection renders the System (at start) section.
//
// Params:
//   - snap: snapshot containing system data.
//
// Returns:
//   - string: rendered system section content.
func (r *RawRenderer) renderSystemSection(snap *model.Snapshot) string {
	system := screen.NewSystemRenderer(r.width)
	// Delegate rendering to system renderer and return.
	return system.RenderForRaw(snap)
}

// renderSystemAndSandboxesSideBySide renders System and Sandboxes side by side.
//
// Params:
//   - snap: snapshot containing system and sandbox data.
//
// Returns:
//   - string: rendered side-by-side layout content.
func (r *RawRenderer) renderSystemAndSandboxesSideBySide(snap *model.Snapshot) string {
	// Calculate widths: each panel gets half minus 1 for gap.
	halfWidth := (r.width - 1) / wideModeMultiplier

	// Build content for both boxes to determine heights.
	systemLines := r.buildSystemContentLines(snap, halfWidth)
	sandboxLines := r.buildSandboxContentLines(snap, halfWidth)

	// Equalize heights by padding shorter content.
	maxContent := max(len(systemLines), len(sandboxLines))
	// Pad system lines to match maximum height.
	for len(systemLines) < maxContent {
		systemLines = append(systemLines, "")
	}
	// Pad sandbox lines to match maximum height.
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

	// Merge side by side and return.
	return mergeSideBySide(systemBox.Render(), sandboxBox.Render(), " ")
}

// hasAnySandboxes checks if there are any sandboxes to display.
//
// Params:
//   - snap: snapshot containing sandbox data.
//
// Returns:
//   - bool: true if sandboxes exist, false otherwise.
func (r *RawRenderer) hasAnySandboxes(snap *model.Snapshot) bool {
	// Check if sandboxes slice is non-empty and return result.
	return len(snap.Sandboxes) > 0
}

// buildSystemContentLines builds the content lines for system section.
//
// Params:
//   - snap: snapshot containing system data.
//   - width: available width for rendering.
//
// Returns:
//   - []string: system content lines for display.
func (r *RawRenderer) buildSystemContentLines(snap *model.Snapshot, width int) []string {
	sys := snap.System

	barWidth := max(width-barReservedSpace, minBarWidth)

	// Build metric lines with progress bars.
	lines := r.buildSystemMetricLines(sys, barWidth)

	// Add limits if present.
	if snap.Limits.HasLimits {
		limitsLine := r.buildLimitsLine(snap.Limits)
		// Append limits line if not empty.
		if limitsLine != "" {
			lines = append(lines, limitsLine)
		}
	}

	// Return assembled system content lines.
	return lines
}

// buildSystemMetricLines builds the system metric progress bar lines.
//
// Params:
//   - sys: system metrics data.
//   - barWidth: width for progress bars.
//
// Returns:
//   - []string: metric lines with progress bars.
func (r *RawRenderer) buildSystemMetricLines(sys model.SystemMetrics, barWidth int) []string {
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

	// Build load average string.
	loadAvgStr := "  Load: " + strconv.FormatFloat(sys.LoadAvg1, 'f', floatPrecision, floatBitSize) +
		" " + strconv.FormatFloat(sys.LoadAvg5, 'f', floatPrecision, floatBitSize)

	// Return metric lines.
	return []string{
		"  " + cpuBar.Render() + loadAvgStr,
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render() + "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
		"  " + diskBar.Render() + "  " + widget.FormatBytes(sys.DiskUsed) + " / " + widget.FormatBytes(sys.DiskTotal),
	}
}

// buildLimitsLine builds the limits line if any limits are set.
//
// Params:
//   - limits: resource limits data.
//
// Returns:
//   - string: formatted limits line or empty string.
func (r *RawRenderer) buildLimitsLine(limits model.ResourceLimits) string {
	// Pre-allocate for max 3 limit parts.
	limitParts := make([]string, 0, maxLimitParts)

	// Add CPUSet limit if configured.
	if limits.CPUSet != "" {
		limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
	}
	// Add CPU quota limit if configured.
	if limits.CPUQuota > 0 {
		limitParts = append(limitParts, "CPU: "+strconv.FormatFloat(limits.CPUQuota, 'f', cpuQuotaPrecision, floatBitSize)+" cores")
	}
	// Add memory limit if configured.
	if limits.MemoryMax > 0 {
		limitParts = append(limitParts, "MEM: "+widget.FormatBytes(limits.MemoryMax))
	}

	// Return empty if no limits.
	if len(limitParts) == 0 {
		// No limits to display.
		return ""
	}

	// Return formatted limits line.
	return "  " + r.theme.Muted + "Limits: " + strings.Join(limitParts, " │ ") + ansi.Reset
}

// buildSandboxContentLines builds the content lines for sandboxes section.
// Pre-allocates slice capacity for efficiency.
//
// Params:
//   - snap: snapshot containing sandbox data.
//   - width: available width for rendering.
//
// Returns:
//   - []string: sandbox content lines for display.
//
// NOTE: Tests needed (KTN-TEST-SPLIT).
func (r *RawRenderer) buildSandboxContentLines(snap *model.Snapshot, width int) []string {
	status := widget.NewStatusIndicator()
	// Pre-allocate for all sandboxes.
	lines := make([]string, 0, len(snap.Sandboxes))

	maxEndpoint := width - endpointReserved
	detectedIcon := status.Detected(true)
	notDetectedIcon := status.Detected(false)

	// Show detected sandboxes first.
	for _, sb := range snap.Sandboxes {
		// Only show detected sandboxes in first pass.
		if sb.Detected {
			endpoint := sb.Endpoint
			// Truncate endpoint if too long.
			if len(endpoint) > maxEndpoint && maxEndpoint > ellipsisLength {
				endpoint = endpoint[:maxEndpoint-ellipsisLength] + "..."
			}
			lines = append(lines, r.formatSandboxLine(detectedIcon, sb.Name, endpoint))
		}
	}

	// Then show not detected ones.
	notDetectedText := r.theme.Muted + "not detected" + ansi.Reset
	// Show not detected sandboxes in second pass.
	for _, sb := range snap.Sandboxes {
		// Only show not detected sandboxes in second pass.
		if !sb.Detected {
			lines = append(lines, r.formatSandboxLine(notDetectedIcon, sb.Name, notDetectedText))
		}
	}

	// Return assembled sandbox content lines.
	return lines
}

// formatSandboxLine formats a sandbox line with strings.Builder.
//
// Params:
//   - icon: status indicator icon.
//   - name: sandbox name.
//   - endpoint: sandbox endpoint or status text.
//
// Returns:
//   - string: formatted sandbox line.
func (r *RawRenderer) formatSandboxLine(icon, name, endpoint string) string {
	var sb strings.Builder
	sb.Grow(stringBuilderGrowth + len(endpoint))
	sb.WriteString("  ")
	sb.WriteString(icon)
	sb.WriteByte(' ')
	sb.WriteString(name)
	// Pad name to 12 chars.
	for i := len(name); i < sandboxNameWidth; i++ {
		sb.WriteByte(' ')
	}
	sb.WriteByte(' ')
	sb.WriteString(endpoint)
	// Return formatted sandbox line.
	return sb.String()
}

// renderSandboxes renders the sandboxes section.
//
// Params:
//   - snap: snapshot containing sandbox data.
//
// Returns:
//   - string: rendered sandboxes section content.
func (r *RawRenderer) renderSandboxes(snap *model.Snapshot) string {
	// Delegate to width-specific renderer and return.
	return r.renderSandboxesWithWidth(snap, r.width)
}

// renderSandboxesWithWidth renders sandboxes with a specific width.
//
// Params:
//   - snap: snapshot containing sandbox data.
//   - width: available width for rendering.
//
// Returns:
//   - string: rendered sandboxes content.
func (r *RawRenderer) renderSandboxesWithWidth(snap *model.Snapshot, width int) string {
	// Reuse buildSandboxContentLines for consistency and efficiency.
	lines := r.buildSandboxContentLines(snap, width)

	box := widget.NewBox(width).
		SetTitle("Sandboxes").
		SetTitleColor(r.theme.Header).
		AddLines(lines)

	// Render box to string and return.
	return box.Render()
}

// mergeSideBySide merges two rendered boxes horizontally.
//
// Params:
//   - left: left panel rendered content.
//   - right: right panel rendered content.
//   - separator: string to place between panels.
//
// Returns:
//   - string: merged side-by-side content.
func mergeSideBySide(left, right, separator string) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	// Calculate left panel width.
	leftWidth := calculateMaxVisibleWidth(leftLines)

	// Remove trailing empty lines.
	leftLines = removeTrailingEmptyLines(leftLines)
	rightLines = removeTrailingEmptyLines(rightLines)

	// Merge lines side by side.
	return mergeLines(leftLines, rightLines, leftWidth, separator)
}

// calculateMaxVisibleWidth calculates the maximum visible width of lines.
//
// Params:
//   - lines: lines to measure.
//
// Returns:
//   - int: maximum visible width.
func calculateMaxVisibleWidth(lines []string) int {
	maxWidth := 0
	// Find maximum visible width.
	for _, line := range lines {
		w := widget.VisibleLen(line)
		// Track maximum width.
		if w > maxWidth {
			maxWidth = w
		}
	}
	// Return maximum width.
	return maxWidth
}

// removeTrailingEmptyLines removes empty lines from the end.
//
// Params:
//   - lines: input lines.
//
// Returns:
//   - []string: lines without trailing empty lines.
func removeTrailingEmptyLines(lines []string) []string {
	// Remove trailing empty lines.
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	// Return trimmed lines.
	return lines
}

// mergeLines merges two line slices side by side.
//
// Params:
//   - leftLines: left column lines.
//   - rightLines: right column lines.
//   - leftWidth: width to pad left column to.
//   - separator: string to place between columns.
//
// Returns:
//   - string: merged content.
func mergeLines(leftLines, rightLines []string, leftWidth int, separator string) string {
	maxLines := max(len(leftLines), len(rightLines))
	var result strings.Builder

	// Merge lines side by side.
	for i := range maxLines {
		leftLine := getLineOrEmpty(leftLines, i)
		leftLine = padLineToWidth(leftLine, leftWidth)
		rightLine := getLineOrEmpty(rightLines, i)

		result.WriteString(leftLine)
		result.WriteString(separator)
		result.WriteString(rightLine)

		// Add newline between lines (but not after last line).
		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	// Return merged content.
	return result.String()
}

// getLineOrEmpty returns the line at index or empty string.
//
// Params:
//   - lines: slice of lines.
//   - idx: index to retrieve.
//
// Returns:
//   - string: line at index or empty string.
func getLineOrEmpty(lines []string, idx int) string {
	// Return line if within bounds.
	if idx < len(lines) {
		// Line exists.
		return lines[idx]
	}
	// Return empty for out of bounds.
	return ""
}

// padLineToWidth pads a line to the specified visible width.
//
// Params:
//   - line: line to pad.
//   - width: target visible width.
//
// Returns:
//   - string: padded line.
func padLineToWidth(line string, width int) string {
	visLen := widget.VisibleLen(line)
	// Return unchanged if already wide enough.
	if visLen >= width {
		// No padding needed.
		return line
	}
	// Pad with spaces.
	return line + strings.Repeat(" ", width-visLen)
}
