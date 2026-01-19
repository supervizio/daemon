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

	tests := []struct {
		name        string
		eventType   lifecycle.Type
		message     string
		wantType    lifecycle.Type
		wantMessage string
	}{
		{
			name:        "creates process started event",
			eventType:   lifecycle.TypeProcessStarted,
			message:     "Process started successfully",
			wantType:    lifecycle.TypeProcessStarted,
			wantMessage: "Process started successfully",
		},
		{
			name:        "creates process stopped event",
			eventType:   lifecycle.TypeProcessStopped,
			message:     "Process stopped gracefully",
			wantType:    lifecycle.TypeProcessStopped,
			wantMessage: "Process stopped gracefully",
		},
		{
			name:        "creates mesh node up event",
			eventType:   lifecycle.TypeMeshNodeUp,
			message:     "Node joined cluster",
			wantType:    lifecycle.TypeMeshNodeUp,
			wantMessage: "Node joined cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := time.Now()
			e := lifecycle.NewEvent(tt.eventType, tt.message)
			after := time.Now()

			assert.Equal(t, tt.wantType, e.Type)
			assert.Equal(t, tt.wantMessage, e.Message)
			assert.True(t, e.Timestamp.After(before) || e.Timestamp.Equal(before))
			assert.True(t, e.Timestamp.Before(after) || e.Timestamp.Equal(after))
		})
	}
}

func TestEvent_WithServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		want        string
	}{
		{"sets service name", "my-service", "my-service"},
		{"sets empty service name", "", ""},
		{"sets service name with special chars", "my-service-123", "my-service-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(lifecycle.TypeProcessStarted, "started").
				WithServiceName(tt.serviceName)

			assert.Equal(t, tt.want, e.ServiceName)
		})
	}
}

func TestEvent_WithNodeID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		nodeID string
		want   string
	}{
		{"sets node ID", "node-123", "node-123"},
		{"sets empty node ID", "", ""},
		{"sets node ID with UUID", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(lifecycle.TypeMeshNodeUp, "node up").
				WithNodeID(tt.nodeID)

			assert.Equal(t, tt.want, e.NodeID)
		})
	}
}

func TestEvent_WithPodName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		podName string
		want    string
	}{
		{"sets pod name", "my-pod-abc123", "my-pod-abc123"},
		{"sets empty pod name", "", ""},
		{"sets long pod name", "my-very-long-pod-name-with-suffix-xyz789", "my-very-long-pod-name-with-suffix-xyz789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(lifecycle.TypeK8sPodCreated, "pod created").
				WithPodName(tt.podName)

			assert.Equal(t, tt.want, e.PodName)
		})
	}
}

func TestEvent_WithStringData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dataKey string
		dataVal string
		wantKey string
		wantVal string
	}{
		{"sets string data", "signal", "SIGKILL", "signal", "SIGKILL"},
		{"sets empty string", "message", "", "message", ""},
		{"sets path string", "path", "/var/log/app.log", "path", "/var/log/app.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(lifecycle.TypeProcessFailed, "failed").
				WithStringData(tt.dataKey, tt.dataVal)

			assert.Equal(t, tt.wantVal, e.Data[tt.wantKey])
		})
	}
}

func TestEvent_WithIntData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dataKey string
		dataVal int
		wantKey string
		wantVal int
	}{
		{"sets integer data", "exit_code", 1, "exit_code", 1},
		{"sets zero value", "count", 0, "count", 0},
		{"sets negative value", "priority", -1, "priority", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(lifecycle.TypeProcessFailed, "failed").
				WithIntData(tt.dataKey, tt.dataVal)

			assert.Equal(t, tt.wantVal, e.Data[tt.wantKey])
		})
	}
}

func TestEvent_WithBoolData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dataKey string
		dataVal bool
		wantKey string
		wantVal bool
	}{
		{"sets boolean true", "graceful", true, "graceful", true},
		{"sets boolean false", "forced", false, "forced", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(lifecycle.TypeProcessFailed, "failed").
				WithBoolData(tt.dataKey, tt.dataVal)

			assert.Equal(t, tt.wantVal, e.Data[tt.wantKey])
		})
	}
}

