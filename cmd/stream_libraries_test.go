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
	"github.com/spf13/cobra"
)

// mockVideoLibraryAPI implements VideoLibraryAPI for testing.
type mockVideoLibraryAPI struct {
	listVideoLibrariesFn        func(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error)
	getVideoLibraryFn           func(ctx context.Context, id int64) (*client.VideoLibrary, error)
	createVideoLibraryFn        func(ctx context.Context, body *client.VideoLibraryCreate) (*client.VideoLibrary, error)
	updateVideoLibraryFn        func(ctx context.Context, id int64, body *client.VideoLibraryUpdate) (*client.VideoLibrary, error)
	deleteVideoLibraryFn        func(ctx context.Context, id int64) error
	resetVideoLibraryApiKeyFn   func(ctx context.Context, id int64) error
	listVideoLibraryLanguagesFn func(ctx context.Context) ([]client.VideoLibraryLanguage, error)
}

func (m *mockVideoLibraryAPI) ListVideoLibraries(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error) {
	return m.listVideoLibrariesFn(ctx, page, perPage, search)
}

func (m *mockVideoLibraryAPI) GetVideoLibrary(ctx context.Context, id int64) (*client.VideoLibrary, error) {
	return m.getVideoLibraryFn(ctx, id)
}

func (m *mockVideoLibraryAPI) CreateVideoLibrary(ctx context.Context, body *client.VideoLibraryCreate) (*client.VideoLibrary, error) {
	return m.createVideoLibraryFn(ctx, body)
}

func (m *mockVideoLibraryAPI) UpdateVideoLibrary(ctx context.Context, id int64, body *client.VideoLibraryUpdate) (*client.VideoLibrary, error) {
	return m.updateVideoLibraryFn(ctx, id, body)
}

func (m *mockVideoLibraryAPI) DeleteVideoLibrary(ctx context.Context, id int64) error {
	return m.deleteVideoLibraryFn(ctx, id)
}

func (m *mockVideoLibraryAPI) ResetVideoLibraryApiKey(ctx context.Context, id int64) error {
	return m.resetVideoLibraryApiKeyFn(ctx, id)
}

func (m *mockVideoLibraryAPI) ListVideoLibraryLanguages(ctx context.Context) ([]client.VideoLibraryLanguage, error) {
	return m.listVideoLibraryLanguagesFn(ctx)
}

func newTestVideoLibraryApp(api VideoLibraryAPI) *App {
	return &App{NewVideoLibraryAPI: func(_ *cobra.Command) (VideoLibraryAPI, error) { return api, nil }}
}

func sampleVideoLibrary() *client.VideoLibrary {
	return &client.VideoLibrary{ //nolint:gosec // G101: test data, not real credentials
		Id:                 100,
		Name:               "my-video-lib",
		VideoCount:         42,
		TrafficUsage:       1024000,
		StorageUsage:       2048000,
		DateCreated:        "2025-01-15T10:30:00Z",
		DateModified:       "2025-06-01T12:00:00Z",
		ApiKey:             "lib-api-key-123",
		ReadOnlyApiKey:     "lib-ro-key-456",
		PullZoneId:         10,
		StorageZoneId:      20,
		EnabledResolutions: "240p,360p,480p,720p,1080p",
		EnableMP4Fallback:  true,
		KeepOriginalFiles:  true,
		EnableDRM:          false,
		AllowDirectPlay:    true,
		ReplicationRegions: []string{"NY", "LA"},
	}
}

// --- stream libraries help ---

func TestStreamLibraries_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "stream", "libraries", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "reset-api-key", "languages"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected libraries help to show %q subcommand", sub)
		}
	}
}

func TestStreamLibraries_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "stream", "lib", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage video libraries") {
		t.Error("expected lib alias to work")
	}
}

// --- stream libraries list ---

func TestStreamLibrariesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		listVideoLibrariesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error) {
			return pagination.PageResponse[*client.VideoLibrary]{
				Items:        []*client.VideoLibrary{sampleVideoLibrary()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-video-lib") {
		t.Error("expected output to contain library name")
	}
}

func TestStreamLibrariesList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		listVideoLibrariesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error) {
			return pagination.PageResponse[*client.VideoLibrary]{
				Items:        []*client.VideoLibrary{sampleVideoLibrary()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "list", "--output", "json")
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

func TestStreamLibrariesList_SearchParam(t *testing.T) {
	t.Parallel()
	var capturedSearch string
	mock := &mockVideoLibraryAPI{
		listVideoLibrariesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error) {
			capturedSearch = search
			return pagination.PageResponse[*client.VideoLibrary]{}, nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	_, _, err := executeCommand(app, "stream", "libraries", "list", "--search", "test-lib")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch != "test-lib" {
		t.Errorf("expected search='test-lib', got %q", capturedSearch)
	}
}

func TestStreamLibrariesList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		listVideoLibrariesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error) {
			return pagination.PageResponse[*client.VideoLibrary]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestVideoLibraryApp(mock)

	_, stderr, err := executeCommand(app, "stream", "libraries", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- stream libraries get ---

func TestStreamLibrariesGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		getVideoLibraryFn: func(_ context.Context, id int64) (*client.VideoLibrary, error) {
			if id != 100 {
				t.Errorf("expected id=100, got %d", id)
			}
			return sampleVideoLibrary(), nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "get", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-video-lib") {
		t.Error("expected output to contain library name")
	}
	if !strings.Contains(out, "lib-api-key-123") {
		t.Error("expected output to contain API key")
	}
}

func TestStreamLibrariesGet_InvalidID(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		getVideoLibraryFn: func(_ context.Context, id int64) (*client.VideoLibrary, error) {
			return sampleVideoLibrary(), nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	_, stderr, err := executeCommand(app, "stream", "libraries", "get", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "invalid video library ID") {
		t.Errorf("expected 'invalid video library ID' error, got %q", stderr)
	}
}

// --- stream libraries create ---

func TestStreamLibrariesCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.VideoLibraryCreate
	mock := &mockVideoLibraryAPI{
		createVideoLibraryFn: func(_ context.Context, body *client.VideoLibraryCreate) (*client.VideoLibrary, error) {
			capturedBody = body
			return &client.VideoLibrary{Id: 200, Name: body.Name}, nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "create", "--name", "new-lib")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Name != "new-lib" {
		t.Errorf("expected name 'new-lib', got %q", capturedBody.Name)
	}
	if !strings.Contains(out, "new-lib") {
		t.Error("expected output to contain created library name")
	}
}

// --- stream libraries update ---

func TestStreamLibrariesUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedBody *client.VideoLibraryUpdate
	mock := &mockVideoLibraryAPI{
		updateVideoLibraryFn: func(_ context.Context, id int64, body *client.VideoLibraryUpdate) (*client.VideoLibrary, error) {
			capturedId = id
			capturedBody = body
			lib := sampleVideoLibrary()
			if body.Name != nil {
				lib.Name = *body.Name
			}
			return lib, nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "update", "100", "--name", "renamed-lib")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 100 {
		t.Errorf("expected id=100, got %d", capturedId)
	}
	if capturedBody.Name == nil || *capturedBody.Name != "renamed-lib" {
		t.Error("expected name to be set in body")
	}
	if !strings.Contains(out, "renamed-lib") {
		t.Error("expected output to show updated name")
	}
}

// --- stream libraries delete ---

func TestStreamLibrariesDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	mock := &mockVideoLibraryAPI{
		deleteVideoLibraryFn: func(_ context.Context, id int64) error {
			deletedId = id
			return nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "delete", "100", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 100 {
		t.Errorf("expected deleted id=100, got %d", deletedId)
	}
	if !strings.Contains(out, "Video library deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestStreamLibrariesDelete_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		deleteVideoLibraryFn: func(_ context.Context, id int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "stream", "libraries", "delete", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

// --- stream libraries reset-api-key ---

func TestStreamLibrariesResetApiKey_Success(t *testing.T) {
	t.Parallel()
	var resetId int64
	mock := &mockVideoLibraryAPI{
		resetVideoLibraryApiKeyFn: func(_ context.Context, id int64) error {
			resetId = id
			return nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "reset-api-key", "100", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resetId != 100 {
		t.Errorf("expected id=100, got %d", resetId)
	}
	if !strings.Contains(out, "API key reset") {
		t.Error("expected reset confirmation message")
	}
}

func TestStreamLibrariesResetApiKey_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		resetVideoLibraryApiKeyFn: func(_ context.Context, id int64) error {
			t.Error("reset should not have been called")
			return nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "stream", "libraries", "reset-api-key", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Operation canceled") {
		t.Error("expected cancellation message")
	}
}

// --- stream libraries languages ---

func TestStreamLibrariesLanguages_Success(t *testing.T) {
	t.Parallel()
	mock := &mockVideoLibraryAPI{
		listVideoLibraryLanguagesFn: func(_ context.Context) ([]client.VideoLibraryLanguage, error) {
			return []client.VideoLibraryLanguage{
				{ShortCode: "en", Name: "English", SupportLevel: 1},
				{ShortCode: "es", Name: "Spanish", SupportLevel: 1},
			}, nil
		},
	}
	app := newTestVideoLibraryApp(mock)

	out, _, err := executeCommand(app, "stream", "libraries", "languages")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "English") {
		t.Error("expected output to contain English")
	}
	if !strings.Contains(out, "Spanish") {
		t.Error("expected output to contain Spanish")
	}
}
