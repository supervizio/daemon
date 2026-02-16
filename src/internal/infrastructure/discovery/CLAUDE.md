<!-- updated: 2026-02-15T21:30:00Z -->
# Discovery - Target Discovery Adapters

Implements `domain/target.Discoverer` for discovering external monitoring targets.

## Structure

```
discovery/
├── factory.go              # Factory for creating discoverers
├── factory_unix.go         # Unix-specific factory (Docker, Podman, port scan)
├── factory_linux.go        # Linux-specific (systemd, OpenRC)
├── factory_bsd.go          # BSD-specific factory
├── factory_other.go        # Fallback (non-Unix)
├── constants.go            # Shared constants
├── static.go               # Static targets from config
├── docker_discoverer.go    # Docker containers via Engine API
├── docker_container.go     # Docker container types
├── docker_port.go          # Docker port mapping
├── container_helper.go     # Shared container helpers
├── podman_discoverer.go    # Podman containers via API
├── systemd.go              # Systemd services (Linux)
├── openrc.go               # OpenRC services (Alpine/Gentoo)
├── bsdrc.go                # BSD rc.d services
├── kubernetes.go           # Kubernetes pods/services
├── kubernetes_types.go     # K8s type definitions
├── kubernetes_auth.go      # K8s auth (kubeconfig, in-cluster)
├── nomad.go                # HashiCorp Nomad allocations
├── nomad_types.go          # Nomad type definitions
├── listeningport.go        # Listening port entities
└── portscandiscoverer.go   # Port scanning discoverer
```

## Key Types

| Type | Description |
|------|-------------|
| `Factory` | Creates discoverers from configuration |
| `DockerDiscoverer` | Docker containers via API (label filtering, port mapping) |
| `PodmanDiscoverer` | Podman containers via compatible API |
| `SystemdDiscoverer` | Systemd services via systemctl (glob patterns) |
| `OpenRCDiscoverer` | OpenRC services (Alpine/Gentoo) |
| `BSDRCDiscoverer` | BSD rc.d services |
| `KubernetesDiscoverer` | K8s pods/services (kubeconfig or in-cluster auth) |
| `NomadDiscoverer` | Nomad allocations via API |
| `StaticDiscoverer` | Targets from YAML configuration |
| `PortScanDiscoverer` | Discovers services by port scanning |

## Platform Support

| Discoverer | Platform | Build Tag |
|------------|----------|-----------|
| Docker, Podman, PortScan | Unix | `//go:build unix` |
| Systemd, OpenRC | Linux | `//go:build linux` |
| BSD rc.d | BSD | `//go:build bsd` |
| Static, Kubernetes, Nomad | All | (none) |

## Dependencies

- Depends on: `domain/target`, `domain/config`, `domain/health`
- Used by: `application/monitoring`, `bootstrap`
