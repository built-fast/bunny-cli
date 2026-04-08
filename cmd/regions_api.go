package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
)

// RegionAPI abstracts the bunny.net region API methods,
// allowing tests to inject mocks without making real API calls.
type RegionAPI interface {
	ListRegions(ctx context.Context) ([]*client.Region, error)
}
