// Package widget provides reusable TUI components.
package widget

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

// sparkCharWidth is the estimated byte width of sparkline UTF-8 characters.
const sparkCharWidth int = 4

// SparkLine renders a mini sparkline chart.
// It visualizes a series of values using 8-level Unicode block characters.
type SparkLine struct {
	Values []float64
	Width  int
	Color  string
}

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
	// check for empty value.
	if len(s.Values) == 0 {
		// return computed result.
		return strings.Repeat(" ", s.Width)
	}

	minVal, maxVal := s.findMinMax()
	valueRange := s.calculateRange(minVal, maxVal)
	start := s.calculateStartIndex()

	var sb strings.Builder
	sb.Grow(s.Width*sparkCharWidth + preAllocGrowSize)
	sb.WriteString(s.Color)
	s.writeSparkChars(&sb, minVal, valueRange, start)
	s.writePadding(&sb, start)
	sb.WriteString(ansi.Reset)

	// return computed result.
	return sb.String()
}

// findMinMax returns the minimum and maximum values in the sparkline.
//
// Returns:
//   - minVal: smallest value in the dataset.
//   - maxVal: largest value in the dataset.
func (s *SparkLine) findMinMax() (minVal, maxVal float64) {
	minVal, maxVal = s.Values[0], s.Values[0]
	// iterate over collection.
	for _, v := range s.Values {
		// evaluate condition.
		if v < minVal {
			minVal = v
		}
		// evaluate condition.
		if v > maxVal {
			maxVal = v
		}
	}
	// return computed result.
	return minVal, maxVal
}

// calculateRange computes the value range, avoiding division by zero.
//
// Params:
//   - minVal: minimum value in the dataset.
//   - maxVal: maximum value in the dataset.
//
// Returns:
//   - float64: range between minVal and maxVal, or 1 if equal.
func (s *SparkLine) calculateRange(minVal, maxVal float64) float64 {
	valueRange := maxVal - minVal
	// check for empty value.
	if valueRange == 0 {
		valueRange = 1
	}
	// return computed result.
	return valueRange
}

// calculateStartIndex determines where to start reading values for width fitting.
//
// Returns:
//   - int: index to start reading from to fit within width.
func (s *SparkLine) calculateStartIndex() int {
	// evaluate condition.
	if len(s.Values) > s.Width {
		// return computed result.
		return len(s.Values) - s.Width
	}
	// return computed result.
	return 0
}

// writeSparkChars writes the spark characters to the builder.
//
// Params:
//   - sb: string builder to write spark characters to.
//   - minVal: minimum value for normalization.
//   - valueRange: range of values for normalization.
//   - start: index to start reading values from.
func (s *SparkLine) writeSparkChars(sb *strings.Builder, minVal, valueRange float64, start int) {
	// execute loop.
	for i := start; i < len(s.Values); i++ {
		idx := s.valueToSparkIndex(s.Values[i], minVal, valueRange)
		sb.WriteString(Sparks[idx])
	}
}

// valueToSparkIndex converts a value to a spark character index.
//
// Params:
//   - value: the value to convert.
//   - minVal: minimum value for normalization.
//   - valueRange: range of values for normalization.
//
// Returns:
//   - int: index into Sparks array (0-7).
func (s *SparkLine) valueToSparkIndex(value, minVal, valueRange float64) int {
	normalized := (value - minVal) / valueRange
	idx := int(normalized * float64(len(Sparks)-1))
	idx = max(idx, 0)
	// evaluate condition.
	if idx >= len(Sparks) {
		idx = len(Sparks) - 1
	}
	// return computed result.
	return idx
}

// writePadding adds trailing spaces if fewer values than width.
//
// Params:
//   - sb: string builder to write padding to.
//   - start: starting index used to calculate rendered count.
func (s *SparkLine) writePadding(sb *strings.Builder, start int) {
	rendered := len(s.Values) - start
	// evaluate condition.
	if rendered < s.Width {
		sb.WriteString(strings.Repeat(" ", s.Width-rendered))
	}
}
