package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

// --- stream statistics ---

func TestStreamStatistics_Success(t *testing.T) {
	t.Parallel()
	var capturedLibraryId int64
	var capturedDateFrom, capturedDateTo string
	mock := &mockStreamAPI{
		getVideoStatsFn: func(_ context.Context, libraryId int64, dateFrom, dateTo string, hourly bool, videoGuid string) (*client.VideoStatistics, error) {
			capturedLibraryId = libraryId
			capturedDateFrom = dateFrom
			capturedDateTo = dateTo
			return &client.VideoStatistics{
				EngagementScore:   75,
				ViewsChart:        map[string]int64{"2025-01-01": 100, "2025-01-02": 150},
				WatchTimeChart:    map[string]int64{"2025-01-01": 3600},
				CountryViewCounts: map[string]int64{"US": 200},
				CountryWatchTime:  map[string]int64{"US": 7200},
			}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "statistics", "100", "--date-from", "2025-01-01", "--date-to", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLibraryId != 100 {
		t.Errorf("expected libraryId=100, got %d", capturedLibraryId)
	}
	if capturedDateFrom != "2025-01-01" {
		t.Errorf("expected dateFrom='2025-01-01', got %q", capturedDateFrom)
	}
	if capturedDateTo != "2025-01-31" {
		t.Errorf("expected dateTo='2025-01-31', got %q", capturedDateTo)
	}
	if !strings.Contains(out, "75") {
		t.Error("expected output to contain engagement score")
	}
}

func TestStreamStatistics_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		getVideoStatsFn: func(_ context.Context, libraryId int64, dateFrom, dateTo string, hourly bool, videoGuid string) (*client.VideoStatistics, error) {
			return nil, fmt.Errorf("stats API error")
		},
	}
	app := newTestStreamApp(mock)

	_, stderr, err := executeCommand(app, "stream", "statistics", "100")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "stats API error") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

func TestStreamStatistics_MissingLibraryID(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "statistics")
	if err == nil {
		t.Fatal("expected error for missing library ID")
	}
}

// --- stream heatmap ---

func TestStreamHeatmap_Success(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		getVideoHeatmapFn: func(_ context.Context, libraryId int64, videoId string) (*client.VideoHeatmap, error) {
			if videoId != "abc-123-def" {
				t.Errorf("expected videoId=abc-123-def, got %s", videoId)
			}
			return &client.VideoHeatmap{
				Heatmap: map[string]int{"0": 100, "1": 85, "2": 70},
			}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "heatmap", "100", "abc-123-def")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "heatmap") {
		t.Error("expected output to contain heatmap data")
	}
}

func TestStreamHeatmap_MissingArgs(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "heatmap", "100")
	if err == nil {
		t.Fatal("expected error for missing video ID")
	}
}
