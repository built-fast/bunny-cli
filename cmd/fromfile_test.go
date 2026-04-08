package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newFromFileTestCmd(runFn func(cmd *cobra.Command, args []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "test",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runFn,
	}
	cmd.Flags().String("name", "", "Name")
	cmd.Flags().String("origin-url", "", "Origin URL")
	cmd.Flags().Bool("enabled", false, "Enabled")
	cmd.Flags().Int("type", 0, "Type")
	return cmd
}

func TestWithFromFile_JSONFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "input.json")
	if err := os.WriteFile(jsonFile, []byte(`{"name":"test-zone","origin_url":"https://example.com"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotName, gotOrigin string
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error {
		gotName, _ = cmd.Flags().GetString("name")
		gotOrigin, _ = cmd.Flags().GetString("origin-url")
		return nil
	})

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", jsonFile})
	if err := wrapped.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != "test-zone" {
		t.Errorf("expected name 'test-zone', got %q", gotName)
	}
	if gotOrigin != "https://example.com" {
		t.Errorf("expected origin-url 'https://example.com', got %q", gotOrigin)
	}
}

func TestWithFromFile_YAMLFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	yamlFile := filepath.Join(dir, "input.yaml")
	if err := os.WriteFile(yamlFile, []byte("name: yaml-zone\norigin_url: https://yaml.example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}

	var gotName, gotOrigin string
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error {
		gotName, _ = cmd.Flags().GetString("name")
		gotOrigin, _ = cmd.Flags().GetString("origin-url")
		return nil
	})

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"-F", yamlFile})
	if err := wrapped.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != "yaml-zone" {
		t.Errorf("expected name 'yaml-zone', got %q", gotName)
	}
	if gotOrigin != "https://yaml.example.com" {
		t.Errorf("expected origin-url 'https://yaml.example.com', got %q", gotOrigin)
	}
}

func TestWithFromFile_CLIFlagsTakePrecedence(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "input.json")
	if err := os.WriteFile(jsonFile, []byte(`{"name":"file-value","origin_url":"https://file.example.com"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotName, gotOrigin string
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error {
		gotName, _ = cmd.Flags().GetString("name")
		gotOrigin, _ = cmd.Flags().GetString("origin-url")
		return nil
	})

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", jsonFile, "--name", "cli-value"})
	if err := wrapped.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != "cli-value" {
		t.Errorf("expected CLI value 'cli-value', got %q", gotName)
	}
	if gotOrigin != "https://file.example.com" {
		t.Errorf("expected file value for origin-url, got %q", gotOrigin)
	}
}

func TestWithFromFile_Stdin(t *testing.T) {
	t.Parallel()
	var gotName string
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error {
		gotName, _ = cmd.Flags().GetString("name")
		return nil
	})

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", "-"})
	wrapped.SetIn(bytes.NewBufferString(`{"name":"stdin-zone"}`))
	if err := wrapped.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != "stdin-zone" {
		t.Errorf("expected name 'stdin-zone', got %q", gotName)
	}
}

func TestWithFromFile_UnknownKey_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "input.json")
	if err := os.WriteFile(jsonFile, []byte(`{"name":"ok","unknown_key":"bad"}`), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error { return nil })

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", jsonFile})
	err := wrapped.Execute()
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("expected 'unknown key' error, got: %s", err.Error())
	}
}

func TestWithFromFile_UnsupportedExtension_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	txtFile := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(txtFile, []byte(`name: test`), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error { return nil })

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", txtFile})
	err := wrapped.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported extension")
	}
	if !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("expected 'unsupported file extension' error, got: %s", err.Error())
	}
}

func TestWithFromFile_EmptyStdin_ReturnsError(t *testing.T) {
	t.Parallel()
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error { return nil })

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", "-"})
	wrapped.SetIn(bytes.NewBufferString(""))
	err := wrapped.Execute()
	if err == nil {
		t.Fatal("expected error for empty stdin")
	}
	if !strings.Contains(err.Error(), "stdin is empty") {
		t.Errorf("expected 'stdin is empty' error, got: %s", err.Error())
	}
}

func TestWithFromFile_NoFile_RunsNormally(t *testing.T) {
	t.Parallel()
	var ranCmd bool
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error {
		ranCmd = true
		return nil
	})

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--name", "direct"})
	if err := wrapped.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ranCmd {
		t.Fatal("expected command to run")
	}
}

func TestWithFromFile_BoolAndIntValues(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "input.json")
	if err := os.WriteFile(jsonFile, []byte(`{"name":"test","enabled":true,"type":1}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotEnabled bool
	var gotType int
	cmd := newFromFileTestCmd(func(cmd *cobra.Command, args []string) error {
		gotEnabled, _ = cmd.Flags().GetBool("enabled")
		gotType, _ = cmd.Flags().GetInt("type")
		return nil
	})

	wrapped := withFromFile(cmd)
	wrapped.SetArgs([]string{"--from-file", jsonFile})
	if err := wrapped.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotEnabled {
		t.Error("expected enabled=true")
	}
	if gotType != 1 {
		t.Errorf("expected type=1, got %d", gotType)
	}
}

func TestKeyToFlagName(t *testing.T) {
	t.Parallel()
	tests := []struct{ input, expected string }{
		{"name", "name"},
		{"origin_url", "origin-url"},
		{"first_name", "first-name"},
		{"already-hyphen", "already-hyphen"},
	}
	for _, tt := range tests {
		got := keyToFlagName(tt.input)
		if got != tt.expected {
			t.Errorf("keyToFlagName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFlattenMap(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"name": "test",
		"address": map[string]any{
			"street1": "123 Main St",
			"city":    "Anytown",
		},
	}
	result := flattenMap(input, "")
	if result["name"] != "test" {
		t.Errorf("expected name='test', got %v", result["name"])
	}
	if result["address-street1"] != "123 Main St" {
		t.Errorf("expected address-street1='123 Main St', got %v", result["address-street1"])
	}
	if result["address-city"] != "Anytown" {
		t.Errorf("expected address-city='Anytown', got %v", result["address-city"])
	}
}

func TestValueToString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "hello"},
		{true, "true"},
		{false, "false"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{nil, ""},
		{[]any{"a", "b", "c"}, "a,b,c"},
	}
	for _, tt := range tests {
		got, err := valueToString(tt.input)
		if err != nil {
			t.Errorf("valueToString(%v) unexpected error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("valueToString(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
