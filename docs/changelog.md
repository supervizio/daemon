# Changelog

All notable changes to superviz.io are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

---

## Recent Changes

### Build & CI

- Optimized CI workflows with parallel x86 and ARM64 pipelines
- Updated Go to 1.25.6
- Moved linter configs to root and fixed ktn-linter issues

### TUI

- Implemented raw mode TUI with full ktn-linter compliance
- Added CLAUDE.md documentation for TUI subdirectories

### Infrastructure

- Portable `chown` syntax for cross-platform E2E compatibility
- Wire dependency injection for compile-time safety

### Architecture

- Hexagonal architecture with strict layer separation
- Domain layer with zero external dependencies
- 10 gRPC RPCs including server-side streaming
- Cross-platform Rust FFI probe for system metrics
- Multi-backend service discovery (Docker, K8s, systemd, Nomad)

---

For the complete commit history, see the [GitHub repository](https://github.com/supervizio/daemon/commits/main).
