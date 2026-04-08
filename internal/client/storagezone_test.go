package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

func TestListStorageZones(t *testing.T) {
	t.Parallel()

	resp := pagination.PageResponse[*StorageZone]{
		Items: []*StorageZone{
			{Id: 1, Name: "zone-1", Region: "DE"},
			{Id: 2, Name: "zone-2", Region: "NY"},
		},
		CurrentPage:  1,
		TotalItems:   2,
		HasMoreItems: false,
	}

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		if r.Header.Get("AccessKey") != "test-key" {
			t.Error("expected AccessKey header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	result, err := c.ListStorageZones(context.Background(), 1, 100, "", false)
	if err != nil {
		t.Fatalf("ListStorageZones: %v", err)
	}

	if capturedPath != "/storagezone?page=1&perPage=100" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Name != "zone-1" {
		t.Errorf("expected zone-1, got %s", result.Items[0].Name)
	}
}

func TestListStorageZones_WithSearch(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagination.PageResponse[*StorageZone]{})
	}))

	_, err := c.ListStorageZones(context.Background(), 1, 20, "myzone", false)
	if err != nil {
		t.Fatalf("ListStorageZones: %v", err)
	}

	if !strings.Contains(capturedPath, "search=myzone") {
		t.Errorf("expected search param in path, got: %s", capturedPath)
	}
}

func TestListStorageZones_IncludeDeleted(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagination.PageResponse[*StorageZone]{})
	}))

	_, err := c.ListStorageZones(context.Background(), 1, 20, "", true)
	if err != nil {
		t.Fatalf("ListStorageZones: %v", err)
	}

	if !strings.Contains(capturedPath, "includeDeleted=true") {
		t.Errorf("expected includeDeleted param in path, got: %s", capturedPath)
	}
}

func TestListStorageZones_PlainArray(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]*StorageZone{ //nolint:gosec // G117: test mock, no real secrets
			{Id: 1, Name: "zone-1"},
		})
	}))

	result, err := c.ListStorageZones(context.Background(), 1, 100, "", false)
	if err != nil {
		t.Fatalf("ListStorageZones: %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.HasMoreItems {
		t.Error("expected HasMoreItems=false for plain array")
	}
}

func TestGetStorageZone(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storagezone/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(StorageZone{Id: 42, Name: "my-storage", Region: "DE", StorageHostname: "storage.bunnycdn.com"}) //nolint:gosec // G117: test mock, no real secrets
	}))

	sz, err := c.GetStorageZone(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetStorageZone: %v", err)
	}
	if sz.Id != 42 || sz.Name != "my-storage" || sz.Region != "DE" {
		t.Errorf("unexpected storage zone: %+v", sz)
	}
}

func TestCreateStorageZone(t *testing.T) {
	t.Parallel()

	var capturedBody StorageZoneCreate
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/storagezone" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(StorageZone{Id: 99, Name: capturedBody.Name, Region: capturedBody.Region}) //nolint:gosec // G117: test mock, no real secrets
	}))

	sz, err := c.CreateStorageZone(context.Background(), &StorageZoneCreate{
		Name:   "new-storage",
		Region: "DE",
	})
	if err != nil {
		t.Fatalf("CreateStorageZone: %v", err)
	}
	if sz.Name != "new-storage" {
		t.Errorf("expected name 'new-storage', got %q", sz.Name)
	}
	if capturedBody.Region != "DE" {
		t.Errorf("expected region 'DE', got %q", capturedBody.Region)
	}
}

func TestUpdateStorageZone(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST (bunny uses POST for update), got %s", r.Method)
		}
		if r.URL.Path != "/storagezone/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	custom404 := "/404.html"
	err := c.UpdateStorageZone(context.Background(), 42, &StorageZoneUpdate{
		Custom404FilePath: &custom404,
	})
	if err != nil {
		t.Fatalf("UpdateStorageZone: %v", err)
	}
}

func TestDeleteStorageZone(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		capturedPath = r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteStorageZone(context.Background(), 42, true)
	if err != nil {
		t.Fatalf("DeleteStorageZone: %v", err)
	}

	if !strings.Contains(capturedPath, "deleteLinkedPullZones=true") {
		t.Errorf("expected deleteLinkedPullZones param, got: %s", capturedPath)
	}
}

func TestDeleteStorageZone_KeepPullZones(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteStorageZone(context.Background(), 42, false)
	if err != nil {
		t.Fatalf("DeleteStorageZone: %v", err)
	}

	if !strings.Contains(capturedPath, "deleteLinkedPullZones=false") {
		t.Errorf("expected deleteLinkedPullZones=false, got: %s", capturedPath)
	}
}

func TestResetStorageZonePassword(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/storagezone/42/resetPassword" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.ResetStorageZonePassword(context.Background(), 42)
	if err != nil {
		t.Fatalf("ResetStorageZonePassword: %v", err)
	}
}

func TestResetStorageZoneReadOnlyPassword(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		capturedPath = r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.ResetStorageZoneReadOnlyPassword(context.Background(), 42)
	if err != nil {
		t.Fatalf("ResetStorageZoneReadOnlyPassword: %v", err)
	}

	if capturedPath != "/storagezone/resetReadOnlyPassword?id=42" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
}

func TestFindStorageZoneByName(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagination.PageResponse[*StorageZone]{
			Items: []*StorageZone{
				{Id: 1, Name: "my-zone", Password: "secret", StorageHostname: "storage.bunnycdn.com"},
				{Id: 2, Name: "my-zone-backup", Password: "other"},
			},
			CurrentPage: 1,
			TotalItems:  2,
		})
	}))

	sz, err := c.FindStorageZoneByName(context.Background(), "my-zone")
	if err != nil {
		t.Fatalf("FindStorageZoneByName: %v", err)
	}
	if sz.Id != 1 || sz.Password != "secret" {
		t.Errorf("unexpected storage zone: %+v", sz)
	}
}

func TestFindStorageZoneByName_NotFound(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagination.PageResponse[*StorageZone]{
			Items: []*StorageZone{
				{Id: 1, Name: "other-zone"},
			},
		})
	}))

	_, err := c.FindStorageZoneByName(context.Background(), "missing-zone")
	if err == nil {
		t.Fatal("expected error for missing zone")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestStorageZoneTierName(t *testing.T) {
	t.Parallel()
	if StorageZoneTierName(0) != "Standard" {
		t.Error("expected Standard for tier 0")
	}
	if StorageZoneTierName(1) != "Edge" {
		t.Error("expected Edge for tier 1")
	}
	if !strings.Contains(StorageZoneTierName(99), "Unknown") {
		t.Error("expected Unknown for unrecognized tier")
	}
}
