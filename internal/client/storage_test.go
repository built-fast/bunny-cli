package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestStorageClient(t *testing.T, handler http.Handler) *StorageClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewStorageClient(StorageClientConfig{
		Password: "zone-password",
		Hostname: srv.URL,
	})
	if err != nil {
		t.Fatalf("NewStorageClient: %v", err)
	}
	return c
}

func TestNewStorageClient_MissingPassword(t *testing.T) {
	t.Parallel()
	_, err := NewStorageClient(StorageClientConfig{
		Hostname: "storage.bunnycdn.com",
	})
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestNewStorageClient_MissingHostname(t *testing.T) {
	t.Parallel()
	_, err := NewStorageClient(StorageClientConfig{
		Password: "pass",
	})
	if err == nil {
		t.Fatal("expected error for missing hostname")
	}
}

func TestNewStorageClient_AddsHTTPS(t *testing.T) {
	t.Parallel()
	c, err := NewStorageClient(StorageClientConfig{
		Password: "pass",
		Hostname: "storage.bunnycdn.com",
	})
	if err != nil {
		t.Fatalf("NewStorageClient: %v", err)
	}
	if c.baseURL != "https://storage.bunnycdn.com" {
		t.Errorf("expected https prefix, got %s", c.baseURL)
	}
}

func TestListFiles(t *testing.T) {
	t.Parallel()

	objects := []StorageObject{
		{ObjectName: "images", IsDirectory: true, Length: 0},
		{ObjectName: "readme.txt", IsDirectory: false, Length: 1024},
	}

	var capturedPath string
	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("AccessKey") != "zone-password" {
			t.Error("expected AccessKey header with zone password")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(objects)
	}))

	result, err := c.ListFiles(context.Background(), "my-zone", "")
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}

	if capturedPath != "/my-zone/" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 objects, got %d", len(result))
	}
	if result[0].ObjectName != "images" || !result[0].IsDirectory {
		t.Errorf("expected directory 'images', got %+v", result[0])
	}
}

func TestListFiles_WithPath(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]StorageObject{})
	}))

	_, err := c.ListFiles(context.Background(), "my-zone", "images/photos")
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}

	if capturedPath != "/my-zone/images/photos/" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
}

func TestDownloadFile(t *testing.T) {
	t.Parallel()

	fileContent := "hello world"
	var capturedPath string
	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("AccessKey") != "zone-password" {
			t.Error("expected AccessKey header")
		}
		w.Header().Set("Content-Length", "11")
		_, _ = w.Write([]byte(fileContent))
	}))

	body, size, err := c.DownloadFile(context.Background(), "my-zone", "path/file.txt")
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	defer func() { _ = body.Close() }()

	if capturedPath != "/my-zone/path/file.txt" {
		t.Errorf("unexpected path: %s", capturedPath)
	}

	if size != 11 {
		t.Errorf("expected size=11, got %d", size)
	}

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("expected %q, got %q", fileContent, string(data))
	}
}

func TestDownloadFile_NotFound(t *testing.T) {
	t.Parallel()

	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	_, _, err := c.DownloadFile(context.Background(), "my-zone", "missing.txt")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	fileContent := "upload content"
	var capturedPath string
	var capturedBody []byte
	var capturedChecksum string

	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("AccessKey") != "zone-password" {
			t.Error("expected AccessKey header")
		}
		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Errorf("expected Content-Type application/octet-stream, got %s", r.Header.Get("Content-Type"))
		}
		capturedChecksum = r.Header.Get("Checksum")
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))

	body := strings.NewReader(fileContent)
	err := c.UploadFile(context.Background(), "my-zone", "path/file.txt", body, int64(len(fileContent)), "ABC123")
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	if capturedPath != "/my-zone/path/file.txt" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
	if string(capturedBody) != fileContent {
		t.Errorf("expected body %q, got %q", fileContent, string(capturedBody))
	}
	if capturedChecksum != "ABC123" {
		t.Errorf("expected checksum ABC123, got %q", capturedChecksum)
	}
}

func TestUploadFile_NoChecksum(t *testing.T) {
	t.Parallel()

	var capturedChecksum string
	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedChecksum = r.Header.Get("Checksum")
		w.WriteHeader(http.StatusCreated)
	}))

	err := c.UploadFile(context.Background(), "my-zone", "file.txt", bytes.NewReader([]byte("data")), 4, "")
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	if capturedChecksum != "" {
		t.Errorf("expected no Checksum header, got %q", capturedChecksum)
	}
}

func TestUploadFile_Error(t *testing.T) {
	t.Parallel()

	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	err := c.UploadFile(context.Background(), "my-zone", "file.txt", bytes.NewReader([]byte("data")), 4, "")
	if err == nil {
		t.Fatal("expected error for 400")
	}
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.Header.Get("AccessKey") != "zone-password" {
			t.Error("expected AccessKey header")
		}
		w.WriteHeader(http.StatusOK)
	}))

	err := c.DeleteFile(context.Background(), "my-zone", "path/file.txt")
	if err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	if capturedPath != "/my-zone/path/file.txt" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
}

func TestDeleteFile_Error(t *testing.T) {
	t.Parallel()

	c := newTestStorageClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	err := c.DeleteFile(context.Background(), "my-zone", "file.txt")
	if err == nil {
		t.Fatal("expected error for 400")
	}
}
