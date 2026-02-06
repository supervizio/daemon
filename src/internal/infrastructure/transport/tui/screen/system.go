// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// defaultTerminalRows is the default number of terminal rows for layout calculation.
const defaultTerminalRows int = 24

// compactGrowSize is the pre-allocated buffer size for compact rendering.
const compactGrowSize int = 48

// normalBarPadding is the padding subtracted from width for normal progress bars.
const normalBarPadding int = 20

// loadBufferSize is the pre-allocated buffer size for load average string.
const loadBufferSize int = 32

// strconvBase is the base used for integer to string conversion.
const strconvBase int = 10

// wideBarPadding is the padding subtracted from half-width for wide progress bars.
const wideBarPadding int = 15

// interactiveBarPadding is the padding subtracted from width for interactive/raw progress bars.
const interactiveBarPadding int = 45

// minBarWidth is the minimum width for progress bars to ensure readability.
const minBarWidth int = 10

// floatBufferSize is the buffer size for float formatting operations.
const floatBufferSize int = 16

// floatBitSize is the bit size for float64 conversion.
const floatBitSize int = 64

// floatPrecision0 is the precision for formatting floats with 0 decimal places.
const floatPrecision0 int = 0

// floatPrecision1 is the precision for formatting floats with 1 decimal place.
const floatPrecision1 int = 1

// floatPrecision2 is the precision for formatting floats with 2 decimal places.
const floatPrecision2 int = 2

// percentMultiplier is the multiplier to convert ratio to percentage.
const percentMultiplier float64 = 100

// widthDivisor is the divisor for calculating wide mode bar width (half terminal).
const widthDivisor int = 2

// metricBarCount is the number of metric bars displayed (CPU, RAM, Swap, Disk).
const metricBarCount int = 4

// cpuBarIndex is the array index for the CPU progress bar.
const cpuBarIndex int = 0

// ramBarIndex is the array index for the RAM progress bar.
const ramBarIndex int = 1

// swapBarIndex is the array index for the Swap progress bar.
const swapBarIndex int = 2

// diskBarIndex is the array index for the Disk progress bar.
const diskBarIndex int = 3

// SystemRenderer renders system metrics for terminal display.
// It supports multiple output formats (compact, normal, wide) based on terminal size.
type SystemRenderer struct {
	theme ansi.Theme
	width int
}

// NewSystemRenderer creates a system renderer.
// Params:
//   - width: terminal width in columns
//
// Returns:
//   - *SystemRenderer: configured renderer instance
func NewSystemRenderer(width int) *SystemRenderer {
	// Initialize renderer with default theme and specified width.
	return &SystemRenderer{
		theme: ansi.DefaultTheme(),
		width: width,
	}
}

// SetWidth updates the renderer width.
// Params:
//   - width: new terminal width in columns
func (s *SystemRenderer) SetWidth(width int) {
	s.width = width
}

// Render returns the system metrics section.
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: rendered system metrics section
func (s *SystemRenderer) Render(snap *model.Snapshot) string {
	layout := terminal.GetLayout(terminal.Size{Cols: s.width, Rows: defaultTerminalRows})

	// Select rendering mode based on terminal layout.
	switch layout {
	// Compact mode for small terminals.
	case terminal.LayoutCompact:
		// Return minimal system metrics.
		return s.renderCompact(snap)
	// Normal mode for standard terminals.
	case terminal.LayoutNormal:
		// Return standard system metrics.
		return s.renderNormal(snap)
	// Wide modes for large terminals.
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		// Return expanded system metrics.
		return s.renderWide(snap)
	// handle default case.
	default:
		// Default to normal rendering for unhandled layouts.
		return s.renderNormal(snap)
	}
}

// renderCompact renders minimal system metrics.
// Uses strings.Builder to avoid fmt.Sprintf allocations.
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: rendered compact system metrics
func (s *SystemRenderer) renderCompact(snap *model.Snapshot) string {
	sys := snap.System

	// Single line format with strings.Builder.
	var sb strings.Builder
	sb.Grow(compactGrowSize)
	sb.WriteString("  CPU ")
	sb.WriteString(widget.FormatPercent(sys.CPUPercent))
	sb.WriteString(" │ RAM ")
	sb.WriteString(widget.FormatPercent(sys.MemoryPercent))
	line := sb.String()

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLine(line)

	// Return rendered compact system box.
	return box.Render()
}

