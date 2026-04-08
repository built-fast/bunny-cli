package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

// --- stream captions help ---

func TestStreamCaptions_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "stream", "captions", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"add", "delete"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected captions help to show %q subcommand", sub)
		}
	}
}

// --- stream captions add ---

func TestStreamCaptionsAdd_Success(t *testing.T) {
	t.Parallel()

	// Create a temp caption file
	dir := t.TempDir()
	captionFile := filepath.Join(dir, "subs.vtt")
	if err := os.WriteFile(captionFile, []byte("WEBVTT\n\n00:00:00.000 --> 00:00:05.000\nHello"), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	var capturedSrclang string
	var capturedLabel string
	var capturedHasFile bool
	mock := &mockStreamAPI{
		addCaptionFn: func(_ context.Context, libraryId int64, videoId, srclang string, body *client.CaptionAdd) error {
			capturedSrclang = srclang
			capturedLabel = body.Label
			capturedHasFile = body.CaptionsFile != ""
			return nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "captions", "add", "100", "abc-123-def",
		"--srclang", "en", "--label", "English", "--file", captionFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSrclang != "en" {
		t.Errorf("expected srclang='en', got %q", capturedSrclang)
	}
	if capturedLabel != "English" {
		t.Errorf("expected label='English', got %q", capturedLabel)
	}
	if !capturedHasFile {
		t.Error("expected CaptionsFile to be non-empty (base64 encoded)")
	}
	if !strings.Contains(out, "Caption") && !strings.Contains(out, "added") {
		t.Error("expected confirmation message")
	}
}

func TestStreamCaptionsAdd_RequiresFlags(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "captions", "add", "100", "abc-123-def")
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}

// --- stream captions delete ---

func TestStreamCaptionsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var capturedSrclang string
	mock := &mockStreamAPI{
		deleteCaptionFn: func(_ context.Context, libraryId int64, videoId, srclang string) error {
			capturedSrclang = srclang
			return nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "captions", "delete", "100", "abc-123-def",
		"--srclang", "en", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSrclang != "en" {
		t.Errorf("expected srclang='en', got %q", capturedSrclang)
	}
	if !strings.Contains(out, "deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestStreamCaptionsDelete_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		deleteCaptionFn: func(_ context.Context, libraryId int64, videoId, srclang string) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestStreamApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "stream", "captions", "delete", "100", "abc-123-def",
		"--srclang", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}
