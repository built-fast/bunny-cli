#!/usr/bin/env bats

load "test_helper"

# --- List Pull Zones ---

@test "pullzones list returns success" {
  run "$BUNNY_BINARY" pullzones list --limit 1
  assert_success
}

@test "pullzones list --output json returns valid JSON" {
  run "$BUNNY_BINARY" pullzones list --limit 1 --output json
  assert_success
  is_valid_json
}

@test "pullzones list default output contains table headers" {
  run "$BUNNY_BINARY" pullzones list --limit 1
  assert_success
  assert_output_contains "Id"
  assert_output_contains "Name"
}

# --- Get Pull Zone ---

@test "pullzones get returns success for a valid ID" {
  run "$BUNNY_BINARY" pullzones get 1
  assert_success
}

@test "pullzones get --output json returns valid JSON" {
  run "$BUNNY_BINARY" pullzones get 1 --output json
  assert_success
  is_valid_json
}

@test "pullzones get without ID fails" {
  run "$BUNNY_BINARY" pullzones get
  assert_failure
}

# --- Create Pull Zone ---

@test "pullzones create returns success" {
  run "$BUNNY_BINARY" pullzones create --name "e2e-test-zone" --no-input
  assert_success
}

@test "pullzones create --output json returns valid JSON" {
  run "$BUNNY_BINARY" pullzones create --name "e2e-json-zone" --output json --no-input
  assert_success
  is_valid_json
}

# --- Delete Pull Zone ---

@test "pullzones delete --yes returns success" {
  run "$BUNNY_BINARY" pullzones delete 1 --yes
  assert_success
}

@test "pullzones delete without ID fails" {
  run "$BUNNY_BINARY" pullzones delete --yes
  assert_failure
}

# --- Purge Cache ---

@test "pullzones purge --yes returns success" {
  run "$BUNNY_BINARY" pullzones purge 1 --yes
  assert_success
}

# --- Subcommands ---

@test "pullzones hostnames list returns success" {
  run "$BUNNY_BINARY" pullzones hostnames list 1
  assert_success
}

@test "pullzones edge-rules list returns success" {
  run "$BUNNY_BINARY" pullzones edge-rules list 1
  assert_success
}

# --- Alias ---

@test "pz alias works" {
  run "$BUNNY_BINARY" pz list --limit 1
  assert_success
}
