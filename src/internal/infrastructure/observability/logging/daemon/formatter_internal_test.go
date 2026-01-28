package daemon

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "get builder from pool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sb := getBuilder()
			assert.NotNil(t, sb)
			assert.IsType(t, &strings.Builder{}, sb)
		})
	}
}

func TestPutBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "return builder to pool",
			content: "test content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sb := getBuilder()
			sb.WriteString(tt.content)
			assert.NotZero(t, sb.Len())

			putBuilder(sb)
			assert.Zero(t, sb.Len(), "builder should be reset after return to pool")
		})
	}
}

func TestFormatMetadataToBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		meta     map[string]any
		contains []string
	}{
		{
			name:     "single key-value",
			meta:     map[string]any{"pid": 1234},
			contains: []string{"pid=1234"},
		},
		{
			name:     "multiple key-values",
			meta:     map[string]any{"pid": 1234, "exit_code": 0},
			contains: []string{"pid=1234", "exit_code=0"},
		},
		{
			name:     "empty metadata",
			meta:     map[string]any{},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sb := getBuilder()
			defer putBuilder(sb)

			formatMetadataToBuilder(sb, tt.meta)
			output := sb.String()

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "string value",
			value:    "test",
			expected: "test",
		},
		{
			name:     "int value",
			value:    42,
			expected: "42",
		},
		{
			name:     "int64 value",
			value:    int64(1234),
			expected: "1234",
		},
		{
			name:     "uint64 value",
			value:    uint64(5678),
			expected: "5678",
		},
		{
			name:     "float64 value",
			value:    3.14,
			expected: "3.14",
		},
		{
			name:     "bool value true",
			value:    true,
			expected: "true",
		},
		{
			name:     "bool value false",
			value:    false,
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sb := getBuilder()
			defer putBuilder(sb)

			formatValue(sb, tt.value)
			assert.Contains(t, sb.String(), tt.expected)
		})
	}
}
