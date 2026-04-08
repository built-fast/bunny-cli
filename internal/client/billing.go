package client

import (
	"context"
	"fmt"
)

// BillingDetails represents the full billing response from bunny.net.
type BillingDetails struct {
	Balance                           float64          `json:"Balance"`
	ThisMonthCharges                  float64          `json:"ThisMonthCharges"`
	LastRechargeBalance               float64          `json:"LastRechargeBalance"`
	BillingRecords                    []*BillingRecord `json:"BillingRecords"`
	MonthlyChargesEUTraffic           float64          `json:"MonthlyChargesEUTraffic"`
	MonthlyChargesUSTraffic           float64          `json:"MonthlyChargesUSTraffic"`
	MonthlyChargesASIATraffic         float64          `json:"MonthlyChargesASIATraffic"`
	MonthlyChargesAFTraffic           float64          `json:"MonthlyChargesAFTraffic"`
	MonthlyChargesSATraffic           float64          `json:"MonthlyChargesSATraffic"`
	MonthlyChargesStorage             float64          `json:"MonthlyChargesStorage"`
	MonthlyChargesDNS                 float64          `json:"MonthlyChargesDNS"`
	MonthlyChargesOptimizer           float64          `json:"MonthlyChargesOptimizer"`
	MonthlyChargesTranscribe          float64          `json:"MonthlyChargesTranscribe"`
	MonthlyChargesPremiumEncoding     float64          `json:"MonthlyChargesPremiumEncoding"`
	MonthlyChargesExtraPullZones      float64          `json:"MonthlyChargesExtraPullZones"`
	MonthlyChargesExtraStorageZones   float64          `json:"MonthlyChargesExtraStorageZones"`
	MonthlyChargesExtraDnsZones       float64          `json:"MonthlyChargesExtraDnsZones"`
	MonthlyChargesExtraVideoLibraries float64          `json:"MonthlyChargesExtraVideoLibraries"`
	MonthlyChargesScripting           float64          `json:"MonthlyChargesScripting"`
	MonthlyChargesScriptingRequests   float64          `json:"MonthlyChargesScriptingRequests"`
	MonthlyChargesScriptingCpu        float64          `json:"MonthlyChargesScriptingCpu"`
	MonthlyChargesDrm                 float64          `json:"MonthlyChargesDrm"`
	MonthlyChargesMagicContainers     float64          `json:"MonthlyChargesMagicContainers"`
	MonthlyChargesShield              float64          `json:"MonthlyChargesShield"`
	MonthlyChargesTaxes               float64          `json:"MonthlyChargesTaxes"`
	MonthlyChargesWebSockets          float64          `json:"MonthlyChargesWebSockets"`
	MonthlyChargesDB                  float64          `json:"MonthlyChargesDB"`
}

// BillingRecord represents a single billing record/transaction.
type BillingRecord struct {
	Id                          int64   `json:"Id"`
	PaymentId                   string  `json:"PaymentId"`
	Amount                      float64 `json:"Amount"`
	Payer                       string  `json:"Payer"`
	Timestamp                   string  `json:"Timestamp"`
	Type                        int     `json:"Type"`
	InvoiceAvailable            bool    `json:"InvoiceAvailable"`
	DocumentDownloadUrl         string  `json:"DocumentDownloadUrl"`
	DetailedDocumentDownloadUrl string  `json:"DetailedDocumentDownloadUrl"`
}

// BillingSummaryItem represents per-pullzone billing usage.
type BillingSummaryItem struct {
	PullZoneId           int64   `json:"PullZoneId"`
	MonthlyUsage         float64 `json:"MonthlyUsage"`
	MonthlyBandwidthUsed int64   `json:"MonthlyBandwidthUsed"`
}

// GetBillingDetails returns the full billing details for the account.
func (c *Client) GetBillingDetails(ctx context.Context) (*BillingDetails, error) {
	var details BillingDetails
	if err := c.Get(ctx, "/billing", &details); err != nil {
		return nil, err
	}
	return &details, nil
}

// GetBillingSummary returns per-pullzone billing summary.
func (c *Client) GetBillingSummary(ctx context.Context) ([]*BillingSummaryItem, error) {
	var items []*BillingSummaryItem
	if err := c.Get(ctx, "/billing/summary", &items); err != nil {
		return nil, err
	}
	return items, nil
}

// DownloadInvoice downloads a billing invoice PDF by record ID.
func (c *Client) DownloadInvoice(ctx context.Context, billingRecordId int64) ([]byte, error) {
	return c.DoRaw(ctx, "GET", fmt.Sprintf("/billing/summary/%d/pdf", billingRecordId), "", nil)
}
