package cmd

import (
	"fmt"
	"strconv"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

func newShieldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shield",
		Short: "Manage Shield security features",
	}
	cmd.AddCommand(newShieldZonesCmd())
	cmd.AddCommand(newShieldWafCmd())
	cmd.AddCommand(newShieldRateLimitsCmd())
	cmd.AddCommand(newShieldAccessListsCmd())
	cmd.AddCommand(newShieldBotDetectionCmd())
	cmd.AddCommand(newShieldUploadScanningCmd())
	cmd.AddCommand(newShieldMetricsCmd())
	cmd.AddCommand(newShieldEventLogsCmd())
	return cmd
}

// --- Shield Zones ---

func newShieldZonesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zones",
		Aliases: []string{"zone"},
		Short:   "Manage Shield zones",
	}
	cmd.AddCommand(newShieldZonesListCmd())
	cmd.AddCommand(withWatch(newShieldZonesGetCmd()))
	cmd.AddCommand(newShieldZonesGetByPullZoneCmd())
	cmd.AddCommand(withFromFile(withInteractive(newShieldZonesCreateCmd())))
	cmd.AddCommand(withFromFile(newShieldZonesUpdateCmd()))
	return cmd
}

func newShieldZonesListCmd() *cobra.Command {
	var (
		limit int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Shield zones",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			result, err := pagination.Collect(func(page, perPage int) (pagination.PageResponse[*client.ShieldZone], error) {
				return c.ListShieldZones(cmd.Context(), page, perPage)
			}, limit, all)
			if err != nil {
				return err
			}

			columns := shieldZoneListColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(result.Items))
			for i, z := range result.Items {
				items[i] = z
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

func newShieldZonesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <shield_zone_id>",
		Short: "Get Shield zone details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			z, err := c.GetShieldZone(cmd.Context(), id)
			if err != nil {
				return err
			}

			columns := shieldZoneDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, z)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldZonesGetByPullZoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-by-pullzone <pull_zone_id>",
		Short: "Get Shield zone by Pull Zone ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pullZoneId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pull zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			z, err := c.GetShieldZoneByPullZone(cmd.Context(), pullZoneId)
			if err != nil {
				return err
			}

			columns := shieldZoneDetailColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			formatted, err := output.FormatOne(cfg, columns, z)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newShieldZonesCreateCmd() *cobra.Command {
	var pullZoneId int64

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Shield zone",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.ShieldZoneCreate{}
			if cmd.Flags().Changed("pull-zone-id") {
				body.PullZoneId = pullZoneId
			}

			z, err := c.CreateShieldZone(cmd.Context(), body)
			if err != nil {
				return err
			}

			columns := shieldZoneResponseColumns()

			formatted, err := output.FormatOne(cfg, columns, z)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().Int64Var(&pullZoneId, "pull-zone-id", 0, "Pull Zone ID to protect (required)")
	_ = cmd.MarkFlagRequired("pull-zone-id")

	return cmd
}

