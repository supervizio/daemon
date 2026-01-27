// Package widget provides reusable TUI components.
package widget

import (
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

const (
	// percentThresholdCritical is the percentage at which bars turn red.
	percentThresholdCritical float64 = 90.0
	// percentThresholdWarning is the percentage at which bars turn yellow.
	percentThresholdWarning float64 = 70.0
	// percentMin is the minimum valid percentage value.
	percentMin float64 = 0.0
	// percentMax is the maximum valid percentage value.
	percentMax float64 = 100.0
	// subBlockLevels is the number of granularity levels per character (1/8th).
	subBlockLevels int = 8
	// percentDigitPadding1 is the single digit padding width.
	percentDigitPadding1 int = 10
	// percentDigitPadding2 is the double digit padding width.
	percentDigitPadding2 int = 100
	// preAllocGrowSize is the pre-allocation size for string builder.
	preAllocGrowSize int = 20
	// sparkCharWidth is the estimated byte width of sparkline UTF-8 characters.
	sparkCharWidth int = 4
)

// BarStyle defines the characters for a progress bar.
// It specifies the fill, empty, and bracket characters used in rendering.
type BarStyle struct {
	Full  string
	Empty string
	Left  string
	Right string
}

// BlockBar uses block characters.
var BlockBar BarStyle = BarStyle{
	Full:  "█",
	Empty: " ",
	Left:  "",
	Right: "",
}

// BracketBar uses brackets with block fill.
var BracketBar BarStyle = BarStyle{
	Full:  "█",
	Empty: " ",
	Left:  "[",
	Right: "]",
}

// ASCIIBar uses ASCII characters only.
var ASCIIBar BarStyle = BarStyle{
	Full:  "#",
	Empty: "-",
	Left:  "[",
	Right: "]",
}

// SubBlockChars provides 1/8th granularity for progress bars.
// Index 0 = empty, 1-7 = partial fills, 8 = full.
var SubBlockChars []string = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

// ProgressBar renders a progress bar.
// It supports customizable styles, colors, labels, and 1/8th character granularity.
type ProgressBar struct {
	Style      BarStyle
	Width      int
	Percent    float64
	Label      string
	ShowValue  bool
	Color      string
	EmptyColor string
}

// NewProgressBar creates a new progress bar.
//
// Params:
//   - width: total width of the progress bar including brackets.
//   - percent: percentage value (0-100) to display.
//
// Returns:
//   - *ProgressBar: configured progress bar with default style.
func NewProgressBar(width int, percent float64) *ProgressBar {
	// Return configured progress bar with defaults.
	return &ProgressBar{
		Style:      BracketBar,
		Width:      width,
		Percent:    clamp(percent, percentMin, percentMax),
		ShowValue:  true,
		Color:      ansi.FgGreen,
		EmptyColor: ansi.FgGray,
	}
}

// SetLabel sets the bar label.
//
// Params:
//   - label: text to display before the progress bar.
//
// Returns:
//   - *ProgressBar: self for method chaining.
func (p *ProgressBar) SetLabel(label string) *ProgressBar {
	p.Label = label
	// Return self for method chaining.
	return p
}

// SetStyle sets the bar style.
//
// Params:
//   - style: BarStyle defining fill, empty, and bracket characters.
//
// Returns:
//   - *ProgressBar: self for method chaining.
func (p *ProgressBar) SetStyle(style BarStyle) *ProgressBar {
	p.Style = style
	// Return self for method chaining.
	return p
}

// SetColor sets the fill color.
//
// Params:
//   - color: ANSI color code for the filled portion.
//
// Returns:
//   - *ProgressBar: self for method chaining.
func (p *ProgressBar) SetColor(color string) *ProgressBar {
	p.Color = color
	// Return self for method chaining.
	return p
}

// SetColorByPercent sets color based on percentage thresholds.
// Green for normal (<70%), yellow for warning (70-90%), red for critical (>=90%).
//
// Returns:
//   - *ProgressBar: self for method chaining.
func (p *ProgressBar) SetColorByPercent() *ProgressBar {
	// Determine color based on utilization thresholds.
	switch {
	// Critical level: red color.
	case p.Percent >= percentThresholdCritical:
		p.Color = ansi.FgRed
	// Warning level: yellow color.
	case p.Percent >= percentThresholdWarning:
		p.Color = ansi.FgYellow
	// Normal level: green color.
	default:
		p.Color = ansi.FgGreen
	}
	// Return self for method chaining.
	return p
}

// Render returns the progress bar as a string with 1/8th character granularity.
//
// Returns:
//   - string: rendered progress bar with optional label and percentage.
func (p *ProgressBar) Render() string {
	var sb strings.Builder
	// Pre-allocate for typical bar: label + bracket + bar + bracket + value.
	sb.Grow(len(p.Label) + p.Width + preAllocGrowSize)

	// Add label prefix if present.
	if p.Label != "" {
		sb.WriteString(p.Label)
		sb.WriteString(" ")
	}

	// Add left bracket.
	sb.WriteString(p.Style.Left)

	// Calculate bar width excluding brackets.
	barWidth := p.Width
	// Account for left bracket if present.
	if p.Style.Left != "" {
		barWidth--
	}
	// Account for right bracket if present.
	if p.Style.Right != "" {
		barWidth--
	}
	// Ensure minimum bar width.
	if barWidth < 1 {
		barWidth = 1
	}

	// Calculate fill with 1/8th granularity (8 sub-levels per character).
	totalSubUnits := barWidth * subBlockLevels
	filledSubUnits := int(float64(totalSubUnits) * p.Percent / percentMax)

	// Split into full characters and partial character.
	fullChars := filledSubUnits / subBlockLevels
	partialEighths := filledSubUnits % subBlockLevels
	emptyChars := barWidth - fullChars
	// Adjust empty count if partial character present.
	if partialEighths > 0 {
		emptyChars--
	}

	// Render filled portion with color.
	sb.WriteString(p.Color)
	// Write full characters when present.
	if fullChars > 0 {
		sb.WriteString(strings.Repeat(p.Style.Full, fullChars))
	}

	// Render partial character if any.
	if partialEighths > 0 {
		sb.WriteString(SubBlockChars[partialEighths])
	}

	sb.WriteString(ansi.Reset)

	// Render empty portion without color.
	if emptyChars > 0 {
		sb.WriteString(strings.Repeat(p.Style.Empty, emptyChars))
	}

	// Add right bracket.
	sb.WriteString(p.Style.Right)

	// Add percentage value if enabled.
	if p.ShowValue {
		sb.WriteByte(' ')
		// Pad to 3 digits for alignment.
		pct := int(p.Percent)
		// Add leading spaces for single digit.
		if pct < percentDigitPadding1 {
			sb.WriteString("  ")
			// Add single space for double digit.
		} else if pct < percentDigitPadding2 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strconv.Itoa(pct))
		sb.WriteByte('%')
	}

	// Return the complete rendered bar.
	return sb.String()
}

