// Package widget provides reusable TUI components.
package widget

import (
	"fmt"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

// BarStyle defines the characters for a progress bar.
type BarStyle struct {
	Full  string
	Empty string
	Left  string
	Right string
}

// BlockBar uses block characters.
var BlockBar = BarStyle{
	Full:  "█",
	Empty: "░",
	Left:  "",
	Right: "",
}

// BracketBar uses brackets with block fill.
var BracketBar = BarStyle{
	Full:  "█",
	Empty: "░",
	Left:  "[",
	Right: "]",
}

// ASCIIBar uses ASCII characters only.
var ASCIIBar = BarStyle{
	Full:  "#",
	Empty: "-",
	Left:  "[",
	Right: "]",
}

// ProgressBar renders a progress bar.
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
func NewProgressBar(width int, percent float64) *ProgressBar {
	return &ProgressBar{
		Style:      BracketBar,
		Width:      width,
		Percent:    clamp(percent, 0, 100),
		ShowValue:  true,
		Color:      ansi.FgGreen,
		EmptyColor: ansi.FgGray,
	}
}

// SetLabel sets the bar label.
func (p *ProgressBar) SetLabel(label string) *ProgressBar {
	p.Label = label
	return p
}

// SetStyle sets the bar style.
func (p *ProgressBar) SetStyle(style BarStyle) *ProgressBar {
	p.Style = style
	return p
}

// SetColor sets the fill color.
func (p *ProgressBar) SetColor(color string) *ProgressBar {
	p.Color = color
	return p
}

// SetColorByPercent sets color based on percentage thresholds.
func (p *ProgressBar) SetColorByPercent() *ProgressBar {
	switch {
	case p.Percent >= 90:
		p.Color = ansi.FgRed
	case p.Percent >= 70:
		p.Color = ansi.FgYellow
	default:
		p.Color = ansi.FgGreen
	}
	return p
}

// Render returns the progress bar as a string.
func (p *ProgressBar) Render() string {
	var sb strings.Builder

	// Label.
	if p.Label != "" {
		sb.WriteString(p.Label)
		sb.WriteString(" ")
	}

	// Left bracket.
	sb.WriteString(p.Style.Left)

	// Calculate fill.
	barWidth := p.Width
	if p.Style.Left != "" {
		barWidth--
	}
	if p.Style.Right != "" {
		barWidth--
	}
	if barWidth < 1 {
		barWidth = 1
	}

	filled := int(float64(barWidth) * p.Percent / 100.0)
	empty := barWidth - filled

	// Fill.
	if filled > 0 {
		sb.WriteString(p.Color)
		sb.WriteString(strings.Repeat(p.Style.Full, filled))
	}

	// Empty.
	if empty > 0 {
		sb.WriteString(p.EmptyColor)
		sb.WriteString(strings.Repeat(p.Style.Empty, empty))
	}

	sb.WriteString(ansi.Reset)

	// Right bracket.
	sb.WriteString(p.Style.Right)

	// Percentage value.
	if p.ShowValue {
		sb.WriteString(fmt.Sprintf(" %3.0f%%", p.Percent))
	}

	return sb.String()
}

// clamp restricts value to [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// SparkLine renders a mini sparkline chart.
type SparkLine struct {
	Values []float64
	Width  int
	Color  string
}

// Sparks are the sparkline characters (8 levels).
var Sparks = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// NewSparkLine creates a sparkline.
func NewSparkLine(values []float64, width int) *SparkLine {
	return &SparkLine{
		Values: values,
		Width:  width,
		Color:  ansi.FgCyan,
	}
}

// Render returns the sparkline as a string.
func (s *SparkLine) Render() string {
	if len(s.Values) == 0 {
		return strings.Repeat(" ", s.Width)
	}

	// Find min/max.
	min, max := s.Values[0], s.Values[0]
	for _, v := range s.Values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Normalize and render.
	var sb strings.Builder
	sb.WriteString(s.Color)

	valueRange := max - min
	if valueRange == 0 {
		valueRange = 1
	}

	// Take last Width values.
	start := 0
	if len(s.Values) > s.Width {
		start = len(s.Values) - s.Width
	}

	for i := start; i < len(s.Values); i++ {
		normalized := (s.Values[i] - min) / valueRange
		idx := int(normalized * float64(len(Sparks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(Sparks) {
			idx = len(Sparks) - 1
		}
		sb.WriteString(Sparks[idx])
	}

	// Pad if needed.
	rendered := len(s.Values) - start
	if rendered < s.Width {
		sb.WriteString(strings.Repeat(" ", s.Width-rendered))
	}

	sb.WriteString(ansi.Reset)
	return sb.String()
}
