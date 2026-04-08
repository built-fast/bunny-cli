package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestParseWatchInterval_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"5s", 5 * time.Second},
		{"1m", time.Minute},
		{"30s", 30 * time.Second},
		{"2m30s", 2*time.Minute + 30*time.Second},
		{"1s", time.Second},
	}
	for _, tt := range tests {
		d, err := parseWatchInterval(tt.input)
		if err != nil {
			t.Errorf("parseWatchInterval(%q) unexpected error: %v", tt.input, err)
		}
		if d != tt.expected {
			t.Errorf("parseWatchInterval(%q) = %v, want %v", tt.input, d, tt.expected)
		}
	}
}

func TestParseWatchInterval_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantMsg string
	}{
		{"abc", "invalid watch interval"},
		{"500ms", "at least 1s"},
		{"0s", "at least 1s"},
		{"", "invalid watch interval"},
	}
	for _, tt := range tests {
		_, err := parseWatchInterval(tt.input)
		if err == nil {
			t.Errorf("parseWatchInterval(%q) expected error, got nil", tt.input)
			continue
		}
		if !strings.Contains(err.Error(), tt.wantMsg) {
			t.Errorf("parseWatchInterval(%q) error = %q, want to contain %q", tt.input, err.Error(), tt.wantMsg)
		}
	}
}

func TestWithWatch_DefaultInterval(t *testing.T) {
	t.Parallel()
	cmd := &cobra.Command{
		Use:  "test",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	cmd = withWatch(cmd)

	f := cmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("expected --watch flag to exist")
	}
	if f.NoOptDefVal != "5s" {
		t.Errorf("expected NoOptDefVal=5s, got %q", f.NoOptDefVal)
	}
}

func TestWatchLoop_ContextCancel(t *testing.T) {
	t.Parallel()
	var callCount int
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	runE := func(cmd *cobra.Command, args []string) error {
		callCount++
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "output")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	watchLoop(ctx, cmd, nil, runE, 50*time.Millisecond, false)

	if callCount < 2 {
		t.Errorf("expected at least 2 calls, got %d", callCount)
	}
	if !strings.Contains(stderr.String(), "Watch stopped") {
		t.Error("expected 'Watch stopped' message on cancel")
	}
	if !strings.Contains(stderr.String(), "Last updated") {
		t.Error("expected 'Last updated' timestamp in output")
	}
}

func TestWatchLoop_PipedMode_NoScreenClear(t *testing.T) {
	t.Parallel()
	var callCount int
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	runE := func(cmd *cobra.Command, args []string) error {
		callCount++
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "json-line")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	watchLoop(ctx, cmd, nil, runE, 50*time.Millisecond, true)

	if callCount < 2 {
		t.Errorf("expected at least 2 calls, got %d", callCount)
	}
	if strings.Contains(stdout.String(), "\033[2J") {
		t.Error("should not contain screen-clear escape sequence in piped mode")
	}
	if strings.Contains(stderr.String(), "Last updated") {
		t.Error("should not show timestamp in piped mode")
	}
	if strings.Contains(stderr.String(), "Watch stopped") {
		t.Error("should not show 'Watch stopped' in piped mode")
	}
	lines := strings.Count(stdout.String(), "json-line")
	if lines < 2 {
		t.Errorf("expected multiple json-line outputs, got %d", lines)
	}
}

func TestWatchLoop_ErrorContinues(t *testing.T) {
	t.Parallel()
	var callCount int
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	runE := func(cmd *cobra.Command, args []string) error {
		callCount++
		if callCount == 1 {
			return fmt.Errorf("temporary API error")
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "recovered")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	watchLoop(ctx, cmd, nil, runE, 50*time.Millisecond, false)

	if callCount < 2 {
		t.Errorf("expected at least 2 calls (error should not stop loop), got %d", callCount)
	}
	if !strings.Contains(stderr.String(), "temporary API error") {
		t.Error("expected error message in stderr")
	}
	if !strings.Contains(stdout.String(), "recovered") {
		t.Error("expected recovery output after error")
	}
}

func TestWatchLoop_ScreenClear(t *testing.T) {
	t.Parallel()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	runE := func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "data")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	watchLoop(ctx, cmd, nil, runE, 50*time.Millisecond, false)

	if !strings.Contains(stdout.String(), "\033[2J\033[H") {
		t.Error("expected screen-clear escape sequence in terminal mode")
	}
}

func TestIsJSONFormat(t *testing.T) {
	t.Parallel()
	if !isJSONFormat("json") {
		t.Error("expected json to be JSON format")
	}
	if !isJSONFormat("json-pretty") {
		t.Error("expected json-pretty to be JSON format")
	}
	if isJSONFormat("table") {
		t.Error("expected table to not be JSON format")
	}
}
