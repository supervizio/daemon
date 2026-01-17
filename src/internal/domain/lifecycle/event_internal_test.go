package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeString(t *testing.T) {
	tests := []struct {
		name     string
		typeVal  Type
		expected string
	}{
		{
			name:     "unknown type",
			typeVal:  TypeUnknown,
			expected: "unknown",
		},
		{
			name:     "process started",
			typeVal:  TypeProcessStarted,
			expected: "process.started",
		},
		{
			name:     "daemon started",
			typeVal:  TypeDaemonStarted,
			expected: "daemon.started",
		},
		{
			name:     "unmapped type",
			typeVal:  Type(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typeVal.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeCategory(t *testing.T) {
	tests := []struct {
		name     string
		typeVal  Type
		expected string
	}{
		{
			name:     "process category",
			typeVal:  TypeProcessStarted,
			expected: "process",
		},
		{
			name:     "mesh category",
			typeVal:  TypeMeshNodeUp,
			expected: "mesh",
		},
		{
			name:     "kubernetes category",
			typeVal:  TypeK8sPodCreated,
			expected: "kubernetes",
		},
		{
			name:     "system category",
			typeVal:  TypeSystemHighCPU,
			expected: "system",
		},
		{
			name:     "daemon category",
			typeVal:  TypeDaemonStarted,
			expected: "daemon",
		},
		{
			name:     "unknown category",
			typeVal:  TypeUnknown,
			expected: "unknown",
		},
		{
			name:     "unmapped type returns unknown",
			typeVal:  Type(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typeVal.Category()
			assert.Equal(t, tt.expected, result)
		})
	}
}
