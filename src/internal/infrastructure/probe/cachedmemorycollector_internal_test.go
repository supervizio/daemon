//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachedMemoryCollector_NewCachedMemoryCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsNonNilCollector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewCachedMemoryCollector()
			assert.NotNil(t, collector)
		})
	}
}
