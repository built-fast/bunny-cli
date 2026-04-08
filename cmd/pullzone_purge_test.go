package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPullZonePurge_WithYes(t *testing.T) {
	t.Parallel()
	var purgedId int64
	var purgedTag string
	mock := &mockPullZoneAPI{
		purgePullZoneCacheFn: func(_ context.Context, id int64, cacheTag string) error {
			purgedId = id
			purgedTag = cacheTag
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "purge", "42", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if purgedId != 42 {
		t.Errorf("expected id=42, got %d", purgedId)
	}
	if purgedTag != "" {
		t.Errorf("expected empty tag, got %q", purgedTag)
	}
	if !strings.Contains(out, "Cache purged") {
		t.Error("expected purge confirmation message")
	}
}

func TestPullZonePurge_WithTag(t *testing.T) {
	t.Parallel()
	var purgedTag string
	mock := &mockPullZoneAPI{
		purgePullZoneCacheFn: func(_ context.Context, id int64, cacheTag string) error {
			purgedTag = cacheTag
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "purge", "42", "--tag", "my-tag", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if purgedTag != "my-tag" {
		t.Errorf("expected tag 'my-tag', got %q", purgedTag)
	}
	if !strings.Contains(out, "Cache purged") {
		t.Error("expected purge confirmation message")
	}
}

func TestPullZonePurge_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		purgePullZoneCacheFn: func(_ context.Context, id int64, cacheTag string) error {
			t.Error("purge should not have been called")
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "pullzones", "purge", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Purge canceled") {
		t.Error("expected cancellation message")
	}
}

func TestPullZonePurge_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "pullzones", "purge", "--yes")
	if err == nil {
		t.Fatal("expected error for missing pull zone ID")
	}
}

func TestPullZonePurge_ShowsFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pullzones", "purge", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--tag", "--yes"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}
