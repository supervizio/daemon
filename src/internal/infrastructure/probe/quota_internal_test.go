//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaLimitsInternal(t *testing.T) {
	tests := []struct {
		name   string
		limits *QuotaLimits
	}{
		{
			name:   "EmptyLimits",
			limits: &QuotaLimits{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.limits)
		})
	}
}
