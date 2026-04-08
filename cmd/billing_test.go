package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/spf13/cobra"
)

type mockBillingAPI struct {
	getBillingDetailsFn func(ctx context.Context) (*client.BillingDetails, error)
	getBillingSummaryFn func(ctx context.Context) ([]*client.BillingSummaryItem, error)
	downloadInvoiceFn   func(ctx context.Context, billingRecordId int64) ([]byte, error)
}

func (m *mockBillingAPI) GetBillingDetails(ctx context.Context) (*client.BillingDetails, error) {
	return m.getBillingDetailsFn(ctx)
}

func (m *mockBillingAPI) GetBillingSummary(ctx context.Context) ([]*client.BillingSummaryItem, error) {
	return m.getBillingSummaryFn(ctx)
}

func (m *mockBillingAPI) DownloadInvoice(ctx context.Context, billingRecordId int64) ([]byte, error) {
	return m.downloadInvoiceFn(ctx, billingRecordId)
}

func newTestBillingApp(api BillingAPI) *App {
	return &App{NewBillingAPI: func(_ *cobra.Command) (BillingAPI, error) { return api, nil }}
}

func sampleBillingDetails() *client.BillingDetails {
	return &client.BillingDetails{
		Balance:                 25.50,
		ThisMonthCharges:        12.75,
		MonthlyChargesEUTraffic: 5.00,
		MonthlyChargesUSTraffic: 3.50,
		MonthlyChargesStorage:   2.25,
		MonthlyChargesDNS:       1.00,
		MonthlyChargesTaxes:     1.00,
		BillingRecords: []*client.BillingRecord{
			{Id: 100, Amount: 10.00, Type: 2, Timestamp: "2024-01-01T00:00:00Z", Payer: "card-1", InvoiceAvailable: true},
			{Id: 101, Amount: 15.50, Type: 3, Timestamp: "2024-02-01T00:00:00Z", Payer: "", InvoiceAvailable: false},
		},
	}
}

func sampleBillingSummary() []*client.BillingSummaryItem {
	return []*client.BillingSummaryItem{
		{PullZoneId: 1, MonthlyUsage: 5.50, MonthlyBandwidthUsed: 1073741824},
		{PullZoneId: 2, MonthlyUsage: 3.25, MonthlyBandwidthUsed: 536870912},
	}
}

// --- billing help ---

func TestBilling_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "billing", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"details", "records", "summary", "invoice"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected billing help to show %q subcommand", sub)
		}
	}
}

// --- billing details ---

func TestBillingDetails_Table(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingDetailsFn: func(_ context.Context) (*client.BillingDetails, error) {
			return sampleBillingDetails(), nil
		},
	}
	app := newTestBillingApp(mock)

	out, _, err := executeCommand(app, "billing", "details")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "25.50") {
		t.Error("expected output to contain balance")
	}
	if !strings.Contains(out, "12.75") {
		t.Error("expected output to contain this month charges")
	}
}

func TestBillingDetails_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingDetailsFn: func(_ context.Context) (*client.BillingDetails, error) {
			return sampleBillingDetails(), nil
		},
	}
	app := newTestBillingApp(mock)

	out, _, err := executeCommand(app, "billing", "details", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["Balance"] != 25.5 {
		t.Errorf("expected Balance=25.5, got %v", result["Balance"])
	}
}

func TestBillingDetails_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingDetailsFn: func(_ context.Context) (*client.BillingDetails, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestBillingApp(mock)

	_, stderr, err := executeCommand(app, "billing", "details")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- billing records ---

func TestBillingRecords_Table(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingDetailsFn: func(_ context.Context) (*client.BillingDetails, error) {
			return sampleBillingDetails(), nil
		},
	}
	app := newTestBillingApp(mock)

	out, _, err := executeCommand(app, "billing", "records")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "100") {
		t.Error("expected output to contain billing record ID")
	}
	if !strings.Contains(out, "CreditCard") {
		t.Error("expected output to contain billing record type")
	}
}

func TestBillingRecords_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingDetailsFn: func(_ context.Context) (*client.BillingDetails, error) {
			return sampleBillingDetails(), nil
		},
	}
	app := newTestBillingApp(mock)

	out, _, err := executeCommand(app, "billing", "records", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["object"] != "list" {
		t.Errorf("expected object=list, got %v", result["object"])
	}
}

// --- billing summary ---

func TestBillingSummary_Table(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingSummaryFn: func(_ context.Context) ([]*client.BillingSummaryItem, error) {
			return sampleBillingSummary(), nil
		},
	}
	app := newTestBillingApp(mock)

	out, _, err := executeCommand(app, "billing", "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "5.50") {
		t.Error("expected output to contain monthly usage")
	}
	if !strings.Contains(out, "1.00 GiB") {
		t.Error("expected output to contain formatted bandwidth")
	}
}

func TestBillingSummary_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingSummaryFn: func(_ context.Context) ([]*client.BillingSummaryItem, error) {
			return sampleBillingSummary(), nil
		},
	}
	app := newTestBillingApp(mock)

	out, _, err := executeCommand(app, "billing", "summary", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["object"] != "list" {
		t.Errorf("expected object=list, got %v", result["object"])
	}
}

func TestBillingSummary_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		getBillingSummaryFn: func(_ context.Context) ([]*client.BillingSummaryItem, error) {
			return nil, fmt.Errorf("API unavailable")
		},
	}
	app := newTestBillingApp(mock)

	_, stderr, err := executeCommand(app, "billing", "summary")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- billing invoice ---

func TestBillingInvoice_Download(t *testing.T) {
	t.Parallel()
	pdfData := []byte("%PDF-1.4 test content")
	mock := &mockBillingAPI{
		downloadInvoiceFn: func(_ context.Context, id int64) ([]byte, error) {
			if id != 100 {
				t.Errorf("expected id=100, got %d", id)
			}
			return pdfData, nil
		},
	}
	app := newTestBillingApp(mock)

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test-invoice.pdf")

	out, _, err := executeCommand(app, "billing", "invoice", "100", "-o", outFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, outFile) {
		t.Errorf("expected output to mention file path, got %s", out)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != string(pdfData) {
		t.Error("output file contents don't match expected PDF data")
	}
}

func TestBillingInvoice_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "billing", "invoice")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestBillingInvoice_InvalidID_Fails(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		downloadInvoiceFn: func(_ context.Context, id int64) ([]byte, error) {
			return nil, nil
		},
	}
	app := newTestBillingApp(mock)

	_, stderr, err := executeCommand(app, "billing", "invoice", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "invalid billing record ID") {
		t.Errorf("expected invalid ID error, got %q", stderr)
	}
}

func TestBillingInvoice_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockBillingAPI{
		downloadInvoiceFn: func(_ context.Context, id int64) ([]byte, error) {
			return nil, fmt.Errorf("invoice not found")
		},
	}
	app := newTestBillingApp(mock)

	_, stderr, err := executeCommand(app, "billing", "invoice", "999")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "invoice not found") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}
