package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newDnsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dns",
		Aliases: []string{"dnszone"},
		Short:   "Manage DNS zones",
	}
	cmd.AddCommand(newDnsListCmd())
	cmd.AddCommand(withWatch(newDnsGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newDnsCreateCmd())))
	cmd.AddCommand(withFromFile(newDnsUpdateCmd()))
	cmd.AddCommand(newDnsDeleteCmd())
	cmd.AddCommand(newDnsImportCmd())
	cmd.AddCommand(newDnsExportCmd())
	cmd.AddCommand(newDnsRecordsCmd())
	cmd.AddCommand(newDnsDnssecCmd())
	return cmd
}

func newDnsListCmd() *cobra.Command {
	var (
		limit  int
		all    bool
		search string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS zones",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.DnsZone], error) {
				return c.ListDnsZones(cmd.Context(), page, perPage, search)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := dnsZoneListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, z := range result.Items {
				items[i] = z
			}

			formatted, err := output.FormatList(cfg, columns, items, result.HasMore)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results to return (default 20)")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages of results")
	cmd.Flags().StringVar(&search, "search", "", "Filter results by search term")

	return cmd
}

func newDnsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get DNS zone details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			zone, err := c.GetDnsZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := dnsZoneDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, zone)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newDnsCreateCmd() *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a DNS zone",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.DnsZoneCreate{}

			if cmd.Flags().Changed("domain") {
				body.Domain = domain
			}

			zone, err := c.CreateDnsZone(cmd.Context(), body)
			if err != nil {
				return err
			}

			// API may return 201 with empty body
			if zone.Id == 0 {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "DNS zone created.")
				return err
			}

			columns := dnsZoneDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, zone)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Domain name (required)")
	_ = cmd.MarkFlagRequired("domain")

	return cmd
}

func newDnsUpdateCmd() *cobra.Command {
	var (
		customNameserversEnabled      bool
		nameserver1                   string
		nameserver2                   string
		soaEmail                      string
		loggingEnabled                bool
		logAnonymizationType          int
		loggingIPAnonymizationEnabled bool
		certificateKeyType            int
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a DNS zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.DnsZoneUpdate{}

			if cmd.Flags().Changed("custom-nameservers-enabled") {
				body.CustomNameserversEnabled = &customNameserversEnabled
			}
			if cmd.Flags().Changed("nameserver1") {
				body.Nameserver1 = &nameserver1
			}
			if cmd.Flags().Changed("nameserver2") {
				body.Nameserver2 = &nameserver2
			}
			if cmd.Flags().Changed("soa-email") {
				body.SoaEmail = &soaEmail
			}
			if cmd.Flags().Changed("logging-enabled") {
				body.LoggingEnabled = &loggingEnabled
			}
			if cmd.Flags().Changed("log-anonymization-type") {
				body.LogAnonymizationType = &logAnonymizationType
			}
			if cmd.Flags().Changed("logging-ip-anonymization-enabled") {
				body.LoggingIPAnonymizationEnabled = &loggingIPAnonymizationEnabled
			}
			if cmd.Flags().Changed("certificate-key-type") {
				body.CertificateKeyType = &certificateKeyType
			}

			zone, err := c.UpdateDnsZone(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := dnsZoneDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, zone)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().BoolVar(&customNameserversEnabled, "custom-nameservers-enabled", false, "Enable custom nameservers")
	cmd.Flags().StringVar(&nameserver1, "nameserver1", "", "Primary nameserver")
	cmd.Flags().StringVar(&nameserver2, "nameserver2", "", "Secondary nameserver")
	cmd.Flags().StringVar(&soaEmail, "soa-email", "", "SOA email address")
	cmd.Flags().BoolVar(&loggingEnabled, "logging-enabled", false, "Enable query logging")
	cmd.Flags().IntVar(&logAnonymizationType, "log-anonymization-type", 0, "Log anonymization type (0=OneDigit, 1=Drop)")
	cmd.Flags().BoolVar(&loggingIPAnonymizationEnabled, "logging-ip-anonymization-enabled", false, "Enable IP anonymization in logs")
	cmd.Flags().IntVar(&certificateKeyType, "certificate-key-type", 0, "Certificate key type (0=ECDSA, 1=RSA)")

	return cmd
}

func newDnsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a DNS zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete DNS zone %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteDnsZone(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "DNS zone deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func newDnsImportCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "import <zone_id>",
		Short: "Import DNS records from a zone file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			var reader io.Reader
			if file == "-" {
				reader = cmd.InOrStdin()
			} else {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("opening zone file: %w", err)
				}
				defer func() { _ = f.Close() }()
				reader = f
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			result, err := c.ImportDnsZone(cmd.Context(), id, reader)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Import complete: %d successful, %d failed, %d skipped.\n",
				result.RecordsSuccessful, result.RecordsFailed, result.RecordsSkipped)
			return err
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to zone file (use '-' for stdin) (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newDnsExportCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export <zone_id>",
		Short: "Export DNS zone as a zone file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			data, err := c.ExportDnsZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					return fmt.Errorf("writing zone file: %w", err)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Zone file written to %s.\n", outputFile)
				return err
			}

			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}

	cmd.Flags().StringVar(&outputFile, "output-file", "", "Write zone file to path instead of stdout")

	return cmd
}

