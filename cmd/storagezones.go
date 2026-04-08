package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newStorageZonesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "storagezones",
		Aliases: []string{"sz"},
		Short:   "Manage storage zones",
	}
	cmd.AddCommand(newStorageZonesListCmd())
	cmd.AddCommand(withWatch(newStorageZonesGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newStorageZonesCreateCmd())))
	cmd.AddCommand(withFromFile(newStorageZonesUpdateCmd()))
	cmd.AddCommand(newStorageZonesDeleteCmd())
	cmd.AddCommand(newStorageZonesResetPasswordCmd())
	return cmd
}

func newStorageZonesListCmd() *cobra.Command {
	var (
		limit          int
		all            bool
		search         string
		includeDeleted bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List storage zones",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.StorageZone], error) {
				return c.ListStorageZones(cmd.Context(), page, perPage, search, includeDeleted)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := storageZoneListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, sz := range result.Items {
				items[i] = sz
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
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "Include deleted storage zones")

	return cmd
}

func newStorageZonesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get storage zone details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid storage zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			sz, err := c.GetStorageZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := storageZoneDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, sz)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newStorageZonesCreateCmd() *cobra.Command {
	var (
		name               string
		region             string
		replicationRegions []string
		zoneTier           int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a storage zone",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.StorageZoneCreate{
				Name:   name,
				Region: region,
			}

			if cmd.Flags().Changed("replication-regions") {
				body.ReplicationRegions = replicationRegions
			}
			if cmd.Flags().Changed("zone-tier") {
				body.ZoneTier = zoneTier
			}

			sz, err := c.CreateStorageZone(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := storageZoneDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, sz)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Storage zone name (required)")
	cmd.Flags().StringVar(&region, "region", "", "Primary storage region: DE, NY, LA, SG (required)")
	cmd.Flags().StringSliceVar(&replicationRegions, "replication-regions", nil, "Additional replication regions (comma-separated): DE, NY, LA, SG, SYD")
	cmd.Flags().IntVar(&zoneTier, "zone-tier", 0, "Storage zone tier (0=Standard, 1=Edge)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("region")
	setFlagOptions(cmd, "region", []string{"DE", "NY", "LA", "SG"})

	return cmd
}

func newStorageZonesUpdateCmd() *cobra.Command {
	var (
		originUrl         string
		custom404FilePath string
		rewrite404To200   bool
		replicationZones  []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a storage zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid storage zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.StorageZoneUpdate{}

			if cmd.Flags().Changed("origin-url") {
				body.OriginUrl = &originUrl
			}
			if cmd.Flags().Changed("custom-404-file-path") {
				body.Custom404FilePath = &custom404FilePath
			}
			if cmd.Flags().Changed("rewrite-404-to-200") {
				body.Rewrite404To200 = &rewrite404To200
			}
			if cmd.Flags().Changed("replication-zones") {
				body.ReplicationZones = replicationZones
			}

			if err := c.UpdateStorageZone(cmd.Context(), id, body); err != nil {
				return err
			}

			// Update returns 204, fetch the updated zone to display
			sz, err := c.GetStorageZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := storageZoneDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, sz)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&originUrl, "origin-url", "", "Origin URL")
	cmd.Flags().StringVar(&custom404FilePath, "custom-404-file-path", "", "Custom 404 error file path")
	cmd.Flags().BoolVar(&rewrite404To200, "rewrite-404-to-200", false, "Rewrite 404 responses to 200")
	cmd.Flags().StringSliceVar(&replicationZones, "replication-zones", nil, "Replication zones (comma-separated)")

	return cmd
}

func newStorageZonesDeleteCmd() *cobra.Command {
	var (
		yes                   bool
		deleteLinkedPullZones bool
	)

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a storage zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid storage zone ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete storage zone %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteStorageZone(cmd.Context(), id, deleteLinkedPullZones); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Storage zone deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&deleteLinkedPullZones, "delete-linked-pull-zones", true, "Delete linked pull zones")

	return cmd
}

