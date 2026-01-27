// Package widget provides reusable TUI components.
package widget

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// durationHoursPerDay is the number of hours in a day.
	durationHoursPerDay int = 24
	// durationMinutesPerHour is the number of minutes in an hour.
	durationMinutesPerHour int = 60
	// durationSecondsPerMinute is the number of seconds in a minute.
	durationSecondsPerMinute int = 60
	// byteUnit is the number of bytes in a kilobyte (1024).
	byteUnit uint64 = 1024
	// networkUnit is the number of bits per kilobit (1000 for network).
	networkUnit uint64 = 1000
	// truncateEllipsisLength is the minimum length for ellipsis truncation.
	truncateEllipsisLength int = 3
	// formatBytesDecimalPlaces is the decimal precision for byte formatting.
	formatBytesDecimalPlaces int = 1
	// formatSpeedThreshold is the threshold for decimal places in speed formatting.
	formatSpeedThreshold float64 = 10.0
	// formatPercentThresholdMax is the maximum percent value.
	formatPercentThresholdMax float64 = 100.0
	// formatPercentThresholdDecimal is the threshold for decimal display.
	formatPercentThresholdDecimal float64 = 10.0
	// formatValueThreshold100 is the threshold for large values.
	formatValueThreshold100 float64 = 100.0
	// formatValueThreshold10 is the threshold for medium values.
	formatValueThreshold10 float64 = 10.0
	// formatValuePrecision0 is zero decimal places.
	formatValuePrecision0 int = 0
	// formatValuePrecision1 is one decimal place.
	formatValuePrecision1 int = 1
	// formatValuePrecision2 is two decimal places.
	formatValuePrecision2 int = 2
	// growSizeDuration is the pre-allocation size for duration strings.
	growSizeDuration int = 8
	// growSizeDurationShort is the pre-allocation size for short duration strings.
	growSizeDurationShort int = 6
	// bufferSize32 is the buffer size for numeric conversions.
	bufferSize32 int = 32
	// byteUnitCount is the number of byte unit suffixes (KB through ZB).
	byteUnitCount int = 7
	// speedUnitCount is the number of speed unit suffixes (K through T).
	speedUnitCount int = 4
	// centerPaddingSides is the divisor for splitting center padding.
	centerPaddingSides int = 2
	// strconvDecimalBase is the decimal base for strconv functions.
	strconvDecimalBase int = 10
	// strconvBitSize64 is the 64-bit size for strconv functions.
	strconvBitSize64 int = 64
	// maxCachedSpaces is the maximum number of cached space strings.
	maxCachedSpaces int = 256
	// maxCachedBars is the maximum width for horizontal bars.
	maxCachedBars int = 200
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
// Pre-computed strings for common padding operations to avoid allocations.
var (
	byteUnitsLong  [byteUnitCount]string  = [...]string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB"}
	byteUnitsShort [byteUnitCount]string  = [...]string{"K", "M", "G", "T", "P", "E", "Z"}
	speedUnitsLong [speedUnitCount]string = [...]string{"Kbps", "Mbps", "Gbps", "Tbps"}
	speedUnits     [speedUnitCount]string = [...]string{"K", "M", "G", "T"}
	// spacesCache holds pre-computed strings of spaces for efficient padding.
	spacesCache [maxCachedSpaces + 1]string = initSpacesCache()
	// barsCache holds pre-computed horizontal bar strings for efficient rendering.
	barsCache [maxCachedBars + 1]string = initBarsCache()
)

// initSpacesCache initializes the spaces cache array.
//
// Returns:
//   - [maxCachedSpaces + 1]string: array of pre-computed space strings.
func initSpacesCache() [maxCachedSpaces + 1]string {
	var cache [maxCachedSpaces + 1]string
	// Pre-compute strings of spaces for efficient padding.
	for i := range cache {
		cache[i] = strings.Repeat(" ", i)
	}
	// Return populated cache.
	return cache
}

// initBarsCache initializes the horizontal bars cache array.
//
// Returns:
//   - [maxCachedBars + 1]string: array of pre-computed bar strings.
func initBarsCache() [maxCachedBars + 1]string {
	var cache [maxCachedBars + 1]string
	// Pre-compute horizontal bar strings for efficient rendering.
	for i := range cache {
		cache[i] = strings.Repeat("─", i)
	}
	// Return populated cache.
	return cache
}

