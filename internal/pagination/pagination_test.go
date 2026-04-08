package pagination

import (
	"fmt"
	"testing"
)

// mockFetcher creates a Fetcher from a flat list of items, properly
// paginating them according to the requested perPage size.
func mockFetcher(pages [][]string) Fetcher[string] {
	// Flatten all items into a single list
	var all []string
	for _, p := range pages {
		all = append(all, p...)
	}

	return func(page, perPage int) (PageResponse[string], error) {
		start := (page - 1) * perPage
		if start >= len(all) {
			return PageResponse[string]{
				CurrentPage:  page,
				TotalItems:   len(all),
				HasMoreItems: false,
			}, nil
		}
		end := start + perPage
		if end > len(all) {
			end = len(all)
		}
		return PageResponse[string]{
			Items:        all[start:end],
			CurrentPage:  page,
			TotalItems:   len(all),
			HasMoreItems: end < len(all),
		}, nil
	}
}

func errorFetcher(_ int, _ int) (PageResponse[string], error) {
	return PageResponse[string]{}, fmt.Errorf("api error")
}

func TestCollect_AllPages(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{
		{"a", "b", "c"},
		{"d", "e"},
		{"f"},
	})

	result, err := Collect(fetch, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 6 {
		t.Fatalf("expected 6 results, got %d", len(result.Items))
	}
	expected := []string{"a", "b", "c", "d", "e", "f"}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, v, expected[i])
		}
	}
	if result.HasMore {
		t.Error("expected HasMore=false when all=true")
	}
}

func TestCollect_LimitWithinFirstPage(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{
		{"a", "b", "c", "d", "e"},
		{"f", "g"},
	})

	result, err := Collect(fetch, 3, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Items))
	}
	expected := []string{"a", "b", "c"}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, v, expected[i])
		}
	}
	if !result.HasMore {
		t.Error("expected HasMore=true when results were truncated")
	}
}

func TestCollect_LimitAcrossPages(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
	})

	result, err := Collect(fetch, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 results, got %d", len(result.Items))
	}
	expected := []string{"a", "b", "c", "d", "e"}
	for i, v := range result.Items {
		if v != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, v, expected[i])
		}
	}
	if !result.HasMore {
		t.Error("expected HasMore=true when results were truncated")
	}
}

func TestCollect_DefaultLimit(t *testing.T) {
	t.Parallel()
	page := make([]string, 30)
	for i := range page {
		page[i] = fmt.Sprintf("item-%d", i)
	}
	fetch := mockFetcher([][]string{page})

	result, err := Collect(fetch, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 20 {
		t.Fatalf("expected 20 results (default limit), got %d", len(result.Items))
	}
	if !result.HasMore {
		t.Error("expected HasMore=true when default limit truncates results")
	}
}

func TestCollect_LimitExceedsAvailable(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{
		{"a", "b"},
	})

	result, err := Collect(fetch, 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Items))
	}
	if result.HasMore {
		t.Error("expected HasMore=false when all items fit within limit")
	}
}

func TestCollect_EmptyResponse(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{{}})

	result, err := Collect(fetch, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Items))
	}
	if result.HasMore {
		t.Error("expected HasMore=false for empty response")
	}
}

func TestCollect_FetchError(t *testing.T) {
	t.Parallel()
	result, err := Collect(errorFetcher, 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "api error" {
		t.Errorf("expected 'api error', got %q", err.Error())
	}
	if result.Items != nil {
		t.Errorf("expected nil results on error, got %v", result.Items)
	}
}

func TestCollect_AllIgnoresLimit(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{
		{"a", "b", "c"},
		{"d", "e"},
	})

	result, err := Collect(fetch, 2, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 results (all=true ignores limit), got %d", len(result.Items))
	}
	if result.HasMore {
		t.Error("expected HasMore=false when all=true")
	}
}

func TestCollect_SinglePage(t *testing.T) {
	t.Parallel()
	fetch := mockFetcher([][]string{
		{"only"},
	})

	result, err := Collect(fetch, 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0] != "only" {
		t.Errorf("expected 'only', got %q", result.Items[0])
	}
	if result.HasMore {
		t.Error("expected HasMore=false for single page within limit")
	}
}
