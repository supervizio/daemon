# Implementation Plan: 100% Test Coverage

## Overview

Plan pour atteindre 100% de coverage sur tous les packages du projet, excluant les fichiers générés (`*.pb.go`, `wire_gen.go`).

**Coverage actuel:** 76.5%
**Objectif:** 100%

## Packages/Fichiers à Exclure (Générés/Vendor)

Ces fichiers sont **exclus** de l'objectif de coverage car générés automatiquement:

| Pattern | Package/Fichier | Raison |
|---------|-----------------|--------|
| `**/*.pb.go` | `api/proto/v1/daemon/daemon.pb.go` | Généré par protoc |
| `**/*_grpc.pb.go` | Fichiers gRPC générés | Généré par protoc-gen-go-grpc |
| `**/wire_gen.go` | `internal/bootstrap/wire_gen.go` | Généré par Wire DI |
| `vendor/**` | Dépendances vendorisées | Code tiers |
| `**/mocks/**` | Fichiers mock | Générés pour tests |
| `**/*.mock.go` | Fichiers mock | Générés par mockgen |

**Configuration `.ktn-linter.yaml` existante:**
```yaml
exclude:
  - "**/*.pb.go"
  - "**/wire_gen.go"
  - "**/api/proto/**"
  - "vendor/**"
```

**Note:** Ces exclusions sont déjà configurées. Le coverage réel à atteindre concerne uniquement le code source écrit manuellement.

## Packages Sans Test Files

| Package | Action |
|---------|--------|
| `internal/application/config` | Créer `loader_external_test.go` |

---

## Plan par Package

### 1. cmd/daemon (0.0% → 100%)

**Fichiers:** `main.go`
**Coverage actuel:** 0.0%
**Fonctions non couvertes:**
- `main()` - Point d'entrée

**Actions:**
1. Créer `main_external_test.go`
2. Tester avec process exec (ou refactor pour testabilité)

**Difficulté:** HAUTE (main() difficile à tester directement)
**Alternative:** Exclure de la couverture via build tag

---

### 2. internal/bootstrap (40.0% → 100%)

**Coverage actuel:** 40.0%
**Fonctions non couvertes:**
- `Run()` - 0.0%
- `run()` - 25.0%
- `WaitForSignals()` - 87.5%

**Actions:**
1. Ajouter tests pour `Run()` avec mock de dépendances
2. Compléter tests `run()` pour tous les chemins
3. Tester `WaitForSignals()` signal handling

**Fichiers à modifier:**
- `app_internal_test.go`
- `app_external_test.go`

---

### 3. internal/application/health (88.0% → 100%)

**Coverage actuel:** 88.0%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `AddListenerWithBinding` | 90.0% |
| `createProberFromBinding` | 81.8% |
| `Start` | 66.7% |
| `runProber` | 66.7% |
| `performProbe` | 90.9% |
| `updateListenerState` | 92.3% |
| `sendEventIfChanged` | 83.3% |
| `Health` | 54.5% |

**Actions:**
1. Ajouter cas de test pour erreurs dans `createProberFromBinding`
2. Tester `Start` avec contexte annulé
3. Compléter `runProber` avec erreurs de probe
4. Tester `Health` avec différents états de listeners
5. Couvrir edge cases dans `sendEventIfChanged`

**Fichiers à modifier:**
- `monitor_internal_test.go`
- `monitor_external_test.go`

---

### 4. internal/application/metrics (95.0% → 100%)

**Coverage actuel:** 95.0%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `Unsubscribe` | 42.9% |
| `UpdateState` | 90.9% |
| `UpdateHealth` | 85.7% |

**Actions:**
1. Tester `Unsubscribe` avec subscriber inexistant
2. Compléter `UpdateState` edge cases
3. Tester `UpdateHealth` avec différents états

**Fichiers à modifier:**
- `tracker_external_test.go`

---

### 5. internal/application/supervisor (96.6% → 100%)

**Coverage actuel:** 96.6%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `stopAll` | 87.5% |
| `Reload` | 94.1% |
| `updateServices` | 76.9% |
| `removeDeletedServices` | 87.5% |

