package logging

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLineWriter_BufferHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		writes    []string
		expectBuf bool
	}{
		{
			name:      "incomplete line stays in buffer",
			writes:    []string{"incomplete"},
			expectBuf: true,
		},
		{
			name:      "complete line clears buffer",
			writes:    []string{"complete\n"},
			expectBuf: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			lw := &LineWriter{
				writer:      buf,
				prefix:      "",
				prefixBytes: []byte{},
				writeBuf:    make([]byte, 0, defaultWriteBufCapacity),
			}

			for _, write := range tt.writes {
				_, err := lw.Write([]byte(write))
				require.NoError(t, err)
			}

			if tt.expectBuf {
				assert.NotEmpty(t, lw.buf)
			} else {
				assert.Empty(t, lw.buf)
			}
		})
	}
}
