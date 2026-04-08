package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// StorageZoneAPI abstracts the bunny.net storage zone API methods,
// allowing tests to inject mocks without making real API calls.
type StorageZoneAPI interface {
	ListStorageZones(ctx context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*client.StorageZone], error)
	GetStorageZone(ctx context.Context, id int64) (*client.StorageZone, error)
	CreateStorageZone(ctx context.Context, body *client.StorageZoneCreate) (*client.StorageZone, error)
	UpdateStorageZone(ctx context.Context, id int64, body *client.StorageZoneUpdate) error
	DeleteStorageZone(ctx context.Context, id int64, deleteLinkedPullZones bool) error
	ResetStorageZonePassword(ctx context.Context, id int64) error
	ResetStorageZoneReadOnlyPassword(ctx context.Context, id int64) error
	FindStorageZoneByName(ctx context.Context, name string) (*client.StorageZone, error)
}
