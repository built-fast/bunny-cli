package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newStreamCollectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collections",
		Short: "Manage collections in a stream library",
	}
	cmd.AddCommand(newStreamCollectionsListCmd())
	cmd.AddCommand(withWatch(newStreamCollectionsGetCmd()))
	cmd.AddCommand(withInteractive(newStreamCollectionsCreateCmd()))
	cmd.AddCommand(newStreamCollectionsUpdateCmd())
	cmd.AddCommand(newStreamCollectionsDeleteCmd())
	return cmd
}

func newStreamCollectionsListCmd() *cobra.Command {
	var (
		limit   int
		all     bool
		search  string
		orderBy string
	)

	cmd := &cobra.Command{
		Use:   "list <library_id>",
		Short: "List collections in a library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.Collection], error) {
				return c.ListCollections(cmd.Context(), libraryId, page, perPage, search, orderBy)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := collectionListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, col := range result.Items {
				items[i] = col
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
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Order results by field (default: date)")

	return cmd
}

func newStreamCollectionsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <library_id> <collection_id>",
		Short: "Get collection details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			collectionId := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			col, err := c.GetCollection(cmd.Context(), libraryId, collectionId)
			if err != nil {
				return err
			}

			columns := collectionDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, col)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newStreamCollectionsCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create <library_id>",
		Short: "Create a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.CollectionCreate{Name: name}

			col, err := c.CreateCollection(cmd.Context(), libraryId, body)
			if err != nil {
				return err
			}

			columns := collectionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, col)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Collection name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newStreamCollectionsUpdateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update <library_id> <collection_id>",
		Short: "Update a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			collectionId := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			body := &client.CollectionUpdate{Name: name}

			if err := c.UpdateCollection(cmd.Context(), libraryId, collectionId, body); err != nil {
				return err
			}

			// Fetch and display the updated collection
			cfg := output.FromContext(cmd.Context())

			col, err := c.GetCollection(cmd.Context(), libraryId, collectionId)
			if err != nil {
				return err
			}

			columns := collectionDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, col)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Collection name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newStreamCollectionsDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <library_id> <collection_id>",
		Short: "Delete a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			collectionId := args[1]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete collection %s? [y/N] ", collectionId))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteCollection(cmd.Context(), libraryId, collectionId); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Collection deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// --- Column definitions ---

func collectionListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Collection]{
		output.StringColumn[*client.Collection]("Guid", func(c *client.Collection) string { return c.Guid }),
		output.StringColumn[*client.Collection]("Name", func(c *client.Collection) string { return c.Name }),
		output.IntColumn[*client.Collection]("Videos", func(c *client.Collection) int { return int(c.VideoCount) }),
		output.IntColumn[*client.Collection]("Total Size", func(c *client.Collection) int { return int(c.TotalSize) }),
	})
}

func collectionDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Collection]{
		output.StringColumn[*client.Collection]("Guid", func(c *client.Collection) string { return c.Guid }),
		output.IntColumn[*client.Collection]("Library Id", func(c *client.Collection) int { return int(c.VideoLibraryId) }),
		output.StringColumn[*client.Collection]("Name", func(c *client.Collection) string { return c.Name }),
		output.IntColumn[*client.Collection]("Videos", func(c *client.Collection) int { return int(c.VideoCount) }),
		output.IntColumn[*client.Collection]("Total Size", func(c *client.Collection) int { return int(c.TotalSize) }),
	})
}
