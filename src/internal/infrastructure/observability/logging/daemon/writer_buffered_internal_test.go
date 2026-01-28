package daemon

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferedWriter_BuffersUntilFlush(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []string
	}{
		{
			name:   "buffer and flush events",
			events: []string{"Test started", "Test running"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cw := NewConsoleWriterWithOptions(&buf, &buf, false)
			bw := NewBufferedWriter(cw)

			for _, msg := range tt.events {
				event := logging.NewLogEvent(logging.LevelInfo, "test", "event", msg)
				err := bw.Write(event)
				assert.NoError(t, err)
			}

			assert.Empty(t, buf.String(), "Buffer should be empty before flush")

			err := bw.Flush()
			assert.NoError(t, err)

			output := buf.String()
			for _, msg := range tt.events {
				assert.Contains(t, output, msg)
			}

			// Write after flush should go directly
			buf.Reset()
			event3 := logging.NewLogEvent(logging.LevelInfo, "test", "done", "Test done")
			err = bw.Write(event3)
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "Test done")
		})
	}
}

func TestBufferedWriter_OrderPreserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		eventCount int
	}{
		{
			name:       "preserve event order",
			eventCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cw := NewConsoleWriterWithOptions(&buf, &buf, false)
			bw := NewBufferedWriter(cw)

			for i := 1; i <= tt.eventCount; i++ {
				event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "Event "+string(rune('0'+i)))
				_ = bw.Write(event)
			}

			_ = bw.Flush()

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			assert.Len(t, lines, tt.eventCount)

			for i, line := range lines {
				expected := "Event " + string(rune('1'+i))
				assert.Contains(t, line, expected)
			}
		})
	}
}

func TestBufferedWriter_EnforcesMaxSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		eventCount int
	}{
		{
			name:       "enforce max buffer size",
			eventCount: maxBufferSize + 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cw := NewConsoleWriterWithOptions(&buf, &buf, false)
			bw := NewBufferedWriter(cw)

			for i := range tt.eventCount {
				event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "message")
				event = event.WithMeta("index", i)
				err := bw.Write(event)
				require.NoError(t, err)
			}

			err := bw.Flush()
			require.NoError(t, err)

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			assert.Equal(t, maxBufferSize, len(lines), "Buffer should be limited to maxBufferSize")
		})
	}
}

// Goroutine lifecycle: Each goroutine writes writesPerGoroutine events then terminates.
// Cleanup: wg.Wait() ensures all goroutines finish before verification.
func TestBufferedWriter_ConcurrentWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		numGoroutines      int
		writesPerGoroutine int
	}{
		{
			name:               "concurrent writes thread safe",
			numGoroutines:      50,
			writesPerGoroutine: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cw := NewConsoleWriterWithOptions(&buf, &buf, false)
			bw := NewBufferedWriter(cw)

			var wg sync.WaitGroup
			wg.Add(tt.numGoroutines)

			// Goroutine lifecycle: Each goroutine completes after writesPerGoroutine writes.
			// Cleanup: wg.Wait() ensures all goroutines finish before assertion.
			for range tt.numGoroutines {
				go func() {
					defer wg.Done()
					for range tt.writesPerGoroutine {
						event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "concurrent message")
						_ = bw.Write(event)
					}
				}()
			}

			wg.Wait()

			err := bw.Flush()
			require.NoError(t, err)

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			expected := min(tt.numGoroutines*tt.writesPerGoroutine, maxBufferSize)
			assert.Equal(t, expected, len(lines))
		})
	}
}
