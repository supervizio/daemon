# E2E Testing Matrix - Research Synthesis

## Executive Summary

Research to complete ALL gaps in E2E testing matrix. Goal: VM + Docker tests for BOTH AMD64 and ARM64 on ALL distributions.

---

## Current Coverage (Before)

| Distribution | AMD64 VM | AMD64 Docker | ARM64 VM | ARM64 Docker |
|--------------|:--------:|:------------:|:--------:|:------------:|
| Debian 12 | ✅ Vagrant | ✅ | ✅ virt-install | ✅ |
| Ubuntu 22.04 | ✅ Vagrant | ✅ | ✅ virt-install | ✅ |
| Alpine 3.19 | ✅ Vagrant | ✅ | ✅ virt-install | ✅ |
| Devuan 4 | ✅ Vagrant | ✅ | ❌ | ❌ |
| Void Linux | ❌ | ✅ | ❌ | ✅ |
| FreeBSD 14 | ✅ Vagrant | N/A | ❌ | N/A |
| OpenBSD 7 | ✅ Vagrant | N/A | ❌ | N/A |
| NetBSD 10 | ⚠️ flaky | N/A | ❌ | N/A |
| DragonFlyBSD 6 | ✅ Vagrant | N/A | ❌ | N/A |

**Gaps: 9** (Devuan ARM64 x2, Void VM x2, BSD ARM64 x4, NetBSD reliability)

---

## Research Findings

### 1. cross-platform-actions/action (BSD ARM64 + NetBSD fix)

**Source**: https://github.com/cross-platform-actions/action

**Key Discovery**: GitHub Action supporting BSD on ARM64 via QEMU!

| OS | ARM64 Support | Versions |
|----|:-------------:|----------|
| FreeBSD | ✅ | 13.0 - 15.0 |
| OpenBSD | ✅ | 6.9 - 7.8 |
| NetBSD | ✅ | 10.0, 10.1 |

**Benefits**:
- No Vagrant dependency for BSD
- ARM64 support via QEMU emulation
- Stable (no flaky boxes)
- Also fixes NetBSD reliability issues

**Usage Example**:
```yaml
- uses: cross-platform-actions/action@v0.26.0
  with:
    operating_system: freebsd
    version: '14.2'
    architecture: arm64
    run: |
      sudo ./setup/install.sh
      ./e2e/test-install.sh
```

**Recommendation**: ✅ **IMPLEMENT** - Replaces Vagrant for BSD, adds ARM64

---

### 2. Void Linux VM

**Analysis**:
- No Vagrant box available (libvirt provider)
- No cloud-init compatible qcow2 images
- Only ISOs and rootfs tarballs available

**Options Evaluated**:

| Option | Complexity | Maintenance | Recommendation |
|--------|------------|-------------|----------------|
| Packer custom image | High | High | ❌ Overkill |
| QEMU + ISO install | High | Medium | ❌ Too slow |
| Docker only | Low | Low | ✅ Current solution |

**Conclusion**: Void Linux has no cloud-compatible VM images. Docker-only testing is the pragmatic solution. Void's primary use case (runit init) is properly tested via container.

**Recommendation**: ⚠️ **SKIP VM** - Docker coverage is sufficient

---

### 3. Devuan ARM64

**Analysis**:
- No ARM64 cloud images from Devuan project
- No ARM64 Docker images available
- Devuan embedded project focuses on other architectures

**Options Evaluated**:

| Option | Complexity | Feasibility |
|--------|------------|-------------|
| Build custom ARM64 image | Very High | Possible but not worth effort |
| Cross-compile only | Low | Already done in CI build |

**Conclusion**: Devuan ARM64 is a very niche use case. SysVinit testing on ARM64 can be verified via Debian's sysvinit packages if needed.

**Recommendation**: ⚠️ **SKIP** - Cross-compilation verified, no ARM64 target demand

---

### 4. DragonFlyBSD ARM64

**Analysis**: DragonFlyBSD is x86-64 ONLY by design. ARM64 is not supported.

**Recommendation**: ❌ **IMPOSSIBLE** - Architecture not supported

---

## Implementation Plan

### Phase 1: BSD with cross-platform-actions (HIGH IMPACT)

Replace Vagrant BSD jobs with cross-platform-actions for:
- FreeBSD AMD64 + ARM64
- OpenBSD AMD64 + ARM64
- NetBSD AMD64 + ARM64 (fixes flaky box!)
- DragonFlyBSD AMD64 only (no ARM64)

**Files to modify**: `.github/workflows/e2e.yml`

**Changes**:
1. Create new job `e2e-bsd` using cross-platform-actions
2. Matrix includes both AMD64 and ARM64 for FreeBSD/OpenBSD/NetBSD
3. DragonFlyBSD AMD64 only
4. Remove old `e2e-bsd-amd64` Vagrant-based job

### Phase 2: Documentation Update

Update documentation to reflect final coverage.

---

## Final Coverage (After Implementation)

| Distribution | AMD64 VM | AMD64 Docker | ARM64 VM | ARM64 Docker |
|--------------|:--------:|:------------:|:--------:|:------------:|
| Debian 12 | ✅ Vagrant | ✅ | ✅ virt-install | ✅ |
| Ubuntu 22.04 | ✅ Vagrant | ✅ | ✅ virt-install | ✅ |
| Alpine 3.19 | ✅ Vagrant | ✅ | ✅ virt-install | ✅ |
| Devuan 4 | ✅ Vagrant | ✅ | - | - |
| Void Linux | - | ✅ | - | ✅ |
| FreeBSD 14 | ✅ cross-platform | N/A | ✅ cross-platform | N/A |
| OpenBSD 7 | ✅ cross-platform | N/A | ✅ cross-platform | N/A |
| NetBSD 10 | ✅ cross-platform | N/A | ✅ cross-platform | N/A |
| DragonFlyBSD 6 | ✅ cross-platform | N/A | - | N/A |

**Improvements**:
- ✅ FreeBSD ARM64 added
- ✅ OpenBSD ARM64 added
- ✅ NetBSD ARM64 added + fixed reliability
- ⚠️ Devuan ARM64 skipped (no images exist)
- ⚠️ Void VM skipped (no cloud images exist)
- ❌ DragonFlyBSD ARM64 impossible (arch not supported)

**Total jobs**: 18 (was 15)
- 5 Linux AMD64
- 5 Linux ARM64
- 7 BSD (4 AMD64 + 3 ARM64)
- 1 PID1

---

## Legend

- ✅ Tested
- ⚠️ Skipped (technical limitation documented)
- ❌ Impossible (architecture not supported)
- `-` Not applicable
- N/A Not applicable (BSD doesn't support Docker)