func TestEvent_Chaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		eventType     lifecycle.Type
		message       string
		serviceName   string
		intData       map[string]int
		stringData    map[string]string
		wantType      lifecycle.Type
		wantMessage   string
		wantService   string
		wantDataCount int
	}{
		{
			name:        "chains all methods",
			eventType:   lifecycle.TypeProcessFailed,
			message:     "Process crashed",
			serviceName: "api-server",
			intData: map[string]int{
				"exit_code": 137,
			},
			stringData: map[string]string{
				"reason": "OOM killed",
			},
			wantType:      lifecycle.TypeProcessFailed,
			wantMessage:   "Process crashed",
			wantService:   "api-server",
			wantDataCount: 2,
		},
		{
			name:          "chains with minimal data",
			eventType:     lifecycle.TypeProcessStarted,
			message:       "Started",
			serviceName:   "worker",
			intData:       map[string]int{},
			stringData:    map[string]string{},
			wantType:      lifecycle.TypeProcessStarted,
			wantMessage:   "Started",
			wantService:   "worker",
			wantDataCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := lifecycle.NewEvent(tt.eventType, tt.message).
				WithServiceName(tt.serviceName)

			for k, v := range tt.intData {
				e = e.WithIntData(k, v)
			}
			for k, v := range tt.stringData {
				e = e.WithStringData(k, v)
			}

			assert.Equal(t, tt.wantType, e.Type)
			assert.Equal(t, tt.wantMessage, e.Message)
			assert.Equal(t, tt.wantService, e.ServiceName)
			assert.Len(t, e.Data, tt.wantDataCount)
			for k, v := range tt.intData {
				assert.Equal(t, v, e.Data[k])
			}
			for k, v := range tt.stringData {
				assert.Equal(t, v, e.Data[k])
			}
		})
	}
}

func TestFilterByType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filterTypes []lifecycle.Type
		eventType   lifecycle.Type
		want        bool
	}{
		{
			name:        "matches process started",
			filterTypes: []lifecycle.Type{lifecycle.TypeProcessStarted, lifecycle.TypeProcessStopped},
			eventType:   lifecycle.TypeProcessStarted,
			want:        true,
		},
		{
			name:        "matches process stopped",
			filterTypes: []lifecycle.Type{lifecycle.TypeProcessStarted, lifecycle.TypeProcessStopped},
			eventType:   lifecycle.TypeProcessStopped,
			want:        true,
		},
		{
			name:        "does not match process failed",
			filterTypes: []lifecycle.Type{lifecycle.TypeProcessStarted, lifecycle.TypeProcessStopped},
			eventType:   lifecycle.TypeProcessFailed,
			want:        false,
		},
		{
			name:        "does not match mesh node up",
			filterTypes: []lifecycle.Type{lifecycle.TypeProcessStarted, lifecycle.TypeProcessStopped},
			eventType:   lifecycle.TypeMeshNodeUp,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filter := lifecycle.FilterByType(tt.filterTypes...)
			got := filter(lifecycle.Event{Type: tt.eventType})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterByCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		category  string
		eventType lifecycle.Type
		want      bool
	}{
		{
			name:      "matches process category for started",
			category:  "process",
			eventType: lifecycle.TypeProcessStarted,
			want:      true,
		},
		{
			name:      "matches process category for failed",
			category:  "process",
			eventType: lifecycle.TypeProcessFailed,
			want:      true,
		},
		{
			name:      "does not match process category for mesh",
			category:  "process",
			eventType: lifecycle.TypeMeshNodeUp,
			want:      false,
		},
		{
			name:      "does not match process category for k8s",
			category:  "process",
			eventType: lifecycle.TypeK8sPodCreated,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filter := lifecycle.FilterByCategory(tt.category)
			got := filter(lifecycle.Event{Type: tt.eventType})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterByServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filterName  string
		serviceName string
		want        bool
	}{
		{
			name:        "matches exact service name",
			filterName:  "my-service",
			serviceName: "my-service",
			want:        true,
		},
		{
			name:        "does not match different service name",
			filterName:  "my-service",
			serviceName: "other-service",
			want:        false,
		},
		{
			name:        "does not match empty service name",
			filterName:  "my-service",
			serviceName: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filter := lifecycle.FilterByServiceName(tt.filterName)
			got := filter(lifecycle.Event{ServiceName: tt.serviceName})
			assert.Equal(t, tt.want, got)
		})
	}
}