**Actions:**
1. Tester `stopAll` avec erreurs de stop
2. Compléter `Reload` avec services ajoutés/supprimés
3. Tester `updateServices` avec configurations invalides
4. Couvrir `removeDeletedServices` edge cases

**Fichiers à modifier:**
- `supervisor_internal_test.go`
- `supervisor_external_test.go`

---

### 6. internal/domain/lifecycle (98.0% → 100%)

**Coverage actuel:** 98.0%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `Category` | 66.7% |

**Actions:**
1. Tester tous les event types dans `Category()`

**Fichiers à modifier:**
- `event_internal_test.go`

---

### 7. internal/domain/listener (95.7% → 100%)

**Coverage actuel:** 95.7%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `CanTransitionTo` | 80.0% |

**Actions:**
1. Tester toutes les transitions d'état invalides

**Fichiers à modifier:**
- `state_external_test.go`

---

### 8. internal/infrastructure/observability/healthcheck (92.5% → 100%)

**Coverage actuel:** 92.5%
**Fonctions non couvertes:**
| Fichier | Fonction | Coverage |
|---------|----------|----------|
| `exec.go` | `executeCommand` | 84.2% |
| `grpc.go` | `Probe` | 91.7% |
| `http.go` | `getStatusCode` | 72.2% |
| `icmp.go` | `Probe` | 85.7% |
| `icmp.go` | `tcpPing` | 90.9% |
| `udp.go` | `Probe` | 88.9% |
| `udp.go` | `dialUDP` | 78.6% |
| `udp.go` | `sendAndReceive` | 85.7% |

**Actions:**
1. Tester `executeCommand` avec timeouts et erreurs
2. Tester `getStatusCode` avec redirections et erreurs HTTP
3. Compléter tests UDP avec erreurs réseau
4. Ajouter tests ICMP edge cases

**Fichiers à modifier:**
- `exec_internal_test.go`
- `http_internal_test.go`
- `udp_internal_test.go`
- `icmp_internal_test.go`

---

### 9. internal/infrastructure/observability/logging (99.4% → 100%)

**Coverage actuel:** 99.4%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `rotate` | 92.3% |

**Actions:**
1. Tester `rotate` avec erreurs de fichier

**Fichiers à modifier:**
- `writer_internal_test.go`

---

### 10. internal/infrastructure/persistence/storage/boltdb (88.4% → 100%)

**Coverage actuel:** 88.4%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `NewStore` | 75.0% |
| `initSchema` | 76.9% |
| `WriteSystemCPU` | 88.9% |
| `WriteSystemMemory` | 88.9% |
| `WriteProcessMetrics` | 83.3% |
| `GetSystemCPU` | 93.3% |
| `GetSystemMemory` | 93.3% |
| `GetProcessMetrics` | 94.4% |
| `GetLatestSystemCPU` | 94.1% |
| `GetLatestSystemMemory` | 94.1% |
| `GetLatestProcessMetrics` | 90.0% |
| `pruneTransaction` | 82.4% |
| `pruneProcessMetricsBuckets` | 83.3% |
| `pruneBucketHelper` | 87.5% |
| `encodeSystemCPU` | 75.0% |
| `encodeSystemMemory` | 75.0% |
| `encodeProcessMetrics` | 75.0% |

**Actions:**
1. Tester `NewStore` avec erreurs d'ouverture DB
2. Tester toutes les méthodes d'écriture avec erreurs d'encodage
3. Compléter tests de lecture avec données corrompues
4. Tester pruning avec différentes conditions

**Fichiers à modifier:**
- `store_internal_test.go`

---

### 11. internal/infrastructure/process/control (88.9% → 100%)

**Coverage actuel:** 88.9%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `NewControl` | 0.0% |

**Actions:**
1. Tester `NewControl` (constructeur)

**Fichiers à modifier:**
- `process_unix_external_test.go`

---

### 12. internal/infrastructure/process/credentials (98.6% → 100%)

**Coverage actuel:** 98.6%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `NewManager` | 0.0% |

**Actions:**
1. Tester `NewManager` (constructeur)

**Fichiers à modifier:**
- Créer `credentials_unix_external_test.go`

---

### 13. internal/infrastructure/process/signals (94.3% → 100%)

