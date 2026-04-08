#!/usr/bin/env bats

load "test_helper"

# --- List Storage Zones ---

@test "storagezones list returns success" {
  run "$BUNNY_BINARY" storagezones list --limit 1
  assert_success
}

@test "storagezones list --output json returns valid JSON" {
  run "$BUNNY_BINARY" storagezones list --limit 1 --output json
  assert_success
  is_valid_json
}

@test "storagezones list default output contains table headers" {
  run "$BUNNY_BINARY" storagezones list --limit 1
  assert_success
  assert_output_contains "Id"
  assert_output_contains "Name"
}

# --- Get Storage Zone ---

@test "storagezones get returns success for a valid ID" {
  run "$BUNNY_BINARY" storagezones get 1
  assert_success
}

@test "storagezones get --output json returns valid JSON" {
  run "$BUNNY_BINARY" storagezones get 1 --output json
  assert_success
  is_valid_json
}

@test "storagezones get without ID fails" {
  run "$BUNNY_BINARY" storagezones get
  assert_failure
}

# --- Create Storage Zone ---

@test "storagezones create returns success" {
  run "$BUNNY_BINARY" storagezones create --name "e2e-test-storage" --region "DE" --no-input
  assert_success
}

@test "storagezones create --output json returns valid JSON" {
  run "$BUNNY_BINARY" storagezones create --name "e2e-json-storage" --region "NY" --output json --no-input
  assert_success
  is_valid_json
}

# --- Delete Storage Zone ---

@test "storagezones delete --yes returns success" {
  run "$BUNNY_BINARY" storagezones delete 1 --yes
  assert_success
}

@test "storagezones delete without ID fails" {
  run "$BUNNY_BINARY" storagezones delete --yes
  assert_failure
}

# --- Reset Password ---

@test "storagezones reset-password --yes returns success" {
  run "$BUNNY_BINARY" storagezones reset-password 1 --yes
  assert_success
}

@test "storagezones reset-password without ID fails" {
  run "$BUNNY_BINARY" storagezones reset-password --yes
  assert_failure
}

# --- Alias ---

@test "sz alias works" {
  run "$BUNNY_BINARY" sz list --limit 1
  assert_success
}
