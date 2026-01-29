// Package tui_test provides black-box tests for the tui package.
package tui_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// testMaxLinesLimitLogContent represents log content for testing maxLines limit.
const testMaxLinesLimitLogContent string = `2026-01-28T10:00:00Z [INFO] line1
2026-01-28T10:00:01Z [INFO] line2
2026-01-28T10:00:02Z [INFO] line3
2026-01-28T10:00:03Z [INFO] line4
2026-01-28T10:00:04Z [INFO] line5`

// TestNewLogAdapter tests the default constructor.
func TestNewLogAdapter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "default_constructor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()

			assert.NotNil(t, adapter, "NewLogAdapter should return non-nil adapter")
			assert.NotNil(t, adapter.Buffer(), "Buffer should be initialized")
		})
	}
}

// TestNewLogAdapterWithBuffer tests the custom buffer constructor.
func TestNewLogAdapterWithBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		bufferSize int
	}{
		{
			name:       "default size buffer",
			bufferSize: 100,
		},
		{
			name:       "small buffer",
			bufferSize: 10,
		},
		{
			name:       "large buffer",
			bufferSize: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buffer := tui.NewLogBuffer(tt.bufferSize)
			adapter := tui.NewLogAdapterWithBuffer(buffer)

			assert.NotNil(t, adapter, "adapter should not be nil")
			assert.Equal(t, buffer, adapter.Buffer(), "adapter should use provided buffer")
		})
	}
}

// TestLogAdapter_Summarize tests the Summarize method.
func TestLogAdapter_Summarize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		entries        []model.LogEntry
		expectedInfo   int
		expectedWarn   int
		expectedError  int
		expectedAlerts bool
	}{
		{
			name:           "empty buffer",
			entries:        nil,
			expectedInfo:   0,
			expectedWarn:   0,
			expectedError:  0,
			expectedAlerts: false,
		},
		{
			name: "info logs only",
			entries: []model.LogEntry{
				{Level: "INFO", Message: "test1"},
				{Level: "INFO", Message: "test2"},
			},
			expectedInfo:   2,
			expectedWarn:   0,
			expectedError:  0,
			expectedAlerts: false,
		},
		{
			name: "warn logs only",
			entries: []model.LogEntry{
				{Level: "WARN", Message: "warn1"},
				{Level: "WARNING", Message: "warn2"},
			},
			expectedInfo:   0,
			expectedWarn:   2,
			expectedError:  0,
			expectedAlerts: false,
		},
		{
			name: "error logs trigger alerts",
			entries: []model.LogEntry{
				{Level: "ERROR", Message: "error1"},
				{Level: "ERR", Message: "error2"},
			},
			expectedInfo:   0,
			expectedWarn:   0,
			expectedError:  2,
			expectedAlerts: true,
		},
		{
			name: "mixed log levels",
			entries: []model.LogEntry{
				{Level: "INFO", Message: "info"},
				{Level: "WARN", Message: "warn"},
				{Level: "ERROR", Message: "error"},
			},
			expectedInfo:   1,
			expectedWarn:   1,
			expectedError:  1,
			expectedAlerts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()
			for _, entry := range tt.entries {
				adapter.AddLog(entry)
			}

			summary := adapter.Summarize()

			assert.Equal(t, tt.expectedInfo, summary.InfoCount, "info count mismatch")
			assert.Equal(t, tt.expectedWarn, summary.WarnCount, "warn count mismatch")
			assert.Equal(t, tt.expectedError, summary.ErrorCount, "error count mismatch")
			assert.Equal(t, tt.expectedAlerts, summary.HasAlerts, "alerts flag mismatch")
			assert.Len(t, summary.RecentEntries, len(tt.entries), "recent entries count mismatch")
		})
	}
}

// TestLogAdapter_AddLog tests adding log entries.
func TestLogAdapter_AddLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entry   model.LogEntry
		wantMsg string
	}{
		{
			name: "simple info log",
			entry: model.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Service:   "web",
				Message:   "service started",
			},
			wantMsg: "service started",
		},
		{
			name: "log with metadata",
			entry: model.LogEntry{
				Timestamp: time.Now(),
				Level:     "ERROR",
				Service:   "api",
				Message:   "connection failed",
				Metadata: map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			wantMsg: "connection failed",
		},
		{
			name: "empty message",
			entry: model.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Service:   "daemon",
				Message:   "",
			},
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()
			adapter.AddLog(tt.entry)

			summary := adapter.Summarize()
			assert.Len(t, summary.RecentEntries, 1, "should have one entry")
			assert.Equal(t, tt.wantMsg, summary.RecentEntries[0].Message, "message mismatch")
		})
	}
}

