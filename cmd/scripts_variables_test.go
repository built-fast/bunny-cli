package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

// --- scripts variables help ---

func TestScriptsVariables_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "variables", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "add", "update", "delete"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected variables help to show %q subcommand", sub)
		}
	}
}

func TestScriptsVariables_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "vars", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage edge script variables") {
		t.Error("expected vars alias to work")
	}
}

// --- scripts variables list ---

func TestScriptsVariablesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptFn: func(_ context.Context, id int64) (*client.EdgeScript, error) {
			return sampleEdgeScript(), nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "variables", "list", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "API_URL") {
		t.Error("expected output to contain variable name")
	}
}

func TestScriptsVariablesList_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "variables", "list")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

// --- scripts variables get ---

func TestScriptsVariablesGet_Success(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptVariableFn: func(_ context.Context, scriptId, variableId int64) (*client.EdgeScriptVariable, error) {
			if scriptId != 42 || variableId != 1 {
				t.Errorf("expected scriptId=42, variableId=1, got %d, %d", scriptId, variableId)
			}
			return &client.EdgeScriptVariable{Id: 1, Name: "API_URL", Required: true, DefaultValue: "https://api.example.com"}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "variables", "get", "42", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "API_URL") {
		t.Error("expected output to contain variable name")
	}
}

func TestScriptsVariablesGet_WithoutArgs_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "variables", "get", "42")
	if err == nil {
		t.Fatal("expected error for missing variable ID argument")
	}
}

// --- scripts variables add ---

func TestScriptsVariablesAdd_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.EdgeScriptVariableCreate
	mock := &mockEdgeScriptAPI{
		addEdgeScriptVariableFn: func(_ context.Context, scriptId int64, body *client.EdgeScriptVariableCreate) (*client.EdgeScriptVariable, error) {
			capturedBody = body
			return &client.EdgeScriptVariable{Id: 5, Name: body.Name, Required: body.Required, DefaultValue: body.DefaultValue}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "variables", "add", "42", "--name", "MY_VAR", "--default-value", "hello", "--required")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Name != "MY_VAR" {
		t.Errorf("expected name 'MY_VAR', got %q", capturedBody.Name)
	}
	if capturedBody.DefaultValue != "hello" {
		t.Errorf("expected default value 'hello', got %q", capturedBody.DefaultValue)
	}
	if !capturedBody.Required {
		t.Error("expected required=true")
	}
	if !strings.Contains(out, "MY_VAR") {
		t.Error("expected output to contain variable name")
	}
}

func TestScriptsVariablesAdd_MissingName_Fails(t *testing.T) {
	t.Parallel()
	_, stderr, err := executeCommand(nil, "scripts", "variables", "add", "42")
	if err == nil {
		t.Fatal("expected error for missing required flag")
	}
	if !strings.Contains(stderr, "name") {
		t.Errorf("expected error about missing name flag, got %q", stderr)
	}
}

// --- scripts variables update ---

func TestScriptsVariablesUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.EdgeScriptVariableUpdate
	mock := &mockEdgeScriptAPI{
		updateEdgeScriptVariableFn: func(_ context.Context, scriptId, variableId int64, body *client.EdgeScriptVariableUpdate) error {
			capturedBody = body
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "variables", "update", "42", "1", "--default-value", "new-value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.DefaultValue == nil || *capturedBody.DefaultValue != "new-value" {
		t.Error("expected default value to be set")
	}
	if !strings.Contains(out, "Variable updated") {
		t.Error("expected confirmation message")
	}
}

// --- scripts variables delete ---

func TestScriptsVariablesDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedScriptId, deletedVarId int64
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptVariableFn: func(_ context.Context, scriptId, variableId int64) error {
			deletedScriptId = scriptId
			deletedVarId = variableId
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "variables", "delete", "42", "1", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedScriptId != 42 || deletedVarId != 1 {
		t.Errorf("expected deleted ids 42/1, got %d/%d", deletedScriptId, deletedVarId)
	}
	if !strings.Contains(out, "Variable deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestScriptsVariablesDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptVariableFn: func(_ context.Context, scriptId, variableId int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "scripts", "variables", "delete", "42", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

// --- scripts variables error ---

func TestScriptsVariablesList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptFn: func(_ context.Context, id int64) (*client.EdgeScript, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, stderr, err := executeCommand(app, "scripts", "variables", "list", "42")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}