// renderNormal renders standard system metrics.
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: rendered normal system metrics
func (s *SystemRenderer) renderNormal(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width - normalBarPadding

	// Create progress bars for CPU, RAM, and Swap.
	cpuBar := s.createProgressBar(barWidth, sys.CPUPercent, "CPU ")
	ramBar := s.createProgressBar(barWidth, sys.MemoryPercent, "RAM ")
	swapBar := s.createProgressBar(barWidth, sys.SwapPercent, "Swap")

	lines := []string{
		"  " + cpuBar.Render(),
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + "/" + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render(),
	}

	// Load average with strconv to avoid fmt.Sprintf.
	var loadBuf strings.Builder
	loadBuf.Grow(loadBufferSize)
	loadBuf.WriteString("  Load: ")
	loadBuf.WriteString(formatFloat2(sys.LoadAvg1))
	loadBuf.WriteByte(' ')
	loadBuf.WriteString(formatFloat2(sys.LoadAvg5))
	loadBuf.WriteByte(' ')
	loadBuf.WriteString(formatFloat2(sys.LoadAvg15))
	lines = append(lines, s.theme.Muted+loadBuf.String()+ansi.Reset)

	// Append limits line if cgroup limits are present.
	lines = s.appendLimitsNormal(lines, limits)

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	// Return rendered normal system box.
	return box.Render()
}

// appendLimitsNormal appends cgroup limits to lines for normal mode.
// Params:
//   - lines: existing lines to append to
//   - limits: resource limits data
//
// Returns:
//   - []string: lines with limits appended if present
func (s *SystemRenderer) appendLimitsNormal(lines []string, limits model.ResourceLimits) []string {
	// Skip if no cgroup limits detected.
	if !limits.HasLimits {
		// Return lines unchanged when no limits.
		return lines
	}

	var limitParts []string

	// Add CPU quota if set.
	if limits.CPUQuota > 0 {
		limitParts = append(limitParts, "CPU: "+formatFloat1(limits.CPUQuota)+" cores")
	}

	// Add memory limit if set.
	if limits.MemoryMax > 0 {
		limitParts = append(limitParts, "MEM: "+widget.FormatBytes(limits.MemoryMax))
	}

	// Add PIDs limit if set.
	if limits.PIDsMax > 0 {
		limitParts = append(limitParts, "PIDs: "+strconv.FormatInt(limits.PIDsCurrent, strconvBase)+"/"+strconv.FormatInt(limits.PIDsMax, strconvBase))
	}

	// Append limits line when any limits present.
	if len(limitParts) > 0 {
		lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
	}

	// Return lines with limits appended.
	return lines
}

// renderWide renders expanded system metrics.
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: rendered wide system metrics
func (s *SystemRenderer) renderWide(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width/widthDivisor - wideBarPadding

	// Create progress bars for CPU, RAM, and Swap.
	cpuBar := s.createProgressBar(barWidth, sys.CPUPercent, "CPU ")
	ramBar := s.createProgressBar(barWidth, sys.MemoryPercent, "RAM ")
	swapBar := s.createProgressBar(barWidth, sys.SwapPercent, "Swap")

	lines := []string{
		"  " + cpuBar.Render() + "  Load: " + formatFloat2(sys.LoadAvg1) + " " + formatFloat2(sys.LoadAvg5) + " " + formatFloat2(sys.LoadAvg15),
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render() + "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
	}

	// Append limits line if cgroup limits are present.
	lines = s.appendLimitsWide(lines, limits)

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	// Return rendered wide system box.
	return box.Render()
}

// appendLimitsWide appends cgroup limits to lines for wide mode with full details.
// Params:
//   - lines: existing lines to append to
//   - limits: resource limits data
//
// Returns:
//   - []string: lines with limits appended if present
func (s *SystemRenderer) appendLimitsWide(lines []string, limits model.ResourceLimits) []string {
	// Skip if no cgroup limits detected.
	if !limits.HasLimits {
		// Return lines unchanged when no limits.
		return lines
	}

	var limitParts []string

	// Add CPU quota if set.
	if limits.CPUQuota > 0 {
		limitParts = append(limitParts, "CPU: "+formatFloat1(limits.CPUQuota)+" cores")
	}

	// Add CPU set if configured.
	if limits.CPUSet != "" {
		limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
	}

	// Add memory usage and limit if set.
	if limits.MemoryMax > 0 {
		used := float64(limits.MemoryCurrent) / float64(limits.MemoryMax) * percentMultiplier
		limitParts = append(limitParts, "MEM: "+widget.FormatBytes(limits.MemoryCurrent)+"/"+widget.FormatBytes(limits.MemoryMax)+" ("+formatFloat0(used)+"%)")
	}

	// Add PIDs usage and limit if set.
	if limits.PIDsMax > 0 {
		limitParts = append(limitParts, "PIDs: "+strconv.FormatInt(limits.PIDsCurrent, strconvBase)+"/"+strconv.FormatInt(limits.PIDsMax, strconvBase))
	}

	// Append limits line when any limits present.
	if len(limitParts) > 0 {
		lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
	}

	// Return lines with limits appended.
	return lines
}

