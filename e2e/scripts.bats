#!/usr/bin/env bats

load "test_helper"

# Override API URL to point at the compute Prism instance
setup() {
  TEST_TEMP_DIR="$(mktemp -d)"
  export XDG_CONFIG_HOME="$TEST_TEMP_DIR"
  export BUNNY_API_KEY="test-api-key"
  export BUNNY_API_URL="${PRISM_COMPUTE_URL:?PRISM_COMPUTE_URL must be set}"
}

# --- List Edge Scripts ---

@test "scripts list returns success" {
  run "$BUNNY_BINARY" scripts list --limit 1
  assert_success
}

@test "scripts list --output json returns valid JSON" {
  run "$BUNNY_BINARY" scripts list --limit 1 --output json
  assert_success
  is_valid_json
}

@test "scripts list default output contains table headers" {
  run "$BUNNY_BINARY" scripts list --limit 1
  assert_success
  assert_output_contains "Id"
  assert_output_contains "Name"
}

# --- Get Edge Script ---

@test "scripts get returns success for a valid ID" {
  run "$BUNNY_BINARY" scripts get 1
  assert_success
}

@test "scripts get --output json returns valid JSON" {
  run "$BUNNY_BINARY" scripts get 1 --output json
  assert_success
  is_valid_json
}

@test "scripts get without ID fails" {
  run "$BUNNY_BINARY" scripts get
  assert_failure
}

# --- Delete Edge Script ---

@test "scripts delete --yes returns success" {
  run "$BUNNY_BINARY" scripts delete 1 --yes
  assert_success
}

@test "scripts delete without ID fails" {
  run "$BUNNY_BINARY" scripts delete --yes
  assert_failure
}

# --- Code ---

@test "scripts code get returns success" {
  run "$BUNNY_BINARY" scripts code get 1
  assert_success
}

# --- Releases ---

@test "scripts releases list returns success" {
  run "$BUNNY_BINARY" scripts releases list 1 --limit 1
  assert_success
}

@test "scripts releases list --output json returns valid JSON" {
  run "$BUNNY_BINARY" scripts releases list 1 --limit 1 --output json
  assert_success
  is_valid_json
}

@test "scripts releases active returns success" {
  run "$BUNNY_BINARY" scripts releases active 1
  assert_success
}

# --- Secrets ---

@test "scripts secrets list returns success" {
  run "$BUNNY_BINARY" scripts secrets list 1
  assert_success
}

# --- Alias ---

@test "compute alias works" {
  run "$BUNNY_BINARY" compute list --limit 1
  assert_success
}