// clamp restricts value to [min, max].
//
// Params:
//   - value: the input value to clamp.
//   - min: minimum allowed value.
//   - max: maximum allowed value.
//
// Returns:
//   - float64: clamped value within bounds.
func clamp(value, min, max float64) float64 {
	// Enforce minimum bound.
	if value < min {
		// Return minimum if value is below.
		return min
	}
	// Enforce maximum bound.
	if value > max {
		// Return maximum if value is above.
		return max
	}
	// Return unchanged value within bounds.
	return value
}

// SparkLine renders a mini sparkline chart.
// It visualizes a series of values using 8-level Unicode block characters.
type SparkLine struct {
	Values []float64
	Width  int
	Color  string
}

// Sparks are the sparkline characters (8 levels).
var Sparks []string = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// NewSparkLine creates a sparkline.
//
// Params:
//   - values: slice of float64 values to visualize.
//   - width: maximum width in characters for the sparkline.
//
// Returns:
//   - *SparkLine: configured sparkline with default color.
func NewSparkLine(values []float64, width int) *SparkLine {
	// Return configured sparkline with defaults.
	return &SparkLine{
		Values: values,
		Width:  width,
		Color:  ansi.FgCyan,
	}
}

// Render returns the sparkline as a string.
//
// Returns:
//   - string: rendered sparkline with color and padding.
func (s *SparkLine) Render() string {
	// Handle empty values with spaces.
	if len(s.Values) == 0 {
		// Return blank line for empty data.
		return strings.Repeat(" ", s.Width)
	}

	// Find min and max values for normalization.
	min, max := s.Values[0], s.Values[0]
	// Iterate through values to find extremes.
	for _, v := range s.Values {
		// Update minimum if smaller.
		if v < min {
			min = v
		}
		// Update maximum if larger.
		if v > max {
			max = v
		}
	}

	// Normalize values and render sparkline.
	var sb strings.Builder
	// Pre-allocate for sparkline chars + ANSI codes.
	sb.Grow(s.Width*sparkCharWidth + preAllocGrowSize)
	sb.WriteString(s.Color)

	// Calculate value range, avoiding division by zero.
	valueRange := max - min
	// Use 1 as range for flat data to avoid NaN.
	if valueRange == 0 {
		valueRange = 1
	}

	// Take last Width values if more available.
	start := 0
	// Truncate beginning to fit width.
	if len(s.Values) > s.Width {
		start = len(s.Values) - s.Width
	}

	// Convert each value to spark character.
	for i := start; i < len(s.Values); i++ {
		normalized := (s.Values[i] - min) / valueRange
		idx := int(normalized * float64(len(Sparks)-1))
		// Clamp index to valid range.
		if idx < 0 {
			idx = 0
		}
		// Prevent overflow beyond spark array.
		if idx >= len(Sparks) {
			idx = len(Sparks) - 1
		}
		sb.WriteString(Sparks[idx])
	}

	// Pad with spaces if fewer values than width.
	rendered := len(s.Values) - start
	// Fill remaining width with blanks.
	if rendered < s.Width {
		sb.WriteString(strings.Repeat(" ", s.Width-rendered))
	}

	sb.WriteString(ansi.Reset)
	// Return the complete sparkline.
	return sb.String()
}
