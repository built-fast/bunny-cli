package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newDnsRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "records",
		Short: "Manage DNS records",
	}
	cmd.AddCommand(newDnsRecordsListCmd())
	cmd.AddCommand(withFromFile(withInteractive(newDnsRecordsAddCmd())))
	cmd.AddCommand(withFromFile(newDnsRecordsUpdateCmd()))
	cmd.AddCommand(newDnsRecordsDeleteCmd())
	return cmd
}

func newDnsRecordsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <zone_id>",
		Short: "List DNS records for a zone",
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

			columns := dnsRecordListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(zone.Records))
			for i := range zone.Records {
				items[i] = &zone.Records[i]
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

func newDnsRecordsAddCmd() *cobra.Command {
	var (
		recordType string
		value      string
		name       string
		ttl        int
		weight     int
		priority   int
		port       int
		comment    string
		disabled   bool
	)

	cmd := &cobra.Command{
		Use:   "add <zone_id>",
		Short: "Add a DNS record to a zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			zoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			typeInt, err := client.DnsRecordTypeFromName(recordType)
			if err != nil {
				return err
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.DnsRecordCreate{
				Type:  typeInt,
				Value: value,
			}

			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("ttl") {
				body.Ttl = ttl
			}
			if cmd.Flags().Changed("weight") {
				body.Weight = weight
			}
			if cmd.Flags().Changed("priority") {
				body.Priority = priority
			}
			if cmd.Flags().Changed("port") {
				body.Port = port
			}
			if cmd.Flags().Changed("comment") {
				body.Comment = comment
			}
			if cmd.Flags().Changed("disabled") {
				body.Disabled = disabled
			}

			record, err := c.AddDnsRecord(cmd.Context(), zoneId, body)
			if err != nil {
				return err
			}

			columns := dnsRecordDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, record)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&recordType, "type", "", "Record type: A, AAAA, CNAME, TXT, MX, SRV, CAA, NS, PTR, Redirect, Flatten, PullZone, Script (required)")
	cmd.Flags().StringVar(&value, "value", "", "Record value (required)")
	cmd.Flags().StringVar(&name, "name", "", "Record name (subdomain)")
	cmd.Flags().IntVar(&ttl, "ttl", 0, "Time to live in seconds")
	cmd.Flags().IntVar(&weight, "weight", 0, "Record weight")
	cmd.Flags().IntVar(&priority, "priority", 0, "Record priority (MX, SRV)")
	cmd.Flags().IntVar(&port, "port", 0, "Port number (SRV)")
	cmd.Flags().StringVar(&comment, "comment", "", "Record comment")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Create record in disabled state")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("value")
	setFlagOptions(cmd, "type", []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV", "CAA", "NS", "PTR", "Redirect", "Flatten", "PullZone", "Script"})

	return cmd
}

func newDnsRecordsUpdateCmd() *cobra.Command {
	var (
		recordType string
		value      string
		name       string
		ttl        int
		weight     int
		priority   int
		port       int
		comment    string
		disabled   bool
	)

	cmd := &cobra.Command{
		Use:   "update <zone_id> <record_id>",
		Short: "Update a DNS record",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			zoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			recordId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS record ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewDnsZoneAPI(cmd)
			if err != nil {
				return err
			}

			body := &client.DnsRecordUpdate{}

			if cmd.Flags().Changed("type") {
				typeInt, err := client.DnsRecordTypeFromName(recordType)
				if err != nil {
					return err
				}
				body.Type = &typeInt
			}
			if cmd.Flags().Changed("value") {
				body.Value = &value
			}
			if cmd.Flags().Changed("name") {
				body.Name = &name
			}
			if cmd.Flags().Changed("ttl") {
				body.Ttl = &ttl
			}
			if cmd.Flags().Changed("weight") {
				body.Weight = &weight
			}
			if cmd.Flags().Changed("priority") {
				body.Priority = &priority
			}
			if cmd.Flags().Changed("port") {
				body.Port = &port
			}
			if cmd.Flags().Changed("comment") {
				body.Comment = &comment
			}
			if cmd.Flags().Changed("disabled") {
				body.Disabled = &disabled
			}

			if err := c.UpdateDnsRecord(cmd.Context(), zoneId, recordId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "DNS record updated.")
			return err
		},
	}

	cmd.Flags().StringVar(&recordType, "type", "", "Record type: A, AAAA, CNAME, TXT, MX, SRV, CAA, NS, PTR, Redirect, Flatten, PullZone, Script")
	cmd.Flags().StringVar(&value, "value", "", "Record value")
	cmd.Flags().StringVar(&name, "name", "", "Record name (subdomain)")
	cmd.Flags().IntVar(&ttl, "ttl", 0, "Time to live in seconds")
	cmd.Flags().IntVar(&weight, "weight", 0, "Record weight")
	cmd.Flags().IntVar(&priority, "priority", 0, "Record priority (MX, SRV)")
	cmd.Flags().IntVar(&port, "port", 0, "Port number (SRV)")
	cmd.Flags().StringVar(&comment, "comment", "", "Record comment")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Disable the record")

	return cmd
}

func newDnsRecordsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <zone_id> <record_id>",
		Short: "Delete a DNS record",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			zoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS zone ID: %w", err)
			}

			recordId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid DNS record ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete DNS record %d from zone %d? [y/N] ", recordId, zoneId))
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

			if err := c.DeleteDnsRecord(cmd.Context(), zoneId, recordId); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "DNS record deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Column definitions ---

func dnsRecordListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.DnsRecord]{
		output.IntColumn[*client.DnsRecord]("Id", func(r *client.DnsRecord) int { return int(r.Id) }),
		output.StringColumn[*client.DnsRecord]("Type", func(r *client.DnsRecord) string { return client.DnsRecordTypeName(r.Type) }),
		output.StringColumn[*client.DnsRecord]("Name", func(r *client.DnsRecord) string { return r.Name }),
		output.StringColumn[*client.DnsRecord]("Value", func(r *client.DnsRecord) string { return r.Value }),
		output.IntColumn[*client.DnsRecord]("TTL", func(r *client.DnsRecord) int { return r.Ttl }),
		output.BoolColumn[*client.DnsRecord]("Disabled", func(r *client.DnsRecord) bool { return r.Disabled }),
	})
}

func dnsRecordDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.DnsRecord]{
		output.IntColumn[*client.DnsRecord]("Id", func(r *client.DnsRecord) int { return int(r.Id) }),
		output.StringColumn[*client.DnsRecord]("Type", func(r *client.DnsRecord) string { return client.DnsRecordTypeName(r.Type) }),
		output.StringColumn[*client.DnsRecord]("Name", func(r *client.DnsRecord) string { return r.Name }),
		output.StringColumn[*client.DnsRecord]("Value", func(r *client.DnsRecord) string { return r.Value }),
		output.IntColumn[*client.DnsRecord]("TTL", func(r *client.DnsRecord) int { return r.Ttl }),
		output.IntColumn[*client.DnsRecord]("Weight", func(r *client.DnsRecord) int { return r.Weight }),
		output.IntColumn[*client.DnsRecord]("Priority", func(r *client.DnsRecord) int { return r.Priority }),
		output.IntColumn[*client.DnsRecord]("Port", func(r *client.DnsRecord) int { return r.Port }),
		output.StringColumn[*client.DnsRecord]("Comment", func(r *client.DnsRecord) string { return r.Comment }),
		output.BoolColumn[*client.DnsRecord]("Accelerated", func(r *client.DnsRecord) bool { return r.Accelerated }),
		output.BoolColumn[*client.DnsRecord]("Disabled", func(r *client.DnsRecord) bool { return r.Disabled }),
	})
}
