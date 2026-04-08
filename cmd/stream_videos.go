package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func newStreamVideosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "videos",
		Short: "Manage videos in a stream library",
	}
	cmd.AddCommand(newStreamVideosListCmd())
	cmd.AddCommand(withWatch(newStreamVideosGetCmd()))
	cmd.AddCommand(withFromFile(withInteractive(newStreamVideosCreateCmd())))
	cmd.AddCommand(withFromFile(newStreamVideosUpdateCmd()))
	cmd.AddCommand(newStreamVideosDeleteCmd())
	cmd.AddCommand(newStreamVideosUploadCmd())
	cmd.AddCommand(newStreamVideosFetchCmd())
	cmd.AddCommand(newStreamVideosReencodeCmd())
	cmd.AddCommand(newStreamVideosTranscribeCmd())
	return cmd
}

func newStreamVideosListCmd() *cobra.Command {
	var (
		limit      int
		all        bool
		search     string
		collection string
		orderBy    string
	)

	cmd := &cobra.Command{
		Use:   "list <library_id>",
		Short: "List videos in a library",
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

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.Video], error) {
				return c.ListVideos(cmd.Context(), libraryId, page, perPage, search, collection, orderBy)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := videoListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, v := range result.Items {
				items[i] = v
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
	cmd.Flags().StringVar(&collection, "collection", "", "Filter by collection ID")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Order results by field (default: date)")

	return cmd
}

func newStreamVideosGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <library_id> <video_id>",
		Short: "Get video details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			v, err := c.GetVideo(cmd.Context(), libraryId, videoId)
			if err != nil {
				return err
			}

			columns := videoDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newStreamVideosCreateCmd() *cobra.Command {
	var (
		title         string
		collectionId  string
		thumbnailTime int
	)

	cmd := &cobra.Command{
		Use:   "create <library_id>",
		Short: "Create a video record",
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

			body := &client.VideoCreate{
				Title: title,
			}
			if cmd.Flags().Changed("collection-id") {
				body.CollectionId = collectionId
			}
			if cmd.Flags().Changed("thumbnail-time") {
				body.ThumbnailTime = thumbnailTime
			}

			v, err := c.CreateVideo(cmd.Context(), libraryId, body)
			if err != nil {
				return err
			}

			columns := videoDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Video title (required)")
	cmd.Flags().StringVar(&collectionId, "collection-id", "", "Collection ID")
	cmd.Flags().IntVar(&thumbnailTime, "thumbnail-time", 0, "Thumbnail time in milliseconds")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}

func newStreamVideosUpdateCmd() *cobra.Command {
	var (
		title        string
		collectionId string
	)

	cmd := &cobra.Command{
		Use:   "update <library_id> <video_id>",
		Short: "Update a video",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.VideoUpdate{}

			if cmd.Flags().Changed("title") {
				body.Title = &title
			}
			if cmd.Flags().Changed("collection-id") {
				body.CollectionId = &collectionId
			}

			if err := c.UpdateVideo(cmd.Context(), libraryId, videoId, body); err != nil {
				return err
			}

			// Fetch and display the updated video
			v, err := c.GetVideo(cmd.Context(), libraryId, videoId)
			if err != nil {
				return err
			}

			columns := videoDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Video title")
	cmd.Flags().StringVar(&collectionId, "collection-id", "", "Collection ID")

	return cmd
}

func newStreamVideosDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <library_id> <video_id>",
		Short: "Delete a video",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete video %s? [y/N] ", videoId))
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

			if err := c.DeleteVideo(cmd.Context(), libraryId, videoId); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Video deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func newStreamVideosUploadCmd() *cobra.Command {
	var (
		title        string
		collectionId string
	)

	cmd := &cobra.Command{
		Use:   "upload <library_id> <file>",
		Short: "Upload a video file",
		Long:  "Creates a video record and uploads the file in one step. Title defaults to filename.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			filePath := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			// Default title to filename without extension
			if title == "" {
				title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
			}

			// Step 1: Create video record
			createBody := &client.VideoCreate{
				Title:        title,
				CollectionId: collectionId,
			}
			v, err := c.CreateVideo(cmd.Context(), libraryId, createBody)
			if err != nil {
				return fmt.Errorf("creating video record: %w", err)
			}

			// Step 2: Upload binary
			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer func() { _ = f.Close() }()

			info, err := f.Stat()
			if err != nil {
				return fmt.Errorf("reading file info: %w", err)
			}
			size := info.Size()

			var reader io.Reader = f
			if showProgress(cmd) {
				bar := progressbar.DefaultBytes(size, "uploading")
				reader = io.TeeReader(f, bar)
				defer func() {
					_ = bar.Close()
					fmt.Fprintln(cmd.ErrOrStderr())
				}()
			}

			if err := c.UploadVideo(cmd.Context(), libraryId, v.Guid, reader, size); err != nil {
				return fmt.Errorf("uploading video: %w", err)
			}

			// Step 3: Fetch and display the video
			v, err = c.GetVideo(cmd.Context(), libraryId, v.Guid)
			if err != nil {
				return err
			}

			columns := videoDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Video title (defaults to filename)")
	cmd.Flags().StringVar(&collectionId, "collection-id", "", "Collection ID")

	return cmd
}

func newStreamVideosFetchCmd() *cobra.Command {
	var (
		fetchUrl     string
		title        string
		collectionId string
	)

	cmd := &cobra.Command{
		Use:   "fetch <library_id>",
		Short: "Fetch and import a video from a URL",
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

			body := &client.VideoFetch{
				Url: fetchUrl,
			}
			if title != "" {
				body.Title = title
			}

			status, err := c.FetchVideo(cmd.Context(), libraryId, body, collectionId, 0)
			if err != nil {
				return err
			}

			if status.Success {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "Video fetch initiated.")
			} else {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Video fetch queued: %s\n", status.Message)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&fetchUrl, "url", "", "URL to fetch video from (required)")
	cmd.Flags().StringVar(&title, "title", "", "Video title")
	cmd.Flags().StringVar(&collectionId, "collection-id", "", "Collection ID")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newStreamVideosReencodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reencode <library_id> <video_id>",
		Short: "Re-encode a video from the original file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			v, err := c.ReencodeVideo(cmd.Context(), libraryId, videoId)
			if err != nil {
				return err
			}

			columns := videoDetailColumns()

			formatted, err := output.FormatOne(cfg, columns, v)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newStreamVideosTranscribeCmd() *cobra.Command {
	var (
		languages           []string
		sourceLanguage      string
		generateTitle       bool
		generateDescription bool
		generateChapters    bool
		generateMoments     bool
	)

	cmd := &cobra.Command{
		Use:   "transcribe <library_id> <video_id>",
		Short: "Transcribe a video",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			settings := &client.TranscribeSettings{}
			if cmd.Flags().Changed("languages") {
				settings.TargetLanguages = languages
			}
			if cmd.Flags().Changed("source-language") {
				settings.SourceLanguage = sourceLanguage
			}
			if cmd.Flags().Changed("generate-title") {
				settings.GenerateTitle = generateTitle
			}
			if cmd.Flags().Changed("generate-description") {
				settings.GenerateDescription = generateDescription
			}
			if cmd.Flags().Changed("generate-chapters") {
				settings.GenerateChapters = generateChapters
			}
			if cmd.Flags().Changed("generate-moments") {
				settings.GenerateMoments = generateMoments
			}

			if err := c.TranscribeVideo(cmd.Context(), libraryId, videoId, settings); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Transcription started.")
			return err
		},
	}

	cmd.Flags().StringSliceVar(&languages, "languages", nil, "Target languages (comma-separated ISO 639-1 codes)")
	cmd.Flags().StringVar(&sourceLanguage, "source-language", "", "Source language (ISO 639-1 code)")
	cmd.Flags().BoolVar(&generateTitle, "generate-title", false, "Auto-generate title")
	cmd.Flags().BoolVar(&generateDescription, "generate-description", false, "Auto-generate description")
	cmd.Flags().BoolVar(&generateChapters, "generate-chapters", false, "Auto-generate chapters")
	cmd.Flags().BoolVar(&generateMoments, "generate-moments", false, "Auto-generate moments")

	return cmd
}

// --- Column definitions ---

func videoListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Video]{
		output.StringColumn[*client.Video]("Guid", func(v *client.Video) string { return v.Guid }),
		output.StringColumn[*client.Video]("Title", func(v *client.Video) string { return v.Title }),
		output.IntColumn[*client.Video]("Views", func(v *client.Video) int { return int(v.Views) }),
		output.IntColumn[*client.Video]("Length", func(v *client.Video) int { return v.Length }),
		output.StringColumn[*client.Video]("Status", func(v *client.Video) string { return client.VideoStatusName(v.Status) }),
		output.IntColumn[*client.Video]("Storage Size", func(v *client.Video) int { return int(v.StorageSize) }),
		output.StringColumn[*client.Video]("Date Uploaded", func(v *client.Video) string { return v.DateUploaded }),
	})
}

func videoDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Video]{
		output.StringColumn[*client.Video]("Guid", func(v *client.Video) string { return v.Guid }),
		output.IntColumn[*client.Video]("Library Id", func(v *client.Video) int { return int(v.VideoLibraryId) }),
		output.StringColumn[*client.Video]("Title", func(v *client.Video) string { return v.Title }),
		output.StringColumn[*client.Video]("Description", func(v *client.Video) string { return v.Description }),
		output.IntColumn[*client.Video]("Views", func(v *client.Video) int { return int(v.Views) }),
		output.IntColumn[*client.Video]("Length", func(v *client.Video) int { return v.Length }),
		output.StringColumn[*client.Video]("Status", func(v *client.Video) string { return client.VideoStatusName(v.Status) }),
		output.IntColumn[*client.Video]("Encode Progress", func(v *client.Video) int { return v.EncodeProgress }),
		output.IntColumn[*client.Video]("Width", func(v *client.Video) int { return v.Width }),
		output.IntColumn[*client.Video]("Height", func(v *client.Video) int { return v.Height }),
		output.StringColumn[*client.Video]("Resolutions", func(v *client.Video) string { return v.AvailableResolutions }),
		output.IntColumn[*client.Video]("Storage Size", func(v *client.Video) int { return int(v.StorageSize) }),
		output.BoolColumn[*client.Video]("MP4 Fallback", func(v *client.Video) bool { return v.HasMP4Fallback }),
		output.StringColumn[*client.Video]("Collection Id", func(v *client.Video) string { return v.CollectionId }),
		output.StringColumn[*client.Video]("Category", func(v *client.Video) string { return v.Category }),
		output.IntColumn[*client.Video]("Avg Watch Time", func(v *client.Video) int { return int(v.AverageWatchTime) }),
		output.IntColumn[*client.Video]("Total Watch Time", func(v *client.Video) int { return int(v.TotalWatchTime) }),
		output.StringColumn[*client.Video]("Date Uploaded", func(v *client.Video) string { return v.DateUploaded }),
	})
}
