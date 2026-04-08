package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newPullZoneHostnamesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hostnames",
		Short: "Manage pull zone hostnames",
	}
	cmd.AddCommand(newPullZoneHostnamesListCmd())
	cmd.AddCommand(newPullZoneHostnamesAddCmd())
	cmd.AddCommand(newPullZoneHostnamesRemoveCmd())
	return cmd
}

func newPullZoneHostnamesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <pull_zone_id>",
		Short: "List hostnames for a pull zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			pz, err := c.GetPullZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := hostnameColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(pz.Hostnames))
			for i := range pz.Hostnames {
				items[i] = &pz.Hostnames[i]
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

func newPullZoneHostnamesAddCmd() *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "add <pull_zone_id>",
		Short: "Add a hostname to a pull zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.AddPullZoneHostname(cmd.Context(), id, hostname); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Hostname %q added to pull zone %d.\n", hostname, id)
			return err
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Hostname to add (required)")
	_ = cmd.MarkFlagRequired("hostname")

	return cmd
}

func newPullZoneHostnamesRemoveCmd() *cobra.Command {
	var (
		hostname string
		yes      bool
	)

	cmd := &cobra.Command{
		Use:   "remove <pull_zone_id>",
		Short: "Remove a hostname from a pull zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to remove hostname %q from pull zone %d? [y/N] ", hostname, id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Removal canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.RemovePullZoneHostname(cmd.Context(), id, hostname); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Hostname %q removed from pull zone %d.\n", hostname, id)
			return err
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Hostname to remove (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	_ = cmd.MarkFlagRequired("hostname")

	return cmd
}

func hostnameColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Hostname]{
		output.IntColumn[*client.Hostname]("Id", func(h *client.Hostname) int { return int(h.Id) }),
		output.StringColumn[*client.Hostname]("Hostname", func(h *client.Hostname) string { return h.Value }),
		output.BoolColumn[*client.Hostname]("Force SSL", func(h *client.Hostname) bool { return h.ForceSSL }),
		output.BoolColumn[*client.Hostname]("System", func(h *client.Hostname) bool { return h.IsSystemHostname }),
		output.BoolColumn[*client.Hostname]("Certificate", func(h *client.Hostname) bool { return h.HasCertificate }),
	})
}
