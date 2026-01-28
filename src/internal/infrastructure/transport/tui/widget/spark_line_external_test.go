package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestNewSparkLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []float64
		width  int
	}{
		{"empty", []float64{}, 10},
		{"single", []float64{50}, 10},
		{"multiple", []float64{10, 20, 30, 40, 50}, 10},
		{"more_than_width", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := widget.NewSparkLine(tt.values, tt.width)
			assert.NotNil(t, spark)
			assert.Equal(t, tt.values, spark.Values)
			assert.Equal(t, tt.width, spark.Width)
		})
	}
}

func TestSparkLine_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []float64
		width  int
	}{
		{"empty", []float64{}, 10},
		{"single", []float64{50}, 10},
		{"multiple", []float64{10, 20, 30, 40, 50}, 10},
		{"flat", []float64{50, 50, 50, 50}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spark := widget.NewSparkLine(tt.values, tt.width)
			result := spark.Render()
			assert.NotEmpty(t, result)
		})
	}
}

func TestSparks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sparks  []string
		wantLen int
	}{
		{"sparks_array", widget.Sparks, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Len(t, tt.sparks, tt.wantLen)
		})
	}
}
