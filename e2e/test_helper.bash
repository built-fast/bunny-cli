#!/usr/bin/env bash
# Shared test helper for BATS e2e tests.
# Provides setup/teardown, config helpers, and assertion utilities.

# --- Setup / Teardown ---

setup() {
  # Create an isolated temp config directory per test
  TEST_TEMP_DIR="$(mktemp -d)"
  export XDG_CONFIG_HOME="$TEST_TEMP_DIR"
  export BUNNY_API_KEY="test-api-key"
  export BUNNY_API_URL="${PRISM_URL:?PRISM_URL must be set}"
}

teardown() {
  if [ -n "${TEST_TEMP_DIR:-}" ] && [ -d "$TEST_TEMP_DIR" ]; then
    rm -rf "$TEST_TEMP_DIR"
  fi
}

# --- Assertion Helpers ---

# Asserts that $status is 0.
assert_success() {
  if [ "$status" -ne 0 ]; then
    echo "assert_success failed" >&2
    echo "  expected exit code: 0" >&2
    echo "  actual exit code:   $status" >&2
    echo "  output: $output" >&2
    return 1
  fi
}

# Asserts that $status is non-zero.
assert_failure() {
  if [ "$status" -eq 0 ]; then
    echo "assert_failure failed" >&2
    echo "  expected: non-zero exit code" >&2
    echo "  actual exit code: 0" >&2
    echo "  output: $output" >&2
    return 1
  fi
}

# Asserts that $output contains the given substring.
# Usage: assert_output_contains "expected text"
assert_output_contains() {
  local expected="$1"
  if [[ "$output" != *"$expected"* ]]; then
    echo "assert_output_contains failed" >&2
    echo "  expected to contain: $expected" >&2
    echo "  actual output:       $output" >&2
    return 1
  fi
}

# Asserts that $output does NOT contain the given substring.
# Usage: assert_output_not_contains "unwanted text"
assert_output_not_contains() {
  local expected="$1"
  if [[ "$output" == *"$expected"* ]]; then
    echo "assert_output_not_contains failed" >&2
    echo "  expected NOT to contain: $expected" >&2
    echo "  actual output:           $output" >&2
    return 1
  fi
}

# Asserts that $output is valid JSON.
is_valid_json() {
  if ! echo "$output" | jq empty 2>/dev/null; then
    echo "is_valid_json failed" >&2
    echo "  output is not valid JSON" >&2
    echo "  actual output: $output" >&2
    return 1
  fi
}

# Asserts that $output (JSON) has the given value at the given jq path.
# Usage: assert_json_value ".field" "expected_value"
assert_json_value() {
  local jq_path="$1"
  local expected="$2"
  local actual

  actual="$(echo "$output" | jq -r "$jq_path" 2>/dev/null)" || {
    echo "assert_json_value failed" >&2
    echo "  could not parse JSON or evaluate path: $jq_path" >&2
    echo "  actual output: $output" >&2
    return 1
  }

  if [ "$actual" != "$expected" ]; then
    echo "assert_json_value failed" >&2
    echo "  jq path:        $jq_path" >&2
    echo "  expected value:  $expected" >&2
    echo "  actual value:    $actual" >&2
    return 1
  fi
}
