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

// SystemRenderer renders system metrics.
type SystemRenderer struct {
	theme ansi.Theme
	width int
}

// NewSystemRenderer creates a system renderer.
func NewSystemRenderer(width int) *SystemRenderer {
	return &SystemRenderer{
		theme: ansi.DefaultTheme(),
		width: width,
	}
}

// SetWidth updates the renderer width.
func (s *SystemRenderer) SetWidth(width int) {
	s.width = width
}

// Render returns the system metrics section.
func (s *SystemRenderer) Render(snap *model.Snapshot) string {
	layout := terminal.GetLayout(terminal.Size{Cols: s.width, Rows: 24})

	switch layout {
	case terminal.LayoutCompact:
		return s.renderCompact(snap)
	case terminal.LayoutNormal:
		return s.renderNormal(snap)
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		return s.renderWide(snap)
	}
	return s.renderNormal(snap)
}

// renderCompact renders minimal system metrics.
// Uses strings.Builder to avoid fmt.Sprintf allocations.
func (s *SystemRenderer) renderCompact(snap *model.Snapshot) string {
	sys := snap.System

	// Single line format with strings.Builder.
	var sb strings.Builder
	sb.Grow(48)
	sb.WriteString("  CPU ")
	sb.WriteString(widget.FormatPercent(sys.CPUPercent))
	sb.WriteString(" │ RAM ")
	sb.WriteString(widget.FormatPercent(sys.MemoryPercent))
	line := sb.String()

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLine(line)

	return box.Render()
}

// renderNormal renders standard system metrics.
func (s *SystemRenderer) renderNormal(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width - 20

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

	lines := []string{
		"  " + cpuBar.Render(),
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + "/" + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render(),
	}

	// Load average with strconv to avoid fmt.Sprintf.
	var loadBuf strings.Builder
	loadBuf.Grow(32)
	loadBuf.WriteString("  Load: ")
	loadBuf.WriteString(formatFloat2(sys.LoadAvg1))
	loadBuf.WriteByte(' ')
	loadBuf.WriteString(formatFloat2(sys.LoadAvg5))
	loadBuf.WriteByte(' ')
	loadBuf.WriteString(formatFloat2(sys.LoadAvg15))
	lines = append(lines, s.theme.Muted+loadBuf.String()+ansi.Reset)

	// Limits if present - use string concatenation to avoid fmt.Sprintf.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, "CPU: "+formatFloat1(limits.CPUQuota)+" cores")
		}
		if limits.MemoryMax > 0 {
			limitParts = append(limitParts, "MEM: "+widget.FormatBytes(limits.MemoryMax))
		}
		if limits.PIDsMax > 0 {
			limitParts = append(limitParts, "PIDs: "+strconv.FormatInt(limits.PIDsCurrent, 10)+"/"+strconv.FormatInt(limits.PIDsMax, 10))
		}
		if len(limitParts) > 0 {
			lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
		}
	}

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	return box.Render()
}

// renderWide renders expanded system metrics.
func (s *SystemRenderer) renderWide(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width/2 - 15

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

	lines := []string{
		"  " + cpuBar.Render() + "  Load: " + formatFloat2(sys.LoadAvg1) + " " + formatFloat2(sys.LoadAvg5) + " " + formatFloat2(sys.LoadAvg15),
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render() + "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
	}

	// Limits if present - use string concatenation to avoid fmt.Sprintf.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, "CPU: "+formatFloat1(limits.CPUQuota)+" cores")
		}
		if limits.CPUSet != "" {
			limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
		}
		if limits.MemoryMax > 0 {
			used := float64(limits.MemoryCurrent) / float64(limits.MemoryMax) * 100
			limitParts = append(limitParts, "MEM: "+widget.FormatBytes(limits.MemoryCurrent)+"/"+widget.FormatBytes(limits.MemoryMax)+" ("+formatFloat0(used)+"%)")
		}
		if limits.PIDsMax > 0 {
			limitParts = append(limitParts, "PIDs: "+strconv.FormatInt(limits.PIDsCurrent, 10)+"/"+strconv.FormatInt(limits.PIDsMax, 10))
		}
		if len(limitParts) > 0 {
			lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
		}
	}

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	return box.Render()
}

// RenderInline returns a single-line summary.
// Uses string concatenation to avoid fmt.Sprintf allocation.
func (s *SystemRenderer) RenderInline(snap *model.Snapshot) string {
	sys := snap.System

	return "CPU " + widget.FormatPercent(sys.CPUPercent) +
		" │ RAM " + widget.FormatPercent(sys.MemoryPercent) +
		" │ Load " + formatFloat2(sys.LoadAvg1)
}

