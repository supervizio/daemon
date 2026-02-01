//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []string
		want string
	}{
		{
			name: "EmptySlice",
			opts: []string{},
			want: "",
		},
		{
			name: "SingleOption",
			opts: []string{"rw"},
			want: "rw",
		},
		{
			name: "MultipleOptions",
			opts: []string{"rw", "noexec", "nosuid"},
			want: "rw,noexec,nosuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinOptions(tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainsFlag(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		flag  string
		want  bool
	}{
		{
			name:  "FlagPresent",
			flags: []string{"up", "loopback"},
			flag:  "up",
			want:  true,
		},
		{
			name:  "FlagNotPresent",
			flags: []string{"up", "loopback"},
			flag:  "down",
			want:  false,
		},
		{
			name:  "EmptyFlags",
			flags: []string{},
			flag:  "up",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsFlag(tt.flags, tt.flag)
			assert.Equal(t, tt.want, got)
		})
	}
}
