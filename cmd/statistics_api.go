package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
)

// StatisticsAPI abstracts the bunny.net statistics API methods,
// allowing tests to inject mocks without making real API calls.
type StatisticsAPI interface {
	GetStatistics(ctx context.Context, opts client.StatisticsOptions) (*client.Statistics, error)
}
