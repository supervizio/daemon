package target_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  target.Type
		want string
	}{
		{"systemd", target.TypeSystemd, "systemd"},
		{"openrc", target.TypeOpenRC, "openrc"},
		{"bsd-rc", target.TypeBSDRC, "bsd-rc"},
		{"docker", target.TypeDocker, "docker"},
		{"podman", target.TypePodman, "podman"},
		{"kubernetes", target.TypeKubernetes, "kubernetes"},
		{"nomad", target.TypeNomad, "nomad"},
		{"remote", target.TypeRemote, "remote"},
		{"custom", target.TypeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.typ.String())
		})
	}
}

func TestType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  target.Type
		want bool
	}{
		{"systemd valid", target.TypeSystemd, true},
		{"docker valid", target.TypeDocker, true},
		{"kubernetes valid", target.TypeKubernetes, true},
		{"invalid type", target.Type("invalid"), false},
		{"empty type", target.Type(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.typ.IsValid())
		})
	}
}

func TestType_IsContainerRuntime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  target.Type
		want bool
	}{
		{"docker is container runtime", target.TypeDocker, true},
		{"podman is container runtime", target.TypePodman, true},
		{"systemd not container runtime", target.TypeSystemd, false},
		{"kubernetes not container runtime", target.TypeKubernetes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.typ.IsContainerRuntime())
		})
	}
}

func TestType_IsOrchestrator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  target.Type
		want bool
	}{
		{"kubernetes is orchestrator", target.TypeKubernetes, true},
		{"nomad is orchestrator", target.TypeNomad, true},
		{"docker not orchestrator", target.TypeDocker, false},
		{"systemd not orchestrator", target.TypeSystemd, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.typ.IsOrchestrator())
		})
	}
}

func TestType_IsInitSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  target.Type
		want bool
	}{
		{"systemd is init system", target.TypeSystemd, true},
		{"openrc is init system", target.TypeOpenRC, true},
		{"bsd-rc is init system", target.TypeBSDRC, true},
		{"docker not init system", target.TypeDocker, false},
		{"kubernetes not init system", target.TypeKubernetes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.typ.IsInitSystem())
		})
	}
}
