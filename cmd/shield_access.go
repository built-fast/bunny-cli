package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newShieldAccessListsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "access-lists",
		Aliases: []string{"access-list"},
		Short:   "Manage Shield access lists",
	}
	cmd.AddCommand(newShieldAccessListsListCmd())
	cmd.AddCommand(withWatch(newShieldAccessListsGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newShieldAccessListsCreateCmd())))
	cmd.AddCommand(withFromFile(newShieldAccessListsUpdateCmd()))
	cmd.AddCommand(newShieldAccessListsDeleteCmd())
	cmd.AddCommand(newShieldAccessListsConfigCmd())
	return cmd
}

func newShieldAccessListsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <shield_zone_id>",
		Short: "List access lists for a Shield zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			resp, err := c.ListAccessLists(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			// Combine managed and custom lists for display
			allLists := make([]client.AccessListDetails, 0, len(resp.ManagedLists)+len(resp.CustomLists))
			allLists = append(allLists, resp.ManagedLists...)
			allLists = append(allLists, resp.CustomLists...)

			columns := accessListDetailsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(allLists))
			for i := range allLists {
				items[i] = &allLists[i]
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

func newShieldAccessListsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <shield_zone_id> <id>",
		Short: "Get custom access list details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}
			id, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid access list ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			list, err := c.GetCustomAccessList(cmd.Context(), shieldZoneId, id)
			if err != nil {
				return err
			}

			columns := customAccessListDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, list)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldAccessListsCreateCmd() *cobra.Command {
	var (
		name        string
		description string
		listType    int
		content     string
	)

	cmd := &cobra.Command{
		Use:   "create <shield_zone_id>",
		Short: "Create a custom access list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.CustomAccessListCreate{}
			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("description") {
				body.Description = description
			}
			if cmd.Flags().Changed("type") {
				body.Type = listType
			}
			if cmd.Flags().Changed("content") {
				body.Content = content
			}

			list, err := c.CreateCustomAccessList(cmd.Context(), shieldZoneId, body)
			if err != nil {
				return err
			}

			columns := customAccessListDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, list)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Access list name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Access list description")
	cmd.Flags().IntVar(&listType, "type", 0, "Access list type")
	cmd.Flags().StringVar(&content, "content", "", "Access list content (entries separated by newlines)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func newShieldAccessListsUpdateCmd() *cobra.Command {
	var (
		name    string
		content string
	)

	cmd := &cobra.Command{
		Use:   "update <shield_zone_id> <id>",
		Short: "Update a custom access list",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}
			id, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid access list ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.CustomAccessListUpdate{}
			if cmd.Flags().Changed("name") {
				body.Name = &name
			}
			if cmd.Flags().Changed("content") {
				body.Content = &content
			}

			list, err := c.UpdateCustomAccessList(cmd.Context(), shieldZoneId, id, body)
			if err != nil {
				return err
			}

			columns := customAccessListDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, list)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Access list name")
	cmd.Flags().StringVar(&content, "content", "", "Access list content (entries separated by newlines)")

	return cmd
}

func newShieldAccessListsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <shield_zone_id> <id>",
		Short: "Delete a custom access list",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}
			id, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid access list ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete access list %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteCustomAccessList(cmd.Context(), shieldZoneId, id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Access list deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Access List Config ---

func newShieldAccessListsConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage access list configurations",
	}
	cmd.AddCommand(newShieldAccessListsConfigUpdateCmd())
	return cmd
}

func newShieldAccessListsConfigUpdateCmd() *cobra.Command {
	var (
		isEnabled bool
		action    int
	)

	cmd := &cobra.Command{
		Use:   "update <shield_zone_id> <config_id>",
		Short: "Update an access list configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}
			configId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid config ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			body := &client.AccessListConfigUpdate{}
			if cmd.Flags().Changed("enabled") {
				body.IsEnabled = &isEnabled
			}
			if cmd.Flags().Changed("action") {
				body.Action = &action
			}

			if err := c.UpdateAccessListConfig(cmd.Context(), shieldZoneId, configId, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Access list configuration updated.")
			return err
		},
	}

	cmd.Flags().BoolVar(&isEnabled, "enabled", false, "Enable or disable the access list")
	cmd.Flags().IntVar(&action, "action", 0, "Access list action (0=Allow, 1=Block, 2=Log, 3=Challenge)")

	return cmd
}

// --- Column definitions ---

func accessListDetailsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.AccessListDetails]{
		output.IntColumn[*client.AccessListDetails]("List Id", func(l *client.AccessListDetails) int { return int(l.ListId) }),
		output.IntColumn[*client.AccessListDetails]("Config Id", func(l *client.AccessListDetails) int { return int(l.ConfigurationId) }),
		output.StringColumn[*client.AccessListDetails]("Name", func(l *client.AccessListDetails) string { return l.Name }),
		output.BoolColumn[*client.AccessListDetails]("Enabled", func(l *client.AccessListDetails) bool { return l.IsEnabled }),
		output.StringColumn[*client.AccessListDetails]("Action", func(l *client.AccessListDetails) string {
			return client.AccessListActionName(l.Action)
		}),
		output.IntColumn[*client.AccessListDetails]("Category", func(l *client.AccessListDetails) int { return l.Category }),
		output.IntColumn[*client.AccessListDetails]("Entries", func(l *client.AccessListDetails) int { return int(l.EntryCount) }),
	})
}

func customAccessListDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.CustomAccessList]{
		output.IntColumn[*client.CustomAccessList]("Id", func(l *client.CustomAccessList) int { return int(l.Id) }),
		output.StringColumn[*client.CustomAccessList]("Name", func(l *client.CustomAccessList) string { return l.Name }),
		output.StringColumn[*client.CustomAccessList]("Description", func(l *client.CustomAccessList) string { return l.Description }),
		output.IntColumn[*client.CustomAccessList]("Type", func(l *client.CustomAccessList) int { return l.Type }),
		output.IntColumn[*client.CustomAccessList]("Entries", func(l *client.CustomAccessList) int { return int(l.EntryCount) }),
		output.StringColumn[*client.CustomAccessList]("Checksum", func(l *client.CustomAccessList) string { return l.Checksum }),
		output.StringColumn[*client.CustomAccessList]("Last Modified", func(l *client.CustomAccessList) string { return l.LastModified }),
	})
}
