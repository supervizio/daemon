package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitSpacesCache(t *testing.T) {
	tests := []struct {
		name  string
		index int
		want  int
	}{
		{
			name:  "zero spaces",
			index: 0,
			want:  0,
		},
		{
			name:  "one space",
			index: 1,
			want:  1,
		},
		{
			name:  "max cached spaces",
			index: maxCachedSpaces,
			want:  maxCachedSpaces,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := initSpacesCache()
			got := len(cache[tt.index])
			if got != tt.want {
				t.Errorf("initSpacesCache()[%d] length = %d; want %d", tt.index, got, tt.want)
			}
		})
	}
}

func TestInitBarsCache(t *testing.T) {
	tests := []struct {
		name  string
		index int
		want  int
	}{
		{
			name:  "zero bars",
			index: 0,
			want:  0,
		},
		{
			name:  "one bar",
			index: 1,
			want:  1,
		},
		{
			name:  "max cached bars",
			index: maxCachedBars,
			want:  maxCachedBars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := initBarsCache()
			got := len([]rune(cache[tt.index]))
			if got != tt.want {
				t.Errorf("initBarsCache()[%d] rune length = %d; want %d", tt.index, got, tt.want)
			}
		})
	}
}

func TestByteUnits(t *testing.T) {
	tests := []struct {
		name  string
		index int
		long  string
		short string
	}{
		{
			name:  "kilobyte",
			index: 0,
			long:  "KB",
			short: "K",
		},
		{
			name:  "megabyte",
			index: 1,
			long:  "MB",
			short: "M",
		},
		{
			name:  "gigabyte",
			index: 2,
			long:  "GB",
			short: "G",
		},
		{
			name:  "terabyte",
			index: 3,
			long:  "TB",
			short: "T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.long, byteUnitsLong[tt.index])
			assert.Equal(t, tt.short, byteUnitsShort[tt.index])
		})
	}
}

func TestSpeedUnits(t *testing.T) {
	tests := []struct {
		name  string
		index int
		long  string
		short string
	}{
		{
			name:  "kilobit",
			index: 0,
			long:  "Kbps",
			short: "K",
		},
		{
			name:  "megabit",
			index: 1,
			long:  "Mbps",
			short: "M",
		},
		{
			name:  "gigabit",
			index: 2,
			long:  "Gbps",
			short: "G",
		},
		{
			name:  "terabit",
			index: 3,
			long:  "Tbps",
			short: "T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.long, speedUnitsLong[tt.index])
			assert.Equal(t, tt.short, speedUnits[tt.index])
		})
	}
}

func TestTextConstants(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
		desc string
	}{
		{
			name: "hours per day",
			got:  durationHoursPerDay,
			want: 24,
			desc: "durationHoursPerDay",
		},
		{
			name: "minutes per hour",
			got:  durationMinutesPerHour,
			want: 60,
			desc: "durationMinutesPerHour",
		},
		{
			name: "seconds per minute",
			got:  durationSecondsPerMinute,
			want: 60,
			desc: "durationSecondsPerMinute",
		},
		{
			name: "byte unit",
			got:  byteUnit,
			want: uint64(1024),
			desc: "byteUnit",
		},
		{
			name: "network unit",
			got:  networkUnit,
			want: uint64(1000),
			desc: "networkUnit",
		},
		{
			name: "ellipsis length",
			got:  truncateEllipsisLength,
			want: 3,
			desc: "truncateEllipsisLength",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got, tt.desc)
		})
	}
}
