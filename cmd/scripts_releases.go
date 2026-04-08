package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newScriptsPublishCmd() *cobra.Command {
	var note string

	cmd := &cobra.Command{
		Use:   "publish <script_id> [uuid]",
		Short: "Publish an edge script release",
		Long:  "Publish the current code as a new release, or publish a specific release by UUID.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			if len(args) == 2 {
				if err := c.PublishEdgeScriptRelease(cmd.Context(), id, args[1]); err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Release %s published.\n", args[1])
				return err
			}

			body := &client.EdgeScriptPublish{}
			if cmd.Flags().Changed("note") {
				body.Note = note
			}

			if err := c.PublishEdgeScript(cmd.Context(), id, body); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Edge script published.")
			return err
		},
	}

	cmd.Flags().StringVar(&note, "note", "", "Release note")

	return cmd
}

func newScriptsReleasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases",
		Short: "Manage edge script releases",
	}
	cmd.AddCommand(newScriptsReleasesListCmd())
	cmd.AddCommand(newScriptsReleasesActiveCmd())
	return cmd
}

func newScriptsReleasesListCmd() *cobra.Command {
	var (
		limit int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "list <script_id>",
		Short: "List edge script releases",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error) {
				return c.ListEdgeScriptReleases(cmd.Context(), id, page, perPage)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := edgeScriptReleaseListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, r := range result.Items {
				items[i] = r
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

func newScriptsReleasesActiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active <script_id>",
		Short: "Get the active release for an edge script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid script ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewEdgeScriptAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			release, err := c.GetActiveEdgeScriptRelease(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := edgeScriptReleaseDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, release)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

// --- Release column definitions ---

func edgeScriptReleaseListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptRelease]{
		output.IntColumn[*client.EdgeScriptRelease]("Id", func(r *client.EdgeScriptRelease) int { return int(r.Id) }),
		output.StringColumn[*client.EdgeScriptRelease]("UUID", func(r *client.EdgeScriptRelease) string { return r.Uuid }),
		output.StringColumn[*client.EdgeScriptRelease]("Status", func(r *client.EdgeScriptRelease) string { return client.ReleaseStatusName(r.Status) }),
		output.StringColumn[*client.EdgeScriptRelease]("Author", func(r *client.EdgeScriptRelease) string { return r.Author }),
		output.StringColumn[*client.EdgeScriptRelease]("Note", func(r *client.EdgeScriptRelease) string { return r.Note }),
		output.StringColumn[*client.EdgeScriptRelease]("Date Published", func(r *client.EdgeScriptRelease) string { return r.DatePublished }),
	})
}

func edgeScriptReleaseDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EdgeScriptRelease]{
		output.IntColumn[*client.EdgeScriptRelease]("Id", func(r *client.EdgeScriptRelease) int { return int(r.Id) }),
		output.StringColumn[*client.EdgeScriptRelease]("UUID", func(r *client.EdgeScriptRelease) string { return r.Uuid }),
		output.StringColumn[*client.EdgeScriptRelease]("Status", func(r *client.EdgeScriptRelease) string { return client.ReleaseStatusName(r.Status) }),
		output.StringColumn[*client.EdgeScriptRelease]("Author", func(r *client.EdgeScriptRelease) string { return r.Author }),
		output.StringColumn[*client.EdgeScriptRelease]("Author Email", func(r *client.EdgeScriptRelease) string { return r.AuthorEmail }),
		output.StringColumn[*client.EdgeScriptRelease]("Commit SHA", func(r *client.EdgeScriptRelease) string { return r.CommitSha }),
		output.StringColumn[*client.EdgeScriptRelease]("Note", func(r *client.EdgeScriptRelease) string { return r.Note }),
		output.StringColumn[*client.EdgeScriptRelease]("Date Released", func(r *client.EdgeScriptRelease) string { return r.DateReleased }),
		output.StringColumn[*client.EdgeScriptRelease]("Date Published", func(r *client.EdgeScriptRelease) string { return r.DatePublished }),
	})
}
