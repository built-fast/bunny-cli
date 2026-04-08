package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

func TestPullZoneEdgeRulesList(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		getPullZoneFn: func(_ context.Context, id int64) (*client.PullZone, error) {
			return samplePullZone(), nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "edge-rules", "list", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "rule-1") {
		t.Error("expected output to contain edge rule GUID")
	}
	if !strings.Contains(out, "Force HTTPS") {
		t.Error("expected output to contain edge rule description")
	}
}

func TestPullZoneEdgeRulesAdd(t *testing.T) {
	t.Parallel()
	var capturedRule *client.EdgeRule
	mock := &mockPullZoneAPI{
		addOrUpdateEdgeRuleFn: func(_ context.Context, pullZoneId int64, rule *client.EdgeRule) error {
			if pullZoneId != 42 {
				t.Errorf("expected pullZoneId=42, got %d", pullZoneId)
			}
			capturedRule = rule
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "edge-rules", "add", "42",
		"--action-type", "1",
		"--action-parameter1", "https://example.com",
		"--description", "Redirect rule")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRule.ActionType != 1 {
		t.Errorf("expected action type 1, got %d", capturedRule.ActionType)
	}
	if capturedRule.Description != "Redirect rule" {
		t.Errorf("expected description 'Redirect rule', got %q", capturedRule.Description)
	}
	if !strings.Contains(out, "Edge rule saved") {
		t.Error("expected success message")
	}
}

func TestPullZoneEdgeRulesAdd_FromFile(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pullzones", "edge-rules", "add", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--from-file") {
		t.Error("expected edge-rules add to have --from-file flag")
	}
}

func TestPullZoneEdgeRulesAdd_FromStdin(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		addOrUpdateEdgeRuleFn: func(_ context.Context, pullZoneId int64, rule *client.EdgeRule) error {
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	stdin := bytes.NewBufferString(`{"action_type":1,"description":"from stdin"}`)
	out, _, err := executeCommandWithStdin(app, stdin, "pullzones", "edge-rules", "add", "42", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Edge rule saved") {
		t.Error("expected success message")
	}
}

func TestPullZoneEdgeRulesDelete_WithYes(t *testing.T) {
	t.Parallel()
	var capturedPzId int64
	var capturedRuleId string
	mock := &mockPullZoneAPI{
		deleteEdgeRuleFn: func(_ context.Context, pullZoneId int64, edgeRuleId string) error {
			capturedPzId = pullZoneId
			capturedRuleId = edgeRuleId
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "edge-rules", "delete", "42", "rule-abc", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPzId != 42 {
		t.Errorf("expected pullZoneId=42, got %d", capturedPzId)
	}
	if capturedRuleId != "rule-abc" {
		t.Errorf("expected ruleId='rule-abc', got %q", capturedRuleId)
	}
	if !strings.Contains(out, "Edge rule deleted") {
		t.Error("expected deletion message")
	}
}

func TestPullZoneEdgeRulesDelete_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		deleteEdgeRuleFn: func(_ context.Context, pullZoneId int64, edgeRuleId string) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "pullzones", "edge-rules", "delete", "42", "rule-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

func TestPullZoneEdgeRulesEnable(t *testing.T) {
	t.Parallel()
	var capturedEnabled bool
	mock := &mockPullZoneAPI{
		setEdgeRuleEnabledFn: func(_ context.Context, pullZoneId int64, edgeRuleId string, enabled bool) error {
			capturedEnabled = enabled
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "edge-rules", "enable", "42", "rule-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !capturedEnabled {
		t.Error("expected enabled=true")
	}
	if !strings.Contains(out, "Edge rule enabled") {
		t.Error("expected enable message")
	}
}

func TestPullZoneEdgeRulesDisable(t *testing.T) {
	t.Parallel()
	var capturedEnabled bool
	capturedEnabled = true // start true to verify it changes
	mock := &mockPullZoneAPI{
		setEdgeRuleEnabledFn: func(_ context.Context, pullZoneId int64, edgeRuleId string, enabled bool) error {
			capturedEnabled = enabled
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "edge-rules", "disable", "42", "rule-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedEnabled {
		t.Error("expected enabled=false")
	}
	if !strings.Contains(out, "Edge rule disabled") {
		t.Error("expected disable message")
	}
}
