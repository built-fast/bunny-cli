#!/usr/bin/env bats

load "test_helper"

# --- List Regions ---

@test "regions returns success" {
  run "$BUNNY_BINARY" regions
  assert_success
}

@test "regions --output json returns valid JSON" {
  run "$BUNNY_BINARY" regions --output json
  assert_success
  is_valid_json
}

@test "regions default output contains table headers" {
  run "$BUNNY_BINARY" regions
  assert_success
  assert_output_contains "Id"
  assert_output_contains "Name"
  assert_output_contains "Region Code"
  assert_output_contains "Continent"
}

@test "regions --field filters output" {
  run "$BUNNY_BINARY" regions -f "Name,Region Code"
  assert_success
  assert_output_contains "Name"
  assert_output_contains "Region Code"
}
