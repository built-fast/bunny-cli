#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

# --- Dependency checks ---
missing=()
for cmd in bats npx jq curl; do
  command -v "$cmd" >/dev/null 2>&1 || missing+=("$cmd")
done
if [ ${#missing[@]} -gt 0 ]; then
  echo "ERROR: Missing required dependencies: ${missing[*]}" >&2
  echo "Please install them before running the e2e test suite." >&2
  exit 1
fi

# --- Build CLI binary ---
echo "==> Building CLI binary..."
make build
BUNNY_BINARY="$PROJECT_ROOT/bin/bunny"

# --- Select a random available port ---
get_random_port() {
  python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()'
}
PRISM_PORT=$(get_random_port)
PRISM_URL="http://127.0.0.1:${PRISM_PORT}"

# --- OpenAPI specs for Prism ---
PRISM_SPEC="$PROJECT_ROOT/openapi/core-platform-api.json"
if [ ! -f "$PRISM_SPEC" ]; then
  echo "ERROR: OpenAPI spec not found at $PRISM_SPEC" >&2
  exit 1
fi

PRISM_COMPUTE_SPEC="$PROJECT_ROOT/openapi/edge-scripting-api.json"
if [ ! -f "$PRISM_COMPUTE_SPEC" ]; then
  echo "ERROR: OpenAPI spec not found at $PRISM_COMPUTE_SPEC" >&2
  exit 1
fi

# --- Select a second random port for compute API ---
PRISM_COMPUTE_PORT=$(get_random_port)
PRISM_COMPUTE_URL="http://127.0.0.1:${PRISM_COMPUTE_PORT}"

# --- Cleanup trap ---
PRISM_PID=""
PRISM_COMPUTE_PID=""
cleanup() {
  if [ -n "$PRISM_PID" ]; then
    kill "$PRISM_PID" 2>/dev/null || true
    wait "$PRISM_PID" 2>/dev/null || true
  fi
  if [ -n "$PRISM_COMPUTE_PID" ]; then
    kill "$PRISM_COMPUTE_PID" 2>/dev/null || true
    wait "$PRISM_COMPUTE_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

# --- Start Prism mock servers ---
echo "==> Starting Prism mock server (platform) on port ${PRISM_PORT}..."
npx @stoplight/prism-cli mock "$PRISM_SPEC" \
  --port "$PRISM_PORT" \
  --host 127.0.0.1 \
  > /dev/null 2>&1 &
PRISM_PID=$!

echo "==> Starting Prism mock server (compute) on port ${PRISM_COMPUTE_PORT}..."
npx @stoplight/prism-cli mock "$PRISM_COMPUTE_SPEC" \
  --port "$PRISM_COMPUTE_PORT" \
  --host 127.0.0.1 \
  > /dev/null 2>&1 &
PRISM_COMPUTE_PID=$!

# --- Wait for Prism to be ready ---
wait_for_prism() {
  local url="$1"
  local name="$2"
  local elapsed=0
  local timeout=30
  echo "==> Waiting for Prism ($name) to be ready..."
  while [ "$elapsed" -lt "$timeout" ]; do
    if curl -so /dev/null "${url}/" 2>&1; then
      echo "==> Prism ($name) is ready."
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  echo "ERROR: Prism ($name) did not become ready within ${timeout} seconds." >&2
  exit 1
}

wait_for_prism "$PRISM_URL" "platform"
wait_for_prism "$PRISM_COMPUTE_URL" "compute"

# --- Run BATS tests ---
echo "==> Running e2e tests..."
export PRISM_URL
export PRISM_COMPUTE_URL
export BUNNY_BINARY

bats e2e/*.bats
