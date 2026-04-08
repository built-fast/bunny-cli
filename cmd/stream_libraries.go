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

func newStreamLibrariesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "libraries",
		Aliases: []string{"lib"},
		Short:   "Manage video libraries",
	}
	cmd.AddCommand(newStreamLibrariesListCmd())
	cmd.AddCommand(withWatch(newStreamLibrariesGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newStreamLibrariesCreateCmd())))
	cmd.AddCommand(withFromFile(newStreamLibrariesUpdateCmd()))
	cmd.AddCommand(newStreamLibrariesDeleteCmd())
	cmd.AddCommand(newStreamLibrariesResetApiKeyCmd())
	cmd.AddCommand(newStreamLibrariesLanguagesCmd())
	return cmd
}

func newStreamLibrariesListCmd() *cobra.Command {
	var (
		limit  int
		all    bool
		search string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List video libraries",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.VideoLibrary], error) {
				return c.ListVideoLibraries(cmd.Context(), page, perPage, search)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := videoLibraryListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, lib := range result.Items {
				items[i] = lib
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

func newStreamLibrariesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get video library details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid video library ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			lib, err := c.GetVideoLibrary(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := videoLibraryDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, lib)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newStreamLibrariesCreateCmd() *cobra.Command {
	var (
		name               string
		replicationRegions []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a video library",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.VideoLibraryCreate{
				Name: name,
			}

			if cmd.Flags().Changed("replication-regions") {
				body.ReplicationRegions = replicationRegions
			}

			lib, err := c.CreateVideoLibrary(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := videoLibraryDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, lib)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Video library name (required)")
	cmd.Flags().StringSliceVar(&replicationRegions, "replication-regions", nil, "Replication regions (comma-separated): DE, NY, LA, SG, SYD")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newStreamLibrariesUpdateCmd() *cobra.Command {
	var (
		name               string
		enabledResolutions string
		enableMP4Fallback  bool
		keepOriginalFiles  bool
		allowDirectPlay    bool
		enableDRM          bool
		webhookUrl         string
		replicationRegions []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a video library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid video library ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.VideoLibraryUpdate{}

			if cmd.Flags().Changed("name") {
				body.Name = &name
			}
			if cmd.Flags().Changed("enabled-resolutions") {
				body.EnabledResolutions = &enabledResolutions
			}
			if cmd.Flags().Changed("enable-mp4-fallback") {
				body.EnableMP4Fallback = &enableMP4Fallback
			}
			if cmd.Flags().Changed("keep-original-files") {
				body.KeepOriginalFiles = &keepOriginalFiles
			}
			if cmd.Flags().Changed("allow-direct-play") {
				body.AllowDirectPlay = &allowDirectPlay
			}
			if cmd.Flags().Changed("enable-drm") {
				body.EnableDRM = &enableDRM
			}
			if cmd.Flags().Changed("webhook-url") {
				body.WebhookUrl = &webhookUrl
			}
			if cmd.Flags().Changed("replication-regions") {
				body.ReplicationRegions = replicationRegions
			}

			lib, err := c.UpdateVideoLibrary(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := videoLibraryDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, lib)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Video library name")
	cmd.Flags().StringVar(&enabledResolutions, "enabled-resolutions", "", "Enabled resolutions (comma-separated: 240p,360p,480p,720p,1080p,1440p,2160p)")
	cmd.Flags().BoolVar(&enableMP4Fallback, "enable-mp4-fallback", false, "Enable MP4 fallback")
	cmd.Flags().BoolVar(&keepOriginalFiles, "keep-original-files", false, "Keep original video files after encoding")
	cmd.Flags().BoolVar(&allowDirectPlay, "allow-direct-play", false, "Allow direct play URLs")
	cmd.Flags().BoolVar(&enableDRM, "enable-drm", false, "Enable DRM")
	cmd.Flags().StringVar(&webhookUrl, "webhook-url", "", "Webhook URL")
	cmd.Flags().StringSliceVar(&replicationRegions, "replication-regions", nil, "Replication regions (comma-separated)")

	return cmd
}

func newStreamLibrariesDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a video library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid video library ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete video library %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.DeleteVideoLibrary(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Video library deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func newStreamLibrariesResetApiKeyCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "reset-api-key <id>",
		Short: "Reset the API key for a video library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid video library ID: %w", err)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to reset the API key for video library %d? [y/N] ", id))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Operation canceled.")
					return err
				}
			}

			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			if err := c.ResetVideoLibraryApiKey(cmd.Context(), id); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "API key reset.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func newStreamLibrariesLanguagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "languages",
		Short: "List supported transcription languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewVideoLibraryAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			langs, err := c.ListVideoLibraryLanguages(cmd.Context())
			if err != nil {
				return err
			}

			columns := videoLibraryLanguageColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(langs))
			for i := range langs {
				items[i] = &langs[i]
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

// --- Column definitions ---

func videoLibraryListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.VideoLibrary]{
		output.IntColumn[*client.VideoLibrary]("Id", func(lib *client.VideoLibrary) int { return int(lib.Id) }),
		output.StringColumn[*client.VideoLibrary]("Name", func(lib *client.VideoLibrary) string { return lib.Name }),
		output.IntColumn[*client.VideoLibrary]("Videos", func(lib *client.VideoLibrary) int { return int(lib.VideoCount) }),
		output.IntColumn[*client.VideoLibrary]("Storage Used", func(lib *client.VideoLibrary) int { return int(lib.StorageUsage) }),
		output.IntColumn[*client.VideoLibrary]("Traffic Used", func(lib *client.VideoLibrary) int { return int(lib.TrafficUsage) }),
		output.StringColumn[*client.VideoLibrary]("Date Created", func(lib *client.VideoLibrary) string { return lib.DateCreated }),
	})
}

func videoLibraryDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.VideoLibrary]{
		output.IntColumn[*client.VideoLibrary]("Id", func(lib *client.VideoLibrary) int { return int(lib.Id) }),
		output.StringColumn[*client.VideoLibrary]("Name", func(lib *client.VideoLibrary) string { return lib.Name }),
		output.IntColumn[*client.VideoLibrary]("Videos", func(lib *client.VideoLibrary) int { return int(lib.VideoCount) }),
		output.IntColumn[*client.VideoLibrary]("Storage Used", func(lib *client.VideoLibrary) int { return int(lib.StorageUsage) }),
		output.IntColumn[*client.VideoLibrary]("Traffic Used", func(lib *client.VideoLibrary) int { return int(lib.TrafficUsage) }),
		output.StringColumn[*client.VideoLibrary]("API Key", func(lib *client.VideoLibrary) string { return lib.ApiKey }),
		output.StringColumn[*client.VideoLibrary]("Read-Only API Key", func(lib *client.VideoLibrary) string { return lib.ReadOnlyApiKey }),
		output.IntColumn[*client.VideoLibrary]("Pull Zone Id", func(lib *client.VideoLibrary) int { return int(lib.PullZoneId) }),
		output.IntColumn[*client.VideoLibrary]("Storage Zone Id", func(lib *client.VideoLibrary) int { return int(lib.StorageZoneId) }),
		output.StringColumn[*client.VideoLibrary]("Resolutions", func(lib *client.VideoLibrary) string { return lib.EnabledResolutions }),
		output.BoolColumn[*client.VideoLibrary]("MP4 Fallback", func(lib *client.VideoLibrary) bool { return lib.EnableMP4Fallback }),
		output.BoolColumn[*client.VideoLibrary]("Keep Originals", func(lib *client.VideoLibrary) bool { return lib.KeepOriginalFiles }),
		output.BoolColumn[*client.VideoLibrary]("Direct Play", func(lib *client.VideoLibrary) bool { return lib.AllowDirectPlay }),
		output.BoolColumn[*client.VideoLibrary]("DRM", func(lib *client.VideoLibrary) bool { return lib.EnableDRM }),
		output.StringColumn[*client.VideoLibrary]("Webhook URL", func(lib *client.VideoLibrary) string { return lib.WebhookUrl }),
		output.StringColumn[*client.VideoLibrary]("Replication Regions", func(lib *client.VideoLibrary) string {
			return strings.Join(lib.ReplicationRegions, ", ")
		}),
		output.StringColumn[*client.VideoLibrary]("Date Created", func(lib *client.VideoLibrary) string { return lib.DateCreated }),
		output.StringColumn[*client.VideoLibrary]("Date Modified", func(lib *client.VideoLibrary) string { return lib.DateModified }),
	})
}

func videoLibraryLanguageColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.VideoLibraryLanguage]{
		output.StringColumn[*client.VideoLibraryLanguage]("Code", func(l *client.VideoLibraryLanguage) string { return l.ShortCode }),
		output.StringColumn[*client.VideoLibraryLanguage]("Name", func(l *client.VideoLibraryLanguage) string { return l.Name }),
		output.IntColumn[*client.VideoLibraryLanguage]("Support Level", func(l *client.VideoLibraryLanguage) int { return l.SupportLevel }),
	})
}
