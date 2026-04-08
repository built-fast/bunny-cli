package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/viper"
)

// --- pullzones help ---

func TestPullZones_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pullzones", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "hostnames", "purge", "edge-rules"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected pullzones help to show %q subcommand", sub)
		}
	}
}

func TestPullZones_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pz", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage pull zones") {
		t.Error("expected pz alias to work")
	}
}

// --- pullzones list ---

func TestPullZonesList_ShowsFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pullzones", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--search"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestPullZonesList_NoAPIKey_ReturnsError(t *testing.T) {
	viper.Reset()
	t.Setenv("BUNNY_API_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, stderr, err := executeCommand(nil, "pullzones", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestPullZonesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		listPullZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error) {
			return pagination.PageResponse[*client.PullZone]{
				Items:        []*client.PullZone{samplePullZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-zone") {
		t.Error("expected output to contain zone name")
	}
	if !strings.Contains(out, "Premium") {
		t.Error("expected output to contain type name")
	}
}

func TestPullZonesList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		listPullZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error) {
			return pagination.PageResponse[*client.PullZone]{
				Items:        []*client.PullZone{samplePullZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "list", "--output", "json")
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

func TestPullZonesList_SearchParam(t *testing.T) {
	t.Parallel()
	var capturedSearch string
	mock := &mockPullZoneAPI{
		listPullZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error) {
			capturedSearch = search
			return pagination.PageResponse[*client.PullZone]{}, nil
		},
	}
	app := newTestPullZoneApp(mock)

	_, _, err := executeCommand(app, "pullzones", "list", "--search", "test-zone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch != "test-zone" {
		t.Errorf("expected search='test-zone', got %q", capturedSearch)
	}
}

func TestPullZonesList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		listPullZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error) {
			return pagination.PageResponse[*client.PullZone]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestPullZoneApp(mock)

	_, stderr, err := executeCommand(app, "pullzones", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- pullzones get ---

func TestPullZonesGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		getPullZoneFn: func(_ context.Context, id int64) (*client.PullZone, error) {
			if id != 42 {
				t.Errorf("expected id=42, got %d", id)
			}
			return samplePullZone(), nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-zone") {
		t.Error("expected output to contain zone name")
	}
	if !strings.Contains(out, "origin.example.com") {
		t.Error("expected output to contain origin URL")
	}
}

func TestPullZonesGet_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		getPullZoneFn: func(_ context.Context, id int64) (*client.PullZone, error) {
			return samplePullZone(), nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "get", "42", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["Name"] != "my-zone" {
		t.Errorf("expected Name=my-zone, got %v", result["Name"])
	}
}

func TestPullZonesGet_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "pullzones", "get")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestPullZonesGet_InvalidID_Fails(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		getPullZoneFn: func(_ context.Context, id int64) (*client.PullZone, error) {
			return samplePullZone(), nil
		},
	}
	app := newTestPullZoneApp(mock)

	_, stderr, err := executeCommand(app, "pullzones", "get", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "invalid pull zone ID") {
		t.Errorf("expected 'invalid pull zone ID' error, got %q", stderr)
	}
}

func TestPullZonesGet_WatchFlag(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pullzones", "get", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--watch") {
		t.Error("expected get command to have --watch flag")
	}
}

// --- pullzones create ---

func TestPullZonesCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.PullZoneCreate
	mock := &mockPullZoneAPI{
		createPullZoneFn: func(_ context.Context, body *client.PullZoneCreate) (*client.PullZone, error) {
			capturedBody = body
			return &client.PullZone{Id: 99, Name: body.Name, OriginUrl: body.OriginUrl}, nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "create", "--name", "new-zone", "--origin-url", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Name != "new-zone" {
		t.Errorf("expected name 'new-zone', got %q", capturedBody.Name)
	}
	if capturedBody.OriginUrl != "https://example.com" {
		t.Errorf("expected origin URL, got %q", capturedBody.OriginUrl)
	}
	if !strings.Contains(out, "new-zone") {
		t.Error("expected output to contain created zone name")
	}
}

func TestPullZonesCreate_FromFile(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "pullzones", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--from-file") {
		t.Error("expected create command to have --from-file flag")
	}
}

func TestPullZonesCreate_FromFileStdin(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		createPullZoneFn: func(_ context.Context, body *client.PullZoneCreate) (*client.PullZone, error) {
			return &client.PullZone{Id: 99, Name: body.Name}, nil
		},
	}
	app := newTestPullZoneApp(mock)

	stdin := bytes.NewBufferString(`{"name":"stdin-zone"}`)
	out, _, err := executeCommandWithStdin(app, stdin, "pullzones", "create", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "stdin-zone") {
		t.Error("expected output to contain zone from stdin")
	}
}

// --- pullzones update ---

func TestPullZonesUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedBody *client.PullZoneUpdate
	mock := &mockPullZoneAPI{
		updatePullZoneFn: func(_ context.Context, id int64, body *client.PullZoneUpdate) (*client.PullZone, error) {
			capturedId = id
			capturedBody = body
			return &client.PullZone{Id: id, Name: "updated-zone", OriginUrl: *body.OriginUrl}, nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "update", "42", "--origin-url", "https://new.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 42 {
		t.Errorf("expected id=42, got %d", capturedId)
	}
	if capturedBody.OriginUrl == nil || *capturedBody.OriginUrl != "https://new.example.com" {
		t.Error("expected origin URL to be set in body")
	}
	if !strings.Contains(out, "updated-zone") {
		t.Error("expected output to show updated zone")
	}
}

func TestPullZonesUpdate_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "pullzones", "update")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

// --- pullzones delete ---

func TestPullZonesDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	mock := &mockPullZoneAPI{
		deletePullZoneFn: func(_ context.Context, id int64) error {
			deletedId = id
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	out, _, err := executeCommand(app, "pullzones", "delete", "42", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 42 {
		t.Errorf("expected deleted id=42, got %d", deletedId)
	}
	if !strings.Contains(out, "Pull zone deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestPullZonesDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockPullZoneAPI{
		deletePullZoneFn: func(_ context.Context, id int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestPullZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "pullzones", "delete", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

func TestPullZonesDelete_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "pullzones", "delete", "--yes")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}
