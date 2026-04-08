package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

func TestPullZoneHostnamesList(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		getPullZoneFn: func(_ context.Context, id int64) (*client.PullZone, error) {
			return samplePullZone(), nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "hostnames", "list", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cdn.example.com") {
		t.Error("expected output to contain hostname")
	}
}

func TestPullZoneHostnamesList_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "pullzones", "hostnames", "list")
	if err == nil {
		t.Fatal("expected error for missing pull zone ID")
	}
}

func TestPullZoneHostnamesAdd(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedHostname string
	mock := &mockPullZoneAPI{
		addPullZoneHostnameFn: func(_ context.Context, id int64, hostname string) error {
			capturedId = id
			capturedHostname = hostname
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "hostnames", "add", "42", "--hostname", "cdn.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 42 {
		t.Errorf("expected id=42, got %d", capturedId)
	}
	if capturedHostname != "cdn.example.com" {
		t.Errorf("expected hostname 'cdn.example.com', got %q", capturedHostname)
	}
	if !strings.Contains(out, "added") {
		t.Error("expected success message")
	}
}

func TestPullZoneHostnamesRemove_WithYes(t *testing.T) {
	t.Parallel()
	var capturedHostname string
	mock := &mockPullZoneAPI{
		removePullZoneHostnameFn: func(_ context.Context, id int64, hostname string) error {
			capturedHostname = hostname
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "hostnames", "remove", "42", "--hostname", "cdn.example.com", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedHostname != "cdn.example.com" {
		t.Errorf("expected hostname, got %q", capturedHostname)
	}
	if !strings.Contains(out, "removed") {
		t.Error("expected success message")
	}
}

func TestPullZoneHostnamesRemove_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		removePullZoneHostnameFn: func(_ context.Context, id int64, hostname string) error {
			t.Error("remove should not have been called")
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "pullzones", "hostnames", "remove", "42", "--hostname", "cdn.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Removal canceled") {
		t.Error("expected cancellation message")
	}
}
