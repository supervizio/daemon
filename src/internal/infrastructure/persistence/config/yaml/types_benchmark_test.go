package yaml_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/persistence/config/yaml"
)

// BenchmarkConfigParse measures allocation overhead of YAML → domain conversion.
// This benchmark tests the entire parsing pipeline including the slice allocation
// optimizations in ConfigDTO.ToDomain(), MonitoringConfigDTO.ToDomain(), and
// ServiceConfigDTO.ToDomain().
func BenchmarkConfigParse(b *testing.B) {
	yamlContent := []byte(`
version: "1.0"
services:
  - name: nginx
    command: /usr/bin/nginx
    args: ["-g", "daemon off;"]
    listeners:
      - name: http
        port: 80
        protocol: tcp
      - name: https
        port: 443
        protocol: tcp
  - name: redis
    command: /usr/bin/redis-server
    listeners:
      - name: redis
        port: 6379
        protocol: tcp
  - name: postgres
    command: /usr/bin/postgres
    listeners:
      - name: postgres
        port: 5432
        protocol: tcp
  - name: app
    command: /usr/bin/myapp
    listeners:
      - name: web
        port: 8080
        protocol: tcp
      - name: admin
        port: 8081
        protocol: tcp
      - name: metrics
        port: 9090
        protocol: tcp
  - name: worker
    command: /usr/bin/worker
logging:
  daemon:
    writers:
      - type: console
        level: info
      - type: file
        level: debug
        path: /var/log/daemon.log
      - type: json
        level: info
        path: /var/log/daemon.json
`)

	loader := yaml.NewLoader()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cfg, err := loader.Parse(yamlContent)
		if err != nil {
			b.Fatalf("Failed to parse config: %v", err)
		}
		_ = cfg // Prevent optimization elimination
	}
}

