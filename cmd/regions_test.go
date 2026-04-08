package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/spf13/cobra"
)

type mockRegionAPI struct {
	listRegionsFn func(ctx context.Context) ([]*client.Region, error)
}

func (m *mockRegionAPI) ListRegions(ctx context.Context) ([]*client.Region, error) {
	return m.listRegionsFn(ctx)
}

func newTestRegionApp(api RegionAPI) *App {
	return &App{NewRegionAPI: func(_ *cobra.Command) (RegionAPI, error) { return api, nil }}
}

func sampleRegions() []*client.Region {
	return []*client.Region{
		{Id: 1, Name: "New York", RegionCode: "NY", ContinentCode: "NA", CountryCode: "US", PricePerGigabyte: 0.01, AllowLatencyRouting: true},
		{Id: 2, Name: "London", RegionCode: "LO", ContinentCode: "EU", CountryCode: "GB", PricePerGigabyte: 0.02, AllowLatencyRouting: true},
	}
}

// --- regions help ---

func TestRegions_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "regions") {
		t.Error("expected root help to show regions command")
	}
}

// --- regions ---

func TestRegions_Table(t *testing.T) {
	t.Parallel()
	mock := &mockRegionAPI{
		listRegionsFn: func(_ context.Context) ([]*client.Region, error) {
			return sampleRegions(), nil
		},
	}
	app := newTestRegionApp(mock)

	out, _, err := executeCommand(app, "regions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "New York") {
		t.Error("expected output to contain region name")
	}
	if !strings.Contains(out, "London") {
		t.Error("expected output to contain second region name")
	}
}

func TestRegions_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockRegionAPI{
		listRegionsFn: func(_ context.Context) ([]*client.Region, error) {
			return sampleRegions(), nil
		},
	}
	app := newTestRegionApp(mock)

	out, _, err := executeCommand(app, "regions", "--output", "json")
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

func TestRegions_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockRegionAPI{
		listRegionsFn: func(_ context.Context) ([]*client.Region, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestRegionApp(mock)

	_, stderr, err := executeCommand(app, "regions")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

func TestRegions_FieldSelection(t *testing.T) {
	t.Parallel()
	mock := &mockRegionAPI{
		listRegionsFn: func(_ context.Context) ([]*client.Region, error) {
			return sampleRegions(), nil
		},
	}
	app := newTestRegionApp(mock)

	out, _, err := executeCommand(app, "regions", "-f", "Name,Region Code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "New York") {
		t.Error("expected output to contain region name")
	}
}