**Coverage actuel:** 94.3%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `IsSubreaper` | 80.0% |
| `prctlSubreaper` | 75.0% |

**Actions:**
1. Tester `IsSubreaper` avec différents états
2. Compléter tests `prctlSubreaper`

**Fichiers à modifier:**
- `signals_linux_internal_test.go`

---

### 14. internal/infrastructure/resources/cgroup (86.0% → 100%)

**Coverage actuel:** 86.0%
**Fonctions non couvertes (nombreuses):**

**V1:**
- `NewV1Reader` - 27.3%
- `detectV1Cgroup` - 75.0%
- `CPUUsage` - 77.8%
- `CPULimit` - 88.9%
- `readCPUQuota` - 91.7%
- `readCPUPeriod` - 85.7%
- `MemoryUsage` - 77.8%
- `MemoryLimit` - 62.5%
- `ReadMemoryStat` - 88.9%

**V2:**
- `NewV2Reader` - 87.5%
- `detectCurrentCgroup` - 75.0%
- `CPUUsage` - 83.3%
- `CPULimit` - 84.2%
- `MemoryUsage` - 88.9%
- `MemoryLimit` - 92.9%
- `ReadMemoryStat` - 95.0%

**Detector:**
- `String` - 83.3%
- `IsContainerized` - 28.6%
- `NewReaderWithPath` - 50.0%

**Actions:**
1. Créer mocks pour système de fichiers cgroup
2. Tester tous les chemins de détection V1/V2
3. Tester lecture avec différents formats de fichiers
4. Tester `IsContainerized` avec différents environnements

**Fichiers à modifier:**
- `cgroup_external_test.go`
- `detector_external_test.go`
- `v1_external_test.go` (créer si nécessaire)
- `v2_external_test.go`

---

### 15. internal/infrastructure/resources/metrics (53.3% → 100%)

**Coverage actuel:** 53.3%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `NewSystemCollector` | 40.0% |
| `isBSD` | 66.7% |
| `DetectedPlatform` | 40.0% |

**Actions:**
1. Tester `NewSystemCollector` pour toutes les plateformes
2. Compléter tests `isBSD` et `DetectedPlatform`

**Fichiers à modifier:**
- `factory_internal_test.go`
- `factory_external_test.go`

---

### 16. internal/infrastructure/resources/metrics/linux (95.8% → 100%)

**Coverage actuel:** 95.8%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `CollectSystem` | 80.0% |
| `parseProcessStat` | 93.3% |
| `CollectAllProcesses` | 83.3% |
| `readMemInfo` | 92.9% |
| `readProcessStatus` | 94.7% |
| `CollectAllProcesses` (memory) | 77.8% |

**Actions:**
1. Tester avec fichiers /proc malformés
2. Compléter parsing edge cases
3. Tester avec processus disparus

**Fichiers à modifier:**
- `cpu_external_test.go`
- `memory_external_test.go`

---

### 17. internal/infrastructure/resources/metrics/scratch (71.1% → 100%)

**Coverage actuel:** 71.1%
**Fonctions non couvertes:**
| Fichier | Fonction | Coverage |
|---------|----------|----------|
| `cpu_collector.go` | `CollectSystem` | 66.7% |
| `cpu_collector.go` | `CollectProcess` | 60.0% |
| `cpu_collector.go` | `CollectAllProcesses` | 66.7% |
| `cpu_collector.go` | `CollectLoadAverage` | 66.7% |
| `cpu_collector.go` | `CollectPressure` | 66.7% |
| `disk_collector.go` | `ListPartitions` | 66.7% |
| `disk_collector.go` | `CollectUsage` | 60.0% |
| `disk_collector.go` | `CollectAllUsage` | 66.7% |
| `disk_collector.go` | `CollectIO` | 66.7% |
| `disk_collector.go` | `CollectDeviceIO` | 60.0% |
| `io_collector.go` | `CollectStats` | 66.7% |
| `io_collector.go` | `CollectPressure` | 66.7% |
| `memory_collector.go` | `CollectSystem` | 66.7% |
| `memory_collector.go` | `CollectProcess` | 60.0% |
| `memory_collector.go` | `CollectAllProcesses` | 66.7% |
| `memory_collector.go` | `CollectPressure` | 66.7% |
| `network_collector.go` | `ListInterfaces` | 66.7% |
| `network_collector.go` | `CollectStats` | 60.0% |
| `network_collector.go` | `CollectAllStats` | 66.7% |

