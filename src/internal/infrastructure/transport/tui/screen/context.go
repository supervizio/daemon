// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// decimalBase is the base for decimal number formatting.
const decimalBase int = 10

// ContextRenderer renders the context/environment section.
// Displays system information including host, OS, networking, and cgroup limits.
type ContextRenderer struct {
	theme  ansi.Theme
	width  int
	status *widget.StatusIndicator
}

// NewContextRenderer creates a context renderer.
//
// Params:
//   - width: terminal width for rendering
//
// Returns:
//   - *ContextRenderer: configured renderer instance
func NewContextRenderer(width int) *ContextRenderer {
	// Initialize renderer with default theme and status indicator.
	return &ContextRenderer{
		theme:  ansi.DefaultTheme(),
		width:  width,
		status: widget.NewStatusIndicator(),
	}
}

// SetWidth updates the renderer width.
//
// Params:
//   - width: new terminal width
func (c *ContextRenderer) SetWidth(width int) {
	c.width = width
}

// Render returns the context section (for raw mode).
//
// Params:
//   - snap: system snapshot with context data
//
// Returns:
//   - string: rendered context section
func (c *ContextRenderer) Render(snap *model.Snapshot) string {
	ctx := snap.Context

	// Build mode string with optional container runtime.
	mode := ctx.Mode.String()
	// Append container runtime info if running in container.
	if ctx.ContainerRuntime != "" {
		mode += " (" + ctx.ContainerRuntime + ")"
	}

	// Format DNS servers or show placeholder.
	dns := strings.Join(ctx.DNSServers, ", ")
	// Show placeholder when no DNS servers configured.
	if dns == "" {
		dns = "-"
	}

	// Format DNS search domains or show placeholder.
	search := strings.Join(ctx.DNSSearch, ", ")
	// Show placeholder when no search domains configured.
	if search == "" {
		search = "-"
	}

	const (
		// hostPadding is the padding width for host information display.
		hostPadding int = 30
		// kernelSpace is the spacing allocation for kernel version display.
		kernelSpace int = 20
		// dnsPadding is the padding width for DNS information display.
		dnsPadding int = 28
	)

	// Build lines using string concatenation to avoid fmt.Sprintf allocations.
	lines := []string{
		"  Host: " + widget.PadRight(ctx.Hostname, hostPadding) + " Mode: " + mode,
		"  OS: " + ctx.OS + " " + ctx.Kernel + " " + ctx.Arch +
			strings.Repeat(" ", max(0, kernelSpace-len(ctx.Kernel)-len(ctx.Arch))) +
			" Uptime: " + widget.FormatDuration(ctx.Uptime),
		"  IP: " + widget.PadRight(ctx.PrimaryIP, hostPadding) + " PID: " + strconv.Itoa(ctx.DaemonPID),
		"  DNS: " + widget.PadRight(dns, dnsPadding) + " Search: " + search,
	}

	box := widget.NewBox(c.width).
		SetTitle("Context").
		SetTitleColor(c.theme.Header).
		AddLines(lines)

	// Return rendered box content.
	return box.Render()
}

