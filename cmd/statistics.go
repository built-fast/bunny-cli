package cmd

import (
	"fmt"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newStatisticsCmd() *cobra.Command {
	var (
		dateFrom     string
		dateTo       string
		pullZone     int64
		serverZoneId int64
		hourly       bool
		loadErrors   bool
	)

	cmd := &cobra.Command{
		Use:     "statistics",
		Aliases: []string{"stats"},
		Short:   "Get global CDN statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewStatisticsAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			opts := client.StatisticsOptions{
				DateFrom:     dateFrom,
				DateTo:       dateTo,
				PullZone:     pullZone,
				ServerZoneId: serverZoneId,
				Hourly:       hourly,
				LoadErrors:   loadErrors,
			}

			stats, err := c.GetStatistics(cmd.Context(), opts)
			if err != nil {
				return err
			}

			columns := statisticsColumns()

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

	cmd.Flags().StringVar(&dateFrom, "date-from", "", "Start date (e.g. 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&dateTo, "date-to", "", "End date (e.g. 2024-01-31T23:59:59Z)")
	cmd.Flags().Int64Var(&pullZone, "pull-zone", -1, "Filter by pull zone ID")
	cmd.Flags().Int64Var(&serverZoneId, "server-zone-id", -1, "Filter by server zone/region ID")
	cmd.Flags().BoolVar(&hourly, "hourly", false, "Return hourly grouped data")
	cmd.Flags().BoolVar(&loadErrors, "load-errors", false, "Include non-2xx error data")

	return cmd
}

func statisticsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.Statistics]{
		output.StringColumn[*client.Statistics]("Total Bandwidth", func(s *client.Statistics) string {
			return client.FormatBytes(s.TotalBandwidthUsed)
		}),
		output.StringColumn[*client.Statistics]("Origin Traffic", func(s *client.Statistics) string {
			return client.FormatBytes(s.TotalOriginTraffic)
		}),
		output.IntColumn[*client.Statistics]("Avg Origin Response (ms)", func(s *client.Statistics) int {
			return int(s.AverageOriginResponseTime)
		}),
		output.StringColumn[*client.Statistics]("Total Requests", func(s *client.Statistics) string {
			return fmt.Sprintf("%d", s.TotalRequestsServed)
		}),
		output.StringColumn[*client.Statistics]("Cache Hit Rate", func(s *client.Statistics) string {
			return fmt.Sprintf("%.2f%%", s.CacheHitRate*100)
		}),
	})
}