// TestLogAdapter_AddDomainEvent tests adding domain events.
func TestLogAdapter_AddDomainEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		event     logging.LogEvent
		wantLevel string
		wantMsg   string
	}{
		{
			name: "info event",
			event: logging.NewLogEvent(
				logging.LevelInfo,
				"web",
				"started",
				"Service started successfully",
			),
			wantLevel: "INFO",
			wantMsg:   "Service started successfully",
		},
		{
			name: "warn event",
			event: logging.NewLogEvent(
				logging.LevelWarn,
				"api",
				"retry",
				"Connection retry attempt",
			),
			wantLevel: "WARN",
			wantMsg:   "Connection retry attempt",
		},
		{
			name: "error event",
			event: logging.NewLogEvent(
				logging.LevelError,
				"db",
				"failed",
				"Database connection lost",
			),
			wantLevel: "ERROR",
			wantMsg:   "Database connection lost",
		},
		{
			name: "debug event",
			event: logging.NewLogEvent(
				logging.LevelDebug,
				"cache",
				"debug",
				"Cache hit",
			),
			wantLevel: "DEBUG",
			wantMsg:   "Cache hit",
		},
		{
			name: "event with metadata",
			event: logging.NewLogEvent(
				logging.LevelInfo,
				"worker",
				"completed",
				"Task finished",
			).WithMeta("duration", "2s").WithMeta("status", "ok"),
			wantLevel: "INFO",
			wantMsg:   "Task finished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()
			adapter.AddDomainEvent(tt.event)

			summary := adapter.Summarize()
			assert.Len(t, summary.RecentEntries, 1, "should have one entry")
			assert.Equal(t, tt.wantLevel, summary.RecentEntries[0].Level, "level mismatch")
			assert.Equal(t, tt.wantMsg, summary.RecentEntries[0].Message, "message mismatch")
		})
	}
}

// TestLogAdapter_Buffer tests buffer access.
func TestLogAdapter_Buffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "buffer_access"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			customBuffer := tui.NewLogBuffer(50)
			adapter := tui.NewLogAdapterWithBuffer(customBuffer)

			buffer := adapter.Buffer()

			assert.NotNil(t, buffer, "Buffer should not be nil")
			assert.Equal(t, customBuffer, buffer, "Buffer should return the same instance")
		})
	}
}

// TestLogAdapter_LoadLogHistory tests loading logs from a file.
func TestLogAdapter_LoadLogHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		logContent    string
		maxLines      int
		expectedCount int
		expectError   bool
	}{
		{
			name: "valid log file with metadata",
			logContent: `2026-01-28T10:00:00Z [INFO] web Service started pid=1234 port=8080
2026-01-28T10:00:01Z [WARN] api Connection slow latency=500ms
2026-01-28T10:00:02Z [ERROR] db Connection failed retry=3`,
			maxLines:      100,
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "log file with daemon messages",
			logContent: `2026-01-28T10:00:00Z [INFO] Daemon initialized
2026-01-28T10:00:01Z [WARN] Memory usage high
2026-01-28T10:00:02Z [ERROR] Failed to connect`,
			maxLines:      100,
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "maxLines limit",
			logContent: testMaxLinesLimitLogContent,
			maxLines:      3,
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "malformed lines ignored",
			logContent: `2026-01-28T10:00:00Z [INFO] valid line
this is not a valid log line
2026-01-28T10:00:01Z [ERROR] another valid line
invalid line without timestamp`,
			maxLines:      100,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "empty file",
			logContent: ``,
			maxLines:      100,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "RFC3339 timestamps with timezone",
			logContent: `2026-01-28T10:00:00+01:00 [INFO] web Service started
2026-01-28T10:00:01-05:00 [WARN] api Connection slow`,
			maxLines:      100,
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary log file.
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")
			err := os.WriteFile(logFile, []byte(tt.logContent), 0600)
			assert.NoError(t, err, "failed to create test log file")

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, tt.maxLines)

			if tt.expectError {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "unexpected error")
			}

			summary := adapter.Summarize()
			assert.Equal(t, tt.expectedCount, len(summary.RecentEntries), "entry count mismatch")
		})
	}
}

