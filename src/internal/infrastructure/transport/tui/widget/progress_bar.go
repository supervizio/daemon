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
)

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
	sb.Grow(len(p.Label) + p.Width + preAllocGrowSize)

	p.writeLabel(&sb)
	sb.WriteString(p.Style.Left)
	barWidth := p.calculateBarWidth()
	p.writeBarContent(&sb, barWidth)
	sb.WriteString(p.Style.Right)
	p.writePercentValue(&sb)

	// return computed result.
	return sb.String()
}

// writeLabel writes the label prefix to the builder if present.
//
// Params:
//   - sb: string builder to write the label to.
func (p *ProgressBar) writeLabel(sb *strings.Builder) {
	// evaluate condition.
	if p.Label != "" {
		sb.WriteString(p.Label)
		sb.WriteString(" ")
	}
}

// calculateBarWidth computes the inner bar width excluding brackets.
//
// Returns:
//   - int: inner width available for the bar content.
func (p *ProgressBar) calculateBarWidth() int {
	barWidth := p.Width
	// evaluate condition.
	if p.Style.Left != "" {
		barWidth--
	}
	// evaluate condition.
	if p.Style.Right != "" {
		barWidth--
	}
	// evaluate condition.
	if barWidth < 1 {
		barWidth = 1
	}
	// return computed result.
	return barWidth
}

// writeBarContent writes the filled and empty portions of the bar.
//
// Params:
//   - sb: string builder to write the bar content to.
//   - barWidth: width of the bar area in characters.
func (p *ProgressBar) writeBarContent(sb *strings.Builder, barWidth int) {
	fullChars, partialEighths, emptyChars := p.calculateFillUnits(barWidth)

	sb.WriteString(p.Color)
	// check for positive value.
	if fullChars > 0 {
		sb.WriteString(strings.Repeat(p.Style.Full, fullChars))
	}
	// check for positive value.
	if partialEighths > 0 {
		sb.WriteString(SubBlockChars[partialEighths])
	}
	sb.WriteString(ansi.Reset)
	// check for positive value.
	if emptyChars > 0 {
		sb.WriteString(strings.Repeat(p.Style.Empty, emptyChars))
	}
}

// calculateFillUnits computes full, partial, and empty character counts.
//
// Params:
//   - barWidth: total width of the bar in characters.
//
// Returns:
//   - fullChars: number of fully filled characters.
//   - partialEighths: sub-block index for partial fill (0-7).
//   - emptyChars: number of empty characters.
func (p *ProgressBar) calculateFillUnits(barWidth int) (fullChars, partialEighths, emptyChars int) {
	totalSubUnits := barWidth * subBlockLevels
	filledSubUnits := int(float64(totalSubUnits) * p.Percent / percentMax)

	fullChars = filledSubUnits / subBlockLevels
	partialEighths = filledSubUnits % subBlockLevels
	emptyChars = barWidth - fullChars
	// check for positive value.
	if partialEighths > 0 {
		emptyChars--
	}
	// return computed result.
	return fullChars, partialEighths, emptyChars
}

// writePercentValue writes the percentage value to the builder if enabled.
//
// Params:
//   - sb: string builder to write the percentage to.
func (p *ProgressBar) writePercentValue(sb *strings.Builder) {
	// evaluate condition.
	if !p.ShowValue {
		// return early when value display is disabled.
		return
	}
	sb.WriteByte(' ')
	pct := int(p.Percent)
	// evaluate condition.
	if pct < percentDigitPadding1 {
		sb.WriteString("  ")
		// handle double-digit padding.
	} else if pct < percentDigitPadding2 {
		sb.WriteByte(' ')
	}
	sb.WriteString(strconv.Itoa(pct))
	sb.WriteByte('%')
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
