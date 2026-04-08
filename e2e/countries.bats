#!/usr/bin/env bats

load "test_helper"

# --- List Countries ---

@test "countries returns success" {
  run "$BUNNY_BINARY" countries
  assert_success
}

@test "countries --output json returns valid JSON" {
  run "$BUNNY_BINARY" countries --output json
  assert_success
  is_valid_json
}

@test "countries default output contains table headers" {
  run "$BUNNY_BINARY" countries
  assert_success
  assert_output_contains "Name"
  assert_output_contains "ISO Code"
  assert_output_contains "EU"
}
