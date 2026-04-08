package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// PullZoneAPI abstracts the bunny.net pull zone API methods,
// allowing tests to inject mocks without making real API calls.
type PullZoneAPI interface {
	ListPullZones(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error)
	GetPullZone(ctx context.Context, id int64) (*client.PullZone, error)
	CreatePullZone(ctx context.Context, body *client.PullZoneCreate) (*client.PullZone, error)
	UpdatePullZone(ctx context.Context, id int64, body *client.PullZoneUpdate) (*client.PullZone, error)
	DeletePullZone(ctx context.Context, id int64) error
	AddPullZoneHostname(ctx context.Context, id int64, hostname string) error
	RemovePullZoneHostname(ctx context.Context, id int64, hostname string) error
	PurgePullZoneCache(ctx context.Context, id int64, cacheTag string) error
	AddOrUpdateEdgeRule(ctx context.Context, pullZoneId int64, rule *client.EdgeRule) error
	DeleteEdgeRule(ctx context.Context, pullZoneId int64, edgeRuleId string) error
	SetEdgeRuleEnabled(ctx context.Context, pullZoneId int64, edgeRuleId string, enabled bool) error
}
