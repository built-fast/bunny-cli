#!/usr/bin/env bats

load "test_helper"

# --- Billing Help ---

@test "billing --help shows subcommands" {
  run "$BUNNY_BINARY" billing --help
  assert_success
  assert_output_contains "details"
  assert_output_contains "records"
  assert_output_contains "summary"
  assert_output_contains "invoice"
}

# --- Billing Details ---

@test "billing details returns success" {
  run "$BUNNY_BINARY" billing details
  assert_success
}

@test "billing details --output json returns valid JSON" {
  run "$BUNNY_BINARY" billing details --output json
  assert_success
  is_valid_json
}

@test "billing details default output contains table headers" {
  run "$BUNNY_BINARY" billing details
  assert_success
  assert_output_contains "Balance"
  assert_output_contains "This Month"
}

# --- Billing Records ---

@test "billing records returns success" {
  run "$BUNNY_BINARY" billing records
  assert_success
}

@test "billing records --output json returns valid JSON" {
  run "$BUNNY_BINARY" billing records --output json
  assert_success
  is_valid_json
}

# --- Billing Summary ---

@test "billing summary returns success" {
  run "$BUNNY_BINARY" billing summary
  assert_success
}

@test "billing summary --output json returns valid JSON" {
  run "$BUNNY_BINARY" billing summary --output json
  assert_success
  is_valid_json
}

# --- Billing Invoice ---

@test "billing invoice without ID fails" {
  run "$BUNNY_BINARY" billing invoice
  assert_failure
}

@test "billing invoice downloads PDF" {
  run "$BUNNY_BINARY" billing invoice 1 -o "$TEST_TEMP_DIR/invoice.pdf"
  assert_success
  assert_output_contains "Invoice saved to"
  [ -f "$TEST_TEMP_DIR/invoice.pdf" ]
}
