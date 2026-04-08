package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
)

// BillingAPI abstracts the bunny.net billing API methods,
// allowing tests to inject mocks without making real API calls.
type BillingAPI interface {
	GetBillingDetails(ctx context.Context) (*client.BillingDetails, error)
	GetBillingSummary(ctx context.Context) ([]*client.BillingSummaryItem, error)
	DownloadInvoice(ctx context.Context, billingRecordId int64) ([]byte, error)
}
