package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
)

// CountryAPI abstracts the bunny.net country API methods,
// allowing tests to inject mocks without making real API calls.
type CountryAPI interface {
	ListCountries(ctx context.Context) ([]*client.Country, error)
}