// TestLogAdapter_LoadLogHistory_NonExistentFile tests loading from non-existent file.
func TestLogAdapter_LoadLogHistory_NonExistentFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "non_existent_file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()
			err := adapter.LoadLogHistory("/non/existent/path/to/file.log", 100)

			// Non-existent file should not be an error (documented behavior).
			assert.NoError(t, err, "non-existent file should not cause error")

			summary := adapter.Summarize()
			assert.Empty(t, summary.RecentEntries, "should have no entries")
		})
	}
}

// TestLogAdapter_LoadLogHistory_EdgeCases tests edge cases.
func TestLogAdapter_LoadLogHistory_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		maxLines int
		wantErr  bool
	}{
		{
			name:     "empty path",
			path:     "",
			maxLines: 100,
			wantErr:  false,
		},
		{
			name:     "zero maxLines uses default",
			path:     "/tmp/test.log",
			maxLines: 0,
			wantErr:  false,
		},
		{
			name:     "negative maxLines uses default",
			path:     "/tmp/test.log",
			maxLines: -1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()
			err := adapter.LoadLogHistory(tt.path, tt.maxLines)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogAdapter_ParseLogFormats tests various log formats.
func TestLogAdapter_ParseLogFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		logLine           string
		wantService       string
		wantMessage       string
		wantLevel         string
		wantMetadataCount int
		wantParsed        bool
	}{
		{
			name:              "service with message and metadata",
			logLine:           "2026-01-28T10:00:00Z [INFO] web Service started pid=1234 port=8080",
			wantService:       "web",
			wantMessage:       "Service started",
			wantLevel:         "INFO",
			wantMetadataCount: 2,
			wantParsed:        true,
		},
		{
			name:              "message with System treated as service",
			logLine:           "2026-01-28T10:00:00Z [INFO] System initialized version=1.0",
			wantService:       "System",
			wantMessage:       "initialized",
			wantLevel:         "INFO",
			wantMetadataCount: 1,
			wantParsed:        true,
		},
		{
			name:              "message only no metadata",
			logLine:           "2026-01-28T10:00:00Z [ERROR] Failed to connect",
			wantService:       "",
			wantMessage:       "Failed to connect",
			wantLevel:         "ERROR",
			wantMetadataCount: 0,
			wantParsed:        true,
		},
		{
			name:              "warn level variant",
			logLine:           "2026-01-28T10:00:00Z [WARN] api Low memory warning=true",
			wantService:       "api",
			wantMessage:       "Low memory",
			wantLevel:         "WARN",
			wantMetadataCount: 1,
			wantParsed:        true,
		},
		{
			name:              "metadata only no message",
			logLine:           "2026-01-28T10:00:00Z [INFO] db status=ready",
			wantService:       "db",
			wantMessage:       "",
			wantLevel:         "INFO",
			wantMetadataCount: 1,
			wantParsed:        true,
		},
		{
			name:              "common starter word Service",
			logLine:           "2026-01-28T10:00:00Z [INFO] Service started successfully",
			wantService:       "",
			wantMessage:       "Service started successfully",
			wantLevel:         "INFO",
			wantMetadataCount: 0,
			wantParsed:        true,
		},
		{
			name:              "common starter word Daemon",
			logLine:           "2026-01-28T10:00:00Z [INFO] Daemon initialized",
			wantService:       "",
			wantMessage:       "Daemon initialized",
			wantLevel:         "INFO",
			wantMetadataCount: 0,
			wantParsed:        true,
		},
		{
			name:              "malformed missing level",
			logLine:           "2026-01-28T10:00:00Z web Service started",
			wantService:       "",
			wantMessage:       "",
			wantLevel:         "",
			wantMetadataCount: 0,
			wantParsed:        false,
		},
		{
			name:              "malformed missing timestamp",
			logLine:           "[INFO] web Service started",
			wantService:       "",
			wantMessage:       "",
			wantLevel:         "",
			wantMetadataCount: 0,
			wantParsed:        false,
		},
		{
			name:              "empty line",
			logLine:           "",
			wantService:       "",
			wantMessage:       "",
			wantLevel:         "",
			wantMetadataCount: 0,
			wantParsed:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary file with the log line.
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")
			err := os.WriteFile(logFile, []byte(tt.logLine+"\n"), 0600)
			assert.NoError(t, err)

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, 1)
			assert.NoError(t, err)

			summary := adapter.Summarize()

			if tt.wantParsed {
				assert.Len(t, summary.RecentEntries, 1, "should have parsed one entry")
				entry := summary.RecentEntries[0]
				assert.Equal(t, tt.wantService, entry.Service, "service mismatch")
				assert.Equal(t, tt.wantMessage, entry.Message, "message mismatch")
				assert.Equal(t, tt.wantLevel, entry.Level, "level mismatch")
				assert.Len(t, entry.Metadata, tt.wantMetadataCount, "metadata count mismatch")
			} else {
				assert.Empty(t, summary.RecentEntries, "malformed line should not be parsed")
			}
		})
	}
}

