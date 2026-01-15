package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/application/health"
)

func TestNewProbeBinding(t *testing.T) {
	target := health.ProbeTarget{
		Address: "localhost:8080",
		Path:    "/health",
		Method:  "GET",
	}

	binding := health.NewProbeBinding("http-listener", health.ProbeHTTP, target)

	assert.Equal(t, "http-listener", binding.ListenerName)
	assert.Equal(t, health.ProbeHTTP, binding.Type)
	assert.Equal(t, target, binding.Target)
	// Check default config
	assert.Equal(t, 10*time.Second, binding.Config.Interval)
	assert.Equal(t, 5*time.Second, binding.Config.Timeout)
}

func TestProbeBinding_WithConfig(t *testing.T) {
	binding := health.NewProbeBinding("tcp-listener", health.ProbeTCP, health.ProbeTarget{
		Address: "localhost:3000",
	})

	customConfig := health.ProbeConfig{
		Interval:         30 * time.Second,
		Timeout:          10 * time.Second,
		SuccessThreshold: 2,
		FailureThreshold: 5,
	}

	binding.WithConfig(customConfig)

	assert.Equal(t, customConfig, binding.Config)
}

func TestDefaultProbeConfig(t *testing.T) {
	config := health.DefaultProbeConfig()

	assert.Equal(t, 10*time.Second, config.Interval)
	assert.Equal(t, 5*time.Second, config.Timeout)
	assert.Equal(t, 1, config.SuccessThreshold)
	assert.Equal(t, 3, config.FailureThreshold)
}