// RenderInline returns a single-line summary.
// Uses string concatenation to avoid fmt.Sprintf allocation.
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: single-line system metrics summary
func (s *SystemRenderer) RenderInline(snap *model.Snapshot) string {
	sys := snap.System

	// Return formatted inline summary.
	return "CPU " + widget.FormatPercent(sys.CPUPercent) +
		" │ RAM " + widget.FormatPercent(sys.MemoryPercent) +
		" │ Load " + formatFloat2(sys.LoadAvg1)
}

// RenderForInteractive renders system metrics for interactive mode.
// Same format as raw mode but with "System" title (updates in real-time).
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: rendered interactive system metrics
func (s *SystemRenderer) RenderForInteractive(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	// Calculate bar width with minimum constraint for readability.
	barWidth := max(s.width-interactiveBarPadding, minBarWidth)

	// Create progress bars and format info strings.
	bars, infos := s.createInteractiveBars(barWidth, sys)

	lines := []string{
		"  " + bars[cpuBarIndex].Render() + infos[cpuBarIndex],
		"  " + bars[ramBarIndex].Render() + infos[ramBarIndex],
		"  " + bars[swapBarIndex].Render() + infos[swapBarIndex],
		"  " + bars[diskBarIndex].Render() + infos[diskBarIndex],
	}

	// Append limits line if cgroup limits are present.
	lines = s.appendLimitsInteractive(lines, limits)

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	// Return rendered interactive system box.
	return box.Render()
}

// createInteractiveBars creates progress bars and info strings for interactive/raw modes.
// Params:
//   - barWidth: width for progress bars
//   - sys: system metrics data
//
// Returns:
//   - [metricBarCount]*widget.ProgressBar: array of progress bars (CPU, RAM, Swap, Disk)
//   - [metricBarCount]string: array of info strings for each bar
func (s *SystemRenderer) createInteractiveBars(barWidth int, sys model.SystemMetrics) (bars [metricBarCount]*widget.ProgressBar, info [metricBarCount]string) {
	// CPU bar.
	cpuBar := s.createProgressBar(barWidth, sys.CPUPercent, "CPU  ")

	// RAM bar.
	ramBar := s.createProgressBar(barWidth, sys.MemoryPercent, "RAM  ")

	// Swap bar.
	swapBar := s.createProgressBar(barWidth, sys.SwapPercent, "Swap ")

	// Disk bar.
	diskBar := s.createProgressBar(barWidth, sys.DiskPercent, "Disk ")

	bars = [metricBarCount]*widget.ProgressBar{cpuBar, ramBar, swapBar, diskBar}

	// Format values with padding - string concatenation avoids fmt.Sprintf.
	info = [metricBarCount]string{
		"  Load: " + formatFloat2(sys.LoadAvg1) + " " + formatFloat2(sys.LoadAvg5) + " " + formatFloat2(sys.LoadAvg15),
		"  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
		"  " + widget.FormatBytes(sys.DiskUsed) + " / " + widget.FormatBytes(sys.DiskTotal),
	}

	// Return bars and info strings.
	return bars, info
}

// appendLimitsInteractive appends cgroup limits to lines for interactive mode.
// appendLimits appends cgroup limits to lines if present.
//
// Params:
//   - lines: existing lines to append to
//   - limits: resource limits data
//
// Returns:
//   - []string: lines with limits appended if present
func (s *SystemRenderer) appendLimits(lines []string, limits model.ResourceLimits) []string {
	// Skip if no cgroup limits detected.
	if !limits.HasLimits {
		// Return lines unchanged when no limits.
		return lines
	}

	var limitParts []string

	// Add CPU quota if set.
	if limits.CPUQuota > 0 {
		limitParts = append(limitParts, "CPU: "+formatFloat0(limits.CPUQuota)+" cores")
	}

	// Add memory limit if set.
	if limits.MemoryMax > 0 {
		limitParts = append(limitParts, "Memory: "+widget.FormatBytes(limits.MemoryMax))
	}

	// Add CPU set if configured.
	if limits.CPUSet != "" {
		limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
	}

	// Append limits line when any limits present.
	if len(limitParts) > 0 {
		lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
	}

	// Return lines with limits appended.
	return lines
}

// appendLimitsInteractive appends cgroup limits to lines for interactive mode.
//
// Params:
//   - lines: existing lines to append to
//   - limits: resource limits data
//
// Returns:
//   - []string: lines with limits appended if present
func (s *SystemRenderer) appendLimitsInteractive(lines []string, limits model.ResourceLimits) []string {
	// Delegate to common implementation.
	return s.appendLimits(lines, limits)
}

