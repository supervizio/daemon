package process

import (
	"bytes"
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProcess(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "test-service",
		Command: "echo hello",
	}

	p := New(cfg)

	assert.NotNil(t, p)
	assert.Equal(t, cfg, p.Config)
	assert.Equal(t, StateStopped, p.State())
	assert.Equal(t, 0, p.PID())
}

func TestProcessStartStop(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "sleep-service",
		Command: "sleep 10",
	}

	p := New(cfg)
	ctx := context.Background()

	// Start the process
	err := p.Start(ctx)
	require.NoError(t, err)

	// Wait a bit for process to start
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, StateRunning, p.State())
	assert.Greater(t, p.PID(), 0)

	// Stop the process
	err = p.Stop(5 * time.Second)
	require.NoError(t, err)

	// Wait for process to stop
	<-p.Wait()

	assert.NotEqual(t, StateRunning, p.State())
}

func TestProcessOutput(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "echo-service",
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	p := New(cfg)
	ctx := context.Background()

	var stdout, stderr bytes.Buffer
	p.SetOutput(&stdout, &stderr)

	err := p.Start(ctx)
	require.NoError(t, err)

	// Wait for process to complete
	<-p.Wait()

	assert.Equal(t, 0, p.ExitCode())
	assert.Contains(t, stdout.String(), "hello world")
}

func TestProcessEnvironment(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "env-service",
		Command: "sh",
		Args:    []string{"-c", "echo $TEST_VAR"},
		Environment: map[string]string{
			"TEST_VAR": "test_value",
		},
	}

	p := New(cfg)
	ctx := context.Background()

	var stdout bytes.Buffer
	p.SetOutput(&stdout, nil)

	err := p.Start(ctx)
	require.NoError(t, err)

	<-p.Wait()

	assert.Equal(t, 0, p.ExitCode())
	assert.Contains(t, stdout.String(), "test_value")
}

func TestProcessWorkingDirectory(t *testing.T) {
	tmpDir := os.TempDir()
	cfg := &config.ServiceConfig{
		Name:             "pwd-service",
		Command:          "pwd",
		WorkingDirectory: tmpDir,
	}

	p := New(cfg)
	ctx := context.Background()

	var stdout bytes.Buffer
	p.SetOutput(&stdout, nil)

	err := p.Start(ctx)
	require.NoError(t, err)

	<-p.Wait()

	assert.Equal(t, 0, p.ExitCode())
	assert.Contains(t, stdout.String(), tmpDir)
}

func TestProcessSignal(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "signal-service",
		Command: "sleep 60",
	}

	p := New(cfg)
	ctx := context.Background()

	err := p.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Send SIGTERM
	err = p.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for process to exit
	<-p.Wait()

	// Process should have been terminated by signal
	assert.NotEqual(t, 0, p.ExitCode())
}

func TestProcessUptime(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "uptime-service",
		Command: "sleep 1",
	}

	p := New(cfg)
	ctx := context.Background()

	err := p.Start(ctx)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	uptime := p.Uptime()
	assert.Greater(t, uptime, 100*time.Millisecond)

	p.Stop(5 * time.Second)
}

func TestProcessDoubleStart(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "double-start",
		Command: "sleep 10",
	}

	p := New(cfg)
	ctx := context.Background()

	err := p.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Second start should fail
	err = p.Start(ctx)
	assert.Error(t, err)

	p.Stop(5 * time.Second)
}

func TestProcessStopNotRunning(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "not-running",
		Command: "echo test",
	}

	p := New(cfg)

	// Stop without start should not error
	err := p.Stop(5 * time.Second)
	assert.NoError(t, err)
}

// RestartTracker tests

