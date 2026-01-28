package logging

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWriteCloser struct {
	closed bool
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockWriteCloser) Close() error {
	m.closed = true
	return nil
}

func TestCapture_CloseInternal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "internal close behavior"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &mockWriteCloser{}
			stderr := &mockWriteCloser{}

			capture := &Capture{
				stdout: stdout,
				stderr: stderr,
			}

			err := capture.Close()
			assert.NoError(t, err)
			assert.True(t, stdout.closed)
			assert.True(t, stderr.closed)
			assert.True(t, capture.closed)
		})
	}
}

func TestCapture_Writers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "stdout and stderr access"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &mockWriteCloser{}
			stderr := &mockWriteCloser{}

			capture := &Capture{
				stdout: stdout,
				stderr: stderr,
			}

			stdoutWriter := capture.Stdout()
			assert.NotNil(t, stdoutWriter)
			assert.Implements(t, (*io.Writer)(nil), stdoutWriter)

			stderrWriter := capture.Stderr()
			assert.NotNil(t, stderrWriter)
			assert.Implements(t, (*io.Writer)(nil), stderrWriter)
		})
	}
}
