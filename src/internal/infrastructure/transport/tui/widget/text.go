// Package widget provides reusable TUI components.
package widget

import (
	"fmt"
	"strings"
	"time"
)

// Align specifies text alignment.
type Align int

// Alignment constants.
const (
	AlignLeft Align = iota
	AlignRight
	AlignCenter
)

// Pad pads text to width with specified alignment.
func Pad(text string, width int, align Align) string {
	visLen := VisibleLen(text)
	if visLen >= width {
		return truncateVisible(text, width)
	}

	padding := width - visLen

	switch align {
	case AlignLeft:
		return text + strings.Repeat(" ", padding)
	case AlignRight:
		return strings.Repeat(" ", padding) + text
	case AlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	}
	return text + strings.Repeat(" ", padding)
}

// Truncate truncates text to maxLen, adding ellipsis if needed.
func Truncate(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if VisibleLen(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return truncateVisible(text, maxLen)
	}
	return truncateVisible(text, maxLen-1) + "…"
}

// FormatDuration formats duration in human-readable form.
func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "-"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// FormatDurationShort formats duration in compact form.
func FormatDurationShort(d time.Duration) string {
	if d < 0 {
		return "-"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// FormatBytes formats bytes in human-readable form.
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatBytesShort formats bytes in compact form.
func FormatBytesShort(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"K", "M", "G", "T", "P"}
	value := float64(bytes) / float64(div)

	if value >= 100 {
		return fmt.Sprintf("%.0f%s", value, units[exp])
	}
	if value >= 10 {
		return fmt.Sprintf("%.1f%s", value, units[exp])
	}
	return fmt.Sprintf("%.2f%s", value, units[exp])
}

// FormatBytesPerSec formats bytes per second.
func FormatBytesPerSec(bytes uint64) string {
	return FormatBytesShort(bytes) + "/s"
}

// FormatPercent formats a percentage.
func FormatPercent(percent float64) string {
	if percent < 0 {
		return "-%"
	}
	if percent >= 100 {
		return "100%"
	}
	if percent >= 10 {
		return fmt.Sprintf("%.0f%%", percent)
	}
	return fmt.Sprintf("%.1f%%", percent)
}

// FormatSpeed formats network speed in bits per second.
func FormatSpeed(bitsPerSec uint64) string {
	if bitsPerSec == 0 {
		return "-"
	}

	const unit = 1000 // Network uses 1000, not 1024.
	if bitsPerSec < unit {
		return fmt.Sprintf("%d bps", bitsPerSec)
	}

	div, exp := uint64(unit), 0
	for n := bitsPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"Kbps", "Mbps", "Gbps", "Tbps"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	value := float64(bitsPerSec) / float64(div)
	if value >= 10 {
		return fmt.Sprintf("%.0f %s", value, units[exp])
	}
	return fmt.Sprintf("%.1f %s", value, units[exp])
}

// FormatSpeedShort formats network speed in compact form.
func FormatSpeedShort(bitsPerSec uint64) string {
	if bitsPerSec == 0 {
		return "-"
	}

	const unit = 1000
	if bitsPerSec < unit {
		return fmt.Sprintf("%dbps", bitsPerSec)
	}

	div, exp := uint64(unit), 0
	for n := bitsPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"K", "M", "G", "T"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	value := float64(bitsPerSec) / float64(div)
	suffix := units[exp]
	if value >= 10 {
		return fmt.Sprintf("%.0f%s", value, suffix)
	}
	return fmt.Sprintf("%.1f%s", value, suffix)
}

// RepeatString repeats a string n times.
func RepeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, n)
}

// JoinWithSep joins strings with separator, skipping empty strings.
func JoinWithSep(sep string, parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}
