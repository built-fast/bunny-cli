package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// AccountAPI abstracts the bunny.net account API methods,
// allowing tests to inject mocks without making real API calls.
type AccountAPI interface {
	ListApiKeys(ctx context.Context, page, perPage int) (pagination.PageResponse[*client.ApiKey], error)
	GetAuditLog(ctx context.Context, date string, opts client.AuditLogOptions) (*client.AuditLogResponse, error)
}
