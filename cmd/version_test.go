package cmd

import (
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "bunny-cli dev (commit: unknown, built: unknown)\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestVersionCommand_ContainsBunny(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bunny-cli") {
		t.Errorf("expected output to contain 'bunny-cli', got %q", out)
	}
}
