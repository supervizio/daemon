//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
)

func TestContainerRuntime_String_External(t *testing.T) {
	tests := []struct {
		name string
		rt   probe.ContainerRuntime
		want string
	}{
		{
			name: "None",
			rt:   probe.ContainerRuntimeNone,
			want: "none",
		},
		{
			name: "Docker",
			rt:   probe.ContainerRuntimeDocker,
			want: "docker",
		},
		{
			name: "Podman",
			rt:   probe.ContainerRuntimePodman,
			want: "podman",
		},
		{
			name: "LXC",
			rt:   probe.ContainerRuntimeLXC,
			want: "lxc",
		},
		{
			name: "Kubernetes",
			rt:   probe.ContainerRuntimeKubernetes,
			want: "kubernetes",
		},
		{
			name: "Jail",
			rt:   probe.ContainerRuntimeJail,
			want: "jail",
		},
		{
			name: "Unknown",
			rt:   probe.ContainerRuntimeUnknown,
			want: "unknown",
		},
		{
			name: "UnrecognizedValue",
			rt:   probe.ContainerRuntime(100),
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
