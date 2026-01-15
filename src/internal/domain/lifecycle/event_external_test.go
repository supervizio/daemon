// Package lifecycle_test provides external tests for the lifecycle package.
package lifecycle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/lifecycle"
)

func TestType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType lifecycle.Type
		want      string
	}{
		{"process started", lifecycle.TypeProcessStarted, "process.started"},
		{"process stopped", lifecycle.TypeProcessStopped, "process.stopped"},
		{"process failed", lifecycle.TypeProcessFailed, "process.failed"},
		{"mesh node up", lifecycle.TypeMeshNodeUp, "mesh.node.up"},
		{"mesh leader changed", lifecycle.TypeMeshLeaderChanged, "mesh.leader.changed"},
		{"k8s pod created", lifecycle.TypeK8sPodCreated, "k8s.pod.created"},
		{"system high cpu", lifecycle.TypeSystemHighCPU, "system.cpu.high"},
		{"daemon started", lifecycle.TypeDaemonStarted, "daemon.started"},
		{"unknown", lifecycle.TypeUnknown, "unknown"},
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
		eventType lifecycle.Type
		want      string
	}{
		{"process event", lifecycle.TypeProcessStarted, "process"},
		{"mesh event", lifecycle.TypeMeshNodeUp, "mesh"},
		{"k8s event", lifecycle.TypeK8sPodCreated, "kubernetes"},
		{"system event", lifecycle.TypeSystemHighCPU, "system"},
		{"daemon event", lifecycle.TypeDaemonStarted, "daemon"},
		{"unknown event", lifecycle.TypeUnknown, "unknown"},
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
	e := lifecycle.NewEvent(lifecycle.TypeProcessStarted, "Process started successfully")
	after := time.Now()

	assert.Equal(t, lifecycle.TypeProcessStarted, e.Type)
	assert.Equal(t, "Process started successfully", e.Message)
	assert.True(t, e.Timestamp.After(before) || e.Timestamp.Equal(before))
	assert.True(t, e.Timestamp.Before(after) || e.Timestamp.Equal(after))
}

func TestEvent_WithServiceName(t *testing.T) {
	t.Parallel()

	e := lifecycle.NewEvent(lifecycle.TypeProcessStarted, "started").
		WithServiceName("my-service")

	assert.Equal(t, "my-service", e.ServiceName)
}

func TestEvent_WithNodeID(t *testing.T) {
	t.Parallel()

	e := lifecycle.NewEvent(lifecycle.TypeMeshNodeUp, "node up").
		WithNodeID("node-123")

	assert.Equal(t, "node-123", e.NodeID)
}

func TestEvent_WithPodName(t *testing.T) {
	t.Parallel()

	e := lifecycle.NewEvent(lifecycle.TypeK8sPodCreated, "pod created").
		WithPodName("my-pod-abc123")

	assert.Equal(t, "my-pod-abc123", e.PodName)
}

func TestEvent_WithData(t *testing.T) {
	t.Parallel()

	e := lifecycle.NewEvent(lifecycle.TypeProcessFailed, "failed").
		WithData("exit_code", 1).
		WithData("signal", "SIGKILL")

	assert.Equal(t, 1, e.Data["exit_code"])
	assert.Equal(t, "SIGKILL", e.Data["signal"])
}

func TestEvent_Chaining(t *testing.T) {
	t.Parallel()

	e := lifecycle.NewEvent(lifecycle.TypeProcessFailed, "Process crashed").
		WithServiceName("api-server").
		WithData("exit_code", 137).
		WithData("reason", "OOM killed")

	assert.Equal(t, lifecycle.TypeProcessFailed, e.Type)
	assert.Equal(t, "Process crashed", e.Message)
	assert.Equal(t, "api-server", e.ServiceName)
	assert.Equal(t, 137, e.Data["exit_code"])
	assert.Equal(t, "OOM killed", e.Data["reason"])
}

func TestFilterByType(t *testing.T) {
	t.Parallel()

	filter := lifecycle.FilterByType(lifecycle.TypeProcessStarted, lifecycle.TypeProcessStopped)

	assert.True(t, filter(lifecycle.Event{Type: lifecycle.TypeProcessStarted}))
	assert.True(t, filter(lifecycle.Event{Type: lifecycle.TypeProcessStopped}))
	assert.False(t, filter(lifecycle.Event{Type: lifecycle.TypeProcessFailed}))
	assert.False(t, filter(lifecycle.Event{Type: lifecycle.TypeMeshNodeUp}))
}

func TestFilterByCategory(t *testing.T) {
	t.Parallel()

	filter := lifecycle.FilterByCategory("process")

	assert.True(t, filter(lifecycle.Event{Type: lifecycle.TypeProcessStarted}))
	assert.True(t, filter(lifecycle.Event{Type: lifecycle.TypeProcessFailed}))
	assert.False(t, filter(lifecycle.Event{Type: lifecycle.TypeMeshNodeUp}))
	assert.False(t, filter(lifecycle.Event{Type: lifecycle.TypeK8sPodCreated}))
}

func TestFilterByServiceName(t *testing.T) {
	t.Parallel()

	filter := lifecycle.FilterByServiceName("my-service")

	assert.True(t, filter(lifecycle.Event{ServiceName: "my-service"}))
	assert.False(t, filter(lifecycle.Event{ServiceName: "other-service"}))
	assert.False(t, filter(lifecycle.Event{ServiceName: ""}))
}
