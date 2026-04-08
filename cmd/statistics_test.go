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

type mockStatisticsAPI struct {
	getStatisticsFn func(ctx context.Context, opts client.StatisticsOptions) (*client.Statistics, error)
}

func (m *mockStatisticsAPI) GetStatistics(ctx context.Context, opts client.StatisticsOptions) (*client.Statistics, error) {
	return m.getStatisticsFn(ctx, opts)
}

func newTestStatisticsApp(api StatisticsAPI) *App {
	return &App{NewStatisticsAPI: func(_ *cobra.Command) (StatisticsAPI, error) { return api, nil }}
}

func sampleStatistics() *client.Statistics {
	return &client.Statistics{
		TotalBandwidthUsed:        1073741824, // 1 GiB
		TotalOriginTraffic:        536870912,  // 512 MiB
		AverageOriginResponseTime: 150,
		TotalRequestsServed:       50000,
		CacheHitRate:              0.95,
	}
}

// --- statistics help ---

func TestStatistics_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "statistics") {
		t.Error("expected root help to show statistics command")
	}
}

func TestStatistics_Alias(t *testing.T) {
	t.Parallel()
	mock := &mockStatisticsAPI{
		getStatisticsFn: func(_ context.Context, opts client.StatisticsOptions) (*client.Statistics, error) {
			return sampleStatistics(), nil
		},
	}
	app := newTestStatisticsApp(mock)

	out, _, err := executeCommand(app, "stats")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "1.00 GiB") {
		t.Errorf("expected bandwidth in output, got %s", out)
	}
}

// --- statistics ---

func TestStatistics_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStatisticsAPI{
		getStatisticsFn: func(_ context.Context, opts client.StatisticsOptions) (*client.Statistics, error) {
			return sampleStatistics(), nil
		},
	}
	app := newTestStatisticsApp(mock)

	out, _, err := executeCommand(app, "statistics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "1.00 GiB") {
		t.Errorf("expected total bandwidth in output, got %s", out)
	}
	if !strings.Contains(out, "95.00%") {
		t.Errorf("expected cache hit rate in output, got %s", out)
	}
	if !strings.Contains(out, "50000") {
		t.Errorf("expected total requests in output, got %s", out)
	}
}

func TestStatistics_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockStatisticsAPI{
		getStatisticsFn: func(_ context.Context, opts client.StatisticsOptions) (*client.Statistics, error) {
			return sampleStatistics(), nil
		},
	}
	app := newTestStatisticsApp(mock)

	out, _, err := executeCommand(app, "statistics", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["TotalRequestsServed"] != float64(50000) {
		t.Errorf("expected TotalRequestsServed=50000, got %v", result["TotalRequestsServed"])
	}
}

func TestStatistics_FlagsPassedToAPI(t *testing.T) {
	t.Parallel()
	var captured client.StatisticsOptions
	mock := &mockStatisticsAPI{
		getStatisticsFn: func(_ context.Context, opts client.StatisticsOptions) (*client.Statistics, error) {
			captured = opts
			return sampleStatistics(), nil
		},
	}
	app := newTestStatisticsApp(mock)

	_, _, err := executeCommand(app, "statistics", "--date-from", "2024-01-01", "--date-to", "2024-01-31", "--pull-zone", "42", "--hourly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.DateFrom != "2024-01-01" {
		t.Errorf("expected DateFrom=2024-01-01, got %q", captured.DateFrom)
	}
	if captured.DateTo != "2024-01-31" {
		t.Errorf("expected DateTo=2024-01-31, got %q", captured.DateTo)
	}
	if captured.PullZone != 42 {
		t.Errorf("expected PullZone=42, got %d", captured.PullZone)
	}
	if !captured.Hourly {
		t.Error("expected Hourly=true")
	}
}

func TestStatistics_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockStatisticsAPI{
		getStatisticsFn: func(_ context.Context, opts client.StatisticsOptions) (*client.Statistics, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestStatisticsApp(mock)

	_, stderr, err := executeCommand(app, "statistics")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}
