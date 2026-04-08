package cmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigure_NewConfig(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("")),
		writer: out,
		readPassword: func() (string, error) {
			return "test-api-key-12345", nil
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "API Key:") {
		t.Errorf("expected API Key prompt, got: %s", output)
	}
	if !strings.Contains(output, "Configuration saved to") {
		t.Errorf("expected confirmation message, got: %s", output)
	}

	// Verify config was written
	v := viper.New()
	v.SetConfigFile(filepath.Join(tmp, "bunny", "config.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if got := v.GetString("api_key"); got != "test-api-key-12345" {
		t.Errorf("api_key = %q, want %q", got, "test-api-key-12345")
	}
}

func TestConfigure_ExistingConfigShowsMaskedKey(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Write existing config
	dir := filepath.Join(tmp, "bunny")
	if err := writeTestConfig(dir, "api_key = \"existing-key-abcd\"\n"); err != nil {
		t.Fatal(err)
	}

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("")),
		writer: out,
		readPassword: func() (string, error) {
			return "", nil // Enter to keep existing API key
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "API Key [****abcd]:") {
		t.Errorf("expected masked API key in prompt, got: %s", output)
	}

	// Verify existing value preserved
	v := viper.New()
	v.SetConfigFile(filepath.Join(tmp, "bunny", "config.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := v.GetString("api_key"); got != "existing-key-abcd" {
		t.Errorf("api_key = %q, want %q", got, "existing-key-abcd")
	}
}

func TestConfigure_EmptyAPIKeyRequired(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("")),
		writer: out,
		readPassword: func() (string, error) {
			return "", nil
		},
	}

	err := runConfigure(p)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	if !strings.Contains(err.Error(), "API key is required") {
		t.Errorf("expected 'API key is required' error, got: %v", err)
	}
}

func TestConfigure_ConfirmationMessage(t *testing.T) {
	viper.Reset()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	out := new(bytes.Buffer)
	p := &configPrompter{
		reader: bufio.NewReader(strings.NewReader("")),
		writer: out,
		readPassword: func() (string, error) {
			return "my-key", nil
		},
	}

	if err := runConfigure(p); err != nil {
		t.Fatalf("runConfigure() error: %v", err)
	}

	expected := "Configuration saved to " + filepath.Join(tmp, "bunny", "config.toml")
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected confirmation %q, got: %s", expected, out.String())
	}
}

func TestMaskKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "****bcde"},
		{"test-api-key-12345", "****2345"},
	}
	for _, tt := range tests {
		got := maskKey(tt.input)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func writeTestConfig(dir, content string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0600)
}
