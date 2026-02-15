#!/usr/bin/env bash
# vm-release.sh - Release a Proxmox VM (remove lock)
#
# Usage: vm-release.sh <VMID>
#
# Required env vars:
#   SSH_OPTS         - SSH options string
#   VM_IP            - IP address of the VM (set by vm-acquire.sh)
#
# The script removes the /usebyjob lock file so other jobs can acquire
# the VM. The VM stays running — vm-cleanup resets all VMs to "base"
# snapshot at the end of the pipeline.
set -euo pipefail

VMID="${1:?Usage: vm-release.sh <VMID>}"

# Validate required environment variables
if [ -z "${VM_IP:-}" ] || [ -z "${SSH_OPTS:-}" ]; then
  echo "[vm-release] WARNING: Missing VM_IP or SSH_OPTS, skipping release"
  exit 0
fi

echo "[vm-release] Releasing VM ${VMID} (removing /usebyjob)..."
ssh ${SSH_OPTS} root@"${VM_IP}" "rm -f /usebyjob" 2>/dev/null || true
echo "[vm-release] VM ${VMID} released"