// --- DNSSEC subcommands ---

func newDnsDnssecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dnssec",
		Short: "Manage DNSSEC for a DNS zone",
	}
	cmd.AddCommand(newDnsDnssecEnableCmd())
	cmd.AddCommand(newDnsDnssecDisableCmd())
	return cmd
}

func newDnsDnssecEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <zone_id>",
		Short: "Enable DNSSEC for a DNS zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			info, err := c.EnableDnsSec(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := dnsSecInfoColumns()

			formatted, err := output.FormatOne(cfg, columns, info)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newDnsDnssecDisableCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "disable <zone_id>",
		Short: "Disable DNSSEC for a DNS zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to disable DNSSEC for zone %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Operation canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			if _, err := c.DisableDnsSec(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "DNSSEC disabled.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Column definitions ---

func dnsZoneListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.DnsZone]{
		output.IntColumn[*client.DnsZone]("Id", func(z *client.DnsZone) int { return int(z.Id) }),
		output.StringColumn[*client.DnsZone]("Domain", func(z *client.DnsZone) string { return z.Domain }),
		output.IntColumn[*client.DnsZone]("Records", func(z *client.DnsZone) int { return len(z.Records) }),
		output.BoolColumn[*client.DnsZone]("DNSSEC", func(z *client.DnsZone) bool { return z.DnsSecEnabled }),
		output.BoolColumn[*client.DnsZone]("NS Detected", func(z *client.DnsZone) bool { return z.NameserversDetected }),
		output.StringColumn[*client.DnsZone]("Date Modified", func(z *client.DnsZone) string { return z.DateModified }),
	})
}

func dnsZoneDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.DnsZone]{
		output.IntColumn[*client.DnsZone]("Id", func(z *client.DnsZone) int { return int(z.Id) }),
		output.StringColumn[*client.DnsZone]("Domain", func(z *client.DnsZone) string { return z.Domain }),
		output.IntColumn[*client.DnsZone]("Records", func(z *client.DnsZone) int { return len(z.Records) }),
		output.StringColumn[*client.DnsZone]("Date Created", func(z *client.DnsZone) string { return z.DateCreated }),
		output.StringColumn[*client.DnsZone]("Date Modified", func(z *client.DnsZone) string { return z.DateModified }),
		output.BoolColumn[*client.DnsZone]("NS Detected", func(z *client.DnsZone) bool { return z.NameserversDetected }),
		output.BoolColumn[*client.DnsZone]("Custom NS Enabled", func(z *client.DnsZone) bool { return z.CustomNameserversEnabled }),
		output.StringColumn[*client.DnsZone]("Nameserver 1", func(z *client.DnsZone) string { return z.Nameserver1 }),
		output.StringColumn[*client.DnsZone]("Nameserver 2", func(z *client.DnsZone) string { return z.Nameserver2 }),
		output.StringColumn[*client.DnsZone]("SOA Email", func(z *client.DnsZone) string { return z.SoaEmail }),
		output.BoolColumn[*client.DnsZone]("Logging Enabled", func(z *client.DnsZone) bool { return z.LoggingEnabled }),
		output.BoolColumn[*client.DnsZone]("IP Anonymization", func(z *client.DnsZone) bool { return z.LoggingIPAnonymizationEnabled }),
		output.StringColumn[*client.DnsZone]("Log Anonymization", func(z *client.DnsZone) string {
			return client.LogAnonymizationTypeName(z.LogAnonymizationType)
		}),
		output.BoolColumn[*client.DnsZone]("DNSSEC", func(z *client.DnsZone) bool { return z.DnsSecEnabled }),
		output.StringColumn[*client.DnsZone]("Certificate Key Type", func(z *client.DnsZone) string {
			return client.CertificateKeyTypeName(z.CertificateKeyType)
		}),
	})
}

func dnsSecInfoColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.DnsSecInfo]{
		output.BoolColumn[*client.DnsSecInfo]("Enabled", func(i *client.DnsSecInfo) bool { return i.Enabled }),
		output.StringColumn[*client.DnsSecInfo]("DS Record", func(i *client.DnsSecInfo) string { return i.DsRecord }),
		output.StringColumn[*client.DnsSecInfo]("Digest", func(i *client.DnsSecInfo) string { return i.Digest }),
		output.StringColumn[*client.DnsSecInfo]("Digest Type", func(i *client.DnsSecInfo) string { return i.DigestType }),
		output.IntColumn[*client.DnsSecInfo]("Algorithm", func(i *client.DnsSecInfo) int { return i.Algorithm }),
		output.StringColumn[*client.DnsSecInfo]("Public Key", func(i *client.DnsSecInfo) string { return i.PublicKey }),
		output.IntColumn[*client.DnsSecInfo]("Key Tag", func(i *client.DnsSecInfo) int { return i.KeyTag }),
		output.IntColumn[*client.DnsSecInfo]("Flags", func(i *client.DnsSecInfo) int { return i.Flags }),
		output.BoolColumn[*client.DnsSecInfo]("DS Configured", func(i *client.DnsSecInfo) bool { return i.DsConfigured }),
	})
}