func newShieldZonesUpdateCmd() *cobra.Command {
	var (
		learningMode                         bool
		wafEnabled                           bool
		wafExecutionMode                     int
		wafRequestHeaderLoggingEnabled       bool
		wafRealtimeThreatIntelligenceEnabled bool
		wafProfileId                         int
		dDoSShieldSensitivity                int
		dDoSExecutionMode                    int
		dDoSChallengeWindow                  int
		whitelabelResponsePages              bool
	)

	cmd := &cobra.Command{
		Use:   "update <shield_zone_id>",
		Short: "Update a Shield zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shield zone ID: %w", err)
			}

			c, err := AppFromContext(cmd.Context()).NewShieldAPI(cmd)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			body := &client.ShieldZoneUpdate{}

			if cmd.Flags().Changed("learning-mode") {
				body.LearningMode = &learningMode
			}
			if cmd.Flags().Changed("waf-enabled") {
				body.WafEnabled = &wafEnabled
			}
			if cmd.Flags().Changed("waf-execution-mode") {
				body.WafExecutionMode = &wafExecutionMode
			}
			if cmd.Flags().Changed("waf-request-header-logging") {
				body.WafRequestHeaderLoggingEnabled = &wafRequestHeaderLoggingEnabled
			}
			if cmd.Flags().Changed("waf-realtime-threat-intel") {
				body.WafRealtimeThreatIntelligenceEnabled = &wafRealtimeThreatIntelligenceEnabled
			}
			if cmd.Flags().Changed("waf-profile-id") {
				body.WafProfileId = &wafProfileId
			}
			if cmd.Flags().Changed("ddos-sensitivity") {
				body.DDoSShieldSensitivity = &dDoSShieldSensitivity
			}
			if cmd.Flags().Changed("ddos-execution-mode") {
				body.DDoSExecutionMode = &dDoSExecutionMode
			}
			if cmd.Flags().Changed("ddos-challenge-window") {
				body.DDoSChallengeWindow = &dDoSChallengeWindow
			}
			if cmd.Flags().Changed("whitelabel-response-pages") {
				body.WhitelabelResponsePages = &whitelabelResponsePages
			}

			z, err := c.UpdateShieldZone(cmd.Context(), id, body)
			if err != nil {
				return err
			}

			columns := shieldZoneResponseColumns()

			formatted, err := output.FormatOne(cfg, columns, z)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	cmd.Flags().BoolVar(&learningMode, "learning-mode", false, "Enable or disable learning mode")
	cmd.Flags().BoolVar(&wafEnabled, "waf-enabled", false, "Enable or disable WAF")
	cmd.Flags().IntVar(&wafExecutionMode, "waf-execution-mode", 0, "WAF execution mode (0=Learn, 1=Protect)")
	cmd.Flags().BoolVar(&wafRequestHeaderLoggingEnabled, "waf-request-header-logging", false, "Enable WAF request header logging")
	cmd.Flags().BoolVar(&wafRealtimeThreatIntelligenceEnabled, "waf-realtime-threat-intel", false, "Enable WAF realtime threat intelligence")
	cmd.Flags().IntVar(&wafProfileId, "waf-profile-id", 0, "WAF profile ID")
	cmd.Flags().IntVar(&dDoSShieldSensitivity, "ddos-sensitivity", 0, "DDoS shield sensitivity")
	cmd.Flags().IntVar(&dDoSExecutionMode, "ddos-execution-mode", 0, "DDoS execution mode (0=Learn, 1=Protect)")
	cmd.Flags().IntVar(&dDoSChallengeWindow, "ddos-challenge-window", 0, "DDoS challenge window in seconds")
	cmd.Flags().BoolVar(&whitelabelResponsePages, "whitelabel-response-pages", false, "Enable whitelabel response pages")

	return cmd
}

// --- Column definitions ---

func shieldZoneListColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZone]{
		output.IntColumn[*client.ShieldZone]("Shield Zone Id", func(z *client.ShieldZone) int { return int(z.ShieldZoneId) }),
		output.IntColumn[*client.ShieldZone]("Pull Zone Id", func(z *client.ShieldZone) int { return int(z.PullZoneId) }),
		output.StringColumn[*client.ShieldZone]("Plan", func(z *client.ShieldZone) string { return client.ShieldPlanName(z.PlanType) }),
		output.BoolColumn[*client.ShieldZone]("WAF Enabled", func(z *client.ShieldZone) bool { return z.WafEnabled }),
		output.BoolColumn[*client.ShieldZone]("DDoS Enabled", func(z *client.ShieldZone) bool { return z.DDoSEnabled }),
		output.BoolColumn[*client.ShieldZone]("Learning Mode", func(z *client.ShieldZone) bool { return z.LearningMode }),
		output.IntColumn[*client.ShieldZone]("WAF Rules", func(z *client.ShieldZone) int { return z.TotalWAFCustomRules }),
		output.IntColumn[*client.ShieldZone]("Rate Limits", func(z *client.ShieldZone) int { return z.TotalRateLimitRules }),
	})
}

func shieldZoneDetailColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZone]{
		output.IntColumn[*client.ShieldZone]("Shield Zone Id", func(z *client.ShieldZone) int { return int(z.ShieldZoneId) }),
		output.IntColumn[*client.ShieldZone]("Pull Zone Id", func(z *client.ShieldZone) int { return int(z.PullZoneId) }),
		output.StringColumn[*client.ShieldZone]("Plan", func(z *client.ShieldZone) string { return client.ShieldPlanName(z.PlanType) }),
		output.BoolColumn[*client.ShieldZone]("Learning Mode", func(z *client.ShieldZone) bool { return z.LearningMode }),
		output.StringColumn[*client.ShieldZone]("Learning Mode Until", func(z *client.ShieldZone) string { return z.LearningModeUntil }),
		output.BoolColumn[*client.ShieldZone]("WAF Enabled", func(z *client.ShieldZone) bool { return z.WafEnabled }),
		output.StringColumn[*client.ShieldZone]("WAF Execution Mode", func(z *client.ShieldZone) string {
			return client.ShieldExecutionModeName(z.WafExecutionMode)
		}),
		output.IntColumn[*client.ShieldZone]("WAF Profile Id", func(z *client.ShieldZone) int { return z.WafProfileId }),
		output.BoolColumn[*client.ShieldZone]("WAF Header Logging", func(z *client.ShieldZone) bool { return z.WafRequestHeaderLoggingEnabled }),
		output.BoolColumn[*client.ShieldZone]("WAF Threat Intel", func(z *client.ShieldZone) bool { return z.WafRealtimeThreatIntelligenceEnabled }),
		output.BoolColumn[*client.ShieldZone]("DDoS Enabled", func(z *client.ShieldZone) bool { return z.DDoSEnabled }),
		output.StringColumn[*client.ShieldZone]("DDoS Execution Mode", func(z *client.ShieldZone) string {
			return client.ShieldExecutionModeName(z.DDoSExecutionMode)
		}),
		output.IntColumn[*client.ShieldZone]("DDoS Sensitivity", func(z *client.ShieldZone) int { return z.DDoSShieldSensitivity }),
		output.IntColumn[*client.ShieldZone]("DDoS Challenge Window", func(z *client.ShieldZone) int { return z.DDoSChallengeWindow }),
		output.IntColumn[*client.ShieldZone]("Custom WAF Rules", func(z *client.ShieldZone) int { return z.TotalWAFCustomRules }),
		output.IntColumn[*client.ShieldZone]("Rate Limit Rules", func(z *client.ShieldZone) int { return z.TotalRateLimitRules }),
		output.BoolColumn[*client.ShieldZone]("Whitelabel Pages", func(z *client.ShieldZone) bool { return z.WhitelabelResponsePages }),
		output.StringColumn[*client.ShieldZone]("Created", func(z *client.ShieldZone) string { return z.CreatedDateTime }),
		output.StringColumn[*client.ShieldZone]("Last Modified", func(z *client.ShieldZone) string { return z.LastModified }),
	})
}

func shieldZoneResponseColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.ShieldZoneResponse]{
		output.IntColumn[*client.ShieldZoneResponse]("Shield Zone Id", func(z *client.ShieldZoneResponse) int { return int(z.ShieldZoneId) }),
		output.IntColumn[*client.ShieldZoneResponse]("Pull Zone Id", func(z *client.ShieldZoneResponse) int { return int(z.PullZoneId) }),
		output.StringColumn[*client.ShieldZoneResponse]("Plan", func(z *client.ShieldZoneResponse) string { return client.ShieldPlanName(z.PlanType) }),
		output.BoolColumn[*client.ShieldZoneResponse]("Learning Mode", func(z *client.ShieldZoneResponse) bool { return z.LearningMode }),
		output.BoolColumn[*client.ShieldZoneResponse]("WAF Enabled", func(z *client.ShieldZoneResponse) bool { return z.WafEnabled }),
		output.StringColumn[*client.ShieldZoneResponse]("WAF Execution Mode", func(z *client.ShieldZoneResponse) string {
			return client.ShieldExecutionModeName(z.WafExecutionMode)
		}),
		output.IntColumn[*client.ShieldZoneResponse]("WAF Profile Id", func(z *client.ShieldZoneResponse) int { return z.WafProfileId }),
		output.StringColumn[*client.ShieldZoneResponse]("DDoS Execution Mode", func(z *client.ShieldZoneResponse) string {
			return client.ShieldExecutionModeName(z.DDoSExecutionMode)
		}),
		output.IntColumn[*client.ShieldZoneResponse]("DDoS Sensitivity", func(z *client.ShieldZoneResponse) int { return z.DDoSShieldSensitivity }),
		output.IntColumn[*client.ShieldZoneResponse]("Rate Limit Rules Limit", func(z *client.ShieldZoneResponse) int { return z.RateLimitRulesLimit }),
		output.IntColumn[*client.ShieldZoneResponse]("Custom WAF Rules Limit", func(z *client.ShieldZoneResponse) int { return z.CustomWafRulesLimit }),
	})
}