// Spaces returns a string of n spaces, using cache for common sizes.
//
// Params:
//   - n: the number of spaces to return.
//
// Returns:
//   - string: a string containing n spaces.
func Spaces(n int) string {
	// Handle zero or negative length.
	if n <= 0 {
		// Return empty for invalid input.
		return ""
	}
	// Use cached value if available.
	if n <= maxCachedSpaces {
		// Return from cache.
		return spacesCache[n]
	}
	// Generate string for larger sizes.
	return strings.Repeat(" ", n)
}

// HorizontalBar returns a string of n horizontal bar characters.
// Uses cache for common sizes.
//
// Params:
//   - n: the number of bar characters to return.
//
// Returns:
//   - string: a string containing n horizontal bar characters.
func HorizontalBar(n int) string {
	// Handle zero or negative length.
	if n <= 0 {
		// Return empty for invalid input.
		return ""
	}
	// Use cached value if available.
	if n <= maxCachedBars {
		// Return from cache.
		return barsCache[n]
	}
	// Generate string for larger sizes.
	return strings.Repeat("─", n)
}

// Pad pads text to width with specified alignment.
// Uses cached space strings for efficiency.
//
// Params:
//   - text: the text to pad.
//   - width: the target width.
//   - align: the alignment direction.
//
// Returns:
//   - string: the padded text.
func Pad(text string, width int, align Align) string {
	visLen := VisibleLen(text)
	// Truncate if text exceeds width.
	if visLen >= width {
		// Return truncated text.
		return truncateVisible(text, width)
	}

	padding := width - visLen

	// Apply padding based on alignment.
	switch align {
	// Left alignment.
	case AlignLeft:
		// Pad to the right.
		return text + Spaces(padding)
	// Right alignment.
	case AlignRight:
		// Pad to the left.
		return Spaces(padding) + text
	// Center alignment.
	case AlignCenter:
		// Pad both sides.
		left := padding / centerPaddingSides
		right := padding - left
		// Return centered text.
		return Spaces(left) + text + Spaces(right)
	}
	// Default to left alignment.
	return text + Spaces(padding)
}

// Truncate truncates text to maxLen, adding ellipsis if needed.
//
// Params:
//   - text: the text to truncate.
//   - maxLen: the maximum visible length.
//
// Returns:
//   - string: the truncated text with optional ellipsis.
func Truncate(text string, maxLen int) string {
	// Handle zero or negative length.
	if maxLen <= 0 {
		// Return empty for invalid length.
		return ""
	}
	// Return as-is if within limit.
	if VisibleLen(text) <= maxLen {
		// No truncation needed.
		return text
	}
	// No room for ellipsis, just truncate.
	if maxLen <= truncateEllipsisLength {
		// Truncate without ellipsis.
		return truncateVisible(text, maxLen)
	}
	// Add ellipsis after truncation.
	return truncateVisible(text, maxLen-1) + "…"
}

// FormatDuration formats duration in human-readable form.
// Uses strings.Builder for efficient string concatenation.
//
// Params:
//   - d: the duration to format.
//
// Returns:
//   - string: the human-readable duration string.
func FormatDuration(d time.Duration) string {
	// Handle negative duration.
	if d < 0 {
		// Return placeholder for negative.
		return "-"
	}

	// Convert duration to time components.
	days := int(d.Hours()) / durationHoursPerDay
	hours := int(d.Hours()) % durationHoursPerDay
	minutes := int(d.Minutes()) % durationMinutesPerHour
	seconds := int(d.Seconds()) % durationSecondsPerMinute

	var sb strings.Builder
	sb.Grow(growSizeDuration)

	// Format based on largest non-zero component.
	switch {
	// Days and hours.
	case days > 0:
		sb.WriteString(strconv.Itoa(days))
		sb.WriteString("d ")
		sb.WriteString(strconv.Itoa(hours))
		sb.WriteByte('h')
	// Hours and minutes.
	case hours > 0:
		sb.WriteString(strconv.Itoa(hours))
		sb.WriteString("h ")
		sb.WriteString(strconv.Itoa(minutes))
		sb.WriteByte('m')
	// Minutes and seconds.
	case minutes > 0:
		sb.WriteString(strconv.Itoa(minutes))
		sb.WriteString("m ")
		sb.WriteString(strconv.Itoa(seconds))
		sb.WriteByte('s')
	// Seconds only.
	default:
		sb.WriteString(strconv.Itoa(seconds))
		sb.WriteByte('s')
	}
	// Return formatted duration.
	return sb.String()
}

