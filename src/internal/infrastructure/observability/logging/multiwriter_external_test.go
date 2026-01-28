package logging_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/observability/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWriteCloser struct {
	*bytes.Buffer
	closed bool
}

func (w *testWriteCloser) Close() error {
	w.closed = true
	return nil
}

func TestNewMultiWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		writerCount int
	}{
		{
			name:        "no writers",
			writerCount: 0,
		},
		{
			name:        "single writer",
			writerCount: 1,
		},
		{
			name:        "multiple writers",
			writerCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var writers []io.WriteCloser
			for range tt.writerCount {
				writers = append(writers, &testWriteCloser{Buffer: &bytes.Buffer{}})
			}

			mw := logging.NewMultiWriter(writers...)
			assert.NotNil(t, mw)
		})
	}
}

func TestMultiWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		writers int
	}{
		{
			name:    "write to multiple writers",
			data:    "test message",
			writers: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var writers []io.WriteCloser
			var bufs []*bytes.Buffer
			for range tt.writers {
				buf := &bytes.Buffer{}
				bufs = append(bufs, buf)
				writers = append(writers, &testWriteCloser{Buffer: buf})
			}

			mw := logging.NewMultiWriter(writers...)
			n, err := mw.Write([]byte(tt.data))
			require.NoError(t, err)
			assert.Equal(t, len(tt.data), n)

			for _, buf := range bufs {
				assert.Equal(t, tt.data, buf.String())
			}
		})
	}
}

func TestMultiWriter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		writers int
	}{
		{
			name:    "close all writers",
			writers: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var writers []io.WriteCloser
			var closers []*testWriteCloser
			for range tt.writers {
				w := &testWriteCloser{Buffer: &bytes.Buffer{}}
				closers = append(closers, w)
				writers = append(writers, w)
			}

			mw := logging.NewMultiWriter(writers...)
			err := mw.Close()
			require.NoError(t, err)

			for _, closer := range closers {
				assert.True(t, closer.closed)
			}
		})
	}
}
