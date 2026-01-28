package logging_test

import (
	"bytes"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/observability/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLineWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "with prefix",
			prefix: "[INFO] ",
		},
		{
			name:   "without prefix",
			prefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			lw := logging.NewLineWriter(buf, tt.prefix)
			assert.NotNil(t, lw)
		})
	}
}

func TestLineWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		input    string
		expected string
	}{
		{
			name:     "single line with prefix",
			prefix:   "[INFO] ",
			input:    "test message\n",
			expected: "[INFO] test message\n",
		},
		{
			name:     "single line without prefix",
			prefix:   "",
			input:    "test message\n",
			expected: "test message\n",
		},
		{
			name:     "multiple lines",
			prefix:   "> ",
			input:    "line1\nline2\n",
			expected: "> line1\n> line2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			lw := logging.NewLineWriter(buf, tt.prefix)

			n, err := lw.Write([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, len(tt.input), n)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLineWriter_Flush(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		input    string
		expected string
	}{
		{
			name:     "flush incomplete line",
			prefix:   "> ",
			input:    "incomplete",
			expected: "> incomplete\n",
		},
		{
			name:     "flush empty buffer",
			prefix:   "> ",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			lw := logging.NewLineWriter(buf, tt.prefix)

			if tt.input != "" {
				_, err := lw.Write([]byte(tt.input))
				require.NoError(t, err)
			}

			err := lw.Flush()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
