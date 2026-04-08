#!/usr/bin/env bats

load "test_helper"

# --- Statistics ---

@test "statistics returns success" {
  run "$BUNNY_BINARY" statistics
  assert_success
}

@test "statistics --output json returns valid JSON" {
  run "$BUNNY_BINARY" statistics --output json
  assert_success
  is_valid_json
}

@test "statistics default output contains table headers" {
  run "$BUNNY_BINARY" statistics
  assert_success
  assert_output_contains "Total Bandwidth"
  assert_output_contains "Cache Hit Rate"
}

@test "statistics with date filters returns success" {
  run "$BUNNY_BINARY" statistics --date-from "2024-01-01T00:00:00Z" --date-to "2024-01-31T23:59:59Z"
  assert_success
}

@test "statistics --hourly returns success" {
  run "$BUNNY_BINARY" statistics --hourly
  assert_success
}

# --- Alias ---

@test "stats alias works" {
  run "$BUNNY_BINARY" stats
  assert_success
}
