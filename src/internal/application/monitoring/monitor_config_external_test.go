package monitoring_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/application/monitoring"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

// mockCreator is a mock prober factory.
type mockCreator struct {
	createFunc func(string, time.Duration) (health.Prober, error)
}

func (m *mockCreator) Create(proberType string, timeout time.Duration) (health.Prober, error) {
	if m.createFunc != nil {
		return m.createFunc(proberType, timeout)
	}
	return nil, nil
}

// mockDiscoverer is a mock discoverer.
type mockDiscoverer struct {
	discoverFunc func(context.Context) ([]target.ExternalTarget, error)
	targetType   target.Type
}

func (m *mockDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	if m.discoverFunc != nil {
		return m.discoverFunc(ctx)
	}
	return nil, nil
}

func (m *mockDiscoverer) Type() target.Type {
	return m.targetType
}

// mockWatcher is a mock watcher.
type mockWatcher struct {
	watchFunc  func(context.Context) (<-chan target.Event, error)
	targetType target.Type
}

func (m *mockWatcher) Watch(ctx context.Context) (<-chan target.Event, error) {
	if m.watchFunc != nil {
		return m.watchFunc(ctx)
	}
	return nil, nil
}

func (m *mockWatcher) Type() target.Type {
	return m.targetType
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name                string
		wantInterval        time.Duration
		wantTimeout         time.Duration
		wantSuccessThresh   int
		wantFailureThresh   int
		wantDiscoveryInt    time.Duration
		wantDiscoveryEnable bool
	}{
		{
			name:                "default values",
			wantInterval:        monitoring.DefaultInterval,
			wantTimeout:         monitoring.DefaultTimeout,
			wantSuccessThresh:   monitoring.DefaultSuccessThreshold,
			wantFailureThresh:   monitoring.DefaultFailureThreshold,
			wantDiscoveryInt:    monitoring.DefaultDiscoveryInterval,
			wantDiscoveryEnable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig()

			assert.Equal(t, tt.wantInterval, cfg.Defaults.Interval)
			assert.Equal(t, tt.wantTimeout, cfg.Defaults.Timeout)
			assert.Equal(t, tt.wantSuccessThresh, cfg.Defaults.SuccessThreshold)
			assert.Equal(t, tt.wantFailureThresh, cfg.Defaults.FailureThreshold)
			assert.Equal(t, tt.wantDiscoveryInt, cfg.Discovery.Interval)
			assert.Equal(t, tt.wantDiscoveryEnable, cfg.Discovery.Enabled)
			assert.Nil(t, cfg.Factory)
			assert.Nil(t, cfg.Events)
			assert.Nil(t, cfg.OnHealthChange)
			assert.Nil(t, cfg.OnUnhealthy)
			assert.Nil(t, cfg.OnHealthy)
		})
	}
}

func TestConfig_WithFactory(t *testing.T) {
	tests := []struct {
		name    string
		factory monitoring.Creator
	}{
		{
			name:    "with factory",
			factory: &mockCreator{},
		},
		{
			name:    "with nil factory",
			factory: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig().WithFactory(tt.factory)

			assert.Equal(t, tt.factory, cfg.Factory)
		})
	}
}

func TestConfig_WithEvents(t *testing.T) {
	tests := []struct {
		name   string
		events chan<- target.Event
	}{
		{
			name:   "with events channel",
			events: make(chan target.Event, 10),
		},
		{
			name:   "with nil events",
			events: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig().WithEvents(tt.events)

			assert.Equal(t, tt.events, cfg.Events)
		})
	}
}

func TestConfig_WithCallbacks(t *testing.T) {
	tests := []struct {
		name         string
		onChange     monitoring.HealthCallback
		onUnhealthy  monitoring.UnhealthyCallback
		onHealthy    monitoring.HealthyCallback
		wantOnChange bool
		wantUnheal   bool
		wantHealthy  bool
	}{
		{
			name:         "all callbacks",
			onChange:     func(string, string, string) {},
			onUnhealthy:  func(string, string) {},
			onHealthy:    func(string) {},
			wantOnChange: true,
			wantUnheal:   true,
			wantHealthy:  true,
		},
		{
			name:         "nil callbacks",
			onChange:     nil,
			onUnhealthy:  nil,
			onHealthy:    nil,
			wantOnChange: false,
			wantUnheal:   false,
			wantHealthy:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig().WithCallbacks(tt.onChange, tt.onUnhealthy, tt.onHealthy)

			if tt.wantOnChange {
				assert.NotNil(t, cfg.OnHealthChange)
			} else {
				assert.Nil(t, cfg.OnHealthChange)
			}
			if tt.wantUnheal {
				assert.NotNil(t, cfg.OnUnhealthy)
			} else {
				assert.Nil(t, cfg.OnUnhealthy)
			}
			if tt.wantHealthy {
				assert.NotNil(t, cfg.OnHealthy)
			} else {
				assert.Nil(t, cfg.OnHealthy)
			}
		})
	}
}