// RenderForRaw renders system metrics for raw mode with "at start" indicator.
// Includes CPU, RAM, Swap, Disk, and Limits.
// Params:
//   - snap: snapshot containing system metrics data
//
// Returns:
//   - string: rendered raw mode system metrics
func (s *SystemRenderer) RenderForRaw(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width - interactiveBarPadding

	// Create progress bars and format info strings.
	bars, infos := s.createRawBars(barWidth, sys)

	lines := []string{
		"  " + bars[cpuBarIndex].Render() + infos[cpuBarIndex],
		"  " + bars[ramBarIndex].Render() + infos[ramBarIndex],
		"  " + bars[swapBarIndex].Render() + infos[swapBarIndex],
		"  " + bars[diskBarIndex].Render() + infos[diskBarIndex],
	}

	// Append limits line if cgroup limits are present.
	lines = s.appendLimitsRaw(lines, limits)

	box := widget.NewBox(s.width).
		SetTitle("System (at start)").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	// Return rendered raw mode system box.
	return box.Render()
}

// createRawBars creates progress bars and info strings for raw mode.
// Params:
//   - barWidth: width for progress bars
//   - sys: system metrics data
//
// Returns:
//   - [metricBarCount]*widget.ProgressBar: array of progress bars (CPU, RAM, Swap, Disk)
//   - [metricBarCount]string: array of info strings for each bar
func (s *SystemRenderer) createRawBars(barWidth int, sys model.SystemMetrics) (bars [metricBarCount]*widget.ProgressBar, info [metricBarCount]string) {
	// CPU bar.
	cpuBar := s.createProgressBar(barWidth, sys.CPUPercent, "CPU  ")

	// RAM bar.
	ramBar := s.createProgressBar(barWidth, sys.MemoryPercent, "RAM  ")

	// Swap bar.
	swapBar := s.createProgressBar(barWidth, sys.SwapPercent, "Swap ")

	// Disk bar.
	diskBar := s.createProgressBar(barWidth, sys.DiskPercent, "Disk ")

	bars = [metricBarCount]*widget.ProgressBar{cpuBar, ramBar, swapBar, diskBar}

	// Format values with padding - string concatenation avoids fmt.Sprintf.
	info = [metricBarCount]string{
		"  Load: " + formatFloat2(sys.LoadAvg1) + " " + formatFloat2(sys.LoadAvg5),
		"  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
		"  " + widget.FormatBytes(sys.DiskUsed) + " / " + widget.FormatBytes(sys.DiskTotal),
	}

	// Return bars and info strings.
	return bars, info
}

// appendLimitsRaw appends cgroup limits to lines for raw mode.
// appendLimitsRaw appends cgroup limits to lines for raw mode.
//
// Params:
//   - lines: existing lines to append to
//   - limits: resource limits data
//
// Returns:
//   - []string: lines with limits appended if present
func (s *SystemRenderer) appendLimitsRaw(lines []string, limits model.ResourceLimits) []string {
	// Delegate to common implementation.
	return s.appendLimits(lines, limits)
}

// createProgressBar creates a progress bar with standard configuration.
// Params:
//   - width: width of the progress bar
//   - percent: percentage value (0-100)
//   - label: label text for the bar
//
// Returns:
//   - *widget.ProgressBar: configured progress bar instance
func (s *SystemRenderer) createProgressBar(width int, percent float64, label string) *widget.ProgressBar {
	// Return configured progress bar with color based on percentage.
	return widget.NewProgressBar(width, percent).
		SetLabel(label).
		SetColorByPercent()
}

// formatFloat2 formats a float with 2 decimal places using strconv.
// Avoids fmt.Sprintf allocation overhead.
// Params:
//   - f: float64 value to format
//
// Returns:
//   - string: formatted float with 2 decimal places
func formatFloat2(f float64) string {
	var buf [floatBufferSize]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', floatPrecision2, floatBitSize)

	// Return formatted float string.
	return string(b)
}

// formatFloat1 formats a float with 1 decimal place using strconv.
// Params:
//   - f: float64 value to format
//
// Returns:
//   - string: formatted float with 1 decimal place
func formatFloat1(f float64) string {
	var buf [floatBufferSize]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', floatPrecision1, floatBitSize)

	// Return formatted float string.
	return string(b)
}

// formatFloat0 formats a float with 0 decimal places using strconv.
// Params:
//   - f: float64 value to format
//
// Returns:
//   - string: formatted float with 0 decimal places
func formatFloat0(f float64) string {
	var buf [floatBufferSize]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', floatPrecision0, floatBitSize)

	// Return formatted float string.
	return string(b)
}
