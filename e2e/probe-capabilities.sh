#!/bin/sh
# Extract probe capability report from supervizio --probe output.
# Outputs a JSON capability report for each metric section.
# Usage: probe-capabilities.sh <binary_path> <output_json_path> [platform_override]
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
  cat > "$OUTPUT_FILE" << EOF
{"platform":"$PLATFORM","error":true,"exit_code":$PROBE_EXIT}
EOF
  echo "Probe capabilities for $PLATFORM saved to $OUTPUT_FILE (probe failed)"
  cat "$OUTPUT_FILE"
  exit 0
fi

# Check if a key exists in the JSON output
check() { echo "$PROBE_OUTPUT" | grep -q "\"$1\"" && echo "true" || echo "false"; }

# Generate capability JSON
cat > "$OUTPUT_FILE" << CAPABILITY_EOF
{
  "platform": "$PLATFORM",
  "cpu": $(check cpu),
  "memory": $(check memory),
  "load": $(check load),
  "disk": $(check disk),
  "network": $(check network),
  "io": $(check io),
  "process": $(check process),
  "thermal": $(check thermal),
  "psi": $(echo "$PROBE_OUTPUT" | grep -q '"pressure"' && echo "true" || echo "false"),
  "context_switches": $(check context_switches),
  "connections": $(check connections),
  "quota": $(check quota),
  "container": $(check container),
  "runtime": $(check runtime)
}
CAPABILITY_EOF

echo "Probe capabilities for $PLATFORM saved to $OUTPUT_FILE"
cat "$OUTPUT_FILE"
