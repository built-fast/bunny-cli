package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newPullZonesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullzones",
		Aliases: []string{"pz"},
		Short:   "Manage pull zones",
	}
	cmd.AddCommand(newPullZonesListCmd())
	cmd.AddCommand(withWatch(newPullZonesGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newPullZonesCreateCmd())))
	cmd.AddCommand(withFromFile(newPullZonesUpdateCmd()))
	cmd.AddCommand(newPullZonesDeleteCmd())
	cmd.AddCommand(newPullZoneHostnamesCmd())
	cmd.AddCommand(newPullZonePurgeCmd())
	cmd.AddCommand(newPullZoneEdgeRulesCmd())
	return cmd
}

func newPullZonesListCmd() *cobra.Command {
	var (
		limit  int
		all    bool
		search string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull zones",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.PullZone], error) {
				return c.ListPullZones(cmd.Context(), page, perPage, search)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := pullZoneListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, pz := range result.Items {
				items[i] = pz
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

func newPullZonesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get pull zone details",
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

			columns := pullZoneDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, pz)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newPullZonesCreateCmd() *cobra.Command {
	var (
		name      string
		originUrl string
		pzType    int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull zone",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.PullZoneCreate{}

			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("origin-url") {
				body.OriginUrl = originUrl
			}
			if cmd.Flags().Changed("type") {
				body.Type = pzType
			}

			pz, err := c.CreatePullZone(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := pullZoneDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, pz)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Pull zone name (required)")
	cmd.Flags().StringVar(&originUrl, "origin-url", "", "Origin server URL")
	cmd.Flags().IntVar(&pzType, "type", 0, "Pull zone type (0=Premium, 1=Volume)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newPullZonesUpdateCmd() *cobra.Command {
	var (
		originUrl        string
		originHostHeader string
		addHostHeader    bool
		verifyOriginSSL  bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a pull zone",
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

			body := &client.PullZoneUpdate{}

			if cmd.Flags().Changed("origin-url") {
				body.OriginUrl = &originUrl
			}
			if cmd.Flags().Changed("origin-host-header") {
				body.OriginHostHeader = &originHostHeader
			}
			if cmd.Flags().Changed("add-host-header") {
				body.AddHostHeader = &addHostHeader
			}
			if cmd.Flags().Changed("verify-origin-ssl") {
				body.VerifyOriginSSL = &verifyOriginSSL
			}

			pz, err := c.UpdatePullZone(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := pullZoneDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, pz)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&originUrl, "origin-url", "", "Origin server URL")
	cmd.Flags().StringVar(&originHostHeader, "origin-host-header", "", "Custom host header for origin requests")
	cmd.Flags().BoolVar(&addHostHeader, "add-host-header", false, "Forward host header to origin")
	cmd.Flags().BoolVar(&verifyOriginSSL, "verify-origin-ssl", false, "Verify origin SSL certificate")

	return cmd
}

func newPullZonesDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a pull zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete pull zone %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewPullZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeletePullZone(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Pull zone deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// pullZoneListColumns defines the columns for pull zone list output.
func pullZoneListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.PullZone]{
		output.IntColumn[*client.PullZone]("Id", func(pz *client.PullZone) int { return int(pz.Id) }),
		output.StringColumn[*client.PullZone]("Name", func(pz *client.PullZone) string { return pz.Name }),
		output.StringColumn[*client.PullZone]("Origin URL", func(pz *client.PullZone) string { return pz.OriginUrl }),
		output.BoolColumn[*client.PullZone]("Enabled", func(pz *client.PullZone) bool { return pz.Enabled }),
		output.StringColumn[*client.PullZone]("Type", func(pz *client.PullZone) string { return client.PullZoneTypeName(pz.Type) }),
		output.IntColumn[*client.PullZone]("Bandwidth Used", func(pz *client.PullZone) int { return int(pz.MonthlyBandwidthUsed) }),
		output.FloatColumn[*client.PullZone]("Monthly Charges", func(pz *client.PullZone) float64 { return pz.MonthlyCharges }),
	})
}

// pullZoneDetailColumns defines the columns for pull zone detail output.
func pullZoneDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.PullZone]{
		output.IntColumn[*client.PullZone]("Id", func(pz *client.PullZone) int { return int(pz.Id) }),
		output.StringColumn[*client.PullZone]("Name", func(pz *client.PullZone) string { return pz.Name }),
		output.StringColumn[*client.PullZone]("Origin URL", func(pz *client.PullZone) string { return pz.OriginUrl }),
		output.BoolColumn[*client.PullZone]("Enabled", func(pz *client.PullZone) bool { return pz.Enabled }),
		output.BoolColumn[*client.PullZone]("Suspended", func(pz *client.PullZone) bool { return pz.Suspended }),
		output.StringColumn[*client.PullZone]("Type", func(pz *client.PullZone) string { return client.PullZoneTypeName(pz.Type) }),
		output.StringColumn[*client.PullZone]("CNAME Domain", func(pz *client.PullZone) string { return pz.CnameDomain }),
		output.StringColumn[*client.PullZone]("Origin Host Header", func(pz *client.PullZone) string { return pz.OriginHostHeader }),
		output.IntColumn[*client.PullZone]("Hostnames", func(pz *client.PullZone) int { return len(pz.Hostnames) }),
		output.IntColumn[*client.PullZone]("Edge Rules", func(pz *client.PullZone) int { return len(pz.EdgeRules) }),
		output.IntColumn[*client.PullZone]("Bandwidth Used", func(pz *client.PullZone) int { return int(pz.MonthlyBandwidthUsed) }),
		output.FloatColumn[*client.PullZone]("Monthly Charges", func(pz *client.PullZone) float64 { return pz.MonthlyCharges }),
		output.BoolColumn[*client.PullZone]("Verify Origin SSL", func(pz *client.PullZone) bool { return pz.VerifyOriginSSL }),
		output.BoolColumn[*client.PullZone]("Zone Security", func(pz *client.PullZone) bool { return pz.ZoneSecurityEnabled }),
	})
}
