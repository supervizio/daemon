package widget

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SparkLine_findMinMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		values  []float64
		wantMin float64
		wantMax float64
	}{
		{"ascending", []float64{1, 2, 3, 4, 5}, 1, 5},
		{"descending", []float64{5, 4, 3, 2, 1}, 1, 5},
		{"single", []float64{42}, 42, 42},
		{"same_values", []float64{5, 5, 5, 5}, 5, 5},
		{"negative", []float64{-10, -5, 0, 5, 10}, -10, 10},
		{"mixed", []float64{3, 1, 4, 1, 5, 9, 2, 6}, 1, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := &SparkLine{Values: tt.values}
			min, max := spark.findMinMax()
			assert.Equal(t, tt.wantMin, min)
			assert.Equal(t, tt.wantMax, max)
		})
	}
}

func Test_SparkLine_calculateRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		min       float64
		max       float64
		wantRange float64
	}{
		{"normal_range", 0, 100, 100},
		{"small_range", 10, 20, 10},
		{"zero_range", 50, 50, 1},
		{"negative_range", -50, 50, 100},
		{"decimal_range", 0.1, 0.9, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := &SparkLine{}
			result := spark.calculateRange(tt.min, tt.max)
			assert.InDelta(t, tt.wantRange, result, 0.0001)
		})
	}
}

func Test_SparkLine_calculateStartIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		values    []float64
		width     int
		wantStart int
	}{
		{"fewer_than_width", []float64{1, 2, 3}, 10, 0},
		{"equal_to_width", []float64{1, 2, 3, 4, 5}, 5, 0},
		{"more_than_width", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 5, 5},
		{"empty", []float64{}, 10, 0},
		{"single", []float64{42}, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := &SparkLine{Values: tt.values, Width: tt.width}
			result := spark.calculateStartIndex()
			assert.Equal(t, tt.wantStart, result)
		})
	}
}

func Test_SparkLine_writeSparkChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		values     []float64
		min        float64
		valueRange float64
		start      int
		wantLen    int
	}{
		{"ascending", []float64{0, 25, 50, 75, 100}, 0, 100, 0, 5},
		{"from_start", []float64{10, 20, 30, 40, 50}, 10, 40, 0, 5},
		{"partial_start", []float64{1, 2, 3, 4, 5, 6, 7, 8}, 1, 7, 3, 5},
		{"single", []float64{50}, 50, 1, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := &SparkLine{Values: tt.values}
			var sb strings.Builder
			spark.writeSparkChars(&sb, tt.min, tt.valueRange, tt.start)
			// Each spark char is multi-byte UTF-8.
			assert.NotEmpty(t, sb.String())
		})
	}
}

func Test_SparkLine_valueToSparkIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      float64
		min        float64
		valueRange float64
		wantMin    int
		wantMax    int
	}{
		{"min_value", 0, 0, 100, 0, 0},
		{"max_value", 100, 0, 100, 7, 7},
		{"mid_value", 50, 0, 100, 3, 4},
		{"negative_clamped", -50, 0, 100, 0, 0},
		{"over_max_clamped", 200, 0, 100, 7, 7},
		{"zero_range", 50, 50, 1, 0, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := &SparkLine{}
			result := spark.valueToSparkIndex(tt.value, tt.min, tt.valueRange)
			assert.GreaterOrEqual(t, result, tt.wantMin)
			assert.LessOrEqual(t, result, tt.wantMax)
			assert.GreaterOrEqual(t, result, 0)
			assert.Less(t, result, len(Sparks))
		})
	}
}

func Test_SparkLine_writePadding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		values     []float64
		width      int
		start      int
		wantSpaces int
	}{
		{"need_padding", []float64{1, 2, 3}, 10, 0, 7},
		{"no_padding", []float64{1, 2, 3, 4, 5}, 5, 0, 0},
		{"partial_start", []float64{1, 2, 3, 4, 5, 6, 7}, 5, 2, 0},
		{"exact_fit", []float64{1, 2, 3}, 3, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := &SparkLine{Values: tt.values, Width: tt.width}
			var sb strings.Builder
			spark.writePadding(&sb, tt.start)
			result := sb.String()
			assert.Equal(t, tt.wantSpaces, len(result))
		})
	}
}
