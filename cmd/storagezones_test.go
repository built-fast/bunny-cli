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

// mockStorageZoneAPI implements StorageZoneAPI for testing.
type mockStorageZoneAPI struct {
	listStorageZonesFn             func(ctx context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error)
	getStorageZoneFn               func(ctx context.Context, id int64) (*client.StorageZone, error)
	createStorageZoneFn            func(ctx context.Context, body *client.StorageZoneCreate) (*client.StorageZone, error)
	updateStorageZoneFn            func(ctx context.Context, id int64, body *client.StorageZoneUpdate) error
	deleteStorageZoneFn            func(ctx context.Context, id int64, deleteLinkedPullZones bool) error
	resetStorageZonePasswordFn     func(ctx context.Context, id int64) error
	resetStorageZoneReadOnlyPassFn func(ctx context.Context, id int64) error
	findStorageZoneByNameFn        func(ctx context.Context, name string) (*client.StorageZone, error)
}

func (m *mockStorageZoneAPI) ListStorageZones(ctx context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
	return m.listStorageZonesFn(ctx, page, perPage, search, includeDeleted)
}

func (m *mockStorageZoneAPI) GetStorageZone(ctx context.Context, id int64) (*client.StorageZone, error) {
	return m.getStorageZoneFn(ctx, id)
}

func (m *mockStorageZoneAPI) CreateStorageZone(ctx context.Context, body *client.StorageZoneCreate) (*client.StorageZone, error) {
	return m.createStorageZoneFn(ctx, body)
}

func (m *mockStorageZoneAPI) UpdateStorageZone(ctx context.Context, id int64, body *client.StorageZoneUpdate) error {
	return m.updateStorageZoneFn(ctx, id, body)
}

func (m *mockStorageZoneAPI) DeleteStorageZone(ctx context.Context, id int64, deleteLinkedPullZones bool) error {
	return m.deleteStorageZoneFn(ctx, id, deleteLinkedPullZones)
}

func (m *mockStorageZoneAPI) ResetStorageZonePassword(ctx context.Context, id int64) error {
	return m.resetStorageZonePasswordFn(ctx, id)
}

func (m *mockStorageZoneAPI) ResetStorageZoneReadOnlyPassword(ctx context.Context, id int64) error {
	return m.resetStorageZoneReadOnlyPassFn(ctx, id)
}

func (m *mockStorageZoneAPI) FindStorageZoneByName(ctx context.Context, name string) (*client.StorageZone, error) {
	return m.findStorageZoneByNameFn(ctx, name)
}

func newTestStorageZoneApp(api StorageZoneAPI) *App {
	return &App{NewStorageZoneAPI: func(_ *cobra.Command) (StorageZoneAPI, error) { return api, nil }}
}

func sampleStorageZone() *client.StorageZone {
	return &client.StorageZone{
		Id:                 42,
		Name:               "my-storage",
		Password:           "secret-password",
		ReadOnlyPassword:   "readonly-password",
		DateModified:       "2025-01-15T10:30:00Z",
		StorageUsed:        1024000,
		FilesStored:        150,
		Region:             "DE",
		StorageHostname:    "storage.bunnycdn.com",
		ZoneTier:           0,
		ReplicationRegions: []string{"NY"},
		PullZones: []client.PullZone{
			{Id: 1, Name: "cdn-zone"},
		},
	}
}

// --- storagezones help ---

func TestStorageZones_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "storagezones", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "reset-password"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected storagezones help to show %q subcommand", sub)
		}
	}
}

func TestStorageZones_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "sz", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage storage zones") {
		t.Error("expected sz alias to work")
	}
}

// --- storagezones list ---

func TestStorageZonesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		listStorageZonesFn: func(_ context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
			return pagination.PageResponse[*client.StorageZone]{
				Items:        []*client.StorageZone{sampleStorageZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-storage") {
		t.Error("expected output to contain zone name")
	}
	if !strings.Contains(out, "Standard") {
		t.Error("expected output to contain tier name")
	}
}

func TestStorageZonesList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		listStorageZonesFn: func(_ context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
			return pagination.PageResponse[*client.StorageZone]{
				Items:        []*client.StorageZone{sampleStorageZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "list", "--output", "json")
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

func TestStorageZonesList_SearchParam(t *testing.T) {
	t.Parallel()
	var capturedSearch string
	mock := &mockStorageZoneAPI{
		listStorageZonesFn: func(_ context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
			capturedSearch = search
			return pagination.PageResponse[*client.StorageZone]{}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	_, _, err := executeCommand(app, "storagezones", "list", "--search", "test-zone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch != "test-zone" {
		t.Errorf("expected search='test-zone', got %q", capturedSearch)
	}
}

func TestStorageZonesList_IncludeDeleted(t *testing.T) {
	t.Parallel()
	var capturedIncludeDeleted bool
	mock := &mockStorageZoneAPI{
		listStorageZonesFn: func(_ context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
			capturedIncludeDeleted = includeDeleted
			return pagination.PageResponse[*client.StorageZone]{}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	_, _, err := executeCommand(app, "storagezones", "list", "--include-deleted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !capturedIncludeDeleted {
		t.Error("expected includeDeleted=true")
	}
}

func TestStorageZonesList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		listStorageZonesFn: func(_ context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error) {
			return pagination.PageResponse[*client.StorageZone]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestStorageZoneApp(mock)

	_, stderr, err := executeCommand(app, "storagezones", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- storagezones get ---

func TestStorageZonesGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		getStorageZoneFn: func(_ context.Context, id int64) (*client.StorageZone, error) {
			if id != 42 {
				t.Errorf("expected id=42, got %d", id)
			}
			return sampleStorageZone(), nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-storage") {
		t.Error("expected output to contain zone name")
	}
	if !strings.Contains(out, "secret-password") {
		t.Error("expected output to contain password")
	}
}

func TestStorageZonesGet_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		getStorageZoneFn: func(_ context.Context, id int64) (*client.StorageZone, error) {
			return sampleStorageZone(), nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "get", "42", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["Name"] != "my-storage" {
		t.Errorf("expected Name=my-storage, got %v", result["Name"])
	}
}

func TestStorageZonesGet_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "storagezones", "get")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestStorageZonesGet_InvalidID_Fails(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		getStorageZoneFn: func(_ context.Context, id int64) (*client.StorageZone, error) {
			return sampleStorageZone(), nil
		},
	}
	app := newTestStorageZoneApp(mock)

	_, stderr, err := executeCommand(app, "storagezones", "get", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "invalid storage zone ID") {
		t.Errorf("expected 'invalid storage zone ID' error, got %q", stderr)
	}
}

// --- storagezones create ---

func TestStorageZonesCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.StorageZoneCreate
	mock := &mockStorageZoneAPI{
		createStorageZoneFn: func(_ context.Context, body *client.StorageZoneCreate) (*client.StorageZone, error) {
			capturedBody = body
			return &client.StorageZone{Id: 99, Name: body.Name, Region: body.Region}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "create", "--name", "new-storage", "--region", "DE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Name != "new-storage" {
		t.Errorf("expected name 'new-storage', got %q", capturedBody.Name)
	}
	if capturedBody.Region != "DE" {
		t.Errorf("expected region 'DE', got %q", capturedBody.Region)
	}
	if !strings.Contains(out, "new-storage") {
		t.Error("expected output to contain created zone name")
	}
}

func TestStorageZonesCreate_FromFile(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "storagezones", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--from-file") {
		t.Error("expected create command to have --from-file flag")
	}
}

func TestStorageZonesCreate_FromFileStdin(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		createStorageZoneFn: func(_ context.Context, body *client.StorageZoneCreate) (*client.StorageZone, error) {
			return &client.StorageZone{Id: 99, Name: body.Name, Region: body.Region}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	stdin := bytes.NewBufferString(`{"name":"stdin-zone","region":"NY"}`)
	out, _, err := executeCommandWithStdin(app, stdin, "storagezones", "create", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "stdin-zone") {
		t.Error("expected output to contain zone from stdin")
	}
}

// --- storagezones update ---

func TestStorageZonesUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedBody *client.StorageZoneUpdate
	mock := &mockStorageZoneAPI{
		updateStorageZoneFn: func(_ context.Context, id int64, body *client.StorageZoneUpdate) error {
			capturedId = id
			capturedBody = body
			return nil
		},
		getStorageZoneFn: func(_ context.Context, id int64) (*client.StorageZone, error) {
			return &client.StorageZone{Id: id, Name: "updated-zone", Custom404FilePath: "/404.html"}, nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "update", "42", "--custom-404-file-path", "/404.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 42 {
		t.Errorf("expected id=42, got %d", capturedId)
	}
	if capturedBody.Custom404FilePath == nil || *capturedBody.Custom404FilePath != "/404.html" {
		t.Error("expected custom 404 path to be set in body")
	}
	if !strings.Contains(out, "updated-zone") {
		t.Error("expected output to show updated zone")
	}
}

func TestStorageZonesUpdate_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "storagezones", "update")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

// --- storagezones delete ---

func TestStorageZonesDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	var capturedDeleteLinked bool
	mock := &mockStorageZoneAPI{
		deleteStorageZoneFn: func(_ context.Context, id int64, deleteLinkedPullZones bool) error {
			deletedId = id
			capturedDeleteLinked = deleteLinkedPullZones
			return nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "delete", "42", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 42 {
		t.Errorf("expected deleted id=42, got %d", deletedId)
	}
	if !capturedDeleteLinked {
		t.Error("expected deleteLinkedPullZones=true by default")
	}
	if !strings.Contains(out, "Storage zone deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestStorageZonesDelete_KeepPullZones(t *testing.T) {
	t.Parallel()
	var capturedDeleteLinked bool
	mock := &mockStorageZoneAPI{
		deleteStorageZoneFn: func(_ context.Context, id int64, deleteLinkedPullZones bool) error {
			capturedDeleteLinked = deleteLinkedPullZones
			return nil
		},
	}
	app := newTestStorageZoneApp(mock)

	_, _, err := executeCommand(app, "storagezones", "delete", "42", "--yes", "--delete-linked-pull-zones=false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedDeleteLinked {
		t.Error("expected deleteLinkedPullZones=false")
	}
}

func TestStorageZonesDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		deleteStorageZoneFn: func(_ context.Context, id int64, deleteLinkedPullZones bool) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestStorageZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "storagezones", "delete", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

// --- storagezones reset-password ---

func TestStorageZonesResetPassword_Success(t *testing.T) {
	t.Parallel()
	var resetId int64
	mock := &mockStorageZoneAPI{
		resetStorageZonePasswordFn: func(_ context.Context, id int64) error {
			resetId = id
			return nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "reset-password", "42", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resetId != 42 {
		t.Errorf("expected id=42, got %d", resetId)
	}
	if !strings.Contains(out, "password reset") {
		t.Error("expected reset confirmation message")
	}
}

func TestStorageZonesResetPassword_ReadOnly(t *testing.T) {
	t.Parallel()
	var called bool
	mock := &mockStorageZoneAPI{
		resetStorageZoneReadOnlyPassFn: func(_ context.Context, id int64) error {
			called = true
			return nil
		},
	}
	app := newTestStorageZoneApp(mock)

	out, _, err := executeCommand(app, "storagezones", "reset-password", "42", "--yes", "--read-only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected read-only password reset to be called")
	}
	if !strings.Contains(out, "read-only password reset") {
		t.Error("expected read-only reset confirmation message")
	}
}

func TestStorageZonesResetPassword_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockStorageZoneAPI{
		resetStorageZonePasswordFn: func(_ context.Context, id int64) error {
			t.Error("reset should not have been called")
			return nil
		},
	}
	app := newTestStorageZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "storagezones", "reset-password", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Password reset canceled") {
		t.Error("expected cancellation message")
	}
}
