package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newStreamStatisticsCmd() *cobra.Command {
	var (
		videoGuid string
		dateFrom  string
		dateTo    string
		hourly    bool
	)

	cmd := &cobra.Command{
		Use:   "statistics <library_id>",
		Short: "Get video library statistics",
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

			stats, err := c.GetVideoStatistics(cmd.Context(), libraryId, dateFrom, dateTo, hourly, videoGuid)
			if err != nil {
				return err
			}

			columns := videoStatisticsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, stats)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&videoGuid, "video-guid", "", "Filter statistics for a specific video")
	cmd.Flags().StringVar(&dateFrom, "date-from", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "date-to", "", "End date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&hourly, "hourly", false, "Return hourly statistics")

	return cmd
}

func newStreamHeatmapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "heatmap <library_id> <video_id>",
		Short: "Get video heatmap data",
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

			hm, err := c.GetVideoHeatmap(cmd.Context(), libraryId, videoId)
			if err != nil {
				return err
			}

			// Heatmap data is best expressed as JSON
			data, err := json.MarshalIndent(hm, "", "  ")
			if err != nil {
				return fmt.Errorf("formatting heatmap: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return err
		},
	}

	return cmd
}

// --- Column definitions ---

func videoStatisticsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.VideoStatistics]{
		output.IntColumn[*client.VideoStatistics]("Engagement Score", func(s *client.VideoStatistics) int { return s.EngagementScore }),
		output.StringColumn[*client.VideoStatistics]("Views Chart", func(s *client.VideoStatistics) string {
			data, _ := json.Marshal(s.ViewsChart)
			return string(data)
		}),
		output.StringColumn[*client.VideoStatistics]("Watch Time Chart", func(s *client.VideoStatistics) string {
			data, _ := json.Marshal(s.WatchTimeChart)
			return string(data)
		}),
		output.StringColumn[*client.VideoStatistics]("Country Views", func(s *client.VideoStatistics) string {
			data, _ := json.Marshal(s.CountryViewCounts)
			return string(data)
		}),
		output.StringColumn[*client.VideoStatistics]("Country Watch Time", func(s *client.VideoStatistics) string {
			data, _ := json.Marshal(s.CountryWatchTime)
			return string(data)
		}),
	})
}
