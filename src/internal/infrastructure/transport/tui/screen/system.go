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
func (s *SystemRenderer) renderCompact(snap *model.Snapshot) string {
	sys := snap.System

	// Single line format.
	line := fmt.Sprintf("  CPU %s │ RAM %s",
		widget.FormatPercent(sys.CPUPercent),
		widget.FormatPercent(sys.MemoryPercent))

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

	// Load average.
	loadLine := fmt.Sprintf("  Load: %.2f %.2f %.2f", sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15)
	lines = append(lines, s.theme.Muted+loadLine+ansi.Reset)

	// Limits if present.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, fmt.Sprintf("CPU: %.1f cores", limits.CPUQuota))
		}
		if limits.MemoryMax > 0 {
			limitParts = append(limitParts, fmt.Sprintf("MEM: %s", widget.FormatBytes(limits.MemoryMax)))
		}
		if limits.PIDsMax > 0 {
			limitParts = append(limitParts, fmt.Sprintf("PIDs: %d/%d", limits.PIDsCurrent, limits.PIDsMax))
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
		"  " + cpuBar.Render() + "  Load: " + fmt.Sprintf("%.2f %.2f %.2f", sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15),
		"  " + ramBar.Render() + "  " + widget.FormatBytes(sys.MemoryUsed) + " / " + widget.FormatBytes(sys.MemoryTotal),
		"  " + swapBar.Render() + "  " + widget.FormatBytes(sys.SwapUsed) + " / " + widget.FormatBytes(sys.SwapTotal),
	}

	// Limits if present.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, fmt.Sprintf("CPU: %.1f cores", limits.CPUQuota))
		}
		if limits.CPUSet != "" {
			limitParts = append(limitParts, fmt.Sprintf("CPUSet: %s", limits.CPUSet))
		}
		if limits.MemoryMax > 0 {
			used := float64(limits.MemoryCurrent) / float64(limits.MemoryMax) * 100
			limitParts = append(limitParts, fmt.Sprintf("MEM: %s/%s (%.0f%%)",
				widget.FormatBytes(limits.MemoryCurrent),
				widget.FormatBytes(limits.MemoryMax),
				used))
		}
		if limits.PIDsMax > 0 {
			limitParts = append(limitParts, fmt.Sprintf("PIDs: %d/%d", limits.PIDsCurrent, limits.PIDsMax))
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
func (s *SystemRenderer) RenderInline(snap *model.Snapshot) string {
	sys := snap.System

	return fmt.Sprintf("CPU %s │ RAM %s │ Load %.2f",
		widget.FormatPercent(sys.CPUPercent),
		widget.FormatPercent(sys.MemoryPercent),
		sys.LoadAvg1)
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

	// Format values with padding.
	cpuInfo := fmt.Sprintf("  Load: %.2f %.2f", sys.LoadAvg1, sys.LoadAvg5)
	ramInfo := fmt.Sprintf("  %s / %s", widget.FormatBytes(sys.MemoryUsed), widget.FormatBytes(sys.MemoryTotal))
	swapInfo := fmt.Sprintf("  %s / %s", widget.FormatBytes(sys.SwapUsed), widget.FormatBytes(sys.SwapTotal))
	diskInfo := fmt.Sprintf("  %s / %s", widget.FormatBytes(sys.DiskUsed), widget.FormatBytes(sys.DiskTotal))

	lines := []string{
		"  " + cpuBar.Render() + cpuInfo,
		"  " + ramBar.Render() + ramInfo,
		"  " + swapBar.Render() + swapInfo,
		"  " + diskBar.Render() + diskInfo,
	}

	// Limits if present.
	if limits.HasLimits {
		var limitParts []string
		if limits.CPUQuota > 0 {
			limitParts = append(limitParts, fmt.Sprintf("CPU: %.0f cores", limits.CPUQuota))
		}
		if limits.MemoryMax > 0 {
			limitParts = append(limitParts, fmt.Sprintf("Memory: %s", widget.FormatBytes(limits.MemoryMax)))
		}
		if limits.CPUSet != "" {
			limitParts = append(limitParts, fmt.Sprintf("CPUSet: %s", limits.CPUSet))
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