// TestLogAdapter_MetadataExtraction tests metadata parsing.
func TestLogAdapter_MetadataExtraction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		logLine          string
		wantMetadataKeys []string
	}{
		{
			name:             "multiple metadata pairs",
			logLine:          "2026-01-28T10:00:00Z [INFO] web Started pid=1234 port=8080 host=localhost",
			wantMetadataKeys: []string{"pid", "port", "host"},
		},
		{
			name:             "metadata with equals in value",
			logLine:          "2026-01-28T10:00:00Z [INFO] api Query executed sql=SELECT * FROM users WHERE id=123",
			wantMetadataKeys: []string{"sql"},
		},
		{
			name:             "no metadata",
			logLine:          "2026-01-28T10:00:00Z [INFO] Just a message",
			wantMetadataKeys: []string{},
		},
		{
			name:             "metadata with underscores",
			logLine:          "2026-01-28T10:00:00Z [INFO] worker Task completed task_id=42 user_name=alice",
			wantMetadataKeys: []string{"task_id", "user_name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")
			err := os.WriteFile(logFile, []byte(tt.logLine+"\n"), 0600)
			assert.NoError(t, err)

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, 1)
			assert.NoError(t, err)

			summary := adapter.Summarize()
			if len(tt.wantMetadataKeys) > 0 {
				assert.Len(t, summary.RecentEntries, 1)
				metadata := summary.RecentEntries[0].Metadata

				for _, key := range tt.wantMetadataKeys {
					_, exists := metadata[key]
					assert.True(t, exists, "metadata key %q should exist", key)
				}
			}
		})
	}
}

// TestLogAdapter_BufferRingBehavior tests ring buffer overflow.
func TestLogAdapter_BufferRingBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "ring_buffer_overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			const bufferSize int = 5
			buffer := tui.NewLogBuffer(bufferSize)
			adapter := tui.NewLogAdapterWithBuffer(buffer)

			// Add more entries than buffer size.
			for i := range 10 {
				adapter.AddLog(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Message:   "message " + string(rune('0'+i)),
				})
			}

			summary := adapter.Summarize()

			// Should only keep last bufferSize entries.
			assert.Len(t, summary.RecentEntries, bufferSize, "buffer should respect size limit")

			// Verify entries are the last ones (5-9).
			for i, entry := range summary.RecentEntries {
				expectedChar := '5' + i
				assert.Contains(t, entry.Message, string(rune(expectedChar)), "should contain correct entry")
			}
		})
	}
}

// TestLogAdapter_ConcurrentAccess tests thread safety.
func TestLogAdapter_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "concurrent_writes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapter()
			done := make(chan bool)

			// Start multiple goroutines adding logs.
			for range 10 {
				go func() {
					for range 100 {
						adapter.AddLog(model.LogEntry{
							Timestamp: time.Now(),
							Level:     "INFO",
							Message:   "concurrent message",
						})
					}
					done <- true
				}()
			}

			// Wait for all goroutines to finish.
			for range 10 {
				<-done
			}

			summary := adapter.Summarize()
			assert.NotNil(t, summary, "summary should not be nil")
			assert.GreaterOrEqual(t, summary.InfoCount, 100, "should have many info logs")
		})
	}
}

// TestLogAdapter_NilBuffer tests adapter with nil buffer.
func TestLogAdapter_NilBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "nil_buffer_handling"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := tui.NewLogAdapterWithBuffer(nil)

			// Operations with nil buffer should not panic.
			summary := adapter.Summarize()
			assert.Equal(t, model.LogSummary{}, summary, "nil buffer should return empty summary")

			adapter.AddLog(model.LogEntry{Level: "INFO", Message: "test"})
			adapter.AddDomainEvent(logging.NewLogEvent(logging.LevelInfo, "test", "test", "message"))

			assert.Nil(t, adapter.Buffer(), "buffer should be nil")
		})
	}
}

