package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
)

func newShieldMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "View Shield metrics and statistics",
	}
	cmd.AddCommand(newShieldMetricsOverviewCmd())
	cmd.AddCommand(newShieldMetricsDetailedCmd())
	cmd.AddCommand(newShieldMetricsRateLimitsCmd())
	cmd.AddCommand(newShieldMetricsWafRuleCmd())
	cmd.AddCommand(newShieldMetricsBotDetectionCmd())
	cmd.AddCommand(newShieldMetricsUploadScanningCmd())
	return cmd
}

func newShieldMetricsOverviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overview <shield_zone_id>",
		Short: "Get Shield zone metrics overview",
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

			metrics, err := c.GetShieldMetricsOverview(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := shieldMetricsOverviewColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, metrics)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldMetricsDetailedCmd() *cobra.Command {
	var (
		startDate  string
		endDate    string
		resolution int
	)

	cmd := &cobra.Command{
		Use:   "detailed <shield_zone_id>",
		Short: "Get detailed Shield zone metrics",
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

			metrics, err := c.GetShieldMetricsDetailed(cmd.Context(), shieldZoneId, startDate, endDate, resolution)
			if err != nil {
				return err
			}

			columns := shieldMetricsDetailedColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, metrics)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&resolution, "resolution", 0, "Metrics resolution")

	return cmd
}

func newShieldMetricsRateLimitsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limits <shield_zone_id>",
		Short: "Get rate limit metrics for a Shield zone",
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

			metrics, err := c.GetShieldRateLimitMetrics(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := shieldRateLimitMetricsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(metrics))
			for i, m := range metrics {
				items[i] = m
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

func newShieldMetricsWafRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "waf-rule <shield_zone_id> <rule_id>",
		Short: "Get metrics for a specific WAF rule",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			shieldZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}
			ruleId, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid rule ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			metrics, err := c.GetShieldWafRuleMetrics(cmd.Context(), shieldZoneId, ruleId)
			if err != nil {
				return err
			}

			columns := wafRuleMetricsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, metrics)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldMetricsBotDetectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bot-detection <shield_zone_id>",
		Short: "Get bot detection metrics for a Shield zone",
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

			metrics, err := c.GetShieldBotDetectionMetrics(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := shieldBotDetectionMetricsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, metrics)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldMetricsUploadScanningCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-scanning <shield_zone_id>",
		Short: "Get upload scanning metrics for a Shield zone",
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

			metrics, err := c.GetShieldUploadScanningMetrics(cmd.Context(), shieldZoneId)
			if err != nil {
				return err
			}

			columns := shieldUploadScanningMetricsColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, metrics)
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

func shieldMetricsOverviewColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZoneMetrics]{
		output.IntColumn[*client.ShieldZoneMetrics]("DDoS Mitigated", func(m *client.ShieldZoneMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.DDoSMitigated)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("WAF Triggered", func(m *client.ShieldZoneMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.WafTriggeredRules)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("Rate Limit Breaches", func(m *client.ShieldZoneMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.RatelimitBreaches)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("Bot Challenged", func(m *client.ShieldZoneMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.BotDetectionChallenged)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("Access List Actions", func(m *client.ShieldZoneMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.AccessListActions)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("Upload Scan Blocks", func(m *client.ShieldZoneMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.UploadScanningBlocks)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("Clean Requests Limit", func(m *client.ShieldZoneMetrics) int {
			return int(m.TotalCleanRequestsLimit)
		}),
		output.IntColumn[*client.ShieldZoneMetrics]("Billable Requests", func(m *client.ShieldZoneMetrics) int {
			return int(m.TotalBillableRequests)
		}),
	})
}

func shieldMetricsDetailedColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldOverviewMetricsData]{
		output.IntColumn[*client.ShieldOverviewMetricsData]("Clean Requests Limit", func(m *client.ShieldOverviewMetricsData) int {
			return int(m.TotalCleanRequestsLimit)
		}),
		output.IntColumn[*client.ShieldOverviewMetricsData]("Billable Requests", func(m *client.ShieldOverviewMetricsData) int {
			return int(m.TotalBillableRequestsThisMonth)
		}),
		output.IntColumn[*client.ShieldOverviewMetricsData]("Resolution", func(m *client.ShieldOverviewMetricsData) int {
			return m.Resolution
		}),
	})
}

func shieldRateLimitMetricsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZoneRateLimitMetrics]{
		output.IntColumn[*client.ShieldZoneRateLimitMetrics]("Rate Limit Id", func(m *client.ShieldZoneRateLimitMetrics) int {
			return int(m.RatelimitId)
		}),
		output.IntColumn[*client.ShieldZoneRateLimitMetrics]("Total Breaches", func(m *client.ShieldZoneRateLimitMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.TotalBreaches)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneRateLimitMetrics]("Logged", func(m *client.ShieldZoneRateLimitMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.LoggedBreaches)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneRateLimitMetrics]("Challenged", func(m *client.ShieldZoneRateLimitMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.ChallengedBreaches)
			}
			return 0
		}),
		output.IntColumn[*client.ShieldZoneRateLimitMetrics]("Blocked", func(m *client.ShieldZoneRateLimitMetrics) int {
			if m.Overview != nil {
				return int(m.Overview.BlockedBreaches)
			}
			return 0
		}),
	})
}

func wafRuleMetricsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.WafRuleMetrics]{
		output.IntColumn[*client.WafRuleMetrics]("Total Triggers", func(m *client.WafRuleMetrics) int { return int(m.TotalTriggers) }),
		output.IntColumn[*client.WafRuleMetrics]("Blocked", func(m *client.WafRuleMetrics) int { return int(m.BlockedRequests) }),
		output.IntColumn[*client.WafRuleMetrics]("Logged", func(m *client.WafRuleMetrics) int { return int(m.LoggedRequests) }),
		output.IntColumn[*client.WafRuleMetrics]("Challenged", func(m *client.WafRuleMetrics) int { return int(m.ChallengedRequests) }),
	})
}

func shieldBotDetectionMetricsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZoneBotDetectionMetrics]{
		output.IntColumn[*client.ShieldZoneBotDetectionMetrics]("Logged Requests", func(m *client.ShieldZoneBotDetectionMetrics) int {
			return int(m.TotalLoggedRequests)
		}),
		output.IntColumn[*client.ShieldZoneBotDetectionMetrics]("Challenged Requests", func(m *client.ShieldZoneBotDetectionMetrics) int {
			return int(m.TotalChallengedRequests)
		}),
	})
}

func shieldUploadScanningMetricsColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZoneUploadScanningMetrics]{
		output.IntColumn[*client.ShieldZoneUploadScanningMetrics]("Logged Requests", func(m *client.ShieldZoneUploadScanningMetrics) int {
			return int(m.TotalLoggedRequests)
		}),
		output.IntColumn[*client.ShieldZoneUploadScanningMetrics]("Blocked Requests", func(m *client.ShieldZoneUploadScanningMetrics) int {
			return int(m.TotalBlockedRequests)
		}),
		output.IntColumn[*client.ShieldZoneUploadScanningMetrics]("Files Scanned", func(m *client.ShieldZoneUploadScanningMetrics) int {
			return int(m.TotalFilesScanned)
		}),
	})
}
