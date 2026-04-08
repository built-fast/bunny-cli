package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

// mockStorageAPI implements StorageAPI for testing.
type mockStorageAPI struct {
	listFilesFn    func(ctx context.Context, zoneName, path string) ([]client.StorageObject, error)
	downloadFileFn func(ctx context.Context, zoneName, path string) (io.ReadCloser, int64, error)
	uploadFileFn   func(ctx context.Context, zoneName, path string, body io.Reader, size int64, checksum string) error
	deleteFileFn   func(ctx context.Context, zoneName, path string) error
}

func (m *mockStorageAPI) ListFiles(ctx context.Context, zoneName, path string) ([]client.StorageObject, error) {
	return m.listFilesFn(ctx, zoneName, path)
}

func (m *mockStorageAPI) DownloadFile(ctx context.Context, zoneName, path string) (io.ReadCloser, int64, error) {
	return m.downloadFileFn(ctx, zoneName, path)
}

func (m *mockStorageAPI) UploadFile(ctx context.Context, zoneName, path string, body io.Reader, size int64, checksum string) error {
	return m.uploadFileFn(ctx, zoneName, path, body, size, checksum)
}

func (m *mockStorageAPI) DeleteFile(ctx context.Context, zoneName, path string) error {
	return m.deleteFileFn(ctx, zoneName, path)
}

// newTestStorageApp creates an App with both StorageZoneAPI (for auto-lookup) and StorageAPI mocked.
func newTestStorageApp(szAPI StorageZoneAPI, sAPI StorageAPI) *App {
	return &App{
		NewStorageZoneAPI: func(_ *cobra.Command) (StorageZoneAPI, error) { return szAPI, nil },
		NewStorageAPI: func(_ *cobra.Command, password, hostname string) (StorageAPI, error) {
			return sAPI, nil
		},
	}
}

func sampleStorageObjects() []client.StorageObject {
	return []client.StorageObject{
		{ObjectName: "images", IsDirectory: true, Length: 0, LastChanged: "2025-01-15T10:00:00Z"},
		{ObjectName: "readme.txt", IsDirectory: false, Length: 1024, LastChanged: "2025-01-15T12:00:00Z"},
	}
}

// mockStorageZoneForLookup returns a mock that supports FindStorageZoneByName.
func mockStorageZoneForLookup() *mockStorageZoneAPI {
	return &mockStorageZoneAPI{
		findStorageZoneByNameFn: func(_ context.Context, name string) (*client.StorageZone, error) {
			return &client.StorageZone{
				Id:              1,
				Name:            name,
				Password:        "zone-pass",
				StorageHostname: "storage.bunnycdn.com",
			}, nil
		},
		// Provide a stub for ListStorageZones since FindStorageZoneByName calls it
		listStorageZonesFn: func(_ context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
			return pagination.PageResponse[*client.StorageZone]{
				Items: []*client.StorageZone{
					{Id: 1, Name: search, Password: "zone-pass", StorageHostname: "storage.bunnycdn.com"},
				},
			}, nil
		},
	}
}

// --- parseZonePath ---

func TestParseZonePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		wantZone string
		wantPath string
	}{
		{"my-zone", "my-zone", ""},
		{"my-zone/", "my-zone", ""},
		{"my-zone/images", "my-zone", "images"},
		{"my-zone/images/photos/pic.jpg", "my-zone", "images/photos/pic.jpg"},
	}

	for _, tt := range tests {
		zone, path := parseZonePath(tt.input)
		if zone != tt.wantZone || path != tt.wantPath {
			t.Errorf("parseZonePath(%q) = (%q, %q), want (%q, %q)", tt.input, zone, path, tt.wantZone, tt.wantPath)
		}
	}
}

// --- storage help ---

func TestStorage_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "storage", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"ls", "cp", "rm"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected storage help to show %q subcommand", sub)
		}
	}
}

func TestStorage_PersistentFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "storage", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--password", "--hostname"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected storage help to contain %s flag", flag)
		}
	}
}

// --- storage ls ---

func TestStorageLs_Table(t *testing.T) {
	t.Parallel()

	storageAPI := &mockStorageAPI{
		listFilesFn: func(_ context.Context, zoneName, path string) ([]client.StorageObject, error) {
			if zoneName != "my-zone" {
				t.Errorf("expected zone 'my-zone', got %q", zoneName)
			}
			return sampleStorageObjects(), nil
		},
	}
	app := newTestStorageApp(mockStorageZoneForLookup(), storageAPI)

	out, _, err := executeCommand(app, "storage", "ls", "my-zone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "images") {
		t.Error("expected output to contain 'images'")
	}
	if !strings.Contains(out, "readme.txt") {
		t.Error("expected output to contain 'readme.txt'")
	}
	if !strings.Contains(out, "dir") {
		t.Error("expected output to contain 'dir' type")
	}
	if !strings.Contains(out, "file") {
		t.Error("expected output to contain 'file' type")
	}
}

