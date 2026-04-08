package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

func sampleCollection() *client.Collection {
	return &client.Collection{
		VideoLibraryId: 100,
		Guid:           "col-abc-123",
		Name:           "My Collection",
		VideoCount:     10,
		TotalSize:      500000,
	}
}

// --- stream collections help ---

func TestStreamCollections_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "stream", "collections", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected collections help to show %q subcommand", sub)
		}
	}
}

// --- stream collections list ---

func TestStreamCollectionsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		listCollectionsFn: func(_ context.Context, libraryId int64, page, itemsPerPage int, search, orderBy string) (pagination.PageResponse[*client.Collection], error) {
			if libraryId != 100 {
				t.Errorf("expected libraryId=100, got %d", libraryId)
			}
			return pagination.PageResponse[*client.Collection]{
				Items:        []*client.Collection{sampleCollection()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "collections", "list", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "My Collection") {
		t.Error("expected output to contain collection name")
	}
}

func TestStreamCollectionsList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		listCollectionsFn: func(_ context.Context, libraryId int64, page, itemsPerPage int, search, orderBy string) (pagination.PageResponse[*client.Collection], error) {
			return pagination.PageResponse[*client.Collection]{}, fmt.Errorf("collections API error")
		},
	}
	app := newTestStreamApp(mock)

	_, stderr, err := executeCommand(app, "stream", "collections", "list", "100")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "collections API error") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- stream collections get ---

func TestStreamCollectionsGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		getCollectionFn: func(_ context.Context, libraryId int64, collectionId string) (*client.Collection, error) {
			if collectionId != "col-abc-123" {
				t.Errorf("expected collectionId=col-abc-123, got %s", collectionId)
			}
			return sampleCollection(), nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "collections", "get", "100", "col-abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "My Collection") {
		t.Error("expected output to contain collection name")
	}
}

// --- stream collections create ---

func TestStreamCollectionsCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedName string
	mock := &mockStreamAPI{
		createCollectionFn: func(_ context.Context, libraryId int64, body *client.CollectionCreate) (*client.Collection, error) {
			capturedName = body.Name
			return &client.Collection{Guid: "new-col-id", Name: body.Name, VideoLibraryId: libraryId}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "collections", "create", "100", "--name", "New Collection")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "New Collection" {
		t.Errorf("expected name='New Collection', got %q", capturedName)
	}
	if !strings.Contains(out, "New Collection") {
		t.Error("expected output to contain collection name")
	}
}

func TestStreamCollectionsCreate_RequiresName(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "collections", "create", "100")
	if err == nil {
		t.Fatal("expected error for missing required --name flag")
	}
}

// --- stream collections update ---

func TestStreamCollectionsUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedName string
	mock := &mockStreamAPI{
		updateCollectionFn: func(_ context.Context, libraryId int64, collectionId string, body *client.CollectionUpdate) error {
			capturedName = body.Name
			return nil
		},
		getCollectionFn: func(_ context.Context, libraryId int64, collectionId string) (*client.Collection, error) {
			return &client.Collection{Guid: collectionId, Name: "Updated Name", VideoLibraryId: libraryId}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "collections", "update", "100", "col-abc-123", "--name", "Updated Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "Updated Name" {
		t.Errorf("expected name='Updated Name', got %q", capturedName)
	}
	if !strings.Contains(out, "Updated Name") {
		t.Error("expected output to show updated name")
	}
}

// --- stream collections delete ---

func TestStreamCollectionsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedCollectionId string
	mock := &mockStreamAPI{
		deleteCollectionFn: func(_ context.Context, libraryId int64, collectionId string) error {
			deletedCollectionId = collectionId
			return nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "collections", "delete", "100", "col-abc-123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedCollectionId != "col-abc-123" {
		t.Errorf("expected deleted collectionId=col-abc-123, got %s", deletedCollectionId)
	}
	if !strings.Contains(out, "Collection deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestStreamCollectionsDelete_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		deleteCollectionFn: func(_ context.Context, libraryId int64, collectionId string) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestStreamApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "stream", "collections", "delete", "100", "col-abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}
