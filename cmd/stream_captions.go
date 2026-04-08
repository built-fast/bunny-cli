package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/spf13/cobra"
)

func newStreamCaptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "captions",
		Short: "Manage video captions",
	}
	cmd.AddCommand(newStreamCaptionsAddCmd())
	cmd.AddCommand(newStreamCaptionsDeleteCmd())
	return cmd
}

func newStreamCaptionsAddCmd() *cobra.Command {
	var (
		srclang string
		label   string
		file    string
	)

	cmd := &cobra.Command{
		Use:   "add <library_id> <video_id>",
		Short: "Add a caption track to a video",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			// Read and base64 encode the captions file
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("reading captions file: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewStreamAPI(cmd)
			if err != nil {
				return err
			}

			body := &client.CaptionAdd{
				Srclang:      srclang,
				Label:        label,
				CaptionsFile: base64.StdEncoding.EncodeToString(data),
			}

			if err := c.AddCaption(cmd.Context(), libraryId, videoId, srclang, body); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Caption %q (%s) added.\n", label, srclang)
			return err
		},
	}

	cmd.Flags().StringVar(&srclang, "srclang", "", "Language code (ISO 639-1, e.g., en, es, fr) (required)")
	cmd.Flags().StringVar(&label, "label", "", "Display label (e.g., English, Spanish) (required)")
	cmd.Flags().StringVar(&file, "file", "", "Path to caption file (.vtt or .srt) (required)")
	_ = cmd.MarkFlagRequired("srclang")
	_ = cmd.MarkFlagRequired("label")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newStreamCaptionsDeleteCmd() *cobra.Command {
	var (
		srclang string
		yes     bool
	)

	cmd := &cobra.Command{
		Use:   "delete <library_id> <video_id>",
		Short: "Delete a caption track from a video",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			libraryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid library ID: %w", err)
			}
			videoId := args[1]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete caption %q from video %s? [y/N] ", srclang, videoId))
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

			if err := c.DeleteCaption(cmd.Context(), libraryId, videoId, srclang); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Caption %q deleted.\n", srclang)
			return err
		},
	}

	cmd.Flags().StringVar(&srclang, "srclang", "", "Language code to delete (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	_ = cmd.MarkFlagRequired("srclang")

	return cmd
}
