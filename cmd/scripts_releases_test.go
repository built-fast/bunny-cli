package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// --- scripts releases help ---

func TestScriptsReleases_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "releases", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "active"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected releases help to show %q subcommand", sub)
		}
	}
}

// --- scripts releases list ---

func TestScriptsReleasesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptReleasesFn: func(_ context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error) {
			return pagination.PageResponse[*client.EdgeScriptRelease]{
				Items: []*client.EdgeScriptRelease{
					{Id: 1, Uuid: "abc-123", Status: 1, Author: "dev", Note: "initial release", DatePublished: "2025-01-15T10:30:00Z"},
					{Id: 2, Uuid: "def-456", Status: 0, Author: "dev", Note: "fix", DatePublished: "2025-01-10T10:30:00Z"},
				},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "releases", "list", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "abc-123") {
		t.Error("expected output to contain release UUID")
	}
	if !strings.Contains(out, "Live") {
		t.Error("expected output to contain release status")
	}
}

func TestScriptsReleasesList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptReleasesFn: func(_ context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error) {
			return pagination.PageResponse[*client.EdgeScriptRelease]{
				Items: []*client.EdgeScriptRelease{
					{Id: 1, Uuid: "abc-123", Status: 1},
				},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "releases", "list", "42", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["object"] != "list" {
		t.Errorf("expected object=list, got %v", result["object"])
	}
}

func TestScriptsReleasesList_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "releases", "list")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestScriptsReleasesList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptReleasesFn: func(_ context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error) {
			return pagination.PageResponse[*client.EdgeScriptRelease]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, stderr, err := executeCommand(app, "scripts", "releases", "list", "42")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- scripts releases active ---

func TestScriptsReleasesActive_Success(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getActiveEdgeScriptReleaseFn: func(_ context.Context, scriptId int64) (*client.EdgeScriptRelease, error) {
			if scriptId != 42 {
				t.Errorf("expected scriptId=42, got %d", scriptId)
			}
			return &client.EdgeScriptRelease{
				Id:            1,
				Uuid:          "abc-123",
				Status:        1,
				Author:        "dev",
				Note:          "current release",
				DatePublished: "2025-01-15T10:30:00Z",
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "releases", "active", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "abc-123") {
		t.Error("expected output to contain release UUID")
	}
	if !strings.Contains(out, "Live") {
		t.Error("expected output to contain release status")
	}
}

func TestScriptsReleasesActive_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "releases", "active")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

// --- scripts publish ---

func TestScriptsPublish_Success(t *testing.T) {
	t.Parallel()
	var capturedNote string
	mock := &mockEdgeScriptAPI{
		publishEdgeScriptFn: func(_ context.Context, scriptId int64, body *client.EdgeScriptPublish) error {
			capturedNote = body.Note
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "publish", "42", "--note", "v1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedNote != "v1.0" {
		t.Errorf("expected note 'v1.0', got %q", capturedNote)
	}
	if !strings.Contains(out, "Edge script published") {
		t.Error("expected publish confirmation message")
	}
}

func TestScriptsPublish_ByUUID(t *testing.T) {
	t.Parallel()
	var capturedUUID string
	mock := &mockEdgeScriptAPI{
		publishEdgeScriptReleaseFn: func(_ context.Context, scriptId int64, uuid string) error {
			capturedUUID = uuid
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "publish", "42", "abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUUID != "abc-123" {
		t.Errorf("expected UUID 'abc-123', got %q", capturedUUID)
	}
	if !strings.Contains(out, "Release abc-123 published") {
		t.Error("expected publish confirmation message with UUID")
	}
}

func TestScriptsPublish_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "publish")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestScriptsPublish_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		publishEdgeScriptFn: func(_ context.Context, scriptId int64, body *client.EdgeScriptPublish) error {
			return fmt.Errorf("publish failed")
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, stderr, err := executeCommand(app, "scripts", "publish", "42")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "publish failed") {
		t.Errorf("expected error in stderr, got %q", stderr)
	}
}
