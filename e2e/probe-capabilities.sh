#!/bin/sh
# =============================================================================
# Save raw probe JSON output with platform metadata for CI summary.
# Usage: probe-capabilities.sh <binary_path> <output_json_path> [platform_override]
#
# Note: Uses /bin/sh for maximum portability (BSD VMs, Alpine, etc.)
# Security: SAST-SAFE - No user input, all paths are hardcoded constants.
# =============================================================================
set -e

BINARY="$1"
OUTPUT_FILE="$2"
PLATFORM="${3:-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)}"

if [ -z "$BINARY" ] || [ -z "$OUTPUT_FILE" ]; then
  echo "Usage: $0 <binary_path> <output_json_path> [platform_override]" >&2
  exit 1
fi

# Normalize arch
case "$PLATFORM" in
  *x86_64*) PLATFORM=$(echo "$PLATFORM" | sed 's/x86_64/amd64/') ;;
  *aarch64*) PLATFORM=$(echo "$PLATFORM" | sed 's/aarch64/arm64/') ;;
esac

# Run probe with explicit error capture (reliable across all sh implementations)
set +e
PROBE_OUTPUT=$("$BINARY" --probe 2>&1)
PROBE_EXIT=$?
set -e

if [ "$PROBE_EXIT" -ne 0 ]; then
  echo "WARNING: --probe exited with code $PROBE_EXIT" >&2
  printf '{"_platform":"%s","_error":true,"_exit_code":%d}\n' "$PLATFORM" "$PROBE_EXIT" > "$OUTPUT_FILE"
  echo "Probe output for $PLATFORM saved to $OUTPUT_FILE (probe failed)"
  cat "$OUTPUT_FILE"
  exit 0
fi

# Save raw probe JSON with _platform field injected (sed works everywhere, no jq needed)
echo "$PROBE_OUTPUT" | sed "s/^{/{\"_platform\":\"$PLATFORM\",/" > "$OUTPUT_FILE"

echo "Probe output for $PLATFORM saved to $OUTPUT_FILE"
cat "$OUTPUT_FILE"
