package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func alwaysFalse() bool { return false }

func TestNewClient_NoAPIKey_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewClient(ClientConfig{})
	if err == nil {
		t.Fatal("expected error when no API key configured")
	}
	expected := "API key not configured, run 'bunny configure' or set BUNNY_API_KEY"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestNewClient_WithAPIKey(t *testing.T) {
	t.Parallel()
	c, err := NewClient(ClientConfig{APIKey: "test-key", IsJSON: alwaysFalse})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_DefaultBaseURL(t *testing.T) {
	t.Parallel()
	c, err := NewClient(ClientConfig{APIKey: "test-key", IsJSON: alwaysFalse})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != BaseURLPlatform {
		t.Errorf("expected base URL %q, got %q", BaseURLPlatform, c.baseURL)
	}
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	t.Parallel()
	c, err := NewClient(ClientConfig{
		APIKey:  "test-key",
		BaseURL: BaseURLStream,
		IsJSON:  alwaysFalse,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != BaseURLStream {
		t.Errorf("expected base URL %q, got %q", BaseURLStream, c.baseURL)
	}
}

func TestNewClient_APIURLOverride(t *testing.T) {
	t.Setenv("BUNNY_API_URL", "http://localhost:4010")
	c, err := NewClient(ClientConfig{APIKey: "test-key", IsJSON: alwaysFalse})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != "http://localhost:4010" {
		t.Errorf("expected base URL %q, got %q", "http://localhost:4010", c.baseURL)
	}
}

func TestClient_Get_SetsAccessKeyHeader(t *testing.T) {
	t.Parallel()
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("AccessKey")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "my-secret-key",
	}

	var result map[string]any
	err := c.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHeader != "my-secret-key" {
		t.Errorf("expected AccessKey header %q, got %q", "my-secret-key", gotHeader)
	}
}

func TestClient_Get_SetsAcceptHeader(t *testing.T) {
	t.Parallel()
	var gotAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	var result map[string]any
	_ = c.Get(context.Background(), "/test", &result)
	if gotAccept != "application/json" {
		t.Errorf("expected Accept header %q, got %q", "application/json", gotAccept)
	}
}

func TestClient_Post_SetsContentType(t *testing.T) {
	t.Parallel()
	var gotContentType string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	body := map[string]string{"name": "test"}
	var result map[string]any
	err := c.Post(context.Background(), "/test", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", gotContentType)
	}
	if !strings.Contains(gotBody, `"name":"test"`) {
		t.Errorf("expected body to contain name field, got: %q", gotBody)
	}
}

func TestClient_Get_DecodesResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "test"})
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	var result map[string]string
	err := c.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id %q, got %q", "123", result["id"])
	}
	if result["name"] != "test" {
		t.Errorf("expected name %q, got %q", "test", result["name"])
	}
}

func TestClient_Get_ErrorResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"ErrorKey": "validation_error",
			"Field":    "Name",
			"Message":  "Name is required",
		})
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	var result map[string]any
	err := c.Get(context.Background(), "/test", &result)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	var apiErr *APIError
	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("expected error to contain field name, got: %v", err)
	}
	_ = apiErr
}

func TestClient_Delete(t *testing.T) {
	t.Parallel()
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	err := c.Delete(context.Background(), "/test/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE method, got %q", gotMethod)
	}
}

func TestClient_Get_URLConstruction(t *testing.T) {
	t.Parallel()
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	c := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	var result map[string]any
	_ = c.Get(context.Background(), "/pullzone/123", &result)
	if gotPath != "/pullzone/123" {
		t.Errorf("expected path %q, got %q", "/pullzone/123", gotPath)
	}
}