// FormatDurationShort formats duration in compact form.
// Uses strings.Builder for efficient string concatenation.
//
// Params:
//   - d: the duration to format.
//
// Returns:
//   - string: the compact duration string.
func FormatDurationShort(d time.Duration) string {
	// Handle negative duration.
	if d < 0 {
		// Return placeholder for negative.
		return "-"
	}

	// Convert duration to time components.
	days := int(d.Hours()) / durationHoursPerDay
	hours := int(d.Hours()) % durationHoursPerDay
	minutes := int(d.Minutes()) % durationMinutesPerHour
	seconds := int(d.Seconds()) % durationSecondsPerMinute

	var sb strings.Builder
	sb.Grow(growSizeDurationShort)

	// Format based on largest non-zero component.
	switch {
	// Days and hours without spaces.
	case days > 0:
		sb.WriteString(strconv.Itoa(days))
		sb.WriteByte('d')
		sb.WriteString(strconv.Itoa(hours))
		sb.WriteByte('h')
	// Hours and minutes without spaces.
	case hours > 0:
		sb.WriteString(strconv.Itoa(hours))
		sb.WriteByte('h')
		sb.WriteString(strconv.Itoa(minutes))
		sb.WriteByte('m')
	// Minutes only.
	case minutes > 0:
		sb.WriteString(strconv.Itoa(minutes))
		sb.WriteByte('m')
	// Seconds only.
	default:
		sb.WriteString(strconv.Itoa(seconds))
		sb.WriteByte('s')
	}
	// Return formatted duration.
	return sb.String()
}

// FormatBytes formats bytes in human-readable form.
// Uses package-level unit array to avoid per-call allocation.
//
// Params:
//   - bytes: the byte count to format.
//
// Returns:
//   - string: the human-readable byte string.
func FormatBytes(bytes uint64) string {
	// Return bytes directly if less than 1KB.
	if bytes < byteUnit {
		// Return raw byte count.
		return strconv.FormatUint(bytes, strconvDecimalBase) + " B"
	}

	// Calculate appropriate unit and divisor.
	div, exp := byteUnit, 0
	// Iterate to find appropriate unit.
	for n := bytes / byteUnit; n >= byteUnit; n /= byteUnit {
		div *= byteUnit
		exp++
	}

	// Clamp to maximum unit index.
	if exp >= len(byteUnitsLong) {
		exp = len(byteUnitsLong) - 1
	}
	// Use strconv.AppendFloat to avoid fmt.Sprintf allocations.
	var buf [bufferSize32]byte
	b := strconv.AppendFloat(buf[:0], float64(bytes)/float64(div), 'f', formatBytesDecimalPlaces, strconvBitSize64)
	b = append(b, ' ')
	b = append(b, byteUnitsLong[exp]...)
	// Return formatted bytes.
	return string(b)
}

// FormatBytesShort formats bytes in compact form.
// Uses package-level unit array to avoid per-call allocation.
//
// Params:
//   - bytes: the byte count to format.
//
// Returns:
//   - string: the compact byte string.
func FormatBytesShort(bytes uint64) string {
	// Return bytes directly if less than 1KB.
	if bytes < byteUnit {
		// Return raw byte count.
		return strconv.FormatUint(bytes, strconvDecimalBase) + "B"
	}

	// Calculate appropriate unit and divisor.
	div, exp := byteUnit, 0
	// Iterate to find appropriate unit.
	for n := bytes / byteUnit; n >= byteUnit; n /= byteUnit {
		div *= byteUnit
		exp++
	}

	// Clamp to maximum unit index.
	if exp >= len(byteUnitsShort) {
		exp = len(byteUnitsShort) - 1
	}
	value := float64(bytes) / float64(div)

	// Use strconv.AppendFloat to avoid fmt.Sprintf allocations.
	var buf [bufferSize32]byte
	var b []byte
	// Adjust precision based on value magnitude.
	switch {
	// Large values: no decimals.
	case value >= formatValueThreshold100:
		b = strconv.AppendFloat(buf[:0], value, 'f', formatValuePrecision0, strconvBitSize64)
	// Medium values: one decimal.
	case value >= formatValueThreshold10:
		b = strconv.AppendFloat(buf[:0], value, 'f', formatValuePrecision1, strconvBitSize64)
	// Small values: two decimals.
	default:
		b = strconv.AppendFloat(buf[:0], value, 'f', formatValuePrecision2, strconvBitSize64)
	}
	b = append(b, byteUnitsShort[exp]...)
	// Return formatted bytes.
	return string(b)
}

