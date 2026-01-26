// Package widget provides reusable TUI components.
package widget

import (
	"fmt"
	"strconv"
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

// Unit suffixes at package level to avoid per-call allocation.
var (
	byteUnitsLong  = [...]string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB"}
	byteUnitsShort = [...]string{"K", "M", "G", "T", "P", "E", "Z"}
	speedUnitsLong = [...]string{"Kbps", "Mbps", "Gbps", "Tbps"}
	speedUnits     = [...]string{"K", "M", "G", "T"}
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
// Uses strconv instead of fmt.Sprintf for integer formatting.
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
		return strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h"
	case hours > 0:
		return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"
	case minutes > 0:
		return strconv.Itoa(minutes) + "m " + strconv.Itoa(seconds) + "s"
	default:
		return strconv.Itoa(seconds) + "s"
	}
}

// FormatDurationShort formats duration in compact form.
// Uses strconv instead of fmt.Sprintf for integer formatting.
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
		return strconv.Itoa(days) + "d" + strconv.Itoa(hours) + "h"
	case hours > 0:
		return strconv.Itoa(hours) + "h" + strconv.Itoa(minutes) + "m"
	case minutes > 0:
		return strconv.Itoa(minutes) + "m"
	default:
		return strconv.Itoa(seconds) + "s"
	}
}

// FormatBytes formats bytes in human-readable form.
// Uses package-level unit array to avoid per-call allocation.
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	if exp >= len(byteUnitsLong) {
		exp = len(byteUnitsLong) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), byteUnitsLong[exp])
}

// FormatBytesShort formats bytes in compact form.
// Uses package-level unit array to avoid per-call allocation.
func FormatBytesShort(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + "B"
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	if exp >= len(byteUnitsShort) {
		exp = len(byteUnitsShort) - 1
	}
	value := float64(bytes) / float64(div)

	if value >= 100 {
		return fmt.Sprintf("%.0f%s", value, byteUnitsShort[exp])
	}
	if value >= 10 {
		return fmt.Sprintf("%.1f%s", value, byteUnitsShort[exp])
	}
	return fmt.Sprintf("%.2f%s", value, byteUnitsShort[exp])
}

// FormatBytesPerSec formats bytes per second.
func FormatBytesPerSec(bytes uint64) string {
	return FormatBytesShort(bytes) + "/s"
}

// FormatPercent formats a percentage.
// Uses strconv for integer-like values to reduce allocations.
func FormatPercent(percent float64) string {
	if percent < 0 {
		return "-%"
	}
	if percent >= 100 {
		return "100%"
	}
	if percent >= 10 {
		return strconv.Itoa(int(percent)) + "%"
	}
	return fmt.Sprintf("%.1f%%", percent)
}

// FormatSpeed formats network speed in bits per second.
// Uses package-level unit array to avoid per-call allocation.
func FormatSpeed(bitsPerSec uint64) string {
	if bitsPerSec == 0 {
		return "-"
	}

	const unit = 1000 // Network uses 1000, not 1024.
	if bitsPerSec < unit {
		return strconv.FormatUint(bitsPerSec, 10) + " bps"
	}

	div, exp := uint64(unit), 0
	for n := bitsPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	if exp >= len(speedUnitsLong) {
		exp = len(speedUnitsLong) - 1
	}

	value := float64(bitsPerSec) / float64(div)
	if value >= 10 {
		return fmt.Sprintf("%.0f %s", value, speedUnitsLong[exp])
	}
	return fmt.Sprintf("%.1f %s", value, speedUnitsLong[exp])
}

// FormatSpeedShort formats network speed in compact form.
// Uses package-level unit array to avoid per-call allocation.
func FormatSpeedShort(bitsPerSec uint64) string {
	if bitsPerSec == 0 {
		return "-"
	}

	const unit = 1000
	if bitsPerSec < unit {
		return strconv.FormatUint(bitsPerSec, 10) + "bps"
	}

	div, exp := uint64(unit), 0
	for n := bitsPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	if exp >= len(speedUnits) {
		exp = len(speedUnits) - 1
	}

	value := float64(bitsPerSec) / float64(div)
	suffix := speedUnits[exp]
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

// TruncateRunes truncates a string to maxRunes runes, adding suffix if truncated.
// This is UTF-8 safe, unlike byte-slicing which can corrupt multi-byte characters.
func TruncateRunes(s string, maxRunes int, suffix string) string {
	if maxRunes <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}

	suffixRunes := []rune(suffix)
	if maxRunes <= len(suffixRunes) {
		return string(suffixRunes[:maxRunes])
	}

	return string(runes[:maxRunes-len(suffixRunes)]) + suffix
}

// PadRight pads a string to the right with spaces (plain text).
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PadLeft pads a string to the left with spaces (plain text).
func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// PadRightAnsi pads an ANSI-colored string to the right based on visible length.
func PadRightAnsi(s string, width int) string {
	visLen := VisibleLen(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// PadLeftAnsi pads an ANSI-colored string to the left based on visible length.
func PadLeftAnsi(s string, width int) string {
	visLen := VisibleLen(s)
	if visLen >= width {
		return s
	}
	return strings.Repeat(" ", width-visLen) + s
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
