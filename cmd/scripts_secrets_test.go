package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

// --- scripts secrets help ---

func TestScriptsSecrets_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "secrets", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "add", "update", "delete"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected secrets help to show %q subcommand", sub)
		}
	}
}

// --- scripts secrets list ---

func TestScriptsSecretsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptSecretsFn: func(_ context.Context, scriptId int64) ([]*client.EdgeScriptSecret, error) {
			return []*client.EdgeScriptSecret{
				{Id: 1, Name: "DB_PASSWORD", LastModified: "2025-01-15T10:30:00Z"},
				{Id: 2, Name: "API_KEY", LastModified: "2025-01-16T11:00:00Z"},
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "secrets", "list", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DB_PASSWORD") {
		t.Error("expected output to contain secret name")
	}
	if !strings.Contains(out, "API_KEY") {
		t.Error("expected output to contain second secret name")
	}
}

func TestScriptsSecretsList_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "secrets", "list")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestScriptsSecretsList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptSecretsFn: func(_ context.Context, scriptId int64) ([]*client.EdgeScriptSecret, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, stderr, err := executeCommand(app, "scripts", "secrets", "list", "42")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- scripts secrets add ---

func TestScriptsSecretsAdd_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.EdgeScriptSecretCreate
	mock := &mockEdgeScriptAPI{
		addEdgeScriptSecretFn: func(_ context.Context, scriptId int64, body *client.EdgeScriptSecretCreate) (*client.EdgeScriptSecret, error) {
			capturedBody = body
			return &client.EdgeScriptSecret{Id: 5, Name: body.Name, LastModified: "2025-01-15T10:30:00Z"}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "secrets", "add", "42", "--name", "MY_SECRET", "--secret", "s3cr3t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Name != "MY_SECRET" {
		t.Errorf("expected name 'MY_SECRET', got %q", capturedBody.Name)
	}
	if capturedBody.Secret != "s3cr3t" {
		t.Errorf("expected secret 's3cr3t', got %q", capturedBody.Secret)
	}
	if !strings.Contains(out, "MY_SECRET") {
		t.Error("expected output to contain secret name")
	}
}

func TestScriptsSecretsAdd_MissingFlags_Fails(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "scripts", "secrets", "add", "42", "--no-input")
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
	if !strings.Contains(stderr, "name") && !strings.Contains(stderr, "secret") {
		t.Errorf("expected error about missing flags, got %q", stderr)
	}
}

// --- scripts secrets update ---

func TestScriptsSecretsUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.EdgeScriptSecretUpdate
	mock := &mockEdgeScriptAPI{
		updateEdgeScriptSecretFn: func(_ context.Context, scriptId, secretId int64, body *client.EdgeScriptSecretUpdate) error {
			capturedBody = body
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "secrets", "update", "42", "1", "--secret", "new-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Secret != "new-secret" {
		t.Errorf("expected secret 'new-secret', got %q", capturedBody.Secret)
	}
	if !strings.Contains(out, "Secret updated") {
		t.Error("expected confirmation message")
	}
}

// --- scripts secrets delete ---

func TestScriptsSecretsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedScriptId, deletedSecretId int64
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptSecretFn: func(_ context.Context, scriptId, secretId int64) error {
			deletedScriptId = scriptId
			deletedSecretId = secretId
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "secrets", "delete", "42", "1", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedScriptId != 42 || deletedSecretId != 1 {
		t.Errorf("expected deleted ids 42/1, got %d/%d", deletedScriptId, deletedSecretId)
	}
	if !strings.Contains(out, "Secret deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestScriptsSecretsDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptSecretFn: func(_ context.Context, scriptId, secretId int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "scripts", "secrets", "delete", "42", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}
