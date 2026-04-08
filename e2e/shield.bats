#!/usr/bin/env bats

load "test_helper"

# Override setup to point at the Shield Prism instance
setup() {
  TEST_TEMP_DIR="$(mktemp -d)"
  export XDG_CONFIG_HOME="$TEST_TEMP_DIR"
  export BUNNY_API_KEY="test-api-key"
  export BUNNY_API_URL="${PRISM_SHIELD_URL:?PRISM_SHIELD_URL must be set}"
}

# --- Shield Zones ---

@test "shield zones list returns success" {
  run "$BUNNY_BINARY" shield zones list --limit 1
  assert_success
}

@test "shield zones list --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield zones list --limit 1 --output json
  assert_success
  is_valid_json
}

@test "shield zones list default output contains table headers" {
  run "$BUNNY_BINARY" shield zones list --limit 1
  assert_success
  assert_output_contains "Shield Zone Id"
  assert_output_contains "Pull Zone Id"
}

@test "shield zones get returns success for a valid ID" {
  run "$BUNNY_BINARY" shield zones get 1
  assert_success
}

@test "shield zones get --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield zones get 1 --output json
  assert_success
  is_valid_json
}

@test "shield zones get without ID fails" {
  run "$BUNNY_BINARY" shield zones get
  assert_failure
}

@test "shield zones get-by-pullzone returns success" {
  run "$BUNNY_BINARY" shield zones get-by-pullzone 1
  assert_success
}

@test "shield zone alias works" {
  run "$BUNNY_BINARY" shield zone list --limit 1
  assert_success
}

# --- WAF ---

@test "shield waf shows subcommands in help" {
  run "$BUNNY_BINARY" shield waf --help
  assert_success
  assert_output_contains "rules"
  assert_output_contains "custom-rules"
  assert_output_contains "profiles"
  assert_output_contains "engine"
  assert_output_contains "triggered"
}

@test "shield waf custom-rules list returns success" {
  run "$BUNNY_BINARY" shield waf custom-rules list 1 --limit 1
  assert_success
}

@test "shield waf custom-rules list --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield waf custom-rules list 1 --limit 1 --output json
  assert_success
  is_valid_json
}

@test "shield waf custom-rules get returns success" {
  run "$BUNNY_BINARY" shield waf custom-rules get 1
  assert_success
}

@test "shield waf custom-rules delete --yes returns success" {
  run "$BUNNY_BINARY" shield waf custom-rules delete 1 --yes
  assert_success
}

@test "shield waf profiles returns success" {
  run "$BUNNY_BINARY" shield waf profiles
  assert_success
}

@test "shield waf engine returns success" {
  run "$BUNNY_BINARY" shield waf engine
  assert_success
}

@test "shield waf triggered list returns success" {
  run "$BUNNY_BINARY" shield waf triggered list 1
  assert_success
}

# --- Rate Limits ---

@test "shield rate-limits list returns success" {
  run "$BUNNY_BINARY" shield rate-limits list 1 --limit 1
  assert_success
}

@test "shield rate-limits list --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield rate-limits list 1 --limit 1 --output json
  assert_success
  is_valid_json
}

@test "shield rate-limits get returns success" {
  run "$BUNNY_BINARY" shield rate-limits get 1
  assert_success
}

@test "shield rate-limits delete --yes returns success" {
  run "$BUNNY_BINARY" shield rate-limits delete 1 --yes
  assert_success
}

@test "shield ratelimits alias works" {
  run "$BUNNY_BINARY" shield ratelimits list 1 --limit 1
  assert_success
}

# --- Access Lists ---

@test "shield access-lists list returns success" {
  run "$BUNNY_BINARY" shield access-lists list 1
  assert_success
}

@test "shield access-lists get returns success" {
  run "$BUNNY_BINARY" shield access-lists get 1 1
  assert_success
}

@test "shield access-lists delete --yes returns success" {
  run "$BUNNY_BINARY" shield access-lists delete 1 1 --yes
  assert_success
}

# --- Bot Detection ---

@test "shield bot-detection get returns success" {
  run "$BUNNY_BINARY" shield bot-detection get 1
  assert_success
}

@test "shield bot-detection get --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield bot-detection get 1 --output json
  assert_success
  is_valid_json
}

@test "shield bot alias works" {
  run "$BUNNY_BINARY" shield bot get 1
  assert_success
}

# --- Upload Scanning ---

@test "shield upload-scanning get returns success" {
  run "$BUNNY_BINARY" shield upload-scanning get 1
  assert_success
}

@test "shield upload-scanning get --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield upload-scanning get 1 --output json
  assert_success
  is_valid_json
}

@test "shield scanning alias works" {
  run "$BUNNY_BINARY" shield scanning get 1
  assert_success
}

# --- Metrics ---

@test "shield metrics overview returns success" {
  run "$BUNNY_BINARY" shield metrics overview 1
  assert_success
}

@test "shield metrics overview --output json returns valid JSON" {
  run "$BUNNY_BINARY" shield metrics overview 1 --output json
  assert_success
  is_valid_json
}

@test "shield metrics detailed returns success" {
  run "$BUNNY_BINARY" shield metrics detailed 1
  assert_success
}

@test "shield metrics rate-limits returns success" {
  run "$BUNNY_BINARY" shield metrics rate-limits 1
  assert_success
}

@test "shield metrics bot-detection returns success" {
  run "$BUNNY_BINARY" shield metrics bot-detection 1
  assert_success
}

@test "shield metrics upload-scanning returns success" {
  run "$BUNNY_BINARY" shield metrics upload-scanning 1
  assert_success
}

# --- Event Logs ---

@test "shield event-logs without date fails" {
  run "$BUNNY_BINARY" shield event-logs 1
  assert_failure
}

@test "shield event-logs returns success" {
  run "$BUNNY_BINARY" shield event-logs 1 2025-01-15
  assert_success
}

# --- Help ---

@test "shield --help shows all subcommands" {
  run "$BUNNY_BINARY" shield --help
  assert_success
  assert_output_contains "zones"
  assert_output_contains "waf"
  assert_output_contains "rate-limits"
  assert_output_contains "access-lists"
  assert_output_contains "bot-detection"
  assert_output_contains "upload-scanning"
  assert_output_contains "metrics"
  assert_output_contains "event-logs"
}
