package daemon

import (
	"errors"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
)

type failingWriter struct {
	closeErr error
}

func (w *failingWriter) Write(event logging.LogEvent) error {
	return nil
}

func (w *failingWriter) Close() error {
	return w.closeErr
}

func TestMultiLogger_CloseWithError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		writers []logging.Writer
		wantErr bool
	}{
		{
			name: "first writer fails",
			writers: []logging.Writer{
				&failingWriter{closeErr: errors.New("close error")},
				&failingWriter{closeErr: nil},
			},
			wantErr: true,
		},
		{
			name: "all writers succeed",
			writers: []logging.Writer{
				&failingWriter{closeErr: nil},
				&failingWriter{closeErr: nil},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := &MultiLogger{writers: tt.writers}
			err := logger.Close()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
