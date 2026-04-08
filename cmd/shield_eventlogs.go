package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newShieldEventLogsCmd() *cobra.Command {
	var continuationToken string

	cmd := &cobra.Command{
		Use:   "event-logs <shield_zone_id> <date>",
		Short: "Get Shield event logs",
		Long:  "Get Shield event logs for a specific date. Date format: YYYY-MM-DD. Use --continuation-token to paginate through results.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}
			date := args[1]

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			resp, err := c.GetShieldEventLogs(cmd.Context(), shieldZoneId, date, continuationToken)
			if err != nil {
				return err
			}

			columns := eventLogColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(resp.Logs))
			for i := range resp.Logs {
				items[i] = &resp.Logs[i]
			}

			formatted, err := output.FormatList(cfg, columns, items, resp.HasMoreData)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)

			if resp.HasMoreData && resp.ContinuationToken != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nMore data available. Use --continuation-token %s to fetch the next page.\n", resp.ContinuationToken)
			}

			return err
		},
	}

	cmd.Flags().StringVar(&continuationToken, "continuation-token", "", "Continuation token for pagination")

	return cmd
}

// --- Column definitions ---

func eventLogColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.EventLog]{
		output.StringColumn[*client.EventLog]("Log Id", func(l *client.EventLog) string { return l.LogId }),
		output.IntColumn[*client.EventLog]("Timestamp", func(l *client.EventLog) int { return int(l.Timestamp) }),
		output.StringColumn[*client.EventLog]("Log", func(l *client.EventLog) string { return l.Log }),
	})
}