// RenderForInteractive renders system metrics for interactive mode.
// Same format as raw mode but with "System" title (updates in real-time).
func (s *SystemRenderer) RenderForInteractive(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width - 45
	if barWidth < 10 {
		barWidth = 10
	}

	// CPU bar.
	cpuBar := widget.NewProgressBar(barWidth, sys.CPUPercent).
		SetLabel("CPU  ").
		SetColorByPercent()

	// RAM bar.
	ramBar := widget.NewProgressBar(barWidth, sys.MemoryPercent).
		SetLabel("RAM  ").
		SetColorByPercent()

	// Swap bar.
	swapBar := widget.NewProgressBar(barWidth, sys.SwapPercent).
		SetLabel("Swap ").
		SetColorByPercent()

	// Disk bar.
	diskBar := widget.NewProgressBar(barWidth, sys.DiskPercent).
		SetLabel("Disk ").
		SetColorByPercent()

	// Format values with padding - string concatenation avoids fmt.Sprintf.
	cpuInfo := "  Load: " + formatFloat2(sys.LoadAvg1) + " " + formatFloat2(sys.LoadAvg5) + " " + formatFloat2(sys.LoadAvg15)
	ramInfo := "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal)
	swapInfo := "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal)
	diskInfo := "  " + widget.FormatBytes(sys.DiskUsed) + " / " + widget.FormatBytes(sys.DiskTotal)

	lines := []string{
		"  " + cpuBar.Render() + cpuInfo,
		"  " + ramBar.Render() + ramInfo,
		"  " + swapBar.Render() + swapInfo,
		"  " + diskBar.Render() + diskInfo,
	}

	// Limits if present - use string concatenation to avoid fmt.Sprintf.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, "CPU: "+formatFloat0(limits.CPUQuota)+" cores")
		}
		if limits.MemoryMax > 0 {
			limitParts = append(limitParts, "Memory: "+widget.FormatBytes(limits.MemoryMax))
		}
		if limits.CPUSet != "" {
			limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
		}
		if len(limitParts) > 0 {
			lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
		}
	}

	box := widget.NewBox(s.width).
		SetTitle("System").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	return box.Render()
}

// RenderForRaw renders system metrics for raw mode with "at start" indicator.
// Includes CPU, RAM, Swap, Disk, and Limits.
func (s *SystemRenderer) RenderForRaw(snap *model.Snapshot) string {
	sys := snap.System
	limits := snap.Limits

	barWidth := s.width - 45

	// CPU bar.
	cpuBar := widget.NewProgressBar(barWidth, sys.CPUPercent).
		SetLabel("CPU  ").
		SetColorByPercent()

	// RAM bar.
	ramBar := widget.NewProgressBar(barWidth, sys.MemoryPercent).
		SetLabel("RAM  ").
		SetColorByPercent()

	// Swap bar.
	swapBar := widget.NewProgressBar(barWidth, sys.SwapPercent).
		SetLabel("Swap ").
		SetColorByPercent()

	// Disk bar.
	diskBar := widget.NewProgressBar(barWidth, sys.DiskPercent).
		SetLabel("Disk ").
		SetColorByPercent()

	// Format values with padding - string concatenation avoids fmt.Sprintf.
	cpuInfo := "  Load: " + formatFloat2(sys.LoadAvg1) + " " + formatFloat2(sys.LoadAvg5)
	ramInfo := "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal)
	swapInfo := "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal)
	diskInfo := "  " + widget.FormatBytes(sys.DiskUsed) + " / " + widget.FormatBytes(sys.DiskTotal)

	lines := []string{
		"  " + cpuBar.Render() + cpuInfo,
		"  " + ramBar.Render() + ramInfo,
		"  " + swapBar.Render() + swapInfo,
		"  " + diskBar.Render() + diskInfo,
	}

	// Limits if present - use string concatenation to avoid fmt.Sprintf.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, "CPU: "+formatFloat0(limits.CPUQuota)+" cores")
		}
		if limits.MemoryMax > 0 {
			limitParts = append(limitParts, "Memory: "+widget.FormatBytes(limits.MemoryMax))
		}
		if limits.CPUSet != "" {
			limitParts = append(limitParts, "CPUSet: "+limits.CPUSet)
		}
		if len(limitParts) > 0 {
			lines = append(lines, "  "+s.theme.Muted+"Limits: "+strings.Join(limitParts, " │ ")+ansi.Reset)
		}
	}

	box := widget.NewBox(s.width).
		SetTitle("System (at start)").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	return box.Render()
}

// formatFloat2 formats a float with 2 decimal places using strconv.
// Avoids fmt.Sprintf allocation overhead.
func formatFloat2(f float64) string {
	var buf [16]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', 2, 64)
	return string(b)
}

// formatFloat1 formats a float with 1 decimal place using strconv.
func formatFloat1(f float64) string {
	var buf [16]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', 1, 64)
	return string(b)
}

// formatFloat0 formats a float with 0 decimal places using strconv.
func formatFloat0(f float64) string {
	var buf [16]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', 0, 64)
	return string(b)
}
