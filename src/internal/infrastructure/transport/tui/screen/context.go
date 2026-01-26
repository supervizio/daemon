// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// ContextRenderer renders the context/environment section.
type ContextRenderer struct {
	theme  ansi.Theme
	width  int
	status *widget.StatusIndicator
}

// NewContextRenderer creates a context renderer.
func NewContextRenderer(width int) *ContextRenderer {
	return &ContextRenderer{
		theme:  ansi.DefaultTheme(),
		width:  width,
		status: widget.NewStatusIndicator(),
	}
}

// SetWidth updates the renderer width.
func (c *ContextRenderer) SetWidth(width int) {
	c.width = width
}

// Render returns the context section (for raw mode).
func (c *ContextRenderer) Render(snap *model.Snapshot) string {
	ctx := snap.Context

	// Mode string.
	mode := ctx.Mode.String()
	if ctx.ContainerRuntime != "" {
		mode += " (" + ctx.ContainerRuntime + ")"
	}

	// DNS.
	dns := strings.Join(ctx.DNSServers, ", ")
	if dns == "" {
		dns = "-"
	}

	search := strings.Join(ctx.DNSSearch, ", ")
	if search == "" {
		search = "-"
	}

	// Build lines using string concatenation to avoid fmt.Sprintf allocations.
	lines := []string{
		"  Host: " + widget.PadRight(ctx.Hostname, 30) + " Mode: " + mode,
		"  OS: " + ctx.OS + " " + ctx.Kernel + " " + ctx.Arch +
			strings.Repeat(" ", maxInt(0, 20-len(ctx.Kernel)-len(ctx.Arch))) +
			" Uptime: " + widget.FormatDuration(ctx.Uptime),
		"  IP: " + widget.PadRight(ctx.PrimaryIP, 30) + " PID: " + strconv.Itoa(ctx.DaemonPID),
		"  DNS: " + widget.PadRight(dns, 28) + " Search: " + search,
	}

	box := widget.NewBox(c.width).
		SetTitle("Context").
		SetTitleColor(c.theme.Header).
		AddLines(lines)

	return box.Render()
}

// RenderLimits returns the limits section.
func (c *ContextRenderer) RenderLimits(snap *model.Snapshot) string {
	limits := snap.Limits

	if !limits.HasLimits {
		return ""
	}

	// CPU limits - use string concatenation to avoid fmt.Sprintf.
	cpuStr := "-"
	if limits.CPUQuota > 0 {
		cpuStr = formatFloat1(limits.CPUQuota) + " cores (quota " +
			strconv.FormatInt(limits.CPUQuotaRaw, 10) + "/" +
			strconv.FormatInt(limits.CPUPeriod, 10) + ")"
	}

	// Memory limits.
	memStr := "-"
	if limits.MemoryMax > 0 {
		memStr = widget.FormatBytes(limits.MemoryMax) + " max"
	}

	// PIDs.
	pidsStr := "-"
	if limits.PIDsMax > 0 {
		pidsStr = strconv.FormatInt(limits.PIDsCurrent, 10) + "/" + strconv.FormatInt(limits.PIDsMax, 10)
	}

	// CPUSet.
	cpusetStr := limits.CPUSet
	if cpusetStr == "" {
		cpusetStr = "-"
	}

	lines := []string{
		"  CPU: " + widget.PadRight(cpuStr, 35) + " Memory: " + memStr,
		"  PIDs: " + widget.PadRight(pidsStr, 34) + " Cpuset: " + cpusetStr,
	}

	box := widget.NewBox(c.width).
		SetTitle("Limits").
		SetTitleColor(c.theme.Header).
		AddLines(lines)

	return box.Render()
}

// RenderSandboxes returns the sandboxes section.
func (c *ContextRenderer) RenderSandboxes(snap *model.Snapshot) string {
	if len(snap.Sandboxes) == 0 {
		return ""
	}

	// Check if any are detected.
	anyDetected := false
	for _, sb := range snap.Sandboxes {
		if sb.Detected {
			anyDetected = true
			break
		}
	}

	// Format as two columns.
	cols := 2
	perCol := (len(snap.Sandboxes) + cols - 1) / cols

	lines := make([]string, perCol)
	halfWidth := (c.width - 6) / 2

	for i := 0; i < perCol; i++ {
		left := ""
		right := ""

		// Left column.
		if i < len(snap.Sandboxes) {
			sb := snap.Sandboxes[i]
			left = c.formatSandbox(sb, halfWidth)
		}

		// Right column.
		rightIdx := i + perCol
		if rightIdx < len(snap.Sandboxes) {
			sb := snap.Sandboxes[rightIdx]
			right = c.formatSandbox(sb, halfWidth)
		}

		lines[i] = "  " + left + " │ " + right
	}

	// Show even if nothing detected (for completeness in raw mode).
	_ = anyDetected

	box := widget.NewBox(c.width).
		SetTitle("Sandboxes").
		SetTitleColor(c.theme.Header).
		AddLines(lines)

	return box.Render()
}

// formatSandbox formats a single sandbox entry.
func (c *ContextRenderer) formatSandbox(sb model.SandboxInfo, width int) string {
	icon := c.status.Detected(sb.Detected)
	name := widget.Pad(sb.Name, 12, widget.AlignLeft)

	status := c.theme.Muted + "not detected" + ansi.Reset
	if sb.Detected {
		status = widget.Truncate(sb.Endpoint, width-15)
	}

	return icon + " " + name + " " + status
}

// maxInt returns the larger of two ints.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