// FormatBytesPerSec formats bytes per second.
//
// Params:
//   - bytes: the bytes per second to format.
//
// Returns:
//   - string: the formatted bytes per second string.
func FormatBytesPerSec(bytes uint64) string {
	// Return formatted bytes with per-second suffix.
	return FormatBytesShort(bytes) + "/s"
}

// FormatPercent formats a percentage.
// Uses strconv for integer-like values to reduce allocations.
//
// Params:
//   - percent: the percentage value to format.
//
// Returns:
//   - string: the formatted percentage string.
func FormatPercent(percent float64) string {
	// Handle negative percent.
	if percent < 0 {
		// Return placeholder for negative.
		return "-%"
	}
	// Cap at 100%.
	if percent >= formatPercentThresholdMax {
		// Return maximum percent.
		return "100%"
	}
	// Integer format for values >= 10.
	if percent >= formatPercentThresholdDecimal {
		// Return integer format.
		return strconv.Itoa(int(percent)) + "%"
	}
	// One decimal for small percentages.
	return fmt.Sprintf("%.1f%%", percent)
}

// FormatSpeed formats network speed in bits per second.
// Uses package-level unit array to avoid per-call allocation.
//
// Params:
//   - bitsPerSec: the bits per second to format.
//
// Returns:
//   - string: the human-readable speed string.
func FormatSpeed(bitsPerSec uint64) string {
	// Handle zero speed.
	if bitsPerSec == 0 {
		// Return placeholder for zero.
		return "-"
	}

	// Return bits directly if less than 1Kbps.
	if bitsPerSec < networkUnit {
		// Return raw bit count.
		return strconv.FormatUint(bitsPerSec, strconvDecimalBase) + " bps"
	}

	// Calculate appropriate unit and divisor.
	div, exp := networkUnit, 0
	// Iterate to find appropriate unit.
	for n := bitsPerSec / networkUnit; n >= networkUnit; n /= networkUnit {
		div *= networkUnit
		exp++
	}

	// Clamp to maximum unit index.
	if exp >= len(speedUnitsLong) {
		exp = len(speedUnitsLong) - 1
	}

	value := float64(bitsPerSec) / float64(div)
	// Format with appropriate decimal places.
	if value >= formatSpeedThreshold {
		// Return integer format for large values.
		return fmt.Sprintf("%.0f %s", value, speedUnitsLong[exp])
	}
	// Return decimal format for small values.
	return fmt.Sprintf("%.1f %s", value, speedUnitsLong[exp])
}

// FormatSpeedShort formats network speed in compact form.
// Uses package-level unit array to avoid per-call allocation.
//
// Params:
//   - bitsPerSec: the bits per second to format.
//
// Returns:
//   - string: the compact speed string.
func FormatSpeedShort(bitsPerSec uint64) string {
	// Handle zero speed.
	if bitsPerSec == 0 {
		// Return placeholder for zero.
		return "-"
	}

	// Return bits directly if less than 1Kbps.
	if bitsPerSec < networkUnit {
		// Return raw bit count.
		return strconv.FormatUint(bitsPerSec, strconvDecimalBase) + "bps"
	}

	// Calculate appropriate unit and divisor.
	div, exp := networkUnit, 0
	// Iterate to find appropriate unit.
	for n := bitsPerSec / networkUnit; n >= networkUnit; n /= networkUnit {
		div *= networkUnit
		exp++
	}

	// Clamp to maximum unit index.
	if exp >= len(speedUnits) {
		exp = len(speedUnits) - 1
	}

	value := float64(bitsPerSec) / float64(div)
	suffix := speedUnits[exp]
	// Format with appropriate decimal places.
	if value >= formatSpeedThreshold {
		// Return integer format for large values.
		return fmt.Sprintf("%.0f%s", value, suffix)
	}
	// Return decimal format for small values.
	return fmt.Sprintf("%.1f%s", value, suffix)
}

