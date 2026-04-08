package cmd

import (
	"fmt"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage account settings",
	}

	cmd.AddCommand(newApiKeysCmd())
	cmd.AddCommand(newAuditLogCmd())

	return cmd
}

// --- api-keys ---

func newApiKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"apikeys"},
		Short:   "Manage API keys",
	}

	cmd.AddCommand(newApiKeysListCmd())

	return cmd
}

func newApiKeysListCmd() *cobra.Command {
	var (
		limit int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.ApiKey], error) {
				return c.ListApiKeys(cmd.Context(), page, perPage)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := apiKeyListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, k := range result.Items {
				items[i] = k
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

	return cmd
}

// --- audit-log ---

func newAuditLogCmd() *cobra.Command {
	var (
		product      []string
		resourceType []string
		resourceId   []string
		actorId      []string
		order        string
		limit        int
		all          bool
	)

	cmd := &cobra.Command{
		Use:   "audit-log <date>",
		Short: "Get audit log entries for a date",
		Long:  "Retrieve audit log entries for the given date (e.g. 2024-01-15).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			date := args[0]

			c, err := AppFromContext(cmd.Context()).NewAccountAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			opts := client.AuditLogOptions{
				Product:      product,
				ResourceType: resourceType,
				ResourceId:   resourceId,
				ActorId:      actorId,
				Order:        order,
			}

			if limit > 0 && !all {
				opts.Limit = limit
			}

			// Collect all pages using continuation tokens
			var allLogs []*client.AuditLogEntry
			var hasMore bool

			for {
				resp, err := c.GetAuditLog(cmd.Context(), date, opts)
				if err != nil {
					return err
				}

				allLogs = append(allLogs, resp.Logs...)

				if !resp.HasMoreData || !all {
					hasMore = resp.HasMoreData
					break
				}

				opts.ContinuationToken = resp.ContinuationToken
			}

			// Apply limit if set and not fetching all
			if limit > 0 && !all && len(allLogs) > limit {
				allLogs = allLogs[:limit]
				hasMore = true
			}

			columns := auditLogColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(allLogs))
			for i, l := range allLogs {
				items[i] = l
			}

			formatted, err := output.FormatList(cfg, columns, items, hasMore)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringSliceVar(&product, "product", nil, "Filter by product (repeatable)")
	cmd.Flags().StringSliceVar(&resourceType, "resource-type", nil, "Filter by resource type (repeatable)")
	cmd.Flags().StringSliceVar(&resourceId, "resource-id", nil, "Filter by resource ID (repeatable)")
	cmd.Flags().StringSliceVar(&actorId, "actor-id", nil, "Filter by actor ID (repeatable)")
	cmd.Flags().StringVar(&order, "order", "", "Sort order (Ascending or Descending)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results to return")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch all pages of results")

	return cmd
}

// --- columns ---

func apiKeyListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ApiKey]{
		output.IntColumn[*client.ApiKey]("Id", func(k *client.ApiKey) int { return int(k.Id) }),
		output.StringColumn[*client.ApiKey]("Key", func(k *client.ApiKey) string { return k.Key }),
		output.StringColumn[*client.ApiKey]("Roles", func(k *client.ApiKey) string { return client.FormatRoles(k.Roles) }),
	})
}

func auditLogColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.AuditLogEntry]{
		output.StringColumn[*client.AuditLogEntry]("Timestamp", func(e *client.AuditLogEntry) string { return e.Timestamp }),
		output.StringColumn[*client.AuditLogEntry]("Product", func(e *client.AuditLogEntry) string { return e.Product }),
		output.StringColumn[*client.AuditLogEntry]("Resource Type", func(e *client.AuditLogEntry) string { return e.ResourceType }),
		output.StringColumn[*client.AuditLogEntry]("Resource Id", func(e *client.AuditLogEntry) string { return e.ResourceId }),
		output.StringColumn[*client.AuditLogEntry]("Action", func(e *client.AuditLogEntry) string { return e.Action }),
		output.StringColumn[*client.AuditLogEntry]("Actor Id", func(e *client.AuditLogEntry) string { return e.ActorId }),
		output.StringColumn[*client.AuditLogEntry]("Actor Type", func(e *client.AuditLogEntry) string { return e.ActorType }),
	})
}
