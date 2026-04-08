#!/usr/bin/env bats

load "test_helper"

# --- List DNS Zones ---

@test "dns list returns success" {
  run "$BUNNY_BINARY" dns list --limit 1
  assert_success
}

@test "dns list --output json returns valid JSON" {
  run "$BUNNY_BINARY" dns list --limit 1 --output json
  assert_success
  is_valid_json
}

@test "dns list default output contains table headers" {
  run "$BUNNY_BINARY" dns list --limit 1
  assert_success
  assert_output_contains "Id"
  assert_output_contains "Domain"
}

# --- Get DNS Zone ---

@test "dns get returns success for a valid ID" {
  run "$BUNNY_BINARY" dns get 1
  assert_success
}

@test "dns get --output json returns valid JSON" {
  run "$BUNNY_BINARY" dns get 1 --output json
  assert_success
  is_valid_json
}

@test "dns get without ID fails" {
  run "$BUNNY_BINARY" dns get
  assert_failure
}

# --- Create DNS Zone ---

@test "dns create returns success" {
  run "$BUNNY_BINARY" dns create --domain "e2e-test.com" --no-input
  assert_success
}

# --- Delete DNS Zone ---

@test "dns delete --yes returns success" {
  run "$BUNNY_BINARY" dns delete 1 --yes
  assert_success
}

@test "dns delete without ID fails" {
  run "$BUNNY_BINARY" dns delete --yes
  assert_failure
}

# --- Export DNS Zone ---

@test "dns export returns success" {
  run "$BUNNY_BINARY" dns export 1
  assert_success
}

# --- Records ---

@test "dns records list returns success" {
  run "$BUNNY_BINARY" dns records list 1
  assert_success
}

# --- DNSSEC ---

@test "dns dnssec enable returns success" {
  run "$BUNNY_BINARY" dns dnssec enable 1
  assert_success
}

# --- Alias ---

@test "dnszone alias works" {
  run "$BUNNY_BINARY" dnszone list --limit 1
  assert_success
}