// TestLogAdapter_InvalidTimestampFormats tests timestamp parsing edge cases.
func TestLogAdapter_InvalidTimestampFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		logLine    string
		wantParsed bool
	}{
		{
			name:       "invalid timestamp format",
			logLine:    "2026-01-28 10:00:00 [INFO] Invalid timestamp",
			wantParsed: false,
		},
		{
			name:       "missing timezone Z",
			logLine:    "2026-01-28T10:00:00 [INFO] Missing Z",
			wantParsed: false,
		},
		{
			name:       "RFC3339 with milliseconds",
			logLine:    "2026-01-28T10:00:00.123Z [INFO] With milliseconds",
			wantParsed: true,
		},
		{
			name:       "RFC3339 with offset",
			logLine:    "2026-01-28T10:00:00+00:00 [INFO] With offset",
			wantParsed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")
			err := os.WriteFile(logFile, []byte(tt.logLine+"\n"), 0600)
			assert.NoError(t, err)

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, 1)
			assert.NoError(t, err)

			summary := adapter.Summarize()
			if tt.wantParsed {
				assert.Len(t, summary.RecentEntries, 1, "should parse entry")
			} else {
				assert.Empty(t, summary.RecentEntries, "should not parse entry")
			}
		})
	}
}

// TestLogAdapter_ServiceNameDetection tests service name detection edge cases.
func TestLogAdapter_ServiceNameDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		logLine     string
		wantService string
		wantMessage string
	}{
		{
			name:        "Started is not a service name",
			logLine:     "2026-01-28T10:00:00Z [INFO] Started successfully",
			wantService: "",
			wantMessage: "Started successfully",
		},
		{
			name:        "Stopped is not a service name",
			logLine:     "2026-01-28T10:00:00Z [INFO] Stopped gracefully",
			wantService: "",
			wantMessage: "Stopped gracefully",
		},
		{
			name:        "Supervisor is not a service name",
			logLine:     "2026-01-28T10:00:00Z [INFO] Supervisor initialized",
			wantService: "",
			wantMessage: "Supervisor initialized",
		},
		{
			name:        "single word with metadata",
			logLine:     "2026-01-28T10:00:00Z [INFO] status=ready",
			wantService: "",
			wantMessage: "",
		},
		{
			name:        "alphanumeric service name",
			logLine:     "2026-01-28T10:00:00Z [INFO] web-api-01 Service started",
			wantService: "web-api-01",
			wantMessage: "Service started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")
			err := os.WriteFile(logFile, []byte(tt.logLine+"\n"), 0600)
			assert.NoError(t, err)

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, 1)
			assert.NoError(t, err)

			summary := adapter.Summarize()
			if len(summary.RecentEntries) > 0 {
				entry := summary.RecentEntries[0]
				assert.Equal(t, tt.wantService, entry.Service, "service mismatch")
				assert.Equal(t, tt.wantMessage, entry.Message, "message mismatch")
			}
		})
	}
}

// TestLogAdapter_ReadLastLines_LargeFile tests handling of large log files.
func TestLogAdapter_ReadLastLines_LargeFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "large_file_handling"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a large log file.
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "large.log")

			file, err := os.Create(logFile)
			assert.NoError(t, err)
			defer func() { _ = file.Close() }()

			// Write 500 lines.
			for i := range 500 {
				_, err := file.WriteString("2026-01-28T10:00:00Z [INFO] Line " + string(rune('0'+(i%10))) + "\n")
				assert.NoError(t, err)
			}
			_ = file.Close()

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, 50)
			assert.NoError(t, err)

			summary := adapter.Summarize()
			// Should only load the last 50 lines.
			assert.LessOrEqual(t, len(summary.RecentEntries), 50, "should respect maxLines limit")
		})
	}
}

// TestLogAdapter_LoadLogHistory_PermissionError tests permission errors.
func TestLogAdapter_LoadLogHistory_PermissionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "permission_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Skip on systems where we can't set permissions (root can read anything).
			if os.Getuid() == 0 {
				return // Cannot test permission errors as root.
			}

			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "noperm.log")

			err := os.WriteFile(logFile, []byte("2026-01-28T10:00:00Z [INFO] test\n"), 0000)
			assert.NoError(t, err)

			adapter := tui.NewLogAdapter()
			err = adapter.LoadLogHistory(logFile, 10)

			// Should return error for permission issues.
			assert.Error(t, err, "should fail with permission error")
		})
	}
}
