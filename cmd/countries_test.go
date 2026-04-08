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

type mockCountryAPI struct {
	listCountriesFn func(ctx context.Context) ([]*client.Country, error)
}

func (m *mockCountryAPI) ListCountries(ctx context.Context) ([]*client.Country, error) {
	return m.listCountriesFn(ctx)
}

func newTestCountryApp(api CountryAPI) *App {
	return &App{NewCountryAPI: func(_ *cobra.Command) (CountryAPI, error) { return api, nil }}
}

func sampleCountries() []*client.Country {
	return []*client.Country{
		{Name: "United States", IsoCode: "US", IsEU: false, TaxRate: 0, PopList: []string{"NY", "LA", "MI"}},
		{Name: "Germany", IsoCode: "DE", IsEU: true, TaxRate: 19.0, TaxPrefix: "DE", PopList: []string{"FR"}},
	}
}

// --- countries help ---

func TestCountries_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "countries") {
		t.Error("expected root help to show countries command")
	}
}

// --- countries ---

func TestCountries_Table(t *testing.T) {
	t.Parallel()
	mock := &mockCountryAPI{
		listCountriesFn: func(_ context.Context) ([]*client.Country, error) {
			return sampleCountries(), nil
		},
	}
	app := newTestCountryApp(mock)

	out, _, err := executeCommand(app, "countries")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "United States") {
		t.Error("expected output to contain country name")
	}
	if !strings.Contains(out, "Germany") {
		t.Error("expected output to contain second country name")
	}
}

func TestCountries_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockCountryAPI{
		listCountriesFn: func(_ context.Context) ([]*client.Country, error) {
			return sampleCountries(), nil
		},
	}
	app := newTestCountryApp(mock)

	out, _, err := executeCommand(app, "countries", "--output", "json")
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

func TestCountries_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockCountryAPI{
		listCountriesFn: func(_ context.Context) ([]*client.Country, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestCountryApp(mock)

	_, stderr, err := executeCommand(app, "countries")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}
