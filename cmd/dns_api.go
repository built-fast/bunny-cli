package cmd

import (
	"context"
	"io"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// DnsZoneAPI abstracts the bunny.net DNS zone API methods,
// allowing tests to inject mocks without making real API calls.
type DnsZoneAPI interface {
	ListDnsZones(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error)
	GetDnsZone(ctx context.Context, id int64) (*client.DnsZone, error)
	CreateDnsZone(ctx context.Context, body *client.DnsZoneCreate) (*client.DnsZone, error)
	UpdateDnsZone(ctx context.Context, id int64, body *client.DnsZoneUpdate) (*client.DnsZone, error)
	DeleteDnsZone(ctx context.Context, id int64) error
	AddDnsRecord(ctx context.Context, zoneId int64, body *client.DnsRecordCreate) (*client.DnsRecord, error)
	UpdateDnsRecord(ctx context.Context, zoneId, recordId int64, body *client.DnsRecordUpdate) error
	DeleteDnsRecord(ctx context.Context, zoneId, recordId int64) error
	ImportDnsZone(ctx context.Context, zoneId int64, data io.Reader) (*client.DnsZoneImportResult, error)
	ExportDnsZone(ctx context.Context, zoneId int64) ([]byte, error)
	EnableDnsSec(ctx context.Context, zoneId int64) (*client.DnsSecInfo, error)
	DisableDnsSec(ctx context.Context, zoneId int64) (*client.DnsSecInfo, error)
}
