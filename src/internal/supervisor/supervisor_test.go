package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/config"
	"github.com/kodflow/daemon/internal/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSupervisor(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:    "test-service",
				Command: "echo hello",
			},
		},
	}

	sup, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, sup)
	assert.Equal(t, StateStopped, sup.State())
}

func TestSupervisorInvalidConfig(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{}, // Empty - invalid
	}

	sup, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, sup)
}

func TestSupervisorStartStop(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:    "sleep-service",
				Command: "sleep 30",
				Restart: config.RestartConfig{
					Policy: config.RestartNever,
				},
			},
		},
	}

	sup, err := New(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sup.Start(ctx)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, StateRunning, sup.State())

	err = sup.Stop()
	require.NoError(t, err)

	assert.Equal(t, StateStopped, sup.State())
}

func TestSupervisorServices(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:    "service-1",
				Command: "sleep 30",
				Restart: config.RestartConfig{
					Policy: config.RestartNever,
				},
			},
			{
				Name:    "service-2",
				Command: "sleep 30",
				Restart: config.RestartConfig{
					Policy: config.RestartNever,
				},
			},
		},
	}

	sup, err := New(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sup.Start(ctx)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	services := sup.Services()
	assert.Len(t, services, 2)
	assert.Contains(t, services, "service-1")
	assert.Contains(t, services, "service-2")

	sup.Stop()
}

func TestSupervisorGetService(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:    "test-service",
				Command: "echo hello",
			},
		},
	}

	sup, err := New(cfg)
	require.NoError(t, err)

	mgr, ok := sup.Service("test-service")
	assert.True(t, ok)
	assert.NotNil(t, mgr)

	_, ok = sup.Service("nonexistent")
	assert.False(t, ok)
}

func TestSupervisorDoubleStart(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:    "test-service",
				Command: "sleep 30",
				Restart: config.RestartConfig{
					Policy: config.RestartNever,
				},
			},
		},
	}

	sup, err := New(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sup.Start(ctx)
	require.NoError(t, err)

	// Second start should fail
	err = sup.Start(ctx)
	assert.Error(t, err)

	sup.Stop()
}

func TestSupervisorStopNotRunning(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:    "test-service",
				Command: "echo hello",
			},
		},
	}

	sup, err := New(cfg)
	require.NoError(t, err)

	// Stop without start should not error
	err = sup.Stop()
	assert.NoError(t, err)
}

func TestIsPID1(t *testing.T) {
	// In test environment, we're not PID 1
	assert.False(t, kernel.Default.Reaper.IsPID1())
}

func TestReapOnce(t *testing.T) {
	// ReapOnce should not panic
	count := kernel.Default.Reaper.ReapOnce()
	assert.GreaterOrEqual(t, count, 0)
}

func TestKernelReaper(t *testing.T) {
	r := kernel.Default.Reaper
	assert.NotNil(t, r)
}
