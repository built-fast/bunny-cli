package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

type mockAccountAPI struct {
	listApiKeysFn func(ctx context.Context, page, perPage int) (pagination.PageResponse[*client.ApiKey], error)
	getAuditLogFn func(ctx context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error)
}

func (m *mockAccountAPI) ListApiKeys(ctx context.Context, page, perPage int) (pagination.PageResponse[*client.ApiKey], error) {
	return m.listApiKeysFn(ctx, page, perPage)
}

func (m *mockAccountAPI) GetAuditLog(ctx context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error) {
	return m.getAuditLogFn(ctx, date, opts)
}

func newTestAccountApp(api AccountAPI) *App {
	return &App{NewAccountAPI: func(_ *cobra.Command) (AccountAPI, error) { return api, nil }}
}

func sampleApiKeys() []*client.ApiKey {
	return []*client.ApiKey{
		{Id: 1, Key: "abc123", Roles: []string{"User", "UserApi"}},
		{Id: 2, Key: "def456", Roles: []string{"User"}},
	}
}

func sampleAuditLog() *client.AuditLogResponse {
	return &client.AuditLogResponse{
		Logs: []*client.AuditLogEntry{
			{
				Timestamp:    "2024-01-15T10:30:00Z",
				Product:      "CDN",
				ResourceType: "PullZone",
				ResourceId:   "12345",
				Action:       "Updated",
				ActorId:      "user-1",
				ActorType:    "User",
			},
		},
		HasMoreData: false,
	}
}

// --- account help ---

func TestAccount_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "account", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"api-keys", "audit-log"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected account help to show %q subcommand", sub)
		}
	}
}

// --- account api-keys list ---

func TestApiKeysList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockAccountAPI{
		listApiKeysFn: func(_ context.Context, page, perPage int) (pagination.PageResponse[*client.ApiKey], error) {
			return pagination.PageResponse[*client.ApiKey]{
				Items:        sampleApiKeys(),
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestAccountApp(mock)

	out, _, err := executeCommand(app, "account", "api-keys", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "abc123") {
		t.Error("expected output to contain API key")
	}
	if !strings.Contains(out, "User, UserApi") {
		t.Error("expected output to contain roles")
	}
}

func TestApiKeysList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockAccountAPI{
		listApiKeysFn: func(_ context.Context, page, perPage int) (pagination.PageResponse[*client.ApiKey], error) {
			return pagination.PageResponse[*client.ApiKey]{
				Items:        sampleApiKeys(),
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestAccountApp(mock)

	out, _, err := executeCommand(app, "account", "api-keys", "list", "--output", "json")
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

func TestApiKeysList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockAccountAPI{
		listApiKeysFn: func(_ context.Context, page, perPage int) (pagination.PageResponse[*client.ApiKey], error) {
			return pagination.PageResponse[*client.ApiKey]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestAccountApp(mock)

	_, stderr, err := executeCommand(app, "account", "api-keys", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- account audit-log ---

func TestAuditLog_Table(t *testing.T) {
	t.Parallel()
	mock := &mockAccountAPI{
		getAuditLogFn: func(_ context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error) {
			if date != "2024-01-15" {
				t.Errorf("expected date=2024-01-15, got %q", date)
			}
			return sampleAuditLog(), nil
		},
	}
	app := newTestAccountApp(mock)

	out, _, err := executeCommand(app, "account", "audit-log", "2024-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PullZone") {
		t.Error("expected output to contain resource type")
	}
	if !strings.Contains(out, "Updated") {
		t.Error("expected output to contain action")
	}
}

func TestAuditLog_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockAccountAPI{
		getAuditLogFn: func(_ context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error) {
			return sampleAuditLog(), nil
		},
	}
	app := newTestAccountApp(mock)

	out, _, err := executeCommand(app, "account", "audit-log", "2024-01-15", "--output", "json")
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

func TestAuditLog_WithoutDate_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "account", "audit-log")
	if err == nil {
		t.Fatal("expected error for missing date argument")
	}
}

func TestAuditLog_FilterParams(t *testing.T) {
	t.Parallel()
	var captured client.AuditLogOptions
	mock := &mockAccountAPI{
		getAuditLogFn: func(_ context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error) {
			captured = opts
			return sampleAuditLog(), nil
		},
	}
	app := newTestAccountApp(mock)

	_, _, err := executeCommand(app, "account", "audit-log", "2024-01-15",
		"--product", "CDN",
		"--resource-type", "PullZone",
		"--order", "Descending",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.Product) != 1 || captured.Product[0] != "CDN" {
		t.Errorf("expected Product=[CDN], got %v", captured.Product)
	}
	if len(captured.ResourceType) != 1 || captured.ResourceType[0] != "PullZone" {
		t.Errorf("expected ResourceType=[PullZone], got %v", captured.ResourceType)
	}
	if captured.Order != "Descending" {
		t.Errorf("expected Order=Descending, got %q", captured.Order)
	}
}

func TestAuditLog_AllPages(t *testing.T) {
	t.Parallel()
	callCount := 0
	mock := &mockAccountAPI{
		getAuditLogFn: func(_ context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error) {
			callCount++
			if callCount == 1 {
				return &client.AuditLogResponse{
					Logs:              []*client.AuditLogEntry{{Timestamp: "2024-01-15T10:00:00Z", Action: "Created"}},
					HasMoreData:       true,
					ContinuationToken: "token-1",
				}, nil
			}
			if opts.ContinuationToken != "token-1" {
				t.Errorf("expected continuation token 'token-1', got %q", opts.ContinuationToken)
			}
			return &client.AuditLogResponse{
				Logs:        []*client.AuditLogEntry{{Timestamp: "2024-01-15T11:00:00Z", Action: "Updated"}},
				HasMoreData: false,
			}, nil
		},
	}
	app := newTestAccountApp(mock)

	out, _, err := executeCommand(app, "account", "audit-log", "2024-01-15", "--all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
	if !strings.Contains(out, "Created") || !strings.Contains(out, "Updated") {
		t.Error("expected output to contain entries from both pages")
	}
}

func TestAuditLog_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockAccountAPI{
		getAuditLogFn: func(_ context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestAccountApp(mock)

	_, stderr, err := executeCommand(app, "account", "audit-log", "2024-01-15")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}
