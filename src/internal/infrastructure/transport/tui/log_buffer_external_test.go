// Package tui_test provides black-box tests for the tui package.
package tui_test

import (
	"sync"
	"testing"
	"time"

	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestNewLogBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		maxSize     int
		wantDefault bool
	}{
		{
			name:        "positive size",
			maxSize:     50,
			wantDefault: false,
		},
		{
			name:        "zero size uses default",
			maxSize:     0,
			wantDefault: true,
		},
		{
			name:        "negative size uses default",
			maxSize:     -10,
			wantDefault: true,
		},
		{
			name:        "large size",
			maxSize:     10000,
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := tui.NewLogBuffer(tt.maxSize)
			assert.NotNil(t, buf, "NewLogBuffer should return non-nil")

			// Verify buffer starts empty.
			entries := buf.Entries()
			assert.Empty(t, entries, "new buffer should have no entries")

			// Verify summary is empty.
			summary := buf.Summary()
			assert.Equal(t, 0, summary.InfoCount, "info count should be 0")
			assert.Equal(t, 0, summary.WarnCount, "warn count should be 0")
			assert.Equal(t, 0, summary.ErrorCount, "error count should be 0")
			assert.False(t, summary.HasAlerts, "should not have alerts")
			assert.Empty(t, summary.RecentEntries, "recent entries should be empty")
		})
	}
}

