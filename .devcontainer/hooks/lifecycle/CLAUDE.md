# Lifecycle Hooks

## Purpose

DevContainer lifecycle scripts executed at specific container events.

## Scripts

| Script | Event | Description |
|--------|-------|-------------|
| `initialize.sh` | onCreateCommand | Initial setup |
| `onCreate.sh` | onCreate | Container creation |
| `postCreate.sh` | postCreate | After container ready |
| `postAttach.sh` | postAttach | After VS Code attaches |
| `postStart.sh` | postStart | After each start |
| `updateContent.sh` | updateContent | Content updates |

## Execution Order

1. initialize.sh (earliest)
2. onCreate.sh
3. postCreate.sh
4. postStart.sh
5. postAttach.sh (latest)

## Conventions

- All scripts must be executable (`chmod +x`)
- Use bash strict mode: `set -euo pipefail`
- Source `../shared/utils.sh` for common functions
- Exit 0 on success, non-zero on failure
