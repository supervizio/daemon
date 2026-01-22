# supervizio

[![CI](https://github.com/supervizio/daemon/actions/workflows/ci.yml/badge.svg)](https://github.com/supervizio/daemon/actions/workflows/ci.yml)
[![Release](https://github.com/supervizio/daemon/actions/workflows/release.yml/badge.svg)](https://github.com/supervizio/daemon/releases)
[![Codacy Grade](https://app.codacy.com/project/badge/Grade/c66eb99290744de6ac6a6e082f83daaf)](https://app.codacy.com/gh/supervizio/daemon/dashboard)
[![Go 1.25](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](https://go.dev/)
[![License MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

PID1-capable process supervisor for containers and Unix systems.

## Install

```bash
# Debian/Ubuntu
curl -fsSL https://supervizio.github.io/daemon/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/supervizio.gpg
echo "deb [signed-by=/etc/apt/keyrings/supervizio.gpg] https://supervizio.github.io/daemon/apt stable main" | sudo tee /etc/apt/sources.list.d/supervizio.list
sudo apt update && sudo apt install supervizio

# Or download binary
curl -fsSL https://github.com/supervizio/daemon/releases/latest/download/supervizio-linux-amd64 -o supervizio
chmod +x supervizio && sudo mv supervizio /usr/local/bin/
```

**[Full documentation](https://supervizio.github.io/daemon)** — Installation for all platforms, configuration, and usage.

## Features

- Multi-service management with dependency ordering
- Health checks (HTTP, TCP, command)
- Restart policies with exponential backoff
- Log rotation with compression
- PID 1 mode (zombie reaping, signal forwarding)
- Linux, BSD, macOS support

## Quick Start

```yaml
# /etc/supervizio/config.yaml
version: "1"
services:
  - name: myapp
    command: /usr/bin/myapp
    restart:
      policy: always
```

```bash
supervizio --config /etc/supervizio/config.yaml
```

## License

MIT