// RepeatString repeats a string n times.
//
// Params:
//   - s: the string to repeat.
//   - n: the number of repetitions.
//
// Returns:
//   - string: the repeated string.
func RepeatString(s string, n int) string {
	// Handle zero or negative count.
	if n <= 0 {
		// Return empty for invalid count.
		return ""
	}
	// Return repeated string.
	return strings.Repeat(s, n)
}

// TruncateRunes truncates a string to maxRunes runes, adding suffix if truncated.
// This is UTF-8 safe, unlike byte-slicing which can corrupt multi-byte characters.
//
// Params:
//   - s: the string to truncate.
//   - maxRunes: the maximum number of runes.
//   - suffix: the suffix to append when truncated.
//
// Returns:
//   - string: the truncated string with optional suffix.
func TruncateRunes(s string, maxRunes int, suffix string) string {
	// Handle zero or negative length.
	if maxRunes <= 0 {
		// Return empty for invalid length.
		return ""
	}

	runes := []rune(s)
	// Return as-is if within limit.
	if len(runes) <= maxRunes {
		// No truncation needed.
		return s
	}

	suffixRunes := []rune(suffix)
	// No room for suffix, just truncate.
	if maxRunes <= len(suffixRunes) {
		// Truncate suffix itself to fit.
		return string(suffixRunes[:maxRunes])
	}

	// Add suffix after truncation.
	return string(runes[:maxRunes-len(suffixRunes)]) + suffix
}

// PadRight pads a string to the right with spaces (plain text).
// Uses cached space strings for efficiency.
//
// Params:
//   - s: the string to pad.
//   - width: the target width.
//
// Returns:
//   - string: the right-padded string.
func PadRight(s string, width int) string {
	// No padding needed.
	if len(s) >= width {
		// Return original string.
		return s
	}
	// Append spaces to reach width.
	return s + Spaces(width-len(s))
}

// PadLeft pads a string to the left with spaces (plain text).
// Uses cached space strings for efficiency.
//
// Params:
//   - s: the string to pad.
//   - width: the target width.
//
// Returns:
//   - string: the left-padded string.
func PadLeft(s string, width int) string {
	// No padding needed.
	if len(s) >= width {
		// Return original string.
		return s
	}
	// Prepend spaces to reach width.
	return Spaces(width-len(s)) + s
}

// PadRightAnsi pads an ANSI-colored string to the right based on visible length.
// Uses cached space strings for efficiency.
//
// Params:
//   - s: the ANSI-colored string to pad.
//   - width: the target visible width.
//
// Returns:
//   - string: the right-padded string.
func PadRightAnsi(s string, width int) string {
	visLen := VisibleLen(s)
	// No padding needed.
	if visLen >= width {
		// Return original string.
		return s
	}
	// Append spaces based on visible length.
	return s + Spaces(width-visLen)
}

// PadLeftAnsi pads an ANSI-colored string to the left based on visible length.
// Uses cached space strings for efficiency.
//
// Params:
//   - s: the ANSI-colored string to pad.
//   - width: the target visible width.
//
// Returns:
//   - string: the left-padded string.
func PadLeftAnsi(s string, width int) string {
	visLen := VisibleLen(s)
	// No padding needed.
	if visLen >= width {
		// Return original string.
		return s
	}
	// Prepend spaces based on visible length.
	return Spaces(width-visLen) + s
}

// JoinWithSep joins strings with separator, skipping empty strings.
//
// Params:
//   - sep: the separator string.
//   - parts: the strings to join.
//
// Returns:
//   - string: the joined string.
func JoinWithSep(sep string, parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	// Filter out empty strings.
	for _, p := range parts {
		// Only include non-empty parts.
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	// Join filtered parts with separator.
	return strings.Join(nonEmpty, sep)
}
