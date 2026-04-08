package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage billing and invoices",
	}

	cmd.AddCommand(newBillingDetailsCmd())
	cmd.AddCommand(newBillingRecordsCmd())
	cmd.AddCommand(newBillingSummaryCmd())
	cmd.AddCommand(newBillingInvoiceCmd())

	return cmd
}

// --- billing details ---

func newBillingDetailsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "details",
		Short: "Get billing details and monthly charges",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewBillingAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			details, err := c.GetBillingDetails(cmd.Context())
			if err != nil {
				return err
			}

			columns := billingDetailsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, details)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

// --- billing records ---

func newBillingRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "records",
		Short: "List billing records/transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewBillingAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			details, err := c.GetBillingDetails(cmd.Context())
			if err != nil {
				return err
			}

			columns := billingRecordColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(details.BillingRecords))
			for i, r := range details.BillingRecords {
				items[i] = r
			}

			formatted, err := output.FormatList(cfg, columns, items, false)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

// --- billing summary ---

func newBillingSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get per-pullzone billing summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewBillingAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			items, err := c.GetBillingSummary(cmd.Context())
			if err != nil {
				return err
			}

			columns := billingSummaryColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			anyItems := make([]any, len(items))
			for i, s := range items {
				anyItems[i] = s
			}

			formatted, err := output.FormatList(cfg, columns, anyItems, false)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

// --- billing invoice ---

func newBillingInvoiceCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "invoice <billing-record-id>",
		Short: "Download a billing invoice PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid billing record ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewBillingAPI(cmd)
			if err != nil {
				return err
			}

			data, err := c.DownloadInvoice(cmd.Context(), id)
			if err != nil {
				return err
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("invoice-%d.pdf", id)
			}

			if err := os.WriteFile(outputFile, data, 0600); err != nil {
				return fmt.Errorf("writing invoice file: %w", err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Invoice saved to %s\n", outputFile)
			return err
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file path (default: invoice-<id>.pdf)")

	return cmd
}

// --- columns ---

func billingDetailsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.BillingDetails]{
		output.FloatColumn[*client.BillingDetails]("Balance", func(b *client.BillingDetails) float64 { return b.Balance }),
		output.FloatColumn[*client.BillingDetails]("This Month", func(b *client.BillingDetails) float64 { return b.ThisMonthCharges }),
		output.FloatColumn[*client.BillingDetails]("EU Traffic", func(b *client.BillingDetails) float64 { return b.MonthlyChargesEUTraffic }),
		output.FloatColumn[*client.BillingDetails]("US Traffic", func(b *client.BillingDetails) float64 { return b.MonthlyChargesUSTraffic }),
		output.FloatColumn[*client.BillingDetails]("Asia Traffic", func(b *client.BillingDetails) float64 { return b.MonthlyChargesASIATraffic }),
		output.FloatColumn[*client.BillingDetails]("Storage", func(b *client.BillingDetails) float64 { return b.MonthlyChargesStorage }),
		output.FloatColumn[*client.BillingDetails]("DNS", func(b *client.BillingDetails) float64 { return b.MonthlyChargesDNS }),
		output.FloatColumn[*client.BillingDetails]("Scripting", func(b *client.BillingDetails) float64 { return b.MonthlyChargesScripting }),
		output.FloatColumn[*client.BillingDetails]("Shield", func(b *client.BillingDetails) float64 { return b.MonthlyChargesShield }),
		output.FloatColumn[*client.BillingDetails]("Taxes", func(b *client.BillingDetails) float64 { return b.MonthlyChargesTaxes }),
	})
}

func billingRecordColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.BillingRecord]{
		output.IntColumn[*client.BillingRecord]("Id", func(r *client.BillingRecord) int { return int(r.Id) }),
		output.FloatColumn[*client.BillingRecord]("Amount", func(r *client.BillingRecord) float64 { return r.Amount }),
		output.StringColumn[*client.BillingRecord]("Type", func(r *client.BillingRecord) string {
			return client.BillingRecordTypeName(r.Type)
		}),
		output.StringColumn[*client.BillingRecord]("Timestamp", func(r *client.BillingRecord) string { return r.Timestamp }),
		output.StringColumn[*client.BillingRecord]("Payer", func(r *client.BillingRecord) string { return r.Payer }),
		output.BoolColumn[*client.BillingRecord]("Invoice", func(r *client.BillingRecord) bool { return r.InvoiceAvailable }),
	})
}

func billingSummaryColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.BillingSummaryItem]{
		output.IntColumn[*client.BillingSummaryItem]("Pull Zone Id", func(s *client.BillingSummaryItem) int { return int(s.PullZoneId) }),
		output.FloatColumn[*client.BillingSummaryItem]("Monthly Usage", func(s *client.BillingSummaryItem) float64 { return s.MonthlyUsage }),
		output.StringColumn[*client.BillingSummaryItem]("Bandwidth Used", func(s *client.BillingSummaryItem) string {
			return client.FormatBytes(s.MonthlyBandwidthUsed)
		}),
	})
}
