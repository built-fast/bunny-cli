package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/spf13/viper"
)

// executeCommandMu serializes executeCommand calls to avoid data races on
// the global viper singleton (viper.BindPFlag is not concurrent-safe).
var executeCommandMu sync.Mutex

// executeCommand runs a command with an optional *App injected into context.
// Pass nil to use the default App from NewRootCmd.
func executeCommand(app *App, args ...string) (string, string, error) { //nolint:unparam // app is nil in M0 but used for mock injection in M1+
	return executeCommandWithStdin(app, nil, args...)
}

// executeCommandWithStdin runs a command with optional *App and stdin.
func executeCommandWithStdin(app *App, stdin *bytes.Buffer, args ...string) (string, string, error) {
	executeCommandMu.Lock()
	defer executeCommandMu.Unlock()

	viper.Reset()

	cmd := NewRootCmd()
	if app != nil {
		cmd.SetContext(NewAppContext(cmd.Context(), app))
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	if stdin != nil {
		cmd.SetIn(stdin)
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(stderr, client.FormatError(err))
	}
	return stdout.String(), stderr.String(), err
}

func TestRootNoArgs_ShowsHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Bunny CLI") {
		t.Error("expected help output to mention 'Bunny CLI'")
	}
	if !strings.Contains(out, "Available Commands") {
		t.Error("expected help output to list available commands")
	}
	for _, flag := range []string{"--api-key", "--output"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain global flag %q", flag)
		}
	}
}

func TestRootHelp_ShowsHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--api-key", "--output", "--jq", "--field"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain global flag %q", flag)
		}
	}
}

func TestRootVersion_PrintsVersionString(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "bunny-cli dev (commit: unknown, built: unknown)\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestRootUnknownCommand_ReturnsError(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "nonexistent-command")
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected stderr to contain 'unknown command', got %q", stderr)
	}
}

func TestRootJQFlag_InvalidExpression(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "--jq", "invalid[[[", "configure")
	if err == nil {
		t.Fatal("expected error for invalid jq expression")
	}
	if !strings.Contains(stderr, "invalid jq expression") {
		t.Errorf("expected stderr to contain 'invalid jq expression', got %q", stderr)
	}
}

func TestRootJQFlag_MutuallyExclusiveWithTable(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "--jq", ".name", "--output", "table", "configure")
	if err == nil {
		t.Fatal("expected error for --jq with --output table")
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected stderr to contain 'mutually exclusive', got %q", stderr)
	}
}

func TestRootJQFlag_AllowedWithJSON(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "--jq", ".", "--output", "json", "configure")
	if err != nil && strings.Contains(stderr, "mutually exclusive") {
		t.Error("--jq with --output json should not be mutually exclusive")
	}
}

func TestRootJQFlag_AllowedWithJSONPretty(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "--jq", ".", "--output", "json-pretty", "configure")
	if err != nil && strings.Contains(stderr, "mutually exclusive") {
		t.Error("--jq with --output json-pretty should not be mutually exclusive")
	}
}

func TestRootJQFlag_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--jq") {
		t.Error("expected help output to contain --jq flag")
	}
}

func TestRootNoRegionFlag(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "--region") {
		t.Error("bunny-cli should NOT have a --region flag")
	}
}