func newStorageZonesResetPasswordCmd() *cobra.Command {
	var (
		yes      bool
		readOnly bool
	)

	cmd := &cobra.Command{
		Use:   "reset-password <id>",
		Short: "Reset storage zone password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid storage zone ID: %w", err)
			}

			passwordType := "password"
			if readOnly {
				passwordType = "read-only password"
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to reset the %s for storage zone %d? [y/N] ", passwordType, id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Password reset canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
			if err != nil {
				return err
			}

			if readOnly {
				if err := c.ResetStorageZoneReadOnlyPassword(cmd.Context(), id); err != nil {
					return err
				}
			} else {
				if err := c.ResetStorageZonePassword(cmd.Context(), id); err != nil {
					return err
				}
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Storage zone %s reset.\n", passwordType)
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Reset the read-only password instead")

	return cmd
}

// storageZoneListColumns defines the columns for storage zone list output.
func storageZoneListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.StorageZone]{
		output.IntColumn[*client.StorageZone]("Id", func(sz *client.StorageZone) int { return int(sz.Id) }),
		output.StringColumn[*client.StorageZone]("Name", func(sz *client.StorageZone) string { return sz.Name }),
		output.StringColumn[*client.StorageZone]("Region", func(sz *client.StorageZone) string { return sz.Region }),
		output.StringColumn[*client.StorageZone]("Hostname", func(sz *client.StorageZone) string { return sz.StorageHostname }),
		output.IntColumn[*client.StorageZone]("Storage Used", func(sz *client.StorageZone) int { return int(sz.StorageUsed) }),
		output.IntColumn[*client.StorageZone]("Files Stored", func(sz *client.StorageZone) int { return int(sz.FilesStored) }),
		output.StringColumn[*client.StorageZone]("Tier", func(sz *client.StorageZone) string { return client.StorageZoneTierName(sz.ZoneTier) }),
	})
}

// storageZoneDetailColumns defines the columns for storage zone detail output.
func storageZoneDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.StorageZone]{
		output.IntColumn[*client.StorageZone]("Id", func(sz *client.StorageZone) int { return int(sz.Id) }),
		output.StringColumn[*client.StorageZone]("Name", func(sz *client.StorageZone) string { return sz.Name }),
		output.StringColumn[*client.StorageZone]("Region", func(sz *client.StorageZone) string { return sz.Region }),
		output.StringColumn[*client.StorageZone]("Hostname", func(sz *client.StorageZone) string { return sz.StorageHostname }),
		output.StringColumn[*client.StorageZone]("Password", func(sz *client.StorageZone) string { return sz.Password }),
		output.StringColumn[*client.StorageZone]("Read-Only Password", func(sz *client.StorageZone) string { return sz.ReadOnlyPassword }),
		output.IntColumn[*client.StorageZone]("Storage Used", func(sz *client.StorageZone) int { return int(sz.StorageUsed) }),
		output.IntColumn[*client.StorageZone]("Files Stored", func(sz *client.StorageZone) int { return int(sz.FilesStored) }),
		output.StringColumn[*client.StorageZone]("Tier", func(sz *client.StorageZone) string { return client.StorageZoneTierName(sz.ZoneTier) }),
		output.StringColumn[*client.StorageZone]("Replication Regions", func(sz *client.StorageZone) string {
			return strings.Join(sz.ReplicationRegions, ", ")
		}),
		output.IntColumn[*client.StorageZone]("Pull Zones", func(sz *client.StorageZone) int { return len(sz.PullZones) }),
		output.StringColumn[*client.StorageZone]("Custom 404 Path", func(sz *client.StorageZone) string { return sz.Custom404FilePath }),
		output.BoolColumn[*client.StorageZone]("Rewrite 404 to 200", func(sz *client.StorageZone) bool { return sz.Rewrite404To200 }),
		output.BoolColumn[*client.StorageZone]("Deleted", func(sz *client.StorageZone) bool { return sz.Deleted }),
		output.StringColumn[*client.StorageZone]("Date Modified", func(sz *client.StorageZone) string { return sz.DateModified }),
	})
}
