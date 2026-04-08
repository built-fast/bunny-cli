package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient(ClientConfig{
		APIKey:  "test-key",
		BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestListPullZones(t *testing.T) {
	t.Parallel()

	resp := pagination.PageResponse[*PullZone]{
		Items: []*PullZone{
			{Id: 1, Name: "zone-1"},
			{Id: 2, Name: "zone-2"},
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

	result, err := c.ListPullZones(context.Background(), 1, 100, "")
	if err != nil {
		t.Fatalf("ListPullZones: %v", err)
	}

	if capturedPath != "/pullzone?page=1&perPage=100" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Name != "zone-1" {
		t.Errorf("expected zone-1, got %s", result.Items[0].Name)
	}
}

func TestListPullZones_WithSearch(t *testing.T) {
	t.Parallel()

	var capturedPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagination.PageResponse[*PullZone]{})
	}))

	_, err := c.ListPullZones(context.Background(), 1, 20, "myzone")
	if err != nil {
		t.Fatalf("ListPullZones: %v", err)
	}

	if !strings.Contains(capturedPath, "search=myzone") {
		t.Errorf("expected search param in path, got: %s", capturedPath)
	}
}

func TestGetPullZone(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pullzone/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PullZone{Id: 42, Name: "my-zone", Enabled: true})
	}))

	pz, err := c.GetPullZone(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetPullZone: %v", err)
	}
	if pz.Id != 42 || pz.Name != "my-zone" || !pz.Enabled {
		t.Errorf("unexpected pull zone: %+v", pz)
	}
}

func TestCreatePullZone(t *testing.T) {
	t.Parallel()

	var capturedBody PullZoneCreate
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pullzone" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(PullZone{Id: 99, Name: capturedBody.Name, OriginUrl: capturedBody.OriginUrl})
	}))

	pz, err := c.CreatePullZone(context.Background(), &PullZoneCreate{
		Name:      "new-zone",
		OriginUrl: "https://origin.example.com",
	})
	if err != nil {
		t.Fatalf("CreatePullZone: %v", err)
	}
	if pz.Name != "new-zone" {
		t.Errorf("expected name 'new-zone', got %q", pz.Name)
	}
	if capturedBody.Name != "new-zone" {
		t.Errorf("expected request body name 'new-zone', got %q", capturedBody.Name)
	}
}

func TestUpdatePullZone(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST (bunny uses POST for update), got %s", r.Method)
		}
		if r.URL.Path != "/pullzone/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PullZone{Id: 42, Name: "updated-zone", OriginUrl: "https://new-origin.com"})
	}))

	origin := "https://new-origin.com"
	pz, err := c.UpdatePullZone(context.Background(), 42, &PullZoneUpdate{
		OriginUrl: &origin,
	})
	if err != nil {
		t.Fatalf("UpdatePullZone: %v", err)
	}
	if pz.OriginUrl != "https://new-origin.com" {
		t.Errorf("expected updated origin URL, got %q", pz.OriginUrl)
	}
}

func TestDeletePullZone(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/pullzone/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeletePullZone(context.Background(), 42)
	if err != nil {
		t.Fatalf("DeletePullZone: %v", err)
	}
}

func TestAddPullZoneHostname(t *testing.T) {
	t.Parallel()

	var capturedBody map[string]string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pullzone/42/addHostname" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.AddPullZoneHostname(context.Background(), 42, "cdn.example.com")
	if err != nil {
		t.Fatalf("AddPullZoneHostname: %v", err)
	}
	if capturedBody["Hostname"] != "cdn.example.com" {
		t.Errorf("expected hostname 'cdn.example.com', got %q", capturedBody["Hostname"])
	}
}

func TestRemovePullZoneHostname(t *testing.T) {
	t.Parallel()

	var capturedBody map[string]string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.RemovePullZoneHostname(context.Background(), 42, "cdn.example.com")
	if err != nil {
		t.Fatalf("RemovePullZoneHostname: %v", err)
	}
	if capturedBody["Hostname"] != "cdn.example.com" {
		t.Errorf("expected hostname 'cdn.example.com', got %q", capturedBody["Hostname"])
	}
}

func TestPurgePullZoneCache(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pullzone/42/purgeCache" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.PurgePullZoneCache(context.Background(), 42, "")
	if err != nil {
		t.Fatalf("PurgePullZoneCache: %v", err)
	}
}

func TestPurgePullZoneCache_WithTag(t *testing.T) {
	t.Parallel()

	var capturedBody map[string]string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.PurgePullZoneCache(context.Background(), 42, "my-tag")
	if err != nil {
		t.Fatalf("PurgePullZoneCache: %v", err)
	}
	if capturedBody["CacheTag"] != "my-tag" {
		t.Errorf("expected cache tag 'my-tag', got %q", capturedBody["CacheTag"])
	}
}

func TestDeleteEdgeRule(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/pullzone/42/edgerules/rule-abc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteEdgeRule(context.Background(), 42, "rule-abc")
	if err != nil {
		t.Fatalf("DeleteEdgeRule: %v", err)
	}
}

func TestPullZoneTypeName(t *testing.T) {
	t.Parallel()
	if PullZoneTypeName(0) != "Premium" {
		t.Error("expected Premium for type 0")
	}
	if PullZoneTypeName(1) != "Volume" {
		t.Error("expected Volume for type 1")
	}
	if !strings.Contains(PullZoneTypeName(99), "Unknown") {
		t.Error("expected Unknown for unrecognized type")
	}
}
