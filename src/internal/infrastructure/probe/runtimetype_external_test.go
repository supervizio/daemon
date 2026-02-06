//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeType_String_External(t *testing.T) {
	tests := []struct {
		name string
		rt   probe.RuntimeType
		want string
	}{
		{
			name: "None",
			rt:   probe.RuntimeNone,
			want: "none",
		},
		{
			name: "Docker",
			rt:   probe.RuntimeDocker,
			want: "docker",
		},
		{
			name: "Kubernetes",
			rt:   probe.RuntimeKubernetes,
			want: "kubernetes",
		},
		{
			name: "Unknown",
			rt:   probe.RuntimeUnknown,
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rt.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRuntimeType_IsOrchestrator_External(t *testing.T) {
	tests := []struct {
		name string
		rt   probe.RuntimeType
		want bool
	}{
		{
			name: "None_NotOrchestrator",
			rt:   probe.RuntimeNone,
			want: false,
		},
		{
			name: "Docker_NotOrchestrator",
			rt:   probe.RuntimeDocker,
			want: false,
		},
		{
			name: "Kubernetes_IsOrchestrator",
			rt:   probe.RuntimeKubernetes,
			want: true,
		},
		{
			name: "AWSECS_IsOrchestrator",
			rt:   probe.RuntimeAWSECS,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rt.IsOrchestrator()
			assert.Equal(t, tt.want, got)
		})
	}
}
