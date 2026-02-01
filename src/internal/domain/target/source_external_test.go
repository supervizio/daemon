package target_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestSource_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source target.Source
		want   string
	}{
		{"static source", target.SourceStatic, "static"},
		{"discovered source", target.SourceDiscovered, "discovered"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.source.String())
		})
	}
}

func TestSource_IsStatic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source target.Source
		want   bool
	}{
		{"static is static", target.SourceStatic, true},
		{"discovered not static", target.SourceDiscovered, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.source.IsStatic())
		})
	}
}

func TestSource_IsDiscovered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source target.Source
		want   bool
	}{
		{"discovered is discovered", target.SourceDiscovered, true},
		{"static not discovered", target.SourceStatic, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.source.IsDiscovered())
		})
	}
}