// RenderLimits returns the limits section.
//
// Params:
//   - snap: system snapshot with cgroup limits
//
// Returns:
//   - string: rendered limits section or empty if no limits
func (c *ContextRenderer) RenderLimits(snap *model.Snapshot) string {
	limits := snap.Limits

	// Skip rendering if no cgroup limits detected.
	if !limits.HasLimits {
		// Return empty string when no limits to display.
		return ""
	}

	const (
		// cpuPadding is the padding width for CPU quota display.
		cpuPadding int = 35
		// pidsPadding is the padding width for PIDs display.
		pidsPadding int = 34
	)

	// Format CPU quota with limits.
	cpuStr := "-"
	// Build CPU quota string if quota is set.
	if limits.CPUQuota > 0 {
		cpuStr = formatFloat1(limits.CPUQuota) + " cores (quota " +
			strconv.FormatInt(limits.CPUQuotaRaw, decimalBase) + "/" +
			strconv.FormatInt(limits.CPUPeriod, decimalBase) + ")"
	}

	// Format memory limits.
	memStr := "-"
	// Build memory limit string if memory limit is set.
	if limits.MemoryMax > 0 {
		memStr = widget.FormatBytes(limits.MemoryMax) + " max"
	}

	// Format PID usage and limits.
	pidsStr := "-"
	// Build PIDs string showing current/max if limit is set.
	if limits.PIDsMax > 0 {
		pidsStr = strconv.FormatInt(limits.PIDsCurrent, decimalBase) + "/" + strconv.FormatInt(limits.PIDsMax, decimalBase)
	}

	// Format CPU affinity set.
	cpusetStr := limits.CPUSet
	// Show placeholder when no CPU set configured.
	if cpusetStr == "" {
		cpusetStr = "-"
	}

	lines := []string{
		"  CPU: " + widget.PadRight(cpuStr, cpuPadding) + " Memory: " + memStr,
		"  PIDs: " + widget.PadRight(pidsStr, pidsPadding) + " Cpuset: " + cpusetStr,
	}

	box := widget.NewBox(c.width).
		SetTitle("Limits").
		SetTitleColor(c.theme.Header).
		AddLines(lines)

	// Return rendered limits box.
	return box.Render()
}

// RenderSandboxes returns the sandboxes section.
//
// Params:
//   - snap: system snapshot with sandbox detection info
//
// Returns:
//   - string: rendered sandboxes section or empty if none detected
func (c *ContextRenderer) RenderSandboxes(snap *model.Snapshot) string {
	// Skip if no sandboxes to display.
	if len(snap.Sandboxes) == 0 {
		// Return empty string when no sandboxes defined.
		return ""
	}

	const (
		// columnCount is the number of columns for sandbox layout.
		columnCount int = 2
		// boxBorderWidth is the total width consumed by box borders.
		boxBorderWidth int = 6
	)

	// Check if any sandboxes are detected.
	anyDetected := false
	// Iterate through sandboxes to find any detected ones.
	for _, sb := range snap.Sandboxes {
		// Mark as detected if sandbox is active.
		if sb.Detected {
			anyDetected = true
			break
		}
	}

	// Calculate two-column layout.
	perCol := (len(snap.Sandboxes) + columnCount - 1) / columnCount

	lines := make([]string, perCol)
	halfWidth := (c.width - boxBorderWidth) / columnCount

	// Build lines with left and right columns.
	// Iterate over rows to create two-column layout.
	for i := range perCol {
		left := ""
		right := ""

		// Format left column entry.
		// Check if sandbox exists for left column position.
		if i < len(snap.Sandboxes) {
			sb := snap.Sandboxes[i]
			left = c.formatSandbox(sb, halfWidth)
		}

		// Format right column entry.
		rightIdx := i + perCol
		// Check if sandbox exists for right column position.
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

	// Return rendered sandboxes box.
	return box.Render()
}

// formatSandbox formats a single sandbox entry.
//
// Params:
//   - sb: sandbox information
//   - width: available width for formatting
//
// Returns:
//   - string: formatted sandbox entry with status
func (c *ContextRenderer) formatSandbox(sb model.SandboxInfo, width int) string {
	// nameWidth is the width allocated for sandbox names.
	const (
		// statusPadding is the padding for sandbox status text.
		nameWidth     int = 12
		statusPadding int = 15
	)

	icon := c.status.Detected(sb.Detected)
	name := widget.Pad(sb.Name, nameWidth, widget.AlignLeft)

	// Show endpoint if detected, otherwise show not detected message.
	status := c.theme.Muted + "not detected" + ansi.Reset
	// Use endpoint when sandbox is detected.
	if sb.Detected {
		status = widget.Truncate(sb.Endpoint, width-statusPadding)
	}

	// Return formatted sandbox entry string.
	return icon + " " + name + " " + status
}

// NOTE: Tests needed - create context_external_test.go and context_internal_test.go