func TestRestartTrackerShouldRestart(t *testing.T) {
	tests := []struct {
		name       string
		policy     config.RestartPolicy
		exitCode   int
		attempts   int
		maxRetries int
		expected   bool
	}{
		{
			name:       "always policy - success exit",
			policy:     config.RestartAlways,
			exitCode:   0,
			attempts:   0,
			maxRetries: 3,
			expected:   true,
		},
		{
			name:       "always policy - exhausted",
			policy:     config.RestartAlways,
			exitCode:   0,
			attempts:   3,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "on-failure policy - success exit",
			policy:     config.RestartOnFailure,
			exitCode:   0,
			attempts:   0,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "on-failure policy - failure exit",
			policy:     config.RestartOnFailure,
			exitCode:   1,
			attempts:   0,
			maxRetries: 3,
			expected:   true,
		},
		{
			name:       "on-failure policy - exhausted",
			policy:     config.RestartOnFailure,
			exitCode:   1,
			attempts:   3,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "never policy",
			policy:     config.RestartNever,
			exitCode:   1,
			attempts:   0,
			maxRetries: 3,
			expected:   false,
		},
		{
			name:       "unless-stopped policy",
			policy:     config.RestartUnless,
			exitCode:   1,
			attempts:   10,
			maxRetries: 3,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RestartConfig{
				Policy:     tt.policy,
				MaxRetries: tt.maxRetries,
			}

			rt := NewRestartTracker(cfg)
			rt.attempts = tt.attempts

			result := rt.ShouldRestart(tt.exitCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRestartTrackerRecordAttempt(t *testing.T) {
	cfg := &config.RestartConfig{
		Policy:     config.RestartAlways,
		MaxRetries: 3,
	}

	rt := NewRestartTracker(cfg)

	assert.Equal(t, 0, rt.Attempts())

	rt.RecordAttempt()
	assert.Equal(t, 1, rt.Attempts())

	rt.RecordAttempt()
	assert.Equal(t, 2, rt.Attempts())
}

func TestRestartTrackerReset(t *testing.T) {
	cfg := &config.RestartConfig{
		Policy:     config.RestartAlways,
		MaxRetries: 3,
	}

	rt := NewRestartTracker(cfg)
	rt.RecordAttempt()
	rt.RecordAttempt()

	assert.Equal(t, 2, rt.Attempts())

	rt.Reset()
	assert.Equal(t, 0, rt.Attempts())
}

func TestRestartTrackerMaybeReset(t *testing.T) {
	cfg := &config.RestartConfig{
		Policy:     config.RestartAlways,
		MaxRetries: 3,
	}

	rt := NewRestartTracker(cfg)
	rt.RecordAttempt()
	rt.RecordAttempt()

	// Short uptime - should not reset
	rt.MaybeReset(1 * time.Minute)
	assert.Equal(t, 2, rt.Attempts())

	// Long uptime - should reset
	rt.MaybeReset(6 * time.Minute)
	assert.Equal(t, 0, rt.Attempts())
}

func TestRestartTrackerNextDelay(t *testing.T) {
	cfg := &config.RestartConfig{
		Policy:     config.RestartAlways,
		MaxRetries: 5,
		Delay:      config.Duration(1 * time.Second),
		DelayMax:   config.Duration(10 * time.Second),
	}

	rt := NewRestartTracker(cfg)

	// First attempt: 1s * 2^0 = 1s
	assert.Equal(t, 1*time.Second, rt.NextDelay())

	rt.RecordAttempt()
	// Second attempt: 1s * 2^1 = 2s
	assert.Equal(t, 2*time.Second, rt.NextDelay())

	rt.RecordAttempt()
	// Third attempt: 1s * 2^2 = 4s
	assert.Equal(t, 4*time.Second, rt.NextDelay())

	rt.RecordAttempt()
	// Fourth attempt: 1s * 2^3 = 8s
	assert.Equal(t, 8*time.Second, rt.NextDelay())

	rt.RecordAttempt()
	// Fifth attempt: 1s * 2^4 = 16s, capped to 10s
	assert.Equal(t, 10*time.Second, rt.NextDelay())
}

func TestRestartTrackerIsExhausted(t *testing.T) {
	cfg := &config.RestartConfig{
		Policy:     config.RestartAlways,
		MaxRetries: 3,
	}

	rt := NewRestartTracker(cfg)

	assert.False(t, rt.IsExhausted())

	rt.RecordAttempt()
	rt.RecordAttempt()
	assert.False(t, rt.IsExhausted())

	rt.RecordAttempt()
	assert.True(t, rt.IsExhausted())
}

// Signal tests

func TestSignalMap(t *testing.T) {
	sig, ok := SignalMap["SIGTERM"]
	assert.True(t, ok)
	assert.Equal(t, syscall.SIGTERM, sig)

	sig, ok = SignalMap["SIGHUP"]
	assert.True(t, ok)
	assert.Equal(t, syscall.SIGHUP, sig)

	_, ok = SignalMap["INVALID"]
	assert.False(t, ok)
}

func TestIsTermSignal(t *testing.T) {
	assert.True(t, IsTermSignal(syscall.SIGTERM))
	assert.True(t, IsTermSignal(syscall.SIGINT))
	assert.True(t, IsTermSignal(syscall.SIGQUIT))
	assert.False(t, IsTermSignal(syscall.SIGHUP))
	assert.False(t, IsTermSignal(syscall.SIGUSR1))
}

func TestIsReloadSignal(t *testing.T) {
	assert.True(t, IsReloadSignal(syscall.SIGHUP))
	assert.False(t, IsReloadSignal(syscall.SIGTERM))
	assert.False(t, IsReloadSignal(syscall.SIGINT))
}

// Manager tests

func TestNewManager(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "test-service",
		Command: "echo test",
	}

	m := NewManager(cfg)

	assert.NotNil(t, m)
	assert.NotNil(t, m.Events())
}

func TestManagerStartStop(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "managed-service",
		Command: "sleep 10",
		Restart: config.RestartConfig{
			Policy:     config.RestartNever,
			MaxRetries: 0,
		},
	}

	m := NewManager(cfg)

	err := m.Start()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, StateRunning, m.State())

	err = m.Stop()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.NotEqual(t, StateRunning, m.State())
}

func TestManagerEvents(t *testing.T) {
	cfg := &config.ServiceConfig{
		Name:    "event-service",
		Command: "echo done",
		Oneshot: true,
		Restart: config.RestartConfig{
			Policy:     config.RestartNever,
			MaxRetries: 0,
		},
	}

	m := NewManager(cfg)
	events := m.Events()

	err := m.Start()
	require.NoError(t, err)

	// Wait for events
	select {
	case event := <-events:
		assert.Equal(t, EventStarted, event.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for started event")
	}

	// Wait for stopped event
	select {
	case event := <-events:
		assert.Equal(t, EventStopped, event.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for stopped event")
	}
}

// Credentials tests

func TestResolveCredentialsEmpty(t *testing.T) {
	uid, gid, err := resolveCredentials("", "")
	require.NoError(t, err)
	assert.Equal(t, uint32(0), uid)
	assert.Equal(t, uint32(0), gid)
}

func TestResolveCredentialsNumeric(t *testing.T) {
	uid, gid, err := resolveCredentials("1000", "1000")
	require.NoError(t, err)
	assert.Equal(t, uint32(1000), uid)
	assert.Equal(t, uint32(1000), gid)
}

func TestResolveCredentialsRoot(t *testing.T) {
	uid, gid, err := resolveCredentials("root", "root")
	require.NoError(t, err)
	assert.Equal(t, uint32(0), uid)
	assert.Equal(t, uint32(0), gid)
}

func TestLookupUser(t *testing.T) {
	// Root should exist on all systems
	u, err := LookupUser("root")
	require.NoError(t, err)
	assert.Equal(t, "0", u.Uid)

	// Lookup by UID
	u, err = LookupUser("0")
	require.NoError(t, err)
	assert.Equal(t, "root", u.Username)
}

func TestLookupGroup(t *testing.T) {
	// Root/wheel group should exist
	g, err := LookupGroup("root")
	if err != nil {
		// Some systems use wheel instead of root
		g, err = LookupGroup("wheel")
	}
	require.NoError(t, err)
	assert.Equal(t, "0", g.Gid)
}