func TestStorageLs_WithPath(t *testing.T) {
	t.Parallel()

	var capturedPath string
	storageAPI := &mockStorageAPI{
		listFilesFn: func(_ context.Context, zoneName, path string) ([]client.StorageObject, error) {
			capturedPath = path
			return []client.StorageObject{}, nil
		},
	}
	app := newTestStorageApp(mockStorageZoneForLookup(), storageAPI)

	_, _, err := executeCommand(app, "storage", "ls", "my-zone/images/photos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPath != "images/photos" {
		t.Errorf("expected path 'images/photos', got %q", capturedPath)
	}
}

func TestStorageLs_JSON(t *testing.T) {
	t.Parallel()

	storageAPI := &mockStorageAPI{
		listFilesFn: func(_ context.Context, zoneName, path string) ([]client.StorageObject, error) {
			return sampleStorageObjects(), nil
		},
	}
	app := newTestStorageApp(mockStorageZoneForLookup(), storageAPI)

	out, _, err := executeCommand(app, "storage", "ls", "my-zone", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"object":"list"`) {
		t.Error("expected JSON list envelope")
	}
}

func TestStorageLs_WithExplicitCredentials(t *testing.T) {
	t.Parallel()

	storageAPI := &mockStorageAPI{
		listFilesFn: func(_ context.Context, zoneName, path string) ([]client.StorageObject, error) {
			return []client.StorageObject{}, nil
		},
	}
	// No StorageZoneAPI mock needed — explicit credentials skip lookup
	app := &App{
		NewStorageAPI: func(_ *cobra.Command, password, hostname string) (StorageAPI, error) {
			if password != "my-pass" {
				t.Errorf("expected password 'my-pass', got %q", password)
			}
			if hostname != "ny.storage.bunnycdn.com" {
				t.Errorf("expected hostname 'ny.storage.bunnycdn.com', got %q", hostname)
			}
			return storageAPI, nil
		},
	}

	_, _, err := executeCommand(app, "storage", "ls", "my-zone", "--password", "my-pass", "--hostname", "ny.storage.bunnycdn.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageLs_ErrorPropagation(t *testing.T) {
	t.Parallel()

	storageAPI := &mockStorageAPI{
		listFilesFn: func(_ context.Context, zoneName, path string) ([]client.StorageObject, error) {
			return nil, fmt.Errorf("storage API error")
		},
	}
	app := newTestStorageApp(mockStorageZoneForLookup(), storageAPI)

	_, stderr, err := executeCommand(app, "storage", "ls", "my-zone")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "storage API error") {
		t.Errorf("expected error in stderr, got %q", stderr)
	}
}

// --- storage rm ---

func TestStorageRm_WithYes(t *testing.T) {
	t.Parallel()

	var deletedZone, deletedPath string
	storageAPI := &mockStorageAPI{
		deleteFileFn: func(_ context.Context, zoneName, path string) error {
			deletedZone = zoneName
			deletedPath = path
			return nil
		},
	}
	app := newTestStorageApp(mockStorageZoneForLookup(), storageAPI)

	out, _, err := executeCommand(app, "storage", "rm", "my-zone/path/file.txt", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedZone != "my-zone" {
		t.Errorf("expected zone 'my-zone', got %q", deletedZone)
	}
	if deletedPath != "path/file.txt" {
		t.Errorf("expected path 'path/file.txt', got %q", deletedPath)
	}
	if !strings.Contains(out, "File deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestStorageRm_Canceled(t *testing.T) {
	t.Parallel()

	storageAPI := &mockStorageAPI{
		deleteFileFn: func(_ context.Context, zoneName, path string) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestStorageApp(mockStorageZoneForLookup(), storageAPI)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "storage", "rm", "my-zone/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

func TestStorageRm_NoPath_Fails(t *testing.T) {
	t.Parallel()

	app := newTestStorageApp(mockStorageZoneForLookup(), &mockStorageAPI{})

	_, stderr, err := executeCommand(app, "storage", "rm", "my-zone", "--yes")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if !strings.Contains(stderr, "path is required") {
		t.Errorf("expected 'path is required' error, got %q", stderr)
	}
}

// --- storage cp direction detection ---

func TestStorageCp_MissingArgs_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "storage", "cp", "only-one-arg")
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestStorageCp_ChecksumFlag(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "storage", "cp", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--checksum") {
		t.Error("expected cp command to have --checksum flag")
	}
}
