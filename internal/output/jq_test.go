package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/itchyny/gojq"
)

func compileJQ(t *testing.T, expr string) *gojq.Code {
	t.Helper()
	query, err := gojq.Parse(expr)
	if err != nil {
		t.Fatalf("failed to parse jq expression %q: %v", expr, err)
	}
	code, err := gojq.Compile(query)
	if err != nil {
		t.Fatalf("failed to compile jq expression %q: %v", expr, err)
	}
	return code
}

func TestApplyJQ_FieldAccess(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".email")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", out)
	}
}

func TestApplyJQ_ArrayIteration(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".[].name")

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[0] != "Alice" || lines[1] != "Bob" {
		t.Errorf("expected Alice\\nBob, got %q", out)
	}
}

func TestApplyJQ_SelectFilter(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, `[.[] | select(.name == "Bob")] | length`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "1" {
		t.Errorf("expected 1, got %q", out)
	}
}

func TestApplyJQ_Length(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, "length")

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
	}
	out, err := applyJQ(code, items, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "2" {
		t.Errorf("expected 2, got %q", out)
	}
}

func TestApplyJQ_NullResult(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".nonexistent")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "null" {
		t.Errorf("expected null, got %q", out)
	}
}

func TestApplyJQ_BoolResult(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".has_more")

	input := map[string]any{"has_more": true}
	out, err := applyJQ(code, input, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("expected true, got %q", out)
	}
}

func TestApplyJQ_ObjectResult_Compact(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, `{name: .name}`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "\n") {
		t.Errorf("expected compact JSON (no newlines), got %q", out)
	}
	var decoded map[string]string
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %q", decoded["name"])
	}
}

func TestApplyJQ_ObjectResult_Pretty(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, `{name: .name}`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json-pretty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "\n") {
		t.Errorf("expected indented JSON (multi-line), got %q", out)
	}
	if !strings.Contains(out, "  ") {
		t.Errorf("expected 2-space indentation, got %q", out)
	}
}

func TestApplyJQ_EmptyResult(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, `select(.name == "Nobody")`)

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := applyJQ(code, item, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestApplyJQ_RuntimeError(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".name / 0")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	_, err := applyJQ(code, item, "json")
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(err.Error(), "jq:") {
		t.Errorf("expected error prefixed with jq:, got %q", err.Error())
	}
}

func TestFormatList_WithJQ_DataLength(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, `.data | length`)

	items := []any{
		testItem{Name: "Alice", Email: "alice@example.com"},
		testItem{Name: "Bob", Email: "bob@example.com"},
		testItem{Name: "Carol", Email: "carol@example.com"},
	}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "3" {
		t.Errorf("expected 3, got %q", out)
	}
}

func TestFormatList_WithJQ_EnvelopeFields(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".object")

	items := []any{testItem{Name: "Alice", Email: "alice@example.com"}}
	out, err := FormatList(&Config{Format: "json", JQ: code}, testColumns, items, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "list" {
		t.Errorf("expected 'list', got %q", out)
	}
}

func TestFormatOne_WithJQ(t *testing.T) {
	t.Parallel()
	code := compileJQ(t, ".email")

	item := testItem{Name: "Alice", Email: "alice@example.com"}
	out, err := FormatOne(&Config{Format: "json", JQ: code}, testColumns, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", out)
	}
}
