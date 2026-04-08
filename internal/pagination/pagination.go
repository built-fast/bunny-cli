package pagination

// PageResponse holds a single page of results from the bunny.net API.
type PageResponse[T any] struct {
	Items        []T  `json:"Items"`
	CurrentPage  int  `json:"CurrentPage"`
	TotalItems   int  `json:"TotalItems"`
	HasMoreItems bool `json:"HasMoreItems"`
}

// Fetcher is a function that fetches a page of results.
// page starts at 1. perPage is the number of items per page.
type Fetcher[T any] func(page, perPage int) (PageResponse[T], error)

// Result holds the items collected from a paginated API along with
// whether more results exist beyond what was collected.
type Result[T any] struct {
	Items   []T
	HasMore bool
}

// Collect fetches pages from a Fetcher and returns collected results.
// If all is true, it fetches every page. Otherwise, it returns up to limit results.
// Default limit is 20 if limit <= 0 and all is false.
// Uses perPage=1000 (bunny max) to minimize requests.
func Collect[T any](fetch Fetcher[T], limit int, all bool) (Result[T], error) {
	if !all && limit <= 0 {
		limit = 20
	}

	const maxPerPage = 1000
	perPage := maxPerPage
	if !all && limit < perPage {
		perPage = limit
	}

	var results []T
	hasMore := false

	for page := 1; ; page++ {
		resp, err := fetch(page, perPage)
		if err != nil {
			return Result[T]{}, err
		}

		if all {
			results = append(results, resp.Items...)
			if !resp.HasMoreItems {
				break
			}
			continue
		}

		remaining := limit - len(results)
		if len(resp.Items) >= remaining {
			results = append(results, resp.Items[:remaining]...)
			hasMore = len(resp.Items) > remaining || resp.HasMoreItems
			break
		}
		results = append(results, resp.Items...)

		if !resp.HasMoreItems {
			break
		}
	}

	return Result[T]{Items: results, HasMore: hasMore}, nil
}
