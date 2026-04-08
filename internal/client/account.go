package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// ApiKey represents a bunny.net API key.
type ApiKey struct {
	Id    int64    `json:"Id"`
	Key   string   `json:"Key"`
	Roles []string `json:"Roles"`
}

// AuditLogEntry represents a single audit log entry.
type AuditLogEntry struct {
	Timestamp     string `json:"Timestamp"`
	Product       string `json:"Product"`
	ResourceType  string `json:"ResourceType"`
	ResourceId    string `json:"ResourceId"`
	ResourceOwner string `json:"ResourceOwner"`
	Action        string `json:"Action"`
	ActorId       string `json:"ActorId"`
	ActorType     string `json:"ActorType"`
	Diff          string `json:"Diff"`
}

// AuditLogResponse holds the response from the audit log endpoint.
type AuditLogResponse struct {
	Logs              []*AuditLogEntry `json:"Logs"`
	HasMoreData       bool             `json:"HasMoreData"`
	ContinuationToken string           `json:"ContinuationToken"`
}

// AuditLogOptions configures the audit log query.
type AuditLogOptions struct {
	Product           []string
	ResourceType      []string
	ResourceId        []string
	ActorId           []string
	Order             string
	ContinuationToken string
	Limit             int
}

// ListApiKeys returns a paginated list of API keys.
func (c *Client) ListApiKeys(ctx context.Context, page, perPage int) (pagination.PageResponse[*ApiKey], error) {
	if perPage < 5 {
		perPage = 5
	}
	path := fmt.Sprintf("/apikey?page=%d&perPage=%d", page, perPage)

	var resp pagination.PageResponse[*ApiKey]
	if err := c.Get(ctx, path, &resp); err != nil {
		return pagination.PageResponse[*ApiKey]{}, err
	}
	return resp, nil
}

// GetAuditLog returns audit log entries for the given date.
func (c *Client) GetAuditLog(ctx context.Context, date string, opts AuditLogOptions) (*AuditLogResponse, error) {
	params := url.Values{}
	for _, v := range opts.Product {
		params.Add("Product[]", v)
	}
	for _, v := range opts.ResourceType {
		params.Add("ResourceType[]", v)
	}
	for _, v := range opts.ResourceId {
		params.Add("ResourceId[]", v)
	}
	for _, v := range opts.ActorId {
		params.Add("ActorId[]", v)
	}
	if opts.Order != "" {
		params.Set("Order", opts.Order)
	}
	if opts.ContinuationToken != "" {
		params.Set("ContinuationToken", opts.ContinuationToken)
	}
	if opts.Limit > 0 {
		params.Set("Limit", strconv.Itoa(opts.Limit))
	}

	path := "/user/audit/" + date
	if q := params.Encode(); q != "" {
		path += "?" + q
	}

	var resp AuditLogResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BillingRecordTypeName returns a human-readable name for a billing record type.
func BillingRecordTypeName(t int) string {
	switch t {
	case 0:
		return "PayPal"
	case 1:
		return "Crypto"
	case 2:
		return "CreditCard"
	case 3:
		return "MonthlyUsage"
	case 4:
		return "Refund"
	case 5:
		return "CouponCode"
	case 6:
		return "BankTransfer"
	case 7:
		return "AffiliateCredits"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// FormatRoles formats a slice of role strings for display.
func FormatRoles(roles []string) string {
	return strings.Join(roles, ", ")
}
