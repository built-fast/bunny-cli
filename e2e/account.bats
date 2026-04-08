#!/usr/bin/env bats

load "test_helper"

# --- Account Help ---

@test "account --help shows subcommands" {
  run "$BUNNY_BINARY" account --help
  assert_success
  assert_output_contains "api-keys"
  assert_output_contains "audit-log"
}

# --- API Keys ---

@test "account api-keys list returns success" {
  run "$BUNNY_BINARY" account api-keys list
  assert_success
}

@test "account api-keys list --output json returns valid JSON" {
  run "$BUNNY_BINARY" account api-keys list --output json
  assert_success
  is_valid_json
}

@test "account api-keys list default output contains table headers" {
  run "$BUNNY_BINARY" account api-keys list
  assert_success
  assert_output_contains "Id"
  assert_output_contains "Key"
  assert_output_contains "Roles"
}

@test "account apikeys alias works" {
  run "$BUNNY_BINARY" account apikeys list
  assert_success
}

# --- Audit Log ---

@test "account audit-log returns success" {
  run "$BUNNY_BINARY" account audit-log "2024-01-15T00:00:00Z"
  assert_success
}

@test "account audit-log --output json returns valid JSON" {
  run "$BUNNY_BINARY" account audit-log "2024-01-15T00:00:00Z" --output json
  assert_success
  is_valid_json
}

@test "account audit-log without date fails" {
  run "$BUNNY_BINARY" account audit-log
  assert_failure
}

@test "account audit-log with filters returns success" {
  run "$BUNNY_BINARY" account audit-log "2024-01-15T00:00:00Z" --order Descending
  assert_success
}