func TestConfig_WithDiscoverers(t *testing.T) {
	disc1 := &mockDiscoverer{targetType: target.TypeSystemd}
	disc2 := &mockDiscoverer{targetType: target.TypeDocker}

	tests := []struct {
		name        string
		discoverers []target.Discoverer
		wantEnabled bool
		wantCount   int
	}{
		{
			name:        "single discoverer",
			discoverers: []target.Discoverer{disc1},
			wantEnabled: true,
			wantCount:   1,
		},
		{
			name:        "multiple discoverers",
			discoverers: []target.Discoverer{disc1, disc2},
			wantEnabled: true,
			wantCount:   2,
		},
		{
			name:        "empty discoverers",
			discoverers: []target.Discoverer{},
			wantEnabled: false,
			wantCount:   0,
		},
		{
			name:        "nil discoverers",
			discoverers: nil,
			wantEnabled: false,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig().WithDiscoverers(tt.discoverers...)

			assert.Equal(t, tt.wantEnabled, cfg.Discovery.Enabled)
			assert.Len(t, cfg.Discovery.Discoverers, tt.wantCount)
		})
	}
}

func TestConfig_WithWatchers(t *testing.T) {
	watcher1 := &mockWatcher{targetType: target.TypeSystemd}
	watcher2 := &mockWatcher{targetType: target.TypeDocker}

	tests := []struct {
		name      string
		watchers  []target.Watcher
		wantCount int
	}{
		{
			name:      "single watcher",
			watchers:  []target.Watcher{watcher1},
			wantCount: 1,
		},
		{
			name:      "multiple watchers",
			watchers:  []target.Watcher{watcher1, watcher2},
			wantCount: 2,
		},
		{
			name:      "empty watchers",
			watchers:  []target.Watcher{},
			wantCount: 0,
		},
		{
			name:      "nil watchers",
			watchers:  nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig().WithWatchers(tt.watchers...)

			assert.Len(t, cfg.Discovery.Watchers, tt.wantCount)
		})
	}
}

func TestConfig_WithDiscoveryInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
	}{
		{
			name:     "standard interval",
			interval: 60 * time.Second,
		},
		{
			name:     "short interval",
			interval: 5 * time.Second,
		},
		{
			name:     "long interval",
			interval: 5 * time.Minute,
		},
		{
			name:     "zero interval",
			interval: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig().WithDiscoveryInterval(tt.interval)

			assert.Equal(t, tt.interval, cfg.Discovery.Interval)
		})
	}
}

func TestConfig_GetInterval(t *testing.T) {
	tests := []struct {
		name             string
		defaultInterval  time.Duration
		overrideInterval time.Duration
		want             time.Duration
	}{
		{
			name:             "use override",
			defaultInterval:  30 * time.Second,
			overrideInterval: 10 * time.Second,
			want:             10 * time.Second,
		},
		{
			name:             "use default when no override",
			defaultInterval:  45 * time.Second,
			overrideInterval: 0,
			want:             45 * time.Second,
		},
		{
			name:             "use package default when both zero",
			defaultInterval:  0,
			overrideInterval: 0,
			want:             monitoring.DefaultInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig()
			cfg.Defaults.Interval = tt.defaultInterval

			got := cfg.GetInterval(tt.overrideInterval)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_GetTimeout(t *testing.T) {
	tests := []struct {
		name            string
		defaultTimeout  time.Duration
		overrideTimeout time.Duration
		want            time.Duration
	}{
		{
			name:            "use override",
			defaultTimeout:  5 * time.Second,
			overrideTimeout: 2 * time.Second,
			want:            2 * time.Second,
		},
		{
			name:            "use default when no override",
			defaultTimeout:  10 * time.Second,
			overrideTimeout: 0,
			want:            10 * time.Second,
		},
		{
			name:            "use package default when both zero",
			defaultTimeout:  0,
			overrideTimeout: 0,
			want:            monitoring.DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig()
			cfg.Defaults.Timeout = tt.defaultTimeout

			got := cfg.GetTimeout(tt.overrideTimeout)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_GetSuccessThreshold(t *testing.T) {
	tests := []struct {
		name                    string
		defaultSuccessThreshold int
		overrideThreshold       int
		want                    int
	}{
		{
			name:                    "use override",
			defaultSuccessThreshold: 1,
			overrideThreshold:       3,
			want:                    3,
		},
		{
			name:                    "use default when no override",
			defaultSuccessThreshold: 2,
			overrideThreshold:       0,
			want:                    2,
		},
		{
			name:                    "use package default when both zero",
			defaultSuccessThreshold: 0,
			overrideThreshold:       0,
			want:                    monitoring.DefaultSuccessThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig()
			cfg.Defaults.SuccessThreshold = tt.defaultSuccessThreshold

			got := cfg.GetSuccessThreshold(tt.overrideThreshold)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_GetFailureThreshold(t *testing.T) {
	tests := []struct {
		name                    string
		defaultFailureThreshold int
		overrideThreshold       int
		want                    int
	}{
		{
			name:                    "use override",
			defaultFailureThreshold: 3,
			overrideThreshold:       5,
			want:                    5,
		},
		{
			name:                    "use default when no override",
			defaultFailureThreshold: 4,
			overrideThreshold:       0,
			want:                    4,
		},
		{
			name:                    "use package default when both zero",
			defaultFailureThreshold: 0,
			overrideThreshold:       0,
			want:                    monitoring.DefaultFailureThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := monitoring.NewConfig()
			cfg.Defaults.FailureThreshold = tt.defaultFailureThreshold

			got := cfg.GetFailureThreshold(tt.overrideThreshold)

			assert.Equal(t, tt.want, got)
		})
	}
}
