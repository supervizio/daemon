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
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create console writer writing to our buffer
	cw := NewConsoleWriterWithOptions(&buf, &buf, false)

	// Wrap in buffered writer
	bw := NewBufferedWriter(cw)

	// Write some events
	event1 := logging.NewLogEvent(logging.LevelInfo, "test", "started", "Test started")
	event2 := logging.NewLogEvent(logging.LevelInfo, "test", "running", "Test running")

	err := bw.Write(event1)
	assert.NoError(t, err)
	err = bw.Write(event2)
	assert.NoError(t, err)

	// Buffer should be empty (not flushed yet)
	assert.Empty(t, buf.String(), "Buffer should be empty before flush")

	// Flush
	err = bw.Flush()
	assert.NoError(t, err)

	// Now events should be written
	output := buf.String()
	assert.Contains(t, output, "Test started")
	assert.Contains(t, output, "Test running")

	// Write after flush should go directly
	buf.Reset()
	event3 := logging.NewLogEvent(logging.LevelInfo, "test", "done", "Test done")
	err = bw.Write(event3)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Test done")
}

func TestBufferedWriter_OrderPreserved(t *testing.T) {
	var buf bytes.Buffer
	cw := NewConsoleWriterWithOptions(&buf, &buf, false)
	bw := NewBufferedWriter(cw)

	// Write events in order
	for i := 1; i <= 5; i++ {
		event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "Event "+string(rune('0'+i)))
		bw.Write(event)
	}

	bw.Flush()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 5)

	// Check order
	for i, line := range lines {
		expected := "Event " + string(rune('1'+i))
		assert.Contains(t, line, expected)
	}
}

// TestBufferedWriter_EnforcesMaxSize tests that buffer size is limited to prevent OOM.
func TestBufferedWriter_EnforcesMaxSize(t *testing.T) {
	var buf bytes.Buffer
	cw := NewConsoleWriterWithOptions(&buf, &buf, false)
	bw := NewBufferedWriter(cw)

	// Write more than maxBufferSize events.
	for i := range maxBufferSize + 100 {
		event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "message")
		event = event.WithMeta("index", i)
		err := bw.Write(event)
		require.NoError(t, err)
	}

	// Flush and count lines (events).
	err := bw.Flush()
	require.NoError(t, err)

	// Should have exactly maxBufferSize events (oldest dropped).
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, maxBufferSize, len(lines), "Buffer should be limited to maxBufferSize")
}

// TestBufferedWriter_ConcurrentWrites tests thread safety of buffered writes.
func TestBufferedWriter_ConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	cw := NewConsoleWriterWithOptions(&buf, &buf, false)
	bw := NewBufferedWriter(cw)

	const numGoroutines = 50
	const writesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range writesPerGoroutine {
				event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "concurrent message")
				_ = bw.Write(event)
			}
		}()
	}

	wg.Wait()

	// Flush and verify no panic or race.
	err := bw.Flush()
	require.NoError(t, err)

	// Total events should be capped at maxBufferSize.
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	expected := min(numGoroutines*writesPerGoroutine, maxBufferSize)
	assert.Equal(t, expected, len(lines))
}
