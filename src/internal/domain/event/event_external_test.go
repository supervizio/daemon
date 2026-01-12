//go:build linux

// Package event_test provides external tests for the event package.
package event_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/event"
)

func TestType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType event.Type
		want      string
	}{
		{"process started", event.TypeProcessStarted, "process.started"},
		{"process stopped", event.TypeProcessStopped, "process.stopped"},
		{"process failed", event.TypeProcessFailed, "process.failed"},
		{"mesh node up", event.TypeMeshNodeUp, "mesh.node.up"},
		{"mesh leader changed", event.TypeMeshLeaderChanged, "mesh.leader.changed"},
		{"k8s pod created", event.TypeK8sPodCreated, "k8s.pod.created"},
		{"system high cpu", event.TypeSystemHighCPU, "system.cpu.high"},
		{"daemon started", event.TypeDaemonStarted, "daemon.started"},
		{"unknown", event.TypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.eventType.String())
		})
	}
}

func TestType_Category(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType event.Type
		want      string
	}{
		{"process event", event.TypeProcessStarted, "process"},
		{"mesh event", event.TypeMeshNodeUp, "mesh"},
		{"k8s event", event.TypeK8sPodCreated, "kubernetes"},
		{"system event", event.TypeSystemHighCPU, "system"},
		{"daemon event", event.TypeDaemonStarted, "daemon"},
		{"unknown event", event.TypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.eventType.Category())
		})
	}
}

func TestNewEvent(t *testing.T) {
	t.Parallel()

	before := time.Now()
	e := event.NewEvent(event.TypeProcessStarted, "Process started successfully")
	after := time.Now()

	assert.Equal(t, event.TypeProcessStarted, e.Type)
	assert.Equal(t, "Process started successfully", e.Message)
	assert.True(t, e.Timestamp.After(before) || e.Timestamp.Equal(before))
	assert.True(t, e.Timestamp.Before(after) || e.Timestamp.Equal(after))
}

func TestEvent_WithServiceName(t *testing.T) {
	t.Parallel()

	e := event.NewEvent(event.TypeProcessStarted, "started").
		WithServiceName("my-service")

	assert.Equal(t, "my-service", e.ServiceName)
}

func TestEvent_WithNodeID(t *testing.T) {
	t.Parallel()

	e := event.NewEvent(event.TypeMeshNodeUp, "node up").
		WithNodeID("node-123")

	assert.Equal(t, "node-123", e.NodeID)
}

func TestEvent_WithPodName(t *testing.T) {
	t.Parallel()

	e := event.NewEvent(event.TypeK8sPodCreated, "pod created").
		WithPodName("my-pod-abc123")

	assert.Equal(t, "my-pod-abc123", e.PodName)
}

func TestEvent_WithData(t *testing.T) {
	t.Parallel()

	e := event.NewEvent(event.TypeProcessFailed, "failed").
		WithData("exit_code", 1).
		WithData("signal", "SIGKILL")

	assert.Equal(t, 1, e.Data["exit_code"])
	assert.Equal(t, "SIGKILL", e.Data["signal"])
}

func TestEvent_Chaining(t *testing.T) {
	t.Parallel()

	e := event.NewEvent(event.TypeProcessFailed, "Process crashed").
		WithServiceName("api-server").
		WithData("exit_code", 137).
		WithData("reason", "OOM killed")

	assert.Equal(t, event.TypeProcessFailed, e.Type)
	assert.Equal(t, "Process crashed", e.Message)
	assert.Equal(t, "api-server", e.ServiceName)
	assert.Equal(t, 137, e.Data["exit_code"])
	assert.Equal(t, "OOM killed", e.Data["reason"])
}

func TestFilterByType(t *testing.T) {
	t.Parallel()

	filter := event.FilterByType(event.TypeProcessStarted, event.TypeProcessStopped)

	assert.True(t, filter(event.Event{Type: event.TypeProcessStarted}))
	assert.True(t, filter(event.Event{Type: event.TypeProcessStopped}))
	assert.False(t, filter(event.Event{Type: event.TypeProcessFailed}))
	assert.False(t, filter(event.Event{Type: event.TypeMeshNodeUp}))
}

func TestFilterByCategory(t *testing.T) {
	t.Parallel()

	filter := event.FilterByCategory("process")

	assert.True(t, filter(event.Event{Type: event.TypeProcessStarted}))
	assert.True(t, filter(event.Event{Type: event.TypeProcessFailed}))
	assert.False(t, filter(event.Event{Type: event.TypeMeshNodeUp}))
	assert.False(t, filter(event.Event{Type: event.TypeK8sPodCreated}))
}

func TestFilterByServiceName(t *testing.T) {
	t.Parallel()

	filter := event.FilterByServiceName("my-service")

	assert.True(t, filter(event.Event{ServiceName: "my-service"}))
	assert.False(t, filter(event.Event{ServiceName: "other-service"}))
	assert.False(t, filter(event.Event{ServiceName: ""}))
}