**Actions:**
1. Les collectors scratch retournent `ErrNotImplemented`
2. Tester les 3 chemins: success (nil context), cancel, not implemented

**Fichiers à modifier:**
- `cpu_collector_external_test.go` (créer)
- `disk_collector_external_test.go` (créer)
- `io_collector_external_test.go` (créer)
- `memory_collector_external_test.go` (créer)
- `network_collector_external_test.go` (créer)

---

### 18. internal/infrastructure/transport/grpc (86.1% → 100%)

**Coverage actuel:** 86.1%
**Fonctions non couvertes:**
| Fonction | Coverage |
|----------|----------|
| `streamLoop` | 93.3% |
| `Serve` | 87.5% |
| `GetState` | 66.7% |
| `StreamState` | 83.3% |
| `ListProcesses` | 77.8% |
| `GetProcess` | 83.3% |
| `StreamProcessMetrics` | 75.0% |
| `GetSystemMetrics` | 66.7% |
| `StreamSystemMetrics` | 83.3% |
| `StreamAllProcessMetrics` | 58.3% |

**Actions:**
1. Tester tous les endpoints RPC avec erreurs
2. Compléter tests de streaming avec contexte annulé
3. Tester `Serve` avec erreurs de binding

**Fichiers à modifier:**
- `server_external_test.go`

---

## Priorités d'Implémentation

### Phase 1 - Quick Wins (Effort faible, Impact élevé)

| Package | De → À | Effort |
|---------|--------|--------|
| `domain/lifecycle` | 98% → 100% | 1h |
| `domain/listener` | 95.7% → 100% | 1h |
| `observability/logging` | 99.4% → 100% | 30min |
| `process/control` | 88.9% → 100% | 30min |
| `process/credentials` | 98.6% → 100% | 30min |

### Phase 2 - Application Layer

| Package | De → À | Effort |
|---------|--------|--------|
| `application/metrics` | 95% → 100% | 2h |
| `application/supervisor` | 96.6% → 100% | 2h |
| `application/health` | 88% → 100% | 4h |

### Phase 3 - Infrastructure Core

| Package | De → À | Effort |
|---------|--------|--------|
| `healthcheck` | 92.5% → 100% | 3h |
| `boltdb` | 88.4% → 100% | 4h |
| `signals` | 94.3% → 100% | 2h |

### Phase 4 - Resources/Metrics

| Package | De → À | Effort |
|---------|--------|--------|
| `resources/metrics` | 53.3% → 100% | 2h |
| `resources/metrics/linux` | 95.8% → 100% | 2h |
| `resources/metrics/scratch` | 71.1% → 100% | 3h |
| `resources/cgroup` | 86% → 100% | 4h |

### Phase 5 - Transport & Bootstrap

| Package | De → À | Effort |
|---------|--------|--------|
| `transport/grpc` | 86.1% → 100% | 4h |
| `bootstrap` | 40% → 100% | 4h |

### Phase 6 - Entry Point (Optionnel)

| Package | De → À | Effort |
|---------|--------|--------|
| `cmd/daemon` | 0% → 100% | 2h |

---

## Estimation Totale

| Phase | Durée Estimée |
|-------|---------------|
| Phase 1 | 3.5h |
| Phase 2 | 8h |
| Phase 3 | 9h |
| Phase 4 | 11h |
| Phase 5 | 8h |
| Phase 6 | 2h (optionnel) |
| **Total** | **~40h** |

---

## Rollback Plan

Si les tests introduisent des régressions:
1. Revenir au commit précédent
2. Identifier le test problématique
3. Corriger le test sans modifier le code de production

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Tests flaky sur I/O | Utiliser mocks et interfaces |
| Tests dépendant de l'OS | Build tags et tests conditionnels |
| Temps d'exécution des tests | Parallélisation avec `t.Parallel()` |
| Code inatteignable | Refactorer ou marquer explicitement |