func TestLogBuffer_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "single_entry",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				entry := model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test-service",
					EventType: "started",
					Message:   "service started",
					Metadata:  map[string]any{"pid": 1234},
				}

				buf.Add(entry)

				entries := buf.Entries()
				assert.Len(t, entries, 1, "should have 1 entry")
				assert.Equal(t, "INFO", entries[0].Level)
				assert.Equal(t, "test-service", entries[0].Service)
				assert.Equal(t, "started", entries[0].EventType)
				assert.Equal(t, "service started", entries[0].Message)
				assert.Equal(t, 1234, entries[0].Metadata["pid"])
			},
		},
		{
			name: "multiple_entries_below_capacity",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				for i := range 5 {
					buf.Add(model.LogEntry{
						Timestamp: time.Now().Add(time.Duration(i) * time.Second),
						Level:     "INFO",
						Service:   "service",
						EventType: "event",
						Message:   "message",
					})
				}

				entries := buf.Entries()
				assert.Len(t, entries, 5, "should have 5 entries")
			},
		},
		{
			name: "wrap_around_at_capacity",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(5)

				// Add 8 entries (3 more than capacity).
				for i := range 8 {
					buf.Add(model.LogEntry{
						Timestamp: time.Now(),
						Level:     "INFO",
						Service:   "service",
						EventType: "event",
						Message:   "message-" + string(rune('0'+i)),
					})
				}

				entries := buf.Entries()
				assert.Len(t, entries, 5, "should only keep last 5 entries")

				// Verify oldest entries were dropped (should have messages 3-7).
				for i, entry := range entries {
					expected := "message-" + string(rune('0'+i+3))
					assert.Equal(t, expected, entry.Message, "entry %d message mismatch", i)
				}
			},
		},
		{
			name: "level_counting_INFO",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "INFO", Message: "info1"})
				buf.Add(model.LogEntry{Level: "INFO", Message: "info2"})

				summary := buf.Summary()
				assert.Equal(t, 2, summary.InfoCount, "should count INFO entries")
				assert.Equal(t, 0, summary.WarnCount, "should not count WARN entries")
				assert.Equal(t, 0, summary.ErrorCount, "should not count ERROR entries")
				assert.False(t, summary.HasAlerts, "INFO only should not trigger alerts")
			},
		},
		{
			name: "level_counting_WARN_variants",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "WARN", Message: "warn1"})
				buf.Add(model.LogEntry{Level: "WARNING", Message: "warn2"})

				summary := buf.Summary()
				assert.Equal(t, 0, summary.InfoCount)
				assert.Equal(t, 2, summary.WarnCount, "should count WARN and WARNING")
				assert.Equal(t, 0, summary.ErrorCount)
			},
		},
		{
			name: "level_counting_ERROR_variants",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "ERROR", Message: "error1"})
				buf.Add(model.LogEntry{Level: "ERR", Message: "error2"})

				summary := buf.Summary()
				assert.Equal(t, 0, summary.InfoCount)
				assert.Equal(t, 0, summary.WarnCount)
				assert.Equal(t, 2, summary.ErrorCount, "should count ERROR and ERR")
				assert.True(t, summary.HasAlerts, "ERROR should trigger alerts")
			},
		},
		{
			name: "level_counting_mixed",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(20)
				buf.Add(model.LogEntry{Level: "INFO", Message: "info1"})
				buf.Add(model.LogEntry{Level: "INFO", Message: "info2"})
				buf.Add(model.LogEntry{Level: "WARN", Message: "warn1"})
				buf.Add(model.LogEntry{Level: "ERROR", Message: "error1"})
				buf.Add(model.LogEntry{Level: "DEBUG", Message: "debug1"}) // Not counted.

				summary := buf.Summary()
				assert.Equal(t, 2, summary.InfoCount)
				assert.Equal(t, 1, summary.WarnCount)
				assert.Equal(t, 1, summary.ErrorCount)
				assert.True(t, summary.HasAlerts, "ERROR should trigger alerts")
			},
		},
		{
			name: "unknown_levels_not_counted",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "DEBUG", Message: "debug"})
				buf.Add(model.LogEntry{Level: "TRACE", Message: "trace"})
				buf.Add(model.LogEntry{Level: "UNKNOWN", Message: "unknown"})

				summary := buf.Summary()
				assert.Equal(t, 0, summary.InfoCount, "DEBUG/TRACE/UNKNOWN should not be counted")
				assert.Equal(t, 0, summary.WarnCount)
				assert.Equal(t, 0, summary.ErrorCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestLogBuffer_AddFromDomainEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "converts_domain_event_to_log_entry",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				now := time.Now()
				event := domainlogging.NewLogEvent(
					domainlogging.LevelInfo,
					"test-service",
					"started",
					"service started successfully",
				).WithMeta("pid", 5678)

				// Set timestamp manually (NewLogEvent uses time.Now()).
				event.Timestamp = now

				buf.AddFromDomainEvent(event)

				entries := buf.Entries()
				assert.Len(t, entries, 1)

				entry := entries[0]
				assert.Equal(t, now, entry.Timestamp)
				assert.Equal(t, "INFO", entry.Level)
				assert.Equal(t, "test-service", entry.Service)
				assert.Equal(t, "started", entry.EventType)
				assert.Equal(t, "service started successfully", entry.Message)
				assert.Equal(t, 5678, entry.Metadata["pid"])
			},
		},
		{
			name: "converts_all_log_levels",
			testFunc: func(t *testing.T) {
				t.Helper()

				levelTests := []struct {
					level    domainlogging.Level
					expected string
				}{
					{domainlogging.LevelDebug, "DEBUG"},
					{domainlogging.LevelInfo, "INFO"},
					{domainlogging.LevelWarn, "WARN"},
					{domainlogging.LevelError, "ERROR"},
				}

				for _, lt := range levelTests {
					buf := tui.NewLogBuffer(10)
					event := domainlogging.NewLogEvent(lt.level, "service", "event", "message")

					buf.AddFromDomainEvent(event)

					entries := buf.Entries()
					assert.Len(t, entries, 1)
					assert.Equal(t, lt.expected, entries[0].Level)
				}
			},
		},
		{
			name: "preserves_metadata",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				event := domainlogging.NewLogEvent(
					domainlogging.LevelError,
					"service",
					"failed",
					"process crashed",
				).WithMeta("exit_code", 137).WithMeta("signal", "SIGKILL")

				buf.AddFromDomainEvent(event)

				entries := buf.Entries()
				assert.Len(t, entries, 1)
				assert.Equal(t, 137, entries[0].Metadata["exit_code"])
				assert.Equal(t, "SIGKILL", entries[0].Metadata["signal"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestLogBuffer_Entries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "empty_buffer_returns_nil",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				entries := buf.Entries()
				assert.Nil(t, entries, "empty buffer should return nil slice")
			},
		},
		{
			name: "returns_entries_in_FIFO_order",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				now := time.Now()

				for i := range 5 {
					buf.Add(model.LogEntry{
						Timestamp: now.Add(time.Duration(i) * time.Second),
						Level:     "INFO",
						Message:   "msg-" + string(rune('0'+i)),
					})
				}

				entries := buf.Entries()
				assert.Len(t, entries, 5)

				for i, entry := range entries {
					expected := "msg-" + string(rune('0'+i))
					assert.Equal(t, expected, entry.Message, "entry %d should be %s", i, expected)
				}
			},
		},
		{
			name: "FIFO_order_after_wrap_around",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(3)

				// Add 5 entries (wraps twice).
				for i := range 5 {
					buf.Add(model.LogEntry{
						Level:   "INFO",
						Message: "msg-" + string(rune('0'+i)),
					})
				}

				entries := buf.Entries()
				assert.Len(t, entries, 3, "should only keep 3 entries")

				// Should have messages 2, 3, 4 (oldest 0, 1 were dropped).
				expected := []string{"msg-2", "msg-3", "msg-4"}
				for i, entry := range entries {
					assert.Equal(t, expected[i], entry.Message, "entry %d mismatch", i)
				}
			},
		},
		{
			name: "returns_copy_not_reference",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "INFO", Message: "original"})

				entries1 := buf.Entries()
				entries2 := buf.Entries()

				// Modify first slice.
				entries1[0].Message = "modified"

				// Second slice should be unchanged.
				assert.Equal(t, "modified", entries1[0].Message)
				assert.Equal(t, "original", entries2[0].Message, "should return independent copies")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestLogBuffer_Summary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "empty_buffer_summary",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				summary := buf.Summary()

				assert.Equal(t, 5*time.Minute, summary.Period)
				assert.Equal(t, 0, summary.InfoCount)
				assert.Equal(t, 0, summary.WarnCount)
				assert.Equal(t, 0, summary.ErrorCount)
				assert.Nil(t, summary.RecentEntries)
				assert.False(t, summary.HasAlerts)
			},
		},
		{
			name: "includes_recent_entries",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "INFO", Message: "entry1"})
				buf.Add(model.LogEntry{Level: "WARN", Message: "entry2"})

				summary := buf.Summary()

				assert.Len(t, summary.RecentEntries, 2)
				assert.Equal(t, "entry1", summary.RecentEntries[0].Message)
				assert.Equal(t, "entry2", summary.RecentEntries[1].Message)
			},
		},
		{
			name: "counts_persist_across_wrap_around",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(3)

				// Add 5 entries (2 will be dropped from buffer).
				buf.Add(model.LogEntry{Level: "INFO", Message: "info1"})
				buf.Add(model.LogEntry{Level: "INFO", Message: "info2"})
				buf.Add(model.LogEntry{Level: "WARN", Message: "warn1"})
				buf.Add(model.LogEntry{Level: "ERROR", Message: "error1"})
				buf.Add(model.LogEntry{Level: "INFO", Message: "info3"})

				summary := buf.Summary()

				// Counts should include all 5 entries.
				assert.Equal(t, 3, summary.InfoCount, "should count all INFO entries")
				assert.Equal(t, 1, summary.WarnCount, "should count all WARN entries")
				assert.Equal(t, 1, summary.ErrorCount, "should count all ERROR entries")
				assert.True(t, summary.HasAlerts)

				// But recent entries should only have last 3.
				assert.Len(t, summary.RecentEntries, 3)
			},
		},
		{
			name: "has_alerts_only_when_errors_present",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf1 := tui.NewLogBuffer(10)
				buf1.Add(model.LogEntry{Level: "INFO"})
				buf1.Add(model.LogEntry{Level: "WARN"})
				assert.False(t, buf1.Summary().HasAlerts, "no alerts without errors")

				buf2 := tui.NewLogBuffer(10)
				buf2.Add(model.LogEntry{Level: "ERROR"})
				assert.True(t, buf2.Summary().HasAlerts, "alerts when errors present")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestLogBuffer_Clear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "clears_all_entries_and_counts",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)

				// Add various entries.
				buf.Add(model.LogEntry{Level: "INFO", Message: "info"})
				buf.Add(model.LogEntry{Level: "WARN", Message: "warn"})
				buf.Add(model.LogEntry{Level: "ERROR", Message: "error"})

				// Verify data exists.
				assert.Len(t, buf.Entries(), 3)
				summary := buf.Summary()
				assert.Equal(t, 1, summary.InfoCount)
				assert.Equal(t, 1, summary.WarnCount)
				assert.Equal(t, 1, summary.ErrorCount)

				// Clear buffer.
				buf.Clear()

				// Verify everything is reset.
				entries := buf.Entries()
				assert.Nil(t, entries, "entries should be nil after clear")

				summary = buf.Summary()
				assert.Equal(t, 0, summary.InfoCount, "info count should be 0")
				assert.Equal(t, 0, summary.WarnCount, "warn count should be 0")
				assert.Equal(t, 0, summary.ErrorCount, "error count should be 0")
				assert.False(t, summary.HasAlerts, "alerts should be false")
			},
		},
		{
			name: "can_add_entries_after_clear",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				buf.Add(model.LogEntry{Level: "INFO", Message: "before"})
				buf.Clear()
				buf.Add(model.LogEntry{Level: "INFO", Message: "after"})

				entries := buf.Entries()
				assert.Len(t, entries, 1)
				assert.Equal(t, "after", entries[0].Message)
			},
		},
		{
			name: "clear_on_empty_buffer_is_safe",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(10)
				assert.NotPanics(t, func() {
					buf.Clear()
				}, "clearing empty buffer should not panic")

				assert.Nil(t, buf.Entries())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestLogBuffer_Concurrent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "concurrent_adds",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(100)
				var wg sync.WaitGroup

				// Add 100 entries concurrently (10 goroutines * 10 entries each).
				for i := range 10 {
					n := i
					wg.Go(func() {
						for range 10 {
							buf.Add(model.LogEntry{
								Level:   "INFO",
								Message: "concurrent-" + string(rune('0'+n)),
							})
						}
					})
				}
				wg.Wait()

				entries := buf.Entries()
				assert.Len(t, entries, 100, "should have all 100 entries")
			},
		},
		{
			name: "concurrent_add_and_read",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(50)
				var wg sync.WaitGroup

				// Writer goroutines.
				for range 5 {
					wg.Go(func() {
						for range 20 {
							buf.Add(model.LogEntry{
								Level:   "INFO",
								Message: "write",
							})
							time.Sleep(1 * time.Millisecond)
						}
					})
				}

				// Reader goroutines.
				for range 3 {
					wg.Go(func() {
						for range 50 {
							_ = buf.Entries()
							_ = buf.Summary()
							time.Sleep(1 * time.Millisecond)
						}
					})
				}

				wg.Wait()

				// Final state should be consistent.
				entries := buf.Entries()
				assert.LessOrEqual(t, len(entries), 50, "should respect max size")
			},
		},
		{
			name: "concurrent_add_and_clear",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(50)
				var wg sync.WaitGroup

				// Writer.
				wg.Go(func() {
					for range 100 {
						buf.Add(model.LogEntry{Level: "INFO", Message: "write"})
						time.Sleep(1 * time.Millisecond)
					}
				})

				// Clearer.
				wg.Go(func() {
					for range 5 {
						time.Sleep(10 * time.Millisecond)
						buf.Clear()
					}
				})

				wg.Wait()

				// Should not panic or deadlock.
				entries := buf.Entries()
				assert.NotNil(t, entries, "should not panic")
			},
		},
		{
			name: "concurrent_AddFromDomainEvent",
			testFunc: func(t *testing.T) {
				t.Helper()

				buf := tui.NewLogBuffer(100)
				var wg sync.WaitGroup

				levels := []domainlogging.Level{
					domainlogging.LevelInfo,
					domainlogging.LevelWarn,
					domainlogging.LevelError,
				}

				for i := range 10 {
					n := i
					wg.Go(func() {
						level := levels[n%len(levels)]
						event := domainlogging.NewLogEvent(level, "service", "event", "message")
						buf.AddFromDomainEvent(event)
					})
				}
				wg.Wait()

				entries := buf.Entries()
				assert.Len(t, entries, 10, "should have all entries")

				summary := buf.Summary()
				totalCounts := summary.InfoCount + summary.WarnCount + summary.ErrorCount
				assert.Equal(t, 10, totalCounts, "total counts should match entries")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}
